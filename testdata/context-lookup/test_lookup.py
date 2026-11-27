"""Contract checks for the offline #177 experiment."""
import unittest

from lookup import encode, open_section, search, sections


class LookupTests(unittest.TestCase):
    adr = "docs/adr/0048-first.md"
    source = "---\r\nstatus: draft\r\n---\r\n# Title\r\n## Decision\r\nText\r\n```md\r\n## Ignored\r\n```\r\n## Decision\r\nSecond\r\n"

    def test_fences_crlf_and_duplicate_headings(self):
        paths = list(sections(self.adr, self.source))
        self.assertEqual([(x.heading, x.start, x.end) for x in paths],
                         [(("Title",), 4, 4), (("Decision",), 5, 9), (("Decision",), 10, 11)])
        result = open_section("commit", {self.adr: self.source}, self.adr, 10, 11)
        self.assertEqual(result["text"], "## Decision\nSecond")
        with self.assertRaises(ValueError):
            open_section("commit", {self.adr: self.source}, self.adr, 5, 8)
        with self.assertRaises(ValueError):
            open_section("commit", {}, "../outside.md", 5, 9)

    def test_missing_and_ambiguous_exact_id(self):
        docs = {self.adr: self.source, "docs/adr/0048-second.md": self.source}
        self.assertEqual(search("commit", docs, "ADR 9999", 5, 2400)["status"], "no_match")
        self.assertEqual(search("commit", docs, "ADR 0048", 5, 2400)["status"], "ambiguous")

    def test_ties_and_utf8_budget(self):
        docs = {path: "# 同期\n## 同期\n本文" for path in ["docs/z.md", "docs/a.md"]}
        result = search("commit", docs, "同期", 5, 190)
        self.assertEqual(result["results"][0]["path"], "docs/a.md")
        self.assertTrue(result["truncated"])
        self.assertLessEqual(len(encode(result)) + 1, 190)
        with self.assertRaises(ValueError):
            search("commit", docs, "同期", 5, 40)


    def test_variant_abstains_for_missing_ids_embedded_in_text(self):
        docs = {self.adr: self.source}
        result = search("commit", docs, "Show ADR 9876 decision.", 5, 2400, "abstain")
        self.assertEqual(result["status"], "no_match")

    def test_literal_bilingual_aliases(self):
        docs = {self.adr: "# Accept sync conflict\n## Decision\nUse sync."}
        self.assertEqual(search("commit", docs, "同期の衝突", 5, 2400)["status"], "no_match")
        result = search("commit", docs, "同期の衝突", 5, 2400, "vocabulary")
        self.assertEqual(result["results"][0]["path"], self.adr)

    def test_literal_phrase_guard(self):
        docs = {self.adr: "# Notion billing\n## Decision\nNotion is unrelated to page publication."}
        query = "Notion billing-page publication"
        self.assertEqual(search("commit", docs, query, 5, 2400)["status"], "ok")
        result = search("commit", docs, query, 5, 2400, require_literals=[query])
        self.assertEqual(result["status"], "no_match")

    def test_only_explicit_clause_reference_reorders_successor(self):
        successor = "docs/adr/0049-next.md"
        unrelated = "docs/adr/0050-unrelated.md"
        docs = {self.adr: self.source,
                successor: "# New\n### D2 (supersedes ADR 0048 D1)\nNew text.",
                unrelated: "# Other\n## See ADR 0048\nRelated text."}
        result = search("commit", docs, "ADR 0048", 2, 2400, "successor")
        self.assertEqual([x["path"] for x in result["results"]], [self.adr, successor])


if __name__ == "__main__":
    unittest.main()
