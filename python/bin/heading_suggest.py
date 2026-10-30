"""Suggests a kind of section for each heading the rules don't know.

It reads a run's sections, embeds every distinct heading with
BAAI/bge-small-en-v1.5, and has each unmapped heading's ten nearest mapped
headings vote for its kind. It writes a line for each unmapped heading to
OUT, most used first, as unmapped.tsv has them, and prints how often the
vote gives back the rules' own kinds. The suggestions are for reading, to
find phrases for internal/sections/headings.txt, not to apply as they stand.
The work is in jdclab.headings, which python/tests/test_headings.py tests.

From the lab's folder, with direnv on (it puts python/lib on PYTHONPATH):

    python python/bin/heading_suggest.py data/runs/latest suggested.tsv
"""

import sys
from pathlib import Path

import torch
from sentence_transformers import SentenceTransformer

from jdclab.headings import load_mapped, load_unmapped, suggest

MODEL = "BAAI/bge-small-en-v1.5"


def main():
    if len(sys.argv) != 3:
        sys.exit("usage: heading_suggest.py RUN OUT")
    sections = Path(sys.argv[1]) / "sections"
    mapped = load_mapped(sections / "ads.jsonl")
    unmapped = load_unmapped(sections / "unmapped.tsv")
    device = "mps" if torch.backends.mps.is_available() else "cpu"
    model = SentenceTransformer(MODEL, device=device)

    def embed(texts):
        return model.encode(texts, batch_size=256, normalize_embeddings=True)

    summary, rows = suggest(embed, mapped, unmapped)
    Path(sys.argv[2]).write_text("\n".join(rows) + "\n", encoding="utf-8")
    print(
        f"{len(mapped.keys)} mapped and {len(unmapped)} unmapped headings, "
        f"by {MODEL} on {device}\n{summary}",
        end="",
    )


if __name__ == "__main__":
    main()
