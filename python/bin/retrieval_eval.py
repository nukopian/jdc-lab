"""Runs the first test of retrieval: which representation of an ad finds ads for a job.

It reads the test cmd/retrieval-set wrote, embeds each representation of
the ads with BAAI/bge-base-en-v1.5, ranks the ads for each job title, and
writes how each representation scored to DIR/report.txt, printing it too.
The work is in jdclab.retrieval, which python/tests/test_retrieval.py tests.

From the lab's folder, with direnv on (it puts python/lib on PYTHONPATH),
after go run ./cmd/retrieval-set data/runs/latest/ads.jsonl data/runs/latest/retrieval:

    python python/bin/retrieval_eval.py data/runs/latest/retrieval
"""

import sys
from pathlib import Path

import torch
from sentence_transformers import SentenceTransformer

from jdclab.retrieval import load, report, run

MODEL = "BAAI/bge-base-en-v1.5"
# BGE's instruction for short queries looking for passages.
QUERY = "Represent this sentence for searching relevant passages: "


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: retrieval_eval.py DIR")
    directory = Path(sys.argv[1])
    ads, queries = load(directory)
    device = "mps" if torch.backends.mps.is_available() else "cpu"
    model = SentenceTransformer(MODEL, device=device)

    def embed(texts):
        return model.encode(
            list(texts),
            batch_size=32,
            normalize_embeddings=True,
            show_progress_bar=True,
        )

    def embed_queries(texts):
        return model.encode([QUERY + t for t in texts], normalize_embeddings=True)

    text = report(run(ads, queries, embed, embed_queries), len(ads), queries)
    text += (
        f"\nmodel {MODEL}, at most {model.max_seq_length} tokens a text, on {device}\n"
    )
    (directory / "report.txt").write_text(text, encoding="utf-8")
    print(text, end="")


if __name__ == "__main__":
    main()
