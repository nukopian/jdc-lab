// Package onnxdemo runs the lab's small ONNX model from Go, with ONNX Runtime.
//
// The model, saved to models/onnx-demo.onnx by python/bin/onnx_demo.py, takes
// a batch of vectors and returns how alike each pair is (cosine similarity).
// The Python and Perl demos run the same model on the same vectors, and must
// get the same scores.
package onnxdemo

import (
	"errors"
	"os"
	"path/filepath"

	ort "github.com/yalue/onnxruntime_go"
)

// Names and Vectors are the demo's input, the same in every language.
var (
	Names   = []string{"responsibilities", "duties", "benefits"}
	Vectors = [][]float32{{0.9, 0.1, 0.0}, {0.8, 0.3, 0.1}, {0.1, 0.2, 0.9}}
)

// Library returns where ONNX Runtime's shared library is: $ONNXRUNTIME_LIB,
// or else Homebrew's.
func Library() string {
	if lib := os.Getenv("ONNXRUNTIME_LIB"); lib != "" {
		return lib
	}
	return "/opt/homebrew/lib/libonnxruntime.dylib"
}

// Start loads ONNX Runtime from lib. Call Stop when done with it.
func Start(lib string) error {
	ort.SetSharedLibraryPath(lib)
	return ort.InitializeEnvironment()
}

// Stop releases ONNX Runtime.
func Stop() error {
	return ort.DestroyEnvironment()
}

// Version returns ONNX Runtime's version, once it's started.
func Version() string {
	return ort.GetVersion()
}

// LabRoot returns the lab's folder: the nearest one at or above the current
// folder that holds go.mod.
func LabRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("not in the lab: no go.mod here or above")
		}
		dir = parent
	}
}

// ModelPath returns where the Python demo saves the model, in the lab.
func ModelPath(lab string) string {
	return filepath.Join(lab, "models", "onnx-demo.onnx")
}

// Similarities runs the model file on rows, which must all be the same
// length, and returns the score for each pair.
func Similarities(model string, rows [][]float32) ([][]float32, error) {
	n, dim := int64(len(rows)), int64(len(rows[0]))
	flat := make([]float32, 0, n*dim)
	for _, row := range rows {
		flat = append(flat, row...)
	}

	input, err := ort.NewTensor(ort.NewShape(n, dim), flat)
	if err != nil {
		return nil, err
	}
	defer input.Destroy()
	output, err := ort.NewEmptyTensor[float32](ort.NewShape(n, n))
	if err != nil {
		return nil, err
	}
	defer output.Destroy()

	session, err := ort.NewAdvancedSession(model,
		[]string{"vectors"}, []string{"scores"},
		[]ort.Value{input}, []ort.Value{output}, nil)
	if err != nil {
		return nil, err
	}
	defer session.Destroy()
	if err := session.Run(); err != nil {
		return nil, err
	}

	data := output.GetData()
	scores := make([][]float32, n)
	for i := range scores {
		scores[i] = append([]float32(nil), data[int64(i)*n:int64(i+1)*n]...)
	}
	return scores, nil
}
