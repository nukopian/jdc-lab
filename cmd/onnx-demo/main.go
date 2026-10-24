// Command onnx-demo runs a small ONNX model from Go, with ONNX Runtime.
//
// The model, saved to models/onnx-demo.onnx by python/bin/onnx_demo.py, takes
// a batch of vectors and returns how alike each pair is (cosine
// similarity). This program runs it on the same vectors as the Python and
// Perl demos, and should print the same table.
//
// After running the Python demo, from anywhere in the lab:
//
//	go run ./cmd/onnx-demo
//
// or, once installed with `go install ./cmd/...`, just onnx-demo. It finds
// the lab by looking upwards from the current folder for go.mod.
//
// ONNX Runtime's shared library is loaded from $ONNXRUNTIME_LIB, or else
// from Homebrew's /opt/homebrew/lib/libonnxruntime.dylib.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	ort "github.com/yalue/onnxruntime_go"
)

var (
	names   = []string{"responsibilities", "duties", "benefits"}
	vectors = [][]float32{{0.9, 0.1, 0.0}, {0.8, 0.3, 0.1}, {0.1, 0.2, 0.9}}
)

func main() {
	log.SetFlags(0)

	lib := os.Getenv("ONNXRUNTIME_LIB")
	if lib == "" {
		lib = "/opt/homebrew/lib/libonnxruntime.dylib"
	}
	ort.SetSharedLibraryPath(lib)
	if err := ort.InitializeEnvironment(); err != nil {
		log.Fatalf("loading ONNX Runtime from %s: %v", lib, err)
	}
	defer ort.DestroyEnvironment()

	lab, err := labRoot()
	if err != nil {
		log.Fatal(err)
	}
	model := filepath.Join(lab, "models", "onnx-demo.onnx")
	if _, err := os.Stat(model); err != nil {
		log.Fatalf("%v: run python python/bin/onnx_demo.py first", err)
	}

	scores, err := similarities(model, vectors)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ONNX Runtime %s, called from Go %s\n\n", ort.GetVersion(), runtime.Version()[2:])
	fmt.Printf("%-18s", "")
	for _, name := range names {
		fmt.Printf("%18s", name)
	}
	fmt.Println()
	for i, name := range names {
		fmt.Printf("%-18s", name)
		for _, score := range scores[i] {
			fmt.Printf("%18.3f", score)
		}
		fmt.Println()
	}
}

// labRoot returns the lab's folder: the nearest one at or above the
// current folder that holds go.mod.
func labRoot() (string, error) {
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

// similarities runs the model file on rows, which must all be the same
// length, and returns the score for each pair.
func similarities(model string, rows [][]float32) ([][]float32, error) {
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
		scores[i] = data[int64(i)*n : int64(i+1)*n]
	}
	return scores, nil
}
