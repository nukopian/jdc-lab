"""Runs the test of facts as filters: does filtering on a fact find ads better than embedding it?

It reads the test cmd/retrieval-set wrote in DIR (its facts folder), and,
by its first argument:

- rank: embeds every ad's own and whole texts with BAAI/bge-base-en-v1.5,
  ranks the ads for each query both ways, and writes DIR/rankings.jsonl;
- pool: writes DIR/pool.jsonl, the ads each work-pattern query's rankings
  put first, to be judged into DIR/judgements.jsonl, one a line, as
  {"query", "id", "verdict": "for" | "against" | "silent", "reason"};
- score: scores the rankings against the places and the judgements, and
  writes DIR/report.txt, printing it too.

The work is in jdclab.filters, which python/tests/test_filters.py tests.

From the lab's folder, with direnv on (it puts python/lib on PYTHONPATH),
after go run ./cmd/retrieval-set data/runs/latest/ads.jsonl data/runs/latest/retrieval:

    python python/bin/retrieval_facts.py rank data/runs/latest/retrieval/facts
"""

import sys
from pathlib import Path

from jdclab.filters import load, pool, rank, read_lines, report, score, write_lines

MODEL = "BAAI/bge-base-en-v1.5"
# BGE's instruction for short queries looking for passages.
QUERY = "Represent this sentence for searching relevant passages: "


def main():
    if len(sys.argv) != 3 or sys.argv[1] not in ("rank", "pool", "score"):
        sys.exit("usage: retrieval_facts.py rank|pool|score DIR")
    step, directory = sys.argv[1], Path(sys.argv[2])
    ads, queries = load(directory)
    if step == "rank":
        # Imported here, as only ranking needs the model.
        import torch
        from sentence_transformers import SentenceTransformer

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

        write_lines(
            directory / "rankings.jsonl", rank(ads, queries, embed, embed_queries)
        )
        print(
            f"{len(ads)} ads ranked for {len(queries)} queries with {MODEL} on {device}"
        )
    elif step == "pool":
        rows = pool(ads, queries, read_lines(directory / "rankings.jsonl"))
        write_lines(directory / "pool.jsonl", rows)
        print(f"{len(rows)} ads to judge, written to {directory / 'pool.jsonl'}")
    else:
        rankings = read_lines(directory / "rankings.jsonl")
        judgements = read_lines(directory / "judgements.jsonl")
        text = report(score(ads, queries, rankings, judgements), queries, judgements)
        text += f"\nmodel {MODEL}\n"
        (directory / "report.txt").write_text(text, encoding="utf-8")
        print(text, end="")


if __name__ == "__main__":
    main()
