"""Score the offline lookup against the independently annotated #176 cases."""
import argparse
from collections import Counter
import json
from pathlib import Path
from statistics import median
from time import perf_counter

from lookup import VARIANTS, documents, encode, search

ROOT = Path(__file__).resolve().parents[2]
FIXTURE = ROOT / "testdata/context-retrieval/cases.json"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--fixture", type=Path, default=FIXTURE)
    parser.add_argument("--variant", choices=VARIANTS, action="append")
    parser.add_argument("--input", choices=("query", "baseline-terms"), default="query")
    parser.add_argument("--require-literal", action="store_true",
                        help="Require at least one complete baseline search phrase before ranking")
    args = parser.parse_args()
    if args.require_literal and args.input != "baseline-terms":
        parser.error("--require-literal requires --input baseline-terms")
    fixture = json.loads(args.fixture.read_text(encoding="utf-8"))
    if fixture["version"] != 1:
        raise ValueError("unsupported fixture version")
    before_load = perf_counter()
    commit, docs = documents(ROOT, fixture["revision"])
    load_ms = (perf_counter() - before_load) * 1000
    for variant in args.variant or ["original"]:
        groups = {}
        for case in fixture["cases"]:
            before = perf_counter()
            query = case["query"] if args.input == "query" else " ".join(case["baseline_terms"])
            literals = case["baseline_terms"] if args.require_literal else ()
            result = search(commit, docs, query, 5, 2400, variant, require_literals=literals)
            elapsed = (perf_counter() - before) * 1000
            found = {(row["path"], row["heading"][-1]) for row in result["results"]}
            required = set(map(tuple, case["required"]))
            historical = set(map(tuple, case["historical"]))
            counters, sizes, times = groups.setdefault(case["split"], (Counter(), [], []))
            counters["cases"] += 1
            counters["required"] += len(required)
            counters["recovered"] += len(required & found)
            counters["positive_cases"] += bool(required)
            counters["all_evidence"] += bool(required and required <= found)
            counters["correct_empty"] += not required and result["status"] == "no_match"
            counters["no_match_cases"] += not required
            counters["false_candidates"] += not required and bool(found)
            counters["historical_without_successor"] += bool(historical & found and (required - historical) - found)
            sizes.append(len(encode(result)) + 1)
            times.append(elapsed)
            print(case["id"], result["status"], len(required & found), "/", len(required),
                  "bytes", sizes[-1], "ms", round(elapsed, 1))
        for split, (counts, sizes, times) in groups.items():
            print(args.input, "literal-filter" if args.require_literal else "no-filter", variant, split, dict(counts), "median bytes", median(sizes), "median ms", round(median(times), 1))
    print("snapshot", commit, "documents", len(docs), "load ms", round(load_ms, 1))


if __name__ == "__main__":
    main()
