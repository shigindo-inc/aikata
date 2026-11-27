"""Offline section lookup for the #177 comparison. Reads only pinned Git blobs."""
import argparse
from collections import Counter
from dataclasses import dataclass
import json
from math import log
from pathlib import Path
import re
import subprocess
import sys
import unicodedata

VARIANTS = ("original", "abstain", "vocabulary", "successor", "combined")

TOP = ("README.md", "AGENTS.md", "SPEC.md", "ARCHITECTURE.md", "GLOSSARY.md",
       "ROADMAP.md", "CONTRIBUTING.md")
HEADING = re.compile(r"^ {0,3}(#{1,6})[ \t]+(.+?)\s*#*\s*$")
FENCE = re.compile(r"^ {0,3}(`{3,}|~{3,})(.*)$")
ADR_FILE = re.compile(r"docs/adr/(\d{4})-[^/]+\.md$")
ADR_QUERY = re.compile(r"(?:adr[ -]*)?(\d{4})", re.IGNORECASE)
TOKEN = re.compile(r"[a-z0-9][a-z0-9_-]*|[\u3040-\u30ff\u3400-\u9fff]+")
STOP = set("a an and are as at be by do does for how in is it no of on or the to what which why with".split())
GENERIC = set("aikata adr current find contract show check document file where command configuration support".split())
# Literal terms from the original tuning questions and their benchmark search terms.
ALIASES = {
    "同期": ("sync",), "衝突": ("conflict",), "ファイル単位": ("per-file",),
    "受け入れ": ("accept",), "管理対象": ("managed-surface",),
    "既定値": ("default",), "除外": ("doctor.exclude",),
    "追加機能": ("optional features", "do-no-harm"), "有効化": ("opt-in",),
}
ADR_MENTION = re.compile(r"\badr[ -]+(\d{4})\b")
SUCCESSOR = re.compile(r"\bsupersedes\s+adr[ -]+(\d{4})\s+\*{0,2}d\d+\b", re.I)


@dataclass(frozen=True)
class Section:
    path: str
    title: str
    heading: tuple
    start: int
    end: int
    text: str


def git(repo, *args):
    return subprocess.check_output(("git", "-C", str(repo), *args))


def documents(repo, revision):
    commit = git(repo, "rev-parse", "--verify", "--end-of-options", revision + "^{commit}").decode().strip()
    docs = {}
    for entry in git(repo, "ls-tree", "-rz", commit).split(b"\0"):
        if not entry:
            continue
        meta, path_bytes = entry.split(b"\t", 1)
        mode, kind, blob = meta.split()
        path = path_bytes.decode()
        if mode not in (b"100644", b"100755") or kind != b"blob":
            continue
        if path not in TOP and not (path.startswith("docs/") and path.endswith(".md")):
            continue
        docs[path] = git(repo, "cat-file", "blob", blob.decode()).decode()
    return commit, docs


def sections(path, text):
    lines = text.splitlines()
    title = path
    stack = []
    boundaries = []
    active_fence = None
    in_frontmatter = bool(lines and lines[0] == "---")
    for offset, line in enumerate(lines, 1):
        if in_frontmatter:
            if offset > 1 and line == "---":
                in_frontmatter = False
            continue
        marker = FENCE.match(line)
        if active_fence:
            if marker and marker[1][0] == active_fence[0] and len(marker[1]) >= len(active_fence) and not marker[2].strip():
                active_fence = None
            continue
        if marker:
            active_fence = marker[1]
            continue
        match = HEADING.match(line)
        if not match:
            continue
        level, label = len(match[1]), match[2]
        if level == 1:
            title = label
        stack = [(depth, name) for depth, name in stack if depth < level]
        stack.append((level, label))
        heading = tuple(name for depth, name in stack if depth > 1) or (label,)
        boundaries.append((offset, heading))
    if not boundaries:
        boundaries = [(1, (path,))]
    for i, (start, heading) in enumerate(boundaries):
        end = boundaries[i + 1][0] - 1 if i + 1 < len(boundaries) else len(lines)
        if end >= start:
            yield Section(path, title, heading, start, end, "\n".join(lines[start - 1:end]))


def normalize(text):
    return unicodedata.normalize("NFKC", text).casefold()


def terms(text):
    return set(TOKEN.findall(normalize(text))) - STOP


def ids(paths, query):
    direct = ADR_QUERY.fullmatch(normalize(query).strip())
    if not direct:
        return []
    return sorted(path for path in paths if (match := ADR_FILE.fullmatch(path)) and match[1] == direct[1])


def explicit_successors(docs):
    by_number = {}
    for path, text in docs.items():
        if not ADR_FILE.fullmatch(path):
            continue
        for part in sections(path, text):
            for match in SUCCESSOR.finditer(part.heading[-1]):
                by_number.setdefault(match[1], []).append(part)
    return by_number


def expand_successors(scored, docs):
    successors = explicit_successors(docs)
    added = set()
    for score, section, reason in scored:
        key = section.path, section.start
        if key in added:
            continue
        added.add(key)
        yield score, section, reason
        adr = ADR_FILE.fullmatch(section.path)
        if adr:
            for related in successors.get(adr[1], ()):
                other = related.path, related.start
                if other not in added:
                    added.add(other)
                    yield score, related, "heading explicitly names a superseded ADR clause"


