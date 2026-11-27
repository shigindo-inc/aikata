"""Reproduce a bounded, assisted navigation + ripgrep read-only baseline."""
import argparse
from collections import defaultdict
from dataclasses import dataclass
import json
from pathlib import Path
import re
import statistics
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parent.parent
TOP_LEVEL = {"README.md", "AGENTS.md", "SPEC.md", "ARCHITECTURE.md", "GLOSSARY.md", "ROADMAP.md", "CONTRIBUTING.md"}
MAX_CANDIDATES = 5
MAX_BYTES = 2400
HEADING = re.compile(r"^ {0,3}#{1,6}[ \t]+(.+?)[ \t]*#*[ \t]*$")
FENCE = re.compile(r"^ {0,3}(`{3,}|~{3,})(.*)$")
ADR_ID = re.compile(r"\bADR\s*[- ]\s*(\d{4})\b", re.IGNORECASE)


@dataclass(frozen=True)
class Section:
    heading: str
    line: int
    end: int


def git(*args):
    return subprocess.check_output(["git", "-C", str(REPO), *args])


def snapshot(revision, directory):
    filenames = git("ls-tree", "-rz", "--name-only", revision).decode().split("\0")
    for name in filenames:
        if name in TOP_LEVEL or (name.startswith("docs/") and name.endswith(".md")):
            file = directory / name
            file.parent.mkdir(parents=True, exist_ok=True)
            file.write_bytes(git("show", revision + ":" + name))


def sections(path):
    lines = path.read_text(encoding="utf-8").splitlines()
    matches = []
    fence = None
    for number, line in enumerate(lines, 1):
        delimiter = FENCE.match(line)
        if delimiter:
            tick, rest = delimiter.groups()
            if fence is None:
                fence = tick
            elif tick.startswith(fence) and not rest.strip():
                fence = None
            continue
        if fence is None and (heading := HEADING.match(line)):
            matches.append((heading.group(1), number))
    return [
        Section(heading, start, matches[i + 1][1] - 1 if i + 1 < len(matches) else len(lines))
        for i, (heading, start) in enumerate(matches)
    ]


def candidate(path, section, excerpt):
    return {
        "path": path,
        "heading": section.heading,
        "line": section.line,
        "end": section.end,
        "excerpt": excerpt.strip()[:120],
    }


def byte_count(value):
    return len(json.dumps(value, ensure_ascii=False, separators=(",", ":")).encode("utf-8"))


def bounded(rows):
    result = []
    for row in rows:
        trimmed = dict(row)
        while byte_count([*result, trimmed]) > MAX_BYTES and trimmed["excerpt"]:
            trimmed["excerpt"] = trimmed["excerpt"][:-1]
        if byte_count([*result, trimmed]) > MAX_BYTES:
            break
        result.append(trimmed)
        if len(result) >= MAX_CANDIDATES:
            break
    return result


def rg_candidates(case, directory, headings):
    terms = case["baseline_terms"]
    command = ["rg", "--no-config", "--json", "--fixed-strings", "--ignore-case", "--sort", "path"]
    for term in terms:
        command.extend(["-e", term])
    process = subprocess.run([*command, "."], cwd=directory, capture_output=True, text=True, check=False)
    if process.returncode not in (0, 1):
        raise RuntimeError(process.stderr)
    candidates = {}
    scores = defaultdict(set)
    for raw in process.stdout.splitlines():
        record = json.loads(raw)
        if record["type"] != "match":
            continue
        match = record["data"]
        path = match["path"]["text"].removeprefix("./")
        line = match["line_number"]
        section = next((s for s in headings[path] if s.line <= line <= s.end), None)
        if section is None:
            continue
        key = (path, section.line)
        excerpt = match["lines"]["text"]
        candidates.setdefault(key, candidate(path, section, excerpt))
        scores[key].update(term.casefold() for term in terms if term.casefold() in excerpt.casefold())
    # Existing Navigation Matrix directs CLI work to SPEC. A direct navigation
    # question is first checked in AGENTS or ARCHITECTURE before ADRs.
    navigation = ("AGENTS.md", "ARCHITECTURE.md", "SPEC.md")
    ordered = sorted(
        candidates,
        key=lambda key: (
            navigation.index(key[0]) if key[0] in navigation else len(navigation),
            -len(scores[key]),
            *key,
        ),
    )
    return [candidates[key] for key in ordered]


