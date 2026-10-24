"""Builds a small ONNX model, saves it, and runs it with ONNX Runtime.

The model takes a batch of vectors and returns how alike each pair is
(cosine similarity), as the lab does when comparing a heading's embedding
with each section type's. It's built here, so the demo needs no download,
and saved to models/onnx-demo.onnx for the Go demo (cmd/onnx-demo) to run
too: the two should print the same table. perl/bin/onnx-demo.pl does the same
from Perl.

From the lab's folder, with the venv active:

    python python/bin/onnx_demo.py
"""

import sys
from pathlib import Path

import numpy as np
import onnx
import onnxruntime as ort
from onnx import TensorProto, helper

# The lab's folder, two levels up from python/bin, wherever this is run from.
LAB = Path(__file__).resolve().parents[2]
MODEL_PATH = LAB / "models" / "onnx-demo.onnx"

NAMES = ["responsibilities", "duties", "benefits"]
VECTORS = [[0.9, 0.1, 0.0], [0.8, 0.3, 0.1], [0.1, 0.2, 0.9]]


def similarity_model():
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


def main():
    MODEL_PATH.parent.mkdir(exist_ok=True)
    onnx.save(similarity_model(), MODEL_PATH)

    session = ort.InferenceSession(MODEL_PATH)
    scores = session.run(None, {"vectors": np.asarray(VECTORS, dtype=np.float32)})[0]

    python = ".".join(str(n) for n in sys.version_info[:3])
    print(f"ONNX Runtime {ort.__version__}, called from Python {python}")
    print(f"Model saved to {MODEL_PATH.relative_to(LAB)}\n")
    print(f"{'':18}" + "".join(f"{name:>18}" for name in NAMES))
    for name, row in zip(NAMES, scores):
        print(f"{name:18}" + "".join(f"{score:18.3f}" for score in row))


if __name__ == "__main__":
    main()
