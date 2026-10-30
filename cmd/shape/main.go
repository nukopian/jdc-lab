// Command shape shows job ads in each of the shapes the pipeline gives them
// (see internal/shape): the HTML as stored, its blocks with their cues, the
// walls split into sentences, the role headings gives each, its sections
// and its chunks, one stage after another, so you can judge how an ad
// changes shape, and where it goes wrong.
//
//	shape [-run DIR] [-stage NAME,...] AD...
//
// An AD is an ad's ID in the run's ads. -stage shows only the stages named:
// posted, blocks, walls, headings, sections or chunks. The stages run on
// the code as it stands; only the chunks' boilerplate comes from the run, from its
// boilerplate/repeated.tsv, as chunk's does.
//
// From the lab's folder:
//
//	go run ./cmd/shape 225682313
//	go run ./cmd/shape -stage blocks,headings 225682313
//
// or, once installed with `go install ./cmd/...`, just shape.
package main

import (
	"bufio"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/boilerplate"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/shape"
)

func main() {
	log.SetFlags(0)
	run := flag.String("run", "data/runs/latest", "the run whose ads and boilerplate to use")
	stage := flag.String("stage", "", "the stages to show, by name, comma-separated: "+strings.Join(shape.Stages, ", "))
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: shape [-run DIR] [-stage NAME,...] AD...")
	}
	var stages []string
	if *stage != "" {
		stages = strings.Split(*stage, ",")
	}

	path := filepath.Join(*run, "boilerplate", "repeated.tsv")
	verdicts, err := boilerplate.Verdicts(path)
	if errors.Is(err, os.ErrNotExist) {
		log.Fatalf("shape: %s: no such file; run the pipeline first (go run ./cmd/pipeline run latest)", path)
	}
	if err != nil {
		log.Fatal(err)
	}
	isBoilerplate := func(s string) bool { return verdicts[headings.Key(s)] }

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, id := range flag.Args() {
		a, err := ads.Find(filepath.Join(*run, "ads.jsonl"), id)
		if errors.Is(err, ads.ErrNotFound) {
			w.Flush()
			log.Fatalf("shape: %s: no ad of that ID in %s", id, *run)
		}
		if err != nil {
			log.Fatal(err)
		}
		if err := shape.Write(w, a, isBoilerplate, stages...); err != nil {
			w.Flush()
			log.Fatal(err)
		}
	}
}
