// Command xml-to-jsonl converts job ads exported by MySQL's --xml option to
// JSON Lines. Each line of the output is one ad, with the lab's field names:
// id, title, location, location_canonical and html. The work is in
// internal/mysqlxml, which its tests check.
//
// From the lab's folder:
//
//	go run ./cmd/xml-to-jsonl data/ingestion/extract.xml data/ingestion/ads.jsonl
//
// or, once installed with `go install ./cmd/...`, just xml-to-jsonl.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nukopian/jdc-lab/internal/mysqlxml"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: xml-to-jsonl EXPORT.xml OUT.jsonl")
	}
	source, target := os.Args[1], os.Args[2]
	n, err := convert(source, target)
	if err != nil {
		os.Remove(target) // half an output is worse than none
		log.Fatal(err)
	}
	fmt.Printf("%d ads written to %s\n", n, target)
}

// convert converts the export at source into target, so that main has one
// place to handle errors, and the deferred Close runs before it exits.
func convert(source, target string) (n int, err error) {
	in, err := os.Open(source)
	if err != nil {
		return 0, err
	}
	defer in.Close()

	out, err := os.Create(target)
	if err != nil {
		return 0, err
	}
	// A failed Close can mean the file wasn't fully written, so it counts.
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()
	return mysqlxml.Convert(in, out)
}