def ranked(docs, query, variant="original", *, require_literals=()):
    normalized = normalize(query).strip()
    exact = ids(docs, query)
    if ADR_QUERY.fullmatch(normalized) and not exact:
        return "no_match", []
    mentions = set(ADR_MENTION.findall(normalized))
    if variant in ("abstain", "combined") and mentions:
        for ident in sorted(mentions):
            named = ids(docs, ident)
            if len(named) > 1:
                return "ambiguous", named
            if not named:
                return "no_match", []
    if len(exact) > 1:
        return "ambiguous", exact
    query_terms = terms(query)
    if variant in ("vocabulary", "combined"):
        query_terms |= {term for label, matches in ALIASES.items() if label in normalized for term in matches}
    direct_path = query.strip()
    if not query_terms and not exact and direct_path not in docs:
        return "no_match", []
    literals = tuple(" ".join(normalize(term).split()) for term in require_literals if term.strip())
    frequencies = Counter(term for text in docs.values() for term in terms(text))
    scored = []
    for path, body in docs.items():
        if (exact and path != exact[0]) or (direct_path in docs and path != direct_path):
            continue
        for section in sections(path, body):
            title = normalize(section.title + " " + path)
            heading = normalize(" ".join(section.heading))
            text = normalize(section.text)
            if literals and not (exact or direct_path == path):
                haystack = " ".join((title + " " + heading + " " + text).split())
                if not any(literal in haystack for literal in literals):
                    continue
            score = 0.0
            substantive = set()
            for term in sorted(query_terms):
                weight = log((len(docs) + 1) / (frequencies[term] + 1)) + 1
                score += weight * (4 * (term in title) + 3 * (term in heading) + (term in text))
                if term not in GENERIC and not term.isdigit() and (term in title or term in heading or term in text):
                    substantive.add(term)
            if exact or direct_path == path:
                if re.match(r"^(decisions?|決定)(\b|$)", heading) or heading == "決定":
                    score += 100
                score += 1
            if (adr := ADR_FILE.fullmatch(path)) and adr[1] in mentions and not exact:
                score += 10
                if re.match(r"^(decisions?|決定)(\b|$)", heading) or heading == "決定":
                    score += 3
            named_section = bool(adr and adr[1] in mentions)
            if variant in ("abstain", "combined") and not (exact or direct_path == path or named_section):
                if len(substantive) < 2 and not any(term in heading or term in title for term in substantive):
                    continue
            if score:
                reason = "exact ADR ID" if exact else ("exact path" if direct_path == path else "weighted title, heading and text matches")
                scored.append((score, section, reason))
    scored.sort(key=lambda row: (-row[0], row[1].path, row[1].start))
    if scored and variant in ("successor", "combined"):
        scored = list(expand_successors(scored, docs))
    return ("ok", scored) if scored else ("no_match", [])


def encode(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":")).encode()


def search(commit, docs, query, limit, budget, variant="original", *, require_literals=()):
    status, ranked_items = ranked(docs, query, variant, require_literals=require_literals)
    result = {"revision": commit, "status": status, "results": [], "truncated": False}
    if status == "ambiguous":
        result["conflicting_paths"] = []
        for path in ranked_items:
            result["conflicting_paths"].append(path)
            if len(encode(result)) + 1 > budget:
                result["conflicting_paths"].pop()
                result["truncated"] = True
                break
    elif status == "ok":
        for _, section, reason in ranked_items:
            if len(result["results"]) == limit:
                result["truncated"] = True
                break
            result["results"].append({"path": section.path, "heading": list(section.heading),
                                      "start": section.start, "end": section.end, "reason": reason})
            if len(encode(result)) + 1 > budget:
                result["results"].pop()
                result["truncated"] = True
                break
        if not result["results"]:
            raise ValueError("budget is too small for a candidate")
    if len(encode(result)) + 1 > budget:
        raise ValueError("budget is too small for the response envelope")
    return result


def open_section(commit, docs, path, start, end):
    if path not in docs:
        raise ValueError("path is outside the indexed Markdown snapshot")
    for section in sections(path, docs[path]):
        if (section.start, section.end) == (start, end):
            return {"revision": commit, "path": path, "heading": list(section.heading),
                    "start": start, "end": end, "text": section.text}
    raise ValueError("section locator does not match the snapshot")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", type=Path, default=Path.cwd())
    parser.add_argument("--revision", required=True)
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--query")
    group.add_argument("--open-section", nargs=3, metavar=("PATH", "START", "END"))
    parser.add_argument("--variant", choices=VARIANTS, default="original")
    parser.add_argument("--limit", type=int, default=5)
    parser.add_argument("--budget", type=int, default=2400)
    args = parser.parse_args()
    try:
        if args.open_section and not re.fullmatch(r"(?:[0-9a-fA-F]{40}|[0-9a-fA-F]{64})", args.revision):
            raise ValueError("opening a section requires the full revision from its locator")
        commit, docs = documents(args.repo, args.revision)
        if args.query is not None:
            if not args.query.strip() or args.limit < 1 or args.budget < 1:
                raise ValueError("query must be nonblank and limits must be positive")
            result = search(commit, docs, args.query, args.limit, args.budget, args.variant)
        else:
            path, start, end = args.open_section
            result = open_section(commit, docs, path, int(start), int(end))
        sys.stdout.buffer.write(encode(result) + b"\n")
    except (ValueError, subprocess.CalledProcessError) as error:
        parser.exit(2, f"lookup: {error}\n")


if __name__ == "__main__":
    main()
