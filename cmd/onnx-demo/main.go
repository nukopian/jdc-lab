// Command onnx-demo runs the lab's small ONNX model from Go, with ONNX
// Runtime, and prints how alike each pair of its vectors is. The Python and
// Perl demos print the same table. The work is in internal/onnxdemo, which
// its tests check.
//
// After running the Python demo, which saves the model, from anywhere in the
// lab:
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
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/nukopian/jdc-lab/internal/onnxdemo"
)

func main() {
	log.SetFlags(0)

	lab, err := onnxdemo.LabRoot()
	if err != nil {
		log.Fatal(err)
	}
	model := onnxdemo.ModelPath(lab)
	if _, err := os.Stat(model); err != nil {
		log.Fatalf("%v: run python python/bin/onnx_demo.py first", err)
	}

	lib := onnxdemo.Library()
	if err := onnxdemo.Start(lib); err != nil {
		log.Fatalf("loading ONNX Runtime from %s: %v", lib, err)
	}
	defer onnxdemo.Stop()

	scores, err := onnxdemo.Similarities(model, onnxdemo.Vectors)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ONNX Runtime %s, called from Go %s\n\n", onnxdemo.Version(), runtime.Version()[2:])
	fmt.Printf("%-18s", "")
	for _, name := range onnxdemo.Names {
		fmt.Printf("%18s", name)
	}
	fmt.Println()
	for i, name := range onnxdemo.Names {
		fmt.Printf("%-18s", name)
		for _, score := range scores[i] {
			fmt.Printf("%18.3f", score)
		}
		fmt.Println()
	}
}
