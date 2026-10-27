"""Tests for jdclab.retrieval: the scores, and a whole run with made-up vectors."""

import numpy as np
import pytest

from jdclab.retrieval import NOTHING, ad_scores, ndcg_at, precision_at, recall_at, run


def test_ad_scores_take_each_ads_best_vector():
    query = np.array([1.0, 0.0])
    vectors = np.array([[0.6, 0.8], [1.0, 0.0], [0.0, 1.0]])
    owner = np.array([0, 0, 1])  # ad 0 has two vectors, ad 1 one, ad 2 none
    assert list(ad_scores(query, vectors, owner, 3)) == [1.0, 0.0, NOTHING]


def test_precision_recall_and_ndcg():
    ranked = [3, 1, 4, 0, 2]
    relevant = {1, 0}
    assert precision_at(2, ranked, relevant) == 0.5
    assert recall_at(4, ranked, relevant) == 1.0
    assert recall_at(2, ranked, relevant) == 0.5
    # Gains at ranks 2 and 4, against both relevant at ranks 1 and 2.
    want = (1 / np.log2(3) + 1 / np.log2(5)) / (1 + 1 / np.log2(3))
    assert ndcg_at(4, ranked, relevant) == pytest.approx(want)
    assert ndcg_at(2, [1, 0], relevant) == pytest.approx(1.0)


def test_run_scores_each_representation():
    # Two ads; the query is a driver. Each text's vector is chosen by word.
    ads = [
        {
            "id": "a",
            "whole": "driver agency",
            "own": "driver",
            "core": "driver",
            "summary": "",
            "sections": ["driver"],
        },
        {
            "id": "b",
            "whole": "chef agency",
            "own": "chef",
            "core": "",
            "summary": "chef",
            "sections": ["chef", "agency"],
        },
    ]
    queries = [{"query": "driver", "relevant": ["a"]}]
    words = {"driver": [1.0, 0.0], "chef": [0.0, 1.0]}

    def embed(texts):
        return np.array(
            [words.get(t.split()[0], [0.7, 0.7]) for t in texts], dtype=float
        )

    results = run(ads, queries, embed, embed)
    assert set(results) == {"whole", "own", "core", "summary", "sections"}
    assert results["own"].precision_at_10 == 0.1  # the one relevant ad, of ten places
    assert results["core"].vectors_per_ad == 0.5  # b has no core text
    assert results["sections"].vectors_per_ad == 1.5
    assert results["summary"].chars_per_ad == 2.0  # "chef", over two ads
