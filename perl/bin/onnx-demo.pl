#!/usr/bin/env perl
# Runs an ONNX model from Perl, through Inline::Python and ONNX Runtime.
#
# The model is built here, in Python, so the demo needs no download: it
# takes a batch of vectors and returns how alike each pair is (cosine
# similarity), as the lab does when comparing a heading's embedding with
# each section type's. The vectors go in as a Perl array of arrays and the
# scores come back as one. python/bin/onnx_demo.py and cmd/onnx-demo do the
# same from Python and Go, and all three print the same table.
#
# From the lab's folder, with direnv on (see HOWTO.md):
#
#   perl perl/bin/onnx-demo.pl
use v5.36;

use Inline Python => <<'END';
import numpy as np
import onnx
import onnxruntime as ort
from onnx import TensorProto, helper


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
    return model.SerializeToString()


session = ort.InferenceSession(similarity_model())


def similarities(rows):
    return session.run(None, {"vectors": np.asarray(rows, dtype=np.float32)})[0].tolist()


def runtime_version():
    return ort.__version__
END

my @names   = qw(responsibilities duties benefits);
my @vectors = ([0.9, 0.1, 0.0], [0.8, 0.3, 0.1], [0.1, 0.2, 0.9]);

printf "ONNX Runtime %s, called from Perl %vd\n\n", runtime_version(), $^V;
my $scores = similarities(\@vectors);
printf "%-18s%s\n", '', join '', map { sprintf '%18s', $_ } @names;
for my $i (keys @names) {
    printf "%-18s%s\n", $names[$i], join '',
        map { sprintf '%18.3f', $_ } $scores->[$i]->@*;
}
