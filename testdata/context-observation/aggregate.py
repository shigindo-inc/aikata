"""Offline reference for the #178 synthetic observation contract."""

import argparse
import json
import re
from collections import defaultdict
from datetime import date
from pathlib import Path


PHASES = ("start", "implementation", "review", "handoff")
KINDS = {"phase_exposed", "question_observed", "outcome_observed"}
OUTCOMES = {"repeat_search", "no_repeat_search"}
STRATA = ("candidate_positive", "candidate_negative", "unclassified")
SPARSE_BELOW = 10
EVENT_FIELDS = {"event_id", "work_id", "week", "phase", "kind"}
WORK_ID = re.compile(r"w-[0-9a-f]{8}")
EVENT_ID = re.compile(r"e-[0-9a-f]{8}")
QUESTION_ID = re.compile(r"q-[a-z][a-z0-9-]{0,39}")
WEEK = re.compile(r"(\d{4})-W(\d{2})")


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def validate_fields(record, fields, reason):
    require(type(record) is dict and record.keys() == fields, reason)


def validate_pattern(value, pattern, reason):
    require(type(value) is str and pattern.fullmatch(value), reason)


def validate_week(value):
    validate_pattern(value, WEEK, "invalid ISO week")
    year, week = WEEK.fullmatch(value).groups()
    date.fromisocalendar(int(year), int(week), 1)


def event_metrics(event_list, questions):
    require(type(event_list) is list, "events must be a list")
    by_event_id = {}
    exposure = defaultdict(set)
    occurrence = defaultdict(dict)
    outcomes = {}
    for event in event_list:
        require(type(event) is dict, "event must be an object")
        kind = event.get("kind")
        require(type(kind) is str and kind in KINDS, "invalid event kind")
        fields = EVENT_FIELDS.copy()
        if kind != "phase_exposed":
            fields.add("question_id")
        if kind == "outcome_observed":
            fields.add("outcome")
        validate_fields(event, fields, "invalid event fields")
        validate_pattern(event["event_id"], EVENT_ID, "invalid event identifier")
        validate_pattern(event["work_id"], WORK_ID, "invalid work identifier")
        validate_week(event["week"])
        phase = event["phase"]
        require(type(phase) is str and phase in PHASES, "invalid phase")
        if kind != "phase_exposed":
            validate_pattern(event["question_id"], QUESTION_ID, "invalid question identifier")
            require(event["question_id"] in questions, "unknown question identifier")
        if kind == "outcome_observed":
            outcome = event["outcome"]
            require(type(outcome) is str and outcome in OUTCOMES, "invalid outcome")

        event_id = event["event_id"]
        if event_id in by_event_id:
            require(by_event_id[event_id] == event, "conflicting event identifier")
            continue
        by_event_id[event_id] = event

        work = event["work_id"]
        if kind == "phase_exposed":
            exposure[phase].add(work)
            continue

        question = event["question_id"]
        key = (phase, question)
        if kind == "question_observed":
            first_week = occurrence[key].get(work, event["week"])
            occurrence[key][work] = min(first_week, event["week"])
        else:
            outcome_key = (phase, question, work)
            previous = outcomes.setdefault(outcome_key, outcome)
            require(previous == outcome, "conflicting outcomes for one work, phase and question")

    for phase, question, work in outcomes:
        require(work in occurrence[(phase, question)], "outcome without question")

    all_exposed = set().union(*exposure.values())
    rows = []
    for question in sorted(questions):
        topic_weeks = {}
        topic_outcomes = defaultdict(list)
        topic_missing = set()
        for phase in PHASES:
            observations = occurrence[(phase, question)]
            eligible = observations.keys() & exposure[phase]
            missing = observations.keys() - exposure[phase]
            row_outcomes = [outcomes.get((phase, question, work)) for work in eligible]
            rows.append(row(question, phase, exposure[phase], eligible, missing, row_outcomes))
            topic_missing.update(missing)
            for work in eligible:
                first_week = topic_weeks.get(work, observations[work])
                topic_weeks[work] = min(first_week, observations[work])
                topic_outcomes[work].append(outcomes.get((phase, question, work)))

        totals = []
        for work, values in topic_outcomes.items():
            if "repeat_search" in values:
                totals.append("repeat_search")
            elif None not in values and work not in topic_missing:
                totals.append("no_repeat_search")
            else:
                totals.append(None)
        total = row(question, "all", all_exposed, topic_weeks, topic_missing, totals)
        total["distinct_first_occurrence_weeks"] = len(set(topic_weeks.values()))
        rows.append(total)
    return rows


def rate(numerator, denominator):
    if denominator == 0:
        status, value = "unavailable", None
    elif denominator < SPARSE_BELOW:
        status, value = "sparse", None
    else:
        status, value = "available", numerator / denominator
    return {"numerator": numerator, "denominator": denominator, "status": status, "value": value}


def row(question, phase, exposed, observed, missing, outcomes):
    unknown = outcomes.count(None)
    return {
        "question_id": question,
        "phase": phase,
        "missing_exposure_works": len(missing),
        "unknown_outcome_works": unknown,
        "occurrence": rate(len(observed), len(exposed)),
        "repeat_search": rate(outcomes.count("repeat_search"), len(outcomes) - unknown),
    }


def audit_counts(records):
    require(type(records) is list, "audit must be a list")
    counts = {
        stratum: {
            "sampled": 0, "occurrence_confirmed": 0, "occurrence_absent": 0,
            "false_merge": 0, "missed_equivalent": 0,
        }
        for stratum in STRATA
    }
    for record in records:
        validate_fields(record, {"stratum", "occurrence", "relation_error"}, "invalid audit fields")
        stratum = record["stratum"]
        occurrence = record["occurrence"]
        error = record["relation_error"]
        require(type(stratum) is str and stratum in STRATA, "invalid audit stratum")
        require(type(occurrence) is str and occurrence in {"confirmed", "absent"}, "invalid audit outcome")
        require(type(error) is str and error in {"none", "false_merge", "missed_equivalent"}, "invalid audit relation")
        counts[stratum]["sampled"] += 1
        counts[stratum]["occurrence_" + occurrence] += 1
        if error != "none":
            counts[stratum][error] += 1
    return counts


def aggregate(payload):
    validate_fields(payload, {"schema_version", "enrolled", "questions", "events", "audit"}, "invalid envelope fields")
    require(type(payload["schema_version"]) is int and payload["schema_version"] == 1, "unsupported schema")
    require(payload["enrolled"] is True, "explicit enrollment must be true")
    vocabulary = payload["questions"]
    require(type(vocabulary) is list and vocabulary, "questions must be a nonempty list")
    for question in vocabulary:
        validate_pattern(question, QUESTION_ID, "invalid vocabulary identifier")
    questions = set(vocabulary)
    require(len(questions) == len(vocabulary), "duplicate vocabulary identifier")
    return {
        "schema_version": 1,
        "sparse_below": SPARSE_BELOW,
        "rows": event_metrics(payload["events"], questions),
        "audit_by_stratum": audit_counts(payload["audit"]),
    }


def unique_json_object(pairs):
    result = {}
    for key, value in pairs:
        require(key not in result, "duplicate JSON key")
        result[key] = value
    return result


def read_payload(path):
    return json.loads(path.read_text(encoding="utf-8"), object_pairs_hook=unique_json_object)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("fixture", type=Path)
    args = parser.parse_args()
    try:
        result = aggregate(read_payload(args.fixture))
    except (ValueError, OSError) as error:
        parser.exit(2, f"observation fixture: {error}\n")
    print(json.dumps(result, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
