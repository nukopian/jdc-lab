"""Builds the lab's small ONNX model, saves it, and runs it with ONNX Runtime.

It saves the model to models/onnx-demo.onnx for the Go demo (cmd/onnx-demo) to
run too: the two should print the same table. perl/bin/onnx-demo.pl does the
same from Perl. The model itself is in jdclab.onnx_demo, which
python/tests/test_onnx_demo.py tests.

From the lab's folder, with direnv on (it puts python/lib on PYTHONPATH):

    python python/bin/onnx_demo.py
"""

import sys
from pathlib import Path

import onnxruntime as ort

from jdclab.onnx_demo import NAMES, VECTORS, save_model, similarities

# The lab's folder, two levels up from python/bin, wherever this is run from.
LAB = Path(__file__).resolve().parents[2]
MODEL_PATH = LAB / "models" / "onnx-demo.onnx"


def main():
    save_model(MODEL_PATH)
    scores = similarities(ort.InferenceSession(MODEL_PATH), VECTORS)

    python = ".".join(str(n) for n in sys.version_info[:3])
    print(f"ONNX Runtime {ort.__version__}, called from Python {python}")
    print(f"Model saved to {MODEL_PATH.relative_to(LAB)}\n")
    print(f"{'':18}" + "".join(f"{name:>18}" for name in NAMES))
    for name, row in zip(NAMES, scores, strict=True):
        print(f"{name:18}" + "".join(f"{score:18.3f}" for score in row))


if __name__ == "__main__":
    main()
