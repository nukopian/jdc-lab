"""A test of facts as filters: does filtering on a fact find ads better than embedding it?

The Go program cmd/retrieval-set writes the test, in its facts folder:
every distinct ad as a listing (its title, whole and own texts, its place
from the export, and the work patterns its facts found), and queries of a
job title with a fact, a place ("Care Assistant in Leeds") or a work
pattern ("part time Receptionist"). Each query's candidates are the ads
whose title is similar; a place query's relevant ads are those of them in
the place, and a work pattern's are judged, from a pool of what each way
ranks highest.

Each query is ranked two ways, for each representation:

- embedded: the whole query, fact and all, against every ad;
- filtered: the title alone, against the ads the fact lets through: the
  place, or the work pattern among the ad's facts.

The model is given, so the tests can use made-up vectors.
"""

import json
from collections.abc import Callable, Sequence
from dataclasses import dataclass
from pathlib import Path

import numpy as np

from jdclab.retrieval import ndcg_at, precision_at

REPRESENTATIONS = ["own", "whole"]
WAYS = ["embedded", "filtered"]
# How deep each way's ranking is kept, and how much of it is judged.
KEPT = 100
POOLED = 10

Embed = Callable[[Sequence[str]], np.ndarray]


def load(directory: Path) -> tuple[list[dict], list[dict]]:
    """Read the test's listings and queries from directory."""

    def lines(name: str) -> list[dict]:
        with open(directory / name, encoding="utf-8") as f:
            return [json.loads(line) for line in f]

    return lines("ads.jsonl"), lines("queries.jsonl")


def lets_through(query: dict, ad: dict) -> bool:
    """Whether a query's fact lets an ad through its filter."""
    if query["kind"] == "place":
        return ad["place"] == query["value"]
    return query["value"] in ad["patterns"]


def rank(
    ads: list[dict], queries: list[dict], embed: Embed, embed_queries: Embed
) -> list[dict]:
    """Rank the ads for each query, each way and representation.

    Returns one ranking a line: the query, the way, the representation,
    and the IDs of the first KEPT ads, best first, ties in the ads' order.
    A filtered ranking holds only the ads the fact lets through.
    """
    whole = embed_queries([q["query"] for q in queries])
    titles = embed_queries([q["title"] for q in queries])
    rankings = []
    for rep in REPRESENTATIONS:
        vectors = embed([ad[rep] for ad in ads])
        for q, qw, qt in zip(queries, whole, titles, strict=True):
            for way in WAYS:
                if way == "embedded":
                    allowed = np.arange(len(ads))
                    scores = vectors @ qw
                else:
                    allowed = np.array(
                        [i for i, ad in enumerate(ads) if lets_through(q, ad)],
                        dtype=int,
                    )
                    scores = vectors[allowed] @ qt if len(allowed) else np.zeros(0)
                order = np.argsort(-scores, kind="stable")[:KEPT]
                rankings.append(
                    {
                        "query": q["query"],
                        "way": way,
                        "representation": rep,
                        "ranked": [ads[int(allowed[i])]["id"] for i in order],
                    }
                )
    return rankings


def pool(ads: list[dict], queries: list[dict], rankings: list[dict]) -> list[dict]:
    """The ads to judge: for each work-pattern query, the candidates among
    the first POOLED of each of its rankings, each once, with the text a
    judge reads. Ads that aren't candidates, their title not similar, are
    not relevant, so aren't judged."""
    by_id = {ad["id"]: ad for ad in ads}
    out = []
    for q in queries:
        if q["kind"] != "pattern":
            continue
        candidates = set(q["candidates"])
        seen: set[str] = set()
        for r in rankings:
            if r["query"] != q["query"]:
                continue
            for i in r["ranked"][:POOLED]:
                if i in candidates and i not in seen:
                    seen.add(i)
                    ad = by_id[i]
                    out.append(
                        {
                            "query": q["query"],
                            "pattern": q["value"],
                            "id": i,
                            "title": ad["title"],
                            "text": ad["whole"],
                        }
                    )
    return out


@dataclass
class Scores:
    """How one way and representation did, over one kind of query."""

    precision_at_10: float
    ndcg_at_10: float
    recall: float
    # Relevant ads the filter can't let through, as a share of the relevant.
    filtered_out: float


