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

from jdclab import bge
from jdclab.retrieval import load, report, run


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: retrieval_eval.py DIR")
    directory = Path(sys.argv[1])
    ads, queries = load(directory)
    model = bge.load()

    text = report(
        run(ads, queries, model.embed, model.embed_queries), len(ads), queries
    )
    text += f"\nmodel {bge.MODEL}, at most {model.max_tokens} tokens a text, on {model.device}\n"
    (directory / "report.txt").write_text(text, encoding="utf-8")
    print(text, end="")


if __name__ == "__main__":
    main()
