"""BAAI/bge-base-en-v1.5, as the retrieval tests embed with it.

Passages are embedded as written, and queries with BGE's instruction for
short queries looking for passages, every vector unit length.
"""

from collections.abc import Callable, Sequence
from dataclasses import dataclass
from typing import Any

import numpy as np

MODEL = "BAAI/bge-base-en-v1.5"
# BGE's instruction for short queries looking for passages.
QUERY = "Represent this sentence for searching relevant passages: "


@dataclass
class Embedder:
    """A model's two ways of embedding, and what it ran on."""

    embed: Callable[[Sequence[str]], np.ndarray]
    embed_queries: Callable[[Sequence[str]], np.ndarray]
    device: str
    max_tokens: int


def embedder(model: Any, device: str) -> Embedder:
    """The ways of embedding with model, a SentenceTransformer, or anything
    with its encode."""

    def embed(texts: Sequence[str]) -> np.ndarray:
        return model.encode(
            list(texts),
            batch_size=32,
            normalize_embeddings=True,
            show_progress_bar=True,
        )

    def embed_queries(texts: Sequence[str]) -> np.ndarray:
        return model.encode([QUERY + t for t in texts], normalize_embeddings=True)

    return Embedder(embed, embed_queries, device, model.max_seq_length)


def load() -> Embedder:
    """Load MODEL, on the Mac's GPU where there is one."""
    # Imported here, as loading them is slow, and only embedding needs them.
    import torch
    from sentence_transformers import SentenceTransformer

    device = "mps" if torch.backends.mps.is_available() else "cpu"
    return embedder(SentenceTransformer(MODEL, device=device), device)
