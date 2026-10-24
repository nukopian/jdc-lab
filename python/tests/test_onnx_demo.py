"""Tests for jdclab.onnx_demo: the model, and ONNX Runtime running it.

The Go and Perl tests check the same scores (EXPECTED), so all three
languages are held to the same numbers.
"""

import numpy as np
import onnx
import onnxruntime as ort
import pytest

from jdclab.onnx_demo import VECTORS, save_model, similarities, similarity_model

# The scores for each pair of VECTORS, to six places.
EXPECTED = {(0, 1): 0.962805, (0, 2): 0.130990, (1, 2): 0.288312}
TOLERANCE = 1e-6


@pytest.fixture(scope="module")
def scores():
    session = ort.InferenceSession(similarity_model().SerializeToString())
    return similarities(session, VECTORS)


def test_model_passes_the_checker():
    onnx.checker.check_model(similarity_model(), full_check=True)


def test_runtime_reads_the_models_ir_version():
    # onnx's default IR version can be newer than ONNX Runtime reads; loading
    # the model is what fails if it is.
    ort.InferenceSession(similarity_model().SerializeToString())


def test_each_vector_is_identical_to_itself(scores):
    np.testing.assert_allclose(np.diag(scores), 1.0, atol=TOLERANCE)


def test_scores_are_symmetric(scores):
    np.testing.assert_allclose(scores, scores.T, atol=TOLERANCE)


@pytest.mark.parametrize(("pair", "expected"), EXPECTED.items())
def test_known_scores(scores, pair, expected):
    i, j = pair
    assert scores[i][j] == pytest.approx(expected, abs=TOLERANCE)


def test_scores_match_numpy(scores):
    # The same arithmetic, done independently in double precision.
    v = np.asarray(VECTORS, dtype=np.float64)
    unit = v / np.linalg.norm(v, axis=1, keepdims=True)
    np.testing.assert_allclose(scores, unit @ unit.T, atol=TOLERANCE)


def test_saved_model_gives_the_same_scores(tmp_path, scores):
    path = tmp_path / "models" / "onnx-demo.onnx"
    save_model(path)
    np.testing.assert_array_equal(
        similarities(ort.InferenceSession(path), VECTORS), scores
    )
