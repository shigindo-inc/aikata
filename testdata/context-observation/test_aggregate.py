import copy
import json
import unittest
from pathlib import Path

from aggregate import aggregate, read_payload, unique_json_object


FIXTURES = Path(__file__).parent
QUESTION = "q-accept-sync-conflict"


class ObservationTest(unittest.TestCase):
    def setUp(self):
        self.payload = read_payload(FIXTURES / "synthetic.json")

    def result(self, report, phase):
        return next(row for row in report["rows"] if row["question_id"] == QUESTION and row["phase"] == phase)

    def test_golden_and_week_order_independence(self):
        expected = read_payload(FIXTURES / "expected.json")
        self.assertEqual(aggregate(self.payload), expected)
        self.payload["events"].reverse()
        self.assertEqual(aggregate(self.payload), expected)

    def test_counts_are_works_not_events_or_summed_phases(self):
        report = aggregate(self.payload)
        implementation = self.result(report, "implementation")
        self.assertEqual(implementation["occurrence"], {
            "numerator": 2, "denominator": 10, "status": "available", "value": 0.2,
        })
        self.assertEqual(implementation["missing_exposure_works"], 1)
        self.assertEqual(implementation["unknown_outcome_works"], 1)
        self.assertEqual(implementation["repeat_search"], {
            "numerator": 1, "denominator": 1, "status": "sparse", "value": None,
        })

        total = self.result(report, "all")
        self.assertEqual(total["occurrence"]["numerator"], 3)
        self.assertEqual(total["occurrence"]["denominator"], 10)
        self.assertEqual(total["distinct_first_occurrence_weeks"], 2)
        self.assertEqual(total["unknown_outcome_works"], 2)
        self.assertEqual(total["missing_exposure_works"], 2)

    def test_zero_denominator_and_explicit_no_search(self):
        report = aggregate(self.payload)
        start = self.result(report, "start")
        self.assertEqual(start["occurrence"]["numerator"], 0)
        self.assertEqual(start["occurrence"]["denominator"], 1)
        handoff = self.result(report, "handoff")
        self.assertEqual(handoff["occurrence"], {
            "numerator": 0, "denominator": 0, "status": "unavailable", "value": None,
        })
        self.assertEqual(handoff["missing_exposure_works"], 1)

    def test_audit_strata_keep_distinct_error_modes(self):
        audit = aggregate(self.payload)["audit_by_stratum"]
        self.assertEqual(audit["candidate_positive"]["false_merge"], 1)
        self.assertEqual(audit["candidate_negative"]["missed_equivalent"], 1)
        self.assertEqual(audit["unclassified"]["sampled"], 1)

    def test_inconsistent_event_replay_is_rejected(self):
        self.payload["events"][-1] = dict(self.payload["events"][-1], week="2026-W48")
        with self.assertRaisesRegex(ValueError, "conflicting event identifier"):
            aggregate(self.payload)

    def test_conflicting_outcomes_and_orphans_are_rejected(self):
        event = next(row for row in self.payload["events"] if row["kind"] == "outcome_observed")
        self.payload["events"].append(dict(event, event_id="e-aaaaaaaa", outcome="no_repeat_search"))
        with self.assertRaisesRegex(ValueError, "conflicting outcomes"):
            aggregate(self.payload)
        self.payload["events"][-1] = dict(event, event_id="e-aaaaaaaa", work_id="w-bbbbbbbb")
        with self.assertRaisesRegex(ValueError, "outcome without question"):
            aggregate(self.payload)

    def test_enrollment_and_version_are_explicit(self):
        for updates in ({"enrolled": False}, {"enrolled": 1}, {"schema_version": 2}, {"schema_version": True}):
            with self.subTest(updates=updates), self.assertRaises(ValueError):
                aggregate(dict(self.payload, **updates))

    def test_extra_fields_invalid_enums_and_unknown_question_are_rejected(self):
        row = next(row for row in self.payload["events"] if row["kind"] == "outcome_observed")
        for updates in (
            {"path": "/private/path.md"}, {"raw_query": "example"},
            {"kind": "search"}, {"phase": "other"}, {"outcome": "unknown"},
            {"question_id": "q-not-in-vocabulary"}, {"week": "2026-W54"},
            {"work_id": "long-session-id"},
        ):
            payload = copy.deepcopy(self.payload)
            index = payload["events"].index(row)
            payload["events"][index].update(updates)
            with self.subTest(updates=updates), self.assertRaises(ValueError):
                aggregate(payload)

    def test_duplicate_object_keys_are_rejected(self):
        with self.assertRaisesRegex(ValueError, "duplicate JSON key"):
            json.loads('{"enrolled":false,"enrolled":true}', object_pairs_hook=unique_json_object)


if __name__ == "__main__":
    unittest.main()
