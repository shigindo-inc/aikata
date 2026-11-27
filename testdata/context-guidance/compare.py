"""Replay two predeclared navigation routes against a pinned Git tree."""

import argparse
import json
import re
import subprocess
from pathlib import Path


FIXTURES = Path(__file__).resolve().parent
ROOT = FIXTURES.parents[1]
HEADING = re.compile(r"^#{1,6} ")


def git(*args):
    return subprocess.run(["git", *args], cwd=ROOT, check=True, capture_output=True).stdout


def section(document, target):
    lines = document.splitlines(keepends=True)
    matches = [index for index, line in enumerate(lines) if line.rstrip() == target]
    if len(matches) != 1:
        raise ValueError(f"expected one pinned heading: {target}")
    start = matches[0]
    depth = target.index(" ")
    end = len(lines)
    for index, line in enumerate(lines[start + 1:], start + 1):
        heading = HEADING.match(line)
        if heading and len(heading.group()) - 1 <= depth:
            end = index
            break
    return "".join(lines[start:end])


def source_link(source):
    plain_heading = source["heading"].lstrip("# ").lower().replace("`", "")
    anchor = re.sub(r"[^\w -]", "", plain_heading).replace(" ", "-")
    return f"- [{source['heading'].lstrip('# ')}]({source['path']}#{anchor})"


def score_navigation(task, arm, sources, excerpts, nav_bytes, shared_bytes):
    route = task[arm]["route"]
    opened = {entry["source"]: excerpts[entry["source"]] for entry in route}
    recovered = []
    for required in task["required"]:
        source = required["source"]
        excerpt = opened.get(source, "")
        if all(fragment in excerpt for fragment in required["contains"]):
            recovered.append(source)
    section_bytes = sum(len(text.encode("utf-8")) for text in opened.values())
    added_bytes = nav_bytes + section_bytes
    return {
        "navigation_operations": 1,
        "source_sections_opened": len(opened),
        "source_files_opened": len({sources[source]["path"] for source in opened}),
        "required_sections": len(task["required"]),
        "recovered_sections": recovered,
        "navigation_bytes": nav_bytes,
        "source_section_bytes": section_bytes,
        "additional_payload_bytes": added_bytes,
        "payload_including_shared_reads_bytes": shared_bytes + added_bytes,
    }


def compare(configuration):
    revision = configuration["revision"]
    sources = configuration["sources"]
    shared_bytes = sum(len(git("show", f"{revision}:{path}")) for path in configuration["shared_reads"])
    excerpts = {
        name: section(git("show", f"{revision}:{source['path']}").decode("utf-8"), source["heading"])
        for name, source in sources.items()
    }
    results = []
    for task in configuration["tasks"]:
        start = task["current"]["start"]
        matches = git(
            "grep", "--no-color", "-n", "-F", "-C", str(start["context"]), "-e", start["term"],
            revision, "--", start["file"],
        )
        for entry in task["current"]["route"]:
            if entry["found_by"].encode("utf-8") not in matches:
                raise ValueError(f"navigation did not surface a planned reference: {entry['source']}")

        guidance = task["guidance"]
        links = [source_link(sources[entry["source"]]) for entry in guidance["route"]]
        card = "\n".join([guidance["card"], *links]) + "\n"
        current = score_navigation(task, "current", sources, excerpts, len(matches), shared_bytes)
        guided = score_navigation(task, "guidance", sources, excerpts, len(card.encode("utf-8")), shared_bytes)
        results.append({
            "task": task["id"],
            "current": current,
            "guidance": guided,
            "difference_guided_minus_current_bytes": guided["additional_payload_bytes"] - current["additional_payload_bytes"],
            "difference_if_current_uses_shared_reads_only_bytes": guided["navigation_bytes"],
        })
    return {"revision": revision, "shared_read_bytes_per_task": shared_bytes, "tasks": results}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="compare outputs with the staged result")
    args = parser.parse_args()
    configuration = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))
    result = compare(configuration)
    if args.check:
        expected = json.loads((FIXTURES / "expected.json").read_text(encoding="utf-8"))
        if result != expected:
            parser.exit(1, "staged navigation results differ from expected.json\n")
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
