"""A small ONNX model, for checking that ONNX Runtime works from each language.

The model takes a batch of vectors and returns how alike each pair is (cosine
similarity), as the lab does when comparing a heading's embedding with each
section type's. It's built here, so it needs no download. The Go and Perl
demos run the same model, on the same vectors, and must get the same scores.
"""

from pathlib import Path

import numpy as np
import onnx
import onnxruntime as ort
from onnx import TensorProto, helper

NAMES = ["responsibilities", "duties", "benefits"]
VECTORS = [[0.9, 0.1, 0.0], [0.8, 0.3, 0.1], [0.1, 0.2, 0.9]]


def similarity_model() -> onnx.ModelProto:
    """Build the model: vectors in, a table of pairwise cosine similarities out."""
    # Scale each vector to length 1, then multiply the batch by its own
    # transpose: entry [i][j] is the cosine of the angle between i and j.
    vectors = helper.make_tensor_value_info("vectors", TensorProto.FLOAT, [None, None])
    scores = helper.make_tensor_value_info("scores", TensorProto.FLOAT, [None, None])
    graph = helper.make_graph(
        [
            helper.make_node("LpNormalization", ["vectors"], ["unit"], axis=1, p=2),
            helper.make_node("Transpose", ["unit"], ["unit_t"]),
            helper.make_node("MatMul", ["unit", "unit_t"], ["scores"]),
        ],
        "cosine_similarity",
        [vectors],
        [scores],
    )
    model = helper.make_model(graph, opset_imports=[helper.make_opsetid("", 17)])
    # onnx writes its newest IR version by default, which can be newer than
    # ONNX Runtime reads; 9 is all opset 17 needs.
    model.ir_version = 9
    onnx.checker.check_model(model)
    return model


def save_model(path: Path) -> None:
    """Save the model to path, making its folder if need be."""
    path.parent.mkdir(parents=True, exist_ok=True)
    onnx.save(similarity_model(), path)


def similarities(session: ort.InferenceSession, rows) -> np.ndarray:
    """Run the model on rows, all the same length; return the score for each pair."""
    return session.run(None, {"vectors": np.asarray(rows, dtype=np.float32)})[0]
