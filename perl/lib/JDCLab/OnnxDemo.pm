package JDCLab::OnnxDemo;

# Runs the lab's small ONNX model from Perl, through Inline::Python and ONNX
# Runtime. The model is built in Python, so it needs no download: it takes a
# batch of vectors and returns how alike each pair is (cosine similarity).
# The Python and Go demos run the same model on the same vectors, and must
# get the same scores. Needs the venv's packages on PYTHONPATH, as the lab's
# .envrc sets.

use v5.36;

use Exporter qw(import);

our @EXPORT_OK = qw(names vectors similarities runtime_version);

use Inline Python => <<'END';
import numpy as np
import onnx
import onnxruntime as ort
from onnx import TensorProto, helper


def _similarity_model():
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
    return model.SerializeToString()


_session = ort.InferenceSession(_similarity_model())


def _similarities(rows):
    return _session.run(None, {"vectors": np.asarray(rows, dtype=np.float32)})[0].tolist()


def _runtime_version():
    return ort.__version__
END

my @NAMES   = qw(responsibilities duties benefits);
my @VECTORS = ([0.9, 0.1, 0.0], [0.8, 0.3, 0.1], [0.1, 0.2, 0.9]);

# The demo's input, the same in every language: copies, so callers can't
# change it.
sub names { return @NAMES }


sub vectors {
    return map { [@$_] } @VECTORS;
}

# Runs the model on an array of arrays, all the same length, and returns the
# score for each pair as another.
sub similarities ($rows) { return _similarities($rows) }

sub runtime_version { return _runtime_version() }

1;