def find(case, directory, headings):
    # A human can turn a sole, explicit ADR number into a filename search,
    # inspect the heading outline and open Decision. Qualifiers such as
    # "superseded" still go through normal search.
    exact = ADR_ID.fullmatch(case["baseline_terms"][0]) if len(case["baseline_terms"]) == 1 else None
    if exact is not None:
        pattern = f"docs/adr/{exact[1]}-"
        rows = []
        for path, outline in sorted(headings.items()):
            if not path.startswith(pattern):
                continue
            selected = next((s for s in outline if s.heading in ("Decision", "Decisions")), None)
            if selected:
                body = (directory / path).read_text(encoding="utf-8").splitlines()
                excerpt = next((line for line in body[selected.line:selected.end] if line.strip()), "")
                rows.append(candidate(path, selected, excerpt))
        return bounded(rows)
    return bounded(rg_candidates(case, directory, headings))


def score_case(case, candidates):
    seen = {(row["path"], row["heading"]) for row in candidates}
    required = {tuple(anchor) for anchor in case["required"]}
    historical = {tuple(anchor) for anchor in case["historical"]}
    missing = sorted(required - seen)
    return {
        "full": not missing if required else not candidates,
        "missing": missing,
        "required": len(required),
        "recovered": len(required & seen),
        "false_candidates": not required and bool(candidates),
        # Exposure to historical evidence without the current complement is
        # a risk indicator, not a measured answer error.
        "historical_without_successor": bool(historical & seen and (required - historical) - seen),
        "bytes": byte_count(candidates),
    }


def validate(case, headings):
    if case["kind"] == "negative" and case["required"]:
        raise ValueError(f"{case['id']}: negative case has required sections")
    if case["kind"] != "negative" and not case["required"]:
        raise ValueError(f"{case['id']}: missing required sections")
    if not {tuple(anchor) for anchor in case["historical"]} <= {tuple(anchor) for anchor in case["required"]}:
        raise ValueError(f"{case['id']}: historical section must be required")
    for path, title in case["required"]:
        count = sum(1 for s in headings.get(path, []) if s.heading == title)
        if count != 1:
            raise ValueError(f"{case['id']}: expected a unique section {path} / {title!r}, found {count}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true", help="Emit candidate rows and scores")
    parser.add_argument("--fixture", type=Path, default=ROOT / "cases.json")
    args = parser.parse_args()
    fixture = json.loads(args.fixture.read_text(encoding="utf-8"))
    if fixture["version"] != 1:
        raise ValueError("unsupported corpus fixture version")
    results = []
    with tempfile.TemporaryDirectory(prefix="aikata-context-retrieval-") as scratch:
        directory = Path(scratch)
        snapshot(fixture["revision"], directory)
        headings = {p.relative_to(directory).as_posix(): sections(p) for p in sorted(directory.rglob("*.md"))}
        for case in fixture["cases"]:
            validate(case, headings)
        for case in fixture["cases"]:
            start = time.perf_counter_ns()
            candidates = find(case, directory, headings)
            elapsed_ms = (time.perf_counter_ns() - start) / 1_000_000
            results.append({
                "id": case["id"],
                "split": case["split"],
                "lang": case["lang"],
                "kind": case["kind"],
                "candidates": candidates,
                "elapsed_ms": elapsed_ms,
                **score_case(case, candidates),
            })
    output = {"revision": fixture["revision"], "documents": len(headings), "cases": results}
    if args.json:
        print(json.dumps(output, indent=2, ensure_ascii=False))
        return
    print(f"{output['revision']} / {output['documents']} Markdown documents")
    for split in dict.fromkeys(r["split"] for r in results):
        subset = [r for r in results if r["split"] == split]
        positive = [r for r in subset if r["required"]]
        negative = [r for r in subset if not r["required"]]
        print(f"{split}: all evidence {sum(r['full'] for r in positive)}/{len(positive)},"
              f" sections {sum(r['recovered'] for r in positive)}/{sum(r['required'] for r in positive)},"
              f" correct empty {sum(r['full'] for r in negative)}/{len(negative)},"
              f" history-only risk {sum(r['historical_without_successor'] for r in positive)},"
              f" bytes {sum(r['bytes'] for r in subset)}, median {statistics.median(r['elapsed_ms'] for r in subset):.2f} ms")
    for row in results:
        print(row["id"], "ok" if row["full"] else "miss", row["bytes"], row["missing"])


if __name__ == "__main__":
    main()
