package onnxdemo

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// expected holds the scores for each pair of Vectors, to six places: the
// same numbers the Python and Perl tests check.
var expected = map[[2]int]float64{
	{0, 1}: 0.962805,
	{0, 2}: 0.130990,
	{1, 2}: 0.288312,
}

const tolerance = 1e-6

// skip says why the tests can't run here, if they can't: no ONNX Runtime
// library, or no model yet.
var skip string

func TestMain(m *testing.M) {
	lib := Library()
	if err := Start(lib); err != nil {
		skip = fmt.Sprintf("no ONNX Runtime at %s (%v): see docs/SETUP.md", lib, err)
		os.Exit(m.Run())
	}
	code := m.Run()
	Stop()
	os.Exit(code)
}

// scores runs the model, or skips the test if it can't.
func scores(t *testing.T) [][]float32 {
	t.Helper()
	if skip != "" {
		t.Skip(skip)
	}
	lab, err := LabRoot()
	if err != nil {
		t.Fatal(err)
	}
	model := ModelPath(lab)
	if _, err := os.Stat(model); err != nil {
		t.Skipf("no model at %s: run python python/bin/onnx_demo.py first", model)
	}
	s, err := Similarities(model, Vectors)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestEachVectorIsIdenticalToItself(t *testing.T) {
	s := scores(t)
	for i := range s {
		if math.Abs(float64(s[i][i])-1) > tolerance {
			t.Errorf("score[%d][%d] = %v, want 1", i, i, s[i][i])
		}
	}
}

func TestScoresAreSymmetric(t *testing.T) {
	s := scores(t)
	for i := range s {
		for j := range s[i] {
			if math.Abs(float64(s[i][j]-s[j][i])) > tolerance {
				t.Errorf("score[%d][%d] = %v but score[%d][%d] = %v", i, j, s[i][j], j, i, s[j][i])
			}
		}
	}
}

func TestKnownScores(t *testing.T) {
	s := scores(t)
	for pair, want := range expected {
		got := float64(s[pair[0]][pair[1]])
		if math.Abs(got-want) > tolerance {
			t.Errorf("score%v = %.6f, want %.6f", pair, got, want)
		}
	}
}

func TestLabRoot(t *testing.T) {
	lab, err := LabRoot()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(lab, "go.mod")); err != nil {
		t.Errorf("LabRoot() = %s, which has no go.mod", lab)
	}
}
