"""A first test of retrieval: which representation of an ad finds ads for a job.

The Go program cmd/retrieval-set writes the test: a sample's ads, with the
texts each representation embeds (whole, own, core, summary, and sections,
a vector each), and job titles as queries, with the ads relevant to each.
This module embeds them, ranks the ads for each query, and scores the
rankings. The model is given, so the tests can use made-up vectors.
"""

import json
import math
from collections.abc import Callable, Sequence
from dataclasses import dataclass
from pathlib import Path

import numpy as np

# The representations that are one text for each ad, as cmd/retrieval-set
# writes them; "sections" is the other, a text for each section.
SINGLE = ["whole", "own", "core", "summary"]
REPRESENTATIONS = [*SINGLE, "sections"]

# What an ad with nothing for a representation scores: below any cosine.
NOTHING = -2.0

Embed = Callable[[Sequence[str]], np.ndarray]


@dataclass
class Scores:
    """How one representation did, over every query."""

    precision_at_10: float
    ndcg_at_10: float
    recall_at_100: float
    chars_per_ad: float
    vectors_per_ad: float


def load(directory: Path) -> tuple[list[dict], list[dict]]:
    """Read the test's ads and queries from directory."""

    def lines(name: str) -> list[dict]:
        with open(directory / name, encoding="utf-8") as f:
            return [json.loads(line) for line in f]

    return lines("ads.jsonl"), lines("queries.jsonl")


def ad_scores(
    query: np.ndarray, vectors: np.ndarray, owner: np.ndarray, ads: int
) -> np.ndarray:
    """Score each ad for a query: the best cosine of any of its vectors.

    vectors are unit length, one a row, and owner[i] is the ad that row i is
    for; an ad with no vector scores NOTHING. With one vector an ad, this is
    the cosine itself.
    """
    best = np.full(ads, NOTHING)
    if len(vectors):
        np.maximum.at(best, owner, vectors @ query)
    return best


def precision_at(k: int, ranked: Sequence[int], relevant: set[int]) -> float:
    """The share of the first k ranked that are relevant."""
    return sum(1 for i in ranked[:k] if i in relevant) / k


def recall_at(k: int, ranked: Sequence[int], relevant: set[int]) -> float:
    """The share of the relevant that are in the first k ranked."""
    return sum(1 for i in ranked[:k] if i in relevant) / len(relevant)


def ndcg_at(k: int, ranked: Sequence[int], relevant: set[int]) -> float:
    """Discounted cumulative gain of the first k, against the best possible."""
    gain = sum(1 / math.log2(r + 2) for r, i in enumerate(ranked[:k]) if i in relevant)
    best = sum(1 / math.log2(r + 2) for r in range(min(k, len(relevant))))
    return gain / best


def run(
    ads: list[dict], queries: list[dict], embed: Embed, embed_queries: Embed
) -> dict[str, Scores]:
    """Embed every representation and the queries, and score each representation."""
    index = {ad["id"]: i for i, ad in enumerate(ads)}
    q = embed_queries([query["query"] for query in queries])
    relevant = [
        {index[i] for i in query["relevant"] if i in index} for query in queries
    ]
    results = {}
    for rep in REPRESENTATIONS:
        texts: list[str] = []
        owner: list[int] = []
        for i, ad in enumerate(ads):
            parts = (ad[rep] or []) if rep == "sections" else [ad[rep]]
            for text in parts:
                if text.strip():
                    texts.append(text)
                    owner.append(i)
        vectors = embed(texts) if texts else np.zeros((0, q.shape[1]))
        p10, n10, r100 = [], [], []
        for qv, rel in zip(q, relevant, strict=True):
            scores = ad_scores(qv, vectors, np.array(owner, dtype=int), len(ads))
            # Highest first; ties in order of the ads, so runs agree.
            ranked = [int(i) for i in np.argsort(-scores, kind="stable")]
            p10.append(precision_at(10, ranked, rel))
            n10.append(ndcg_at(10, ranked, rel))
            r100.append(recall_at(100, ranked, rel))
        results[rep] = Scores(
            precision_at_10=float(np.mean(p10)),
            ndcg_at_10=float(np.mean(n10)),
            recall_at_100=float(np.mean(r100)),
            chars_per_ad=sum(len(t) for t in texts) / len(ads),
            vectors_per_ad=len(texts) / len(ads),
        )
    return results


def report(results: dict[str, Scores], ads: int, queries: list[dict]) -> str:
    """A table of the results, with what chance would score."""
    relevant = sum(len(q["relevant"]) for q in queries)
    lines = [
        f"{ads} ads, {len(queries)} job titles as queries, {relevant} relevant ads in all",
        f"chance: precision at 10 {relevant / len(queries) / ads:.3f}",
        "",
        f"{'representation':16}{'P@10':>7}{'nDCG@10':>9}{'R@100':>8}{'chars/ad':>10}{'vectors/ad':>12}",
    ]
    for rep, s in results.items():
        lines.append(
            f"{rep:16}{s.precision_at_10:7.3f}{s.ndcg_at_10:9.3f}{s.recall_at_100:8.3f}"
            f"{s.chars_per_ad:10.0f}{s.vectors_per_ad:12.2f}"
        )
    return "\n".join(lines) + "\n"
