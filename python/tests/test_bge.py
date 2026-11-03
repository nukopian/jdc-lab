"""Tests for jdclab.bge: queries get BGE's instruction, passages don't."""

import numpy as np

from jdclab.bge import QUERY, embedder


class Model:
    """Stands in for a SentenceTransformer, keeping what it was asked."""

    max_seq_length = 512

    def __init__(self):
        self.asked = []

    def encode(self, texts, **options):
        self.asked.append((list(texts), options["normalize_embeddings"]))
        return np.zeros((len(texts), 2))


def test_queries_get_the_instruction_and_passages_dont():
    model = Model()
    e = embedder(model, "cpu")
    e.embed(["a chef"])
    e.embed_queries(["Chef"])
    assert model.asked == [(["a chef"], True), ([QUERY + "Chef"], True)]
    assert (e.device, e.max_tokens) == ("cpu", 512)
