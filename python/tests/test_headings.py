"""Tests for jdclab.headings: keys as Go makes them, and the vote."""

import json

import numpy as np
import pytest

from jdclab.headings import Mapped, key, load_mapped, suggest, vote


@pytest.mark.parametrize(
    ("text", "want"),
    [
        ("What You\u2019ll Be Doing:", "what you'll be doing"),
        ("\u2022 Pay & Benefits \u2013 2026", "pay & benefits 2026"),
        ("Day/Night\u00a0 shifts\u2026", "day/night shifts..."),
        ("\u2b50 Why join us? \U0001f680", "why join us"),
    ],
)
def test_key_folds_as_go_does(text, want):
    assert key(text) == want


def test_vote_weighs_by_likeness():
    kinds = ["benefits", "requirements", "requirements", "company"]
    sims = np.array([0.9, 0.5, 0.45, 0.1])
    v = vote(sims, kinds, k=3)
    assert v.kind == "requirements"
    assert v.share == pytest.approx(0.95 / 1.85)
    assert v.nearest == pytest.approx(0.9)
    assert v.voters == [0, 1, 2]


def test_load_mapped_takes_each_keys_commonest_kind(tmp_path):
    ads = [
        {"sections": [{"kind": "summary"}, {"kind": "details", "heading": "Pay:"}]},
        {"sections": [{"kind": "benefits", "heading": "pay"}]},
        {"sections": [{"kind": "details", "heading": "PAY"}]},
        {"sections": [{"kind": "unmapped", "heading": "Please note"}]},
    ]
    path = tmp_path / "ads.jsonl"
    path.write_text("".join(json.dumps(a) + "\n" for a in ads), encoding="utf-8")
    m = load_mapped(path)
    assert (m.keys, m.kinds, m.uses) == (["pay"], ["details"], [3])


def test_suggest_votes_for_each_unmapped_heading():
    # Each text's vector is chosen by its first word.
    axes = {"duties": [1.0, 0.0], "perks": [0.0, 1.0], "tasks": [0.9, 0.1]}
    mapped = Mapped(
        [f"{w} {i}" for w in ("duties", "perks") for i in range(6)],
        ["responsibilities"] * 6 + ["benefits"] * 6,
        [1] * 12,
    )

    def embed(texts):
        vs = np.array([axes[t.split()[0]] for t in texts])
        return vs / np.linalg.norm(vs, axis=1, keepdims=True)

    summary, rows = suggest(embed, mapped, [("tasks", 4, "Tasks:")])
    assert "right for 100.0% of mapped keys" in summary
    assert rows[1].split("\t")[:2] == ["4", "responsibilities"]
    assert rows[1].split("\t")[4] == "Tasks:"
