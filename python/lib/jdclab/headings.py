"""Suggests a kind of section for each unmapped heading, by its nearest mapped ones.

A run's sections place each heading the rules know as a kind (summary,
requirements, benefits and the rest), and list the rest in unmapped.tsv.
Embedded, each unmapped heading's nearest mapped headings vote for their
kinds, each vote weighed by how alike the two are. The suggestions are for
reading, to find phrases for internal/sections/headings.txt: the vote always
picks a kind, even for a heading that names an advertiser, a slogan or a code,
which no kind fits.
"""

import csv
import json
import unicodedata
from collections import Counter, defaultdict
from collections.abc import Callable, Sequence
from dataclasses import dataclass
from pathlib import Path

import numpy as np

# How many mapped headings vote on each heading.
K = 10

# Key's folding of typographic characters, as internal/headings/key.go has it.
_FOLD = {
    **dict.fromkeys(map(chr, (0x2018, 0x2019, 0x201A, 0x201B, 0x2032)), "'"),
    **dict.fromkeys(
        map(chr, (0x201C, 0x201D, 0x201E, 0x201F, 0x2033, 0x00AB, 0x00BB)), '"'
    ),
    **dict.fromkeys(
        map(chr, (0x2010, 0x2011, 0x2012, 0x2013, 0x2014, 0x2015, 0x2212)), " "
    ),
    **dict.fromkeys(
        map(chr, (0x2022, 0x2023, 0x2043, 0x25AA, 0x25CF, 0x00B7, 0x2981, 0xF0B7)),
        "",
    ),
    chr(0x2026): "...",
    **dict.fromkeys(
        map(chr, (0x00A0, *range(0x2000, 0x200B), 0x202F, 0x205F, 0x3000)), " "
    ),
    **dict.fromkeys(
        map(
            chr,
            (0x200B, 0x200C, 0x200D, 0x2060, 0xFEFF, 0x00AD, 0xFE0F, 0xFE0E),
        ),
        "",
    ),
}


def key(text: str) -> str:
    """Returns a heading's key, as Go's headings.Key makes it.

    Letters and digits are kept, in lower case, with ', & and /; symbols are
    dropped; anything else is a space, and runs of spaces are one.
    """
    out = []
    for c in text:
        if c in _FOLD:
            out.append(_FOLD[c])
            continue
        cat = unicodedata.category(c)
        if cat[0] == "L" or cat == "Nd":
            out.append(c.lower())
        elif c in "'&/":
            out.append(c)
        elif cat not in ("So", "Sk", "Co"):
            out.append(" ")
    return " ".join("".join(out).split())


@dataclass
class Mapped:
    """The headings the rules placed: each key, its kind, and its uses."""

    keys: list[str]
    kinds: list[str]
    uses: list[int]


def load_mapped(path: Path) -> Mapped:
    """Reads a run's sections/ads.jsonl for every heading the rules placed.

    A key placed as more than one kind, as a field's label can be, takes the
    kind it was placed as most.
    """
    counts: defaultdict[str, Counter[str]] = defaultdict(Counter)
    with path.open(encoding="utf-8") as f:
        for line in f:
            for section in json.loads(line)["sections"]:
                heading = section.get("heading")
                if heading and section["kind"] != "unmapped" and (k := key(heading)):
                    counts[k][section["kind"]] += 1
    keys = list(counts)
    return Mapped(
        keys,
        [counts[k].most_common(1)[0][0] for k in keys],
        [counts[k].total() for k in keys],
    )


def load_unmapped(path: Path) -> list[tuple[str, int, str]]:
    """Reads a run's sections/unmapped.tsv: each heading's key, ads and text."""
    with path.open(encoding="utf-8") as f:
        return [
            (row["key"], int(row["ads"]), row["heading"])
            for row in csv.DictReader(f, delimiter="\t")
        ]


@dataclass
class Vote:
    """A vote's outcome: the kind, its share of the votes, and the voters.

    nearest is how alike the closest voter is (cosine), and voters are their
    indexes into the mapped headings, closest first.
    """

    kind: str
    share: float
    nearest: float
    voters: list[int]


def vote(sims: np.ndarray, kinds: Sequence[str], k: int = K) -> Vote:
    """Has the k mapped headings most like one heading vote for its kind.

    sims holds how alike the heading is to each mapped heading; a mapped
    heading to leave out of the vote has a similarity below -1.
    """
    voters = np.argpartition(-sims, k)[:k]
    voters = voters[np.argsort(-sims[voters])]
    weight: defaultdict[str, float] = defaultdict(float)
    for i in voters:
        weight[kinds[i]] += float(sims[i])
    kind = max(weight, key=weight.__getitem__)
    return Vote(
        kind,
        weight[kind] / sum(weight.values()),
        float(sims[voters[0]]),
        voters.tolist(),
    )


def leave_one_out(vectors: np.ndarray, mapped: Mapped) -> str:
    """Says how often each mapped heading's neighbours give back its kind.

    This flatters the model: mapped headings share the phrases the rules
    matched, so each one's neighbours are near copies of it. Unmapped
    headings have no such phrases.
    """
    sims = vectors @ vectors.T
    np.fill_diagonal(sims, -2.0)
    votes = [vote(row, mapped.kinds) for row in sims]
    right = np.array([v.kind == k for v, k in zip(votes, mapped.kinds, strict=True)])
    share = np.array([v.share for v in votes])
    uses = np.array(mapped.uses)
    lines = [
        (
            f"leave one out: right for {right.mean():.1%} of mapped keys, "
            f"{(right * uses).sum() / uses.sum():.1%} of their uses"
        )
    ]
    for least in (0.5, 0.7, 0.9):
        sure = share >= least
        lines.append(
            f"  share >= {least}: {sure.mean():.0%} of keys, "
            f"right for {right[sure].mean():.1%}"
        )
    return "\n".join(lines) + "\n"


def suggest(
    embed: Callable[[list[str]], np.ndarray], mapped: Mapped, unmapped
) -> tuple[str, list[str]]:
    """Embeds the headings and votes on each unmapped one.

    Returns the leave-one-out summary, and a TSV line for each unmapped
    heading, in unmapped.tsv's order, with its three closest voters.
    """
    m = embed(mapped.keys)
    u = embed([k for k, _, _ in unmapped])
    rows = ["ads\tkind\tshare\tnearest\theading\tvoters"]
    for sims, (_, ads, heading) in zip(u @ m.T, unmapped, strict=True):
        v = vote(sims, mapped.kinds)
        voters = " | ".join(
            f"{mapped.keys[i]} ({mapped.kinds[i]})" for i in v.voters[:3]
        )
        rows.append(
            f"{ads}\t{v.kind}\t{v.share:.2f}\t{v.nearest:.3f}\t{heading}\t{voters}"
        )
    return leave_one_out(m, mapped), rows