def relevant_sets(queries: list[dict], judgements: list[dict]) -> dict[str, set[str]]:
    """Each query's relevant ads: a place's as given; a work pattern's, the
    candidates judged to be for it."""
    judged_for: dict[str, set[str]] = {}
    for j in judgements:
        if j["verdict"] == "for":
            judged_for.setdefault(j["query"], set()).add(j["id"])
    out = {}
    for q in queries:
        if q["kind"] == "place":
            out[q["query"]] = set(q["relevant"])
        else:
            out[q["query"]] = judged_for.get(q["query"], set()) & set(q["candidates"])
    return out


def score(
    ads: list[dict], queries: list[dict], rankings: list[dict], judgements: list[dict]
) -> dict[tuple[str, str, str], Scores]:
    """Score each kind of query, way and representation.

    Recall is at KEPT for places, whose relevant ads are all known, and at
    POOLED for work patterns, of the relevant ads in the pool: those any
    way found in its first POOLED. A query with no relevant ads is left
    out of its kind's means.
    """
    by_id = {ad["id"]: ad for ad in ads}
    relevant = relevant_sets(queries, judgements)
    by_query = {q["query"]: q for q in queries}
    sums: dict[tuple[str, str, str], list[list[float]]] = {}
    for r in rankings:
        q = by_query.get(r["query"])
        if q is None or not relevant[q["query"]]:
            continue
        rel = relevant[q["query"]]
        depth = KEPT if q["kind"] == "place" else POOLED
        ranked = r["ranked"]
        lost = sum(1 for i in rel if not lets_through(q, by_id[i])) / len(rel)
        key = (q["kind"], r["way"], r["representation"])
        sums.setdefault(key, []).append(
            [
                precision_at(10, ranked, rel),
                ndcg_at(10, ranked, rel),
                sum(1 for i in ranked[:depth] if i in rel) / len(rel),
                lost,
            ]
        )
    return {
        key: Scores(*(float(np.mean(column)) for column in zip(*rows, strict=True)))
        for key, rows in sums.items()
    }


def report(
    results: dict[tuple[str, str, str], Scores],
    queries: list[dict],
    judgements: list[dict],
) -> str:
    """A table of the results, with the queries and judgements behind them."""
    lines = []
    for kind, recall in [("place", f"R@{KEPT}"), ("pattern", f"R@{POOLED} in pool")]:
        qs = [q for q in queries if q["kind"] == kind]
        rel = relevant_sets(qs, judgements)
        scored = [r for r in rel.values() if r]
        # The best precision at 10 can be, as some queries have fewer.
        best = float(np.mean([min(10, len(r)) / 10 for r in scored])) if scored else 0.0
        lines.append(
            f"{kind}: {len(qs)} queries, {len(scored)} with relevant ads, "
            f"{sum(len(r) for r in scored)} relevant in all, best P@10 {best:.3f}"
        )
        lines.append(
            f"  {'way':10}{'rep':7}{'P@10':>7}{'nDCG@10':>9}{recall:>18}{'filtered out':>14}"
        )
        for (k, way, rep), s in sorted(results.items()):
            if k != kind:
                continue
            lost = f"{s.filtered_out:14.3f}" if way == "filtered" else f"{'':14}"
            lines.append(
                f"  {way:10}{rep:7}{s.precision_at_10:7.3f}{s.ndcg_at_10:9.3f}"
                f"{s.recall:18.3f}{lost}"
            )
        lines.append("")
    verdicts: dict[str, int] = {}
    for j in judgements:
        verdicts[j["verdict"]] = verdicts.get(j["verdict"], 0) + 1
    lines.append(
        "judgements: "
        + ", ".join(f"{v} {n}" for v, n in sorted(verdicts.items()))
        + f" ({sum(verdicts.values())} in all)"
    )
    return "\n".join(lines) + "\n"


def write_lines(path: Path, rows: list[dict]) -> None:
    """Write rows to path, one JSON object a line."""
    with open(path, "w", encoding="utf-8") as f:
        f.writelines(json.dumps(row, ensure_ascii=False) + "\n" for row in rows)


def read_lines(path: Path) -> list[dict]:
    """Read one JSON object a line from path."""
    with open(path, encoding="utf-8") as f:
        return [json.loads(line) for line in f]
