"""Tests for jdclab.filters: ranking both ways, the pool, and the scores, with made-up vectors."""

import numpy as np
import pytest

from jdclab.filters import lets_through, pool, rank, relevant_sets, report, score

# Words as vectors: a text's vector is its words' sum, made unit length.
WORDS = {
    "chef": [1.0, 0.0, 0.0],
    "leeds": [0.0, 1.0, 0.0],
    "york": [0.0, 0.0, 1.0],
}


def vectors(texts):
    out = []
    for t in texts:
        v = np.zeros(3)
        for w in t.lower().split():
            v += WORDS.get(w, [0.0, 0.0, 0.0])
        out.append(v / (np.linalg.norm(v) or 1.0))
    return np.array(out)


def ad(i, title, text, place="", patterns=()):
    return {
        "id": i,
        "title": title,
        "own": text,
        "whole": text + " agency",
        "place": place,
        "patterns": list(patterns),
    }


ADS = [
    # A chef in Leeds whose text never says so, and one in York that
    # says Leeds: the filter finds the first, the embedding the second.
    ad("a", "Chef", "chef", place="Leeds", patterns=["part time"]),
    ad("b", "Chef", "chef leeds", place="York"),
    ad("c", "Chef", "chef york", place="York", patterns=["part time"]),
]
PLACE = {
    "query": "Chef in Leeds",
    "title": "Chef",
    "kind": "place",
    "value": "Leeds",
    "candidates": ["a", "b", "c"],
    "relevant": ["a"],
}
PATTERN = {
    "query": "part time Chef",
    "title": "Chef",
    "kind": "pattern",
    "value": "part time",
    "candidates": ["a", "b", "c"],
    "relevant": [],
}


def test_lets_through():
    assert lets_through(PLACE, ADS[0]) and not lets_through(PLACE, ADS[1])
    assert lets_through(PATTERN, ADS[2]) and not lets_through(PATTERN, ADS[1])


def test_rank_both_ways():
    rankings = rank(ADS, [PLACE], vectors, vectors)
    got = {(r["way"], r["representation"]): r["ranked"] for r in rankings}
    # Embedded, "Chef in Leeds" is nearest the text that says Leeds.
    assert got[("embedded", "own")][0] == "b"
    # Filtered, only the ad in Leeds is let through.
    assert got[("filtered", "own")] == ["a"]
    assert set(got) == {
        (w, r) for w in ("embedded", "filtered") for r in ("own", "whole")
    }


def test_pool_takes_candidates_once_and_only_pattern_queries():
    rankings = [
        {
            "query": "part time Chef",
            "way": "embedded",
            "representation": "own",
            "ranked": ["b", "x", "a"],
        },
        {
            "query": "part time Chef",
            "way": "filtered",
            "representation": "own",
            "ranked": ["a", "c"],
        },
        {
            "query": "Chef in Leeds",
            "way": "embedded",
            "representation": "own",
            "ranked": ["a"],
        },
    ]
    rows = pool(ADS + [ad("x", "Driver", "driver")], [PLACE, PATTERN], rankings)
    # x isn't a candidate; a comes once; the place query isn't pooled.
    assert [r["id"] for r in rows] == ["b", "a", "c"]
    assert all(
        r["query"] == "part time Chef" and r["pattern"] == "part time" for r in rows
    )


def test_relevant_sets_keep_judged_candidates():
    judgements = [
        {"query": "part time Chef", "id": "a", "verdict": "for"},
        {"query": "part time Chef", "id": "b", "verdict": "against"},
        {"query": "part time Chef", "id": "z", "verdict": "for"},  # not a candidate
    ]
    rel = relevant_sets([PLACE, PATTERN], judgements)
    assert rel == {"Chef in Leeds": {"a"}, "part time Chef": {"a"}}


def test_score_counts_what_the_filter_loses():
    # b is judged part time, though its facts didn't find it, so the filter
    # can't let it through: half the relevant are filtered out.
    judgements = [
        {"query": "part time Chef", "id": "a", "verdict": "for"},
        {"query": "part time Chef", "id": "b", "verdict": "for"},
    ]
    rankings = [
        {
            "query": "part time Chef",
            "way": "embedded",
            "representation": "own",
            "ranked": ["b", "c", "a"],
        },
        {
            "query": "part time Chef",
            "way": "filtered",
            "representation": "own",
            "ranked": ["a", "c"],
        },
    ]
    got = score(ADS, [PATTERN], rankings, judgements)
    embedded, filtered = (
        got[("pattern", "embedded", "own")],
        got[("pattern", "filtered", "own")],
    )
    assert embedded.precision_at_10 == pytest.approx(0.2)
    assert embedded.recall == 1.0
    assert filtered.recall == 0.5
    assert filtered.filtered_out == 0.5


def test_score_skips_other_queries_and_report_gives_the_best():
    rankings = [
        {
            "query": "Chef in Leeds",
            "way": "filtered",
            "representation": "own",
            "ranked": ["a"],
        },
        {
            "query": "part time Chef",
            "way": "filtered",
            "representation": "own",
            "ranked": ["a"],
        },
    ]
    got = score(ADS, [PLACE], rankings, [])
    assert list(got) == [("place", "filtered", "own")]
    # One relevant ad, so the best precision at 10 is 0.1.
    assert "best P@10 0.100" in report(got, [PLACE], [])
