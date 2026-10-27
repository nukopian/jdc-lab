// Command retrieval-set builds a test of retrieval from a sample's ads
// (see internal/retrieval): 5,000 of its distinct ads, chosen by a fixed
// hash of their IDs, with the texts each representation of them embeds,
// and 50 job titles as queries, with the ads relevant to each. Python
// embeds and scores them: python/bin/retrieval_test.py.
//
// It writes OUTDIR/ads.jsonl and OUTDIR/queries.jsonl, one JSON object a
// line, and prints a summary.
//
// From the lab's folder:
//
//	go run ./cmd/retrieval-set data/runs/latest/ads.jsonl data/runs/latest/retrieval
//
// or, once installed with `go install ./cmd/...`, just retrieval-set.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/boilerplate"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/retrieval"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// The test's size: ads in the sample, queries, and the fewest relevant
// ads a query needs.
const (
	sampled = 5000
	queries = 50
	least   = 5
)

// an ad, with its sections.
type divided struct {
	ad ads.Ad
	ss []sections.Section
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: retrieval-set ADS.jsonl OUTDIR")
	}
	source, dir := os.Args[1], os.Args[2]
	all, ix, err := read(source)
	if err != nil {
		log.Fatal(err)
	}

	byID := map[string]divided{}
	var ids []string
	for _, d := range all {
		byID[d.ad.ID] = d
		ids = append(ids, d.ad.ID)
	}
	sample := retrieval.Sample(ids, sampled)
	titles := map[string]string{}
	var texts []retrieval.Texts
	for _, id := range sample {
		d := byID[id]
		src := boilerplate.SourceOf(d.ad)
		titles[id] = d.ad.Title
		texts = append(texts, retrieval.TextsOf(id, d.ss, func(s string) bool { return ix.Boilerplate(src, s) }))
	}
	qs := retrieval.Queries(titles, queries, least)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	if err := writeLines(filepath.Join(dir, "ads.jsonl"), texts); err != nil {
		log.Fatal(err)
	}
	if err := writeLines(filepath.Join(dir, "queries.jsonl"), qs); err != nil {
		log.Fatal(err)
	}
	relevant := 0
	for _, q := range qs {
		relevant += len(q.Relevant)
	}
	fmt.Printf("%d ads of %d, and %d queries with %d relevant ads in all, written to %s\n",
		len(texts), len(all), len(qs), relevant, dir)
}

// read reads the sample's distinct ads, divided into sections, and
// indexes their texts for boilerplate, as cmd/boilerplate does.
func read(source string) (all []divided, ix boilerplate.Index, err error) {
	var d ads.Distinct
	for a, err := range ads.File(source) {
		if err != nil {
			return nil, ix, err
		}
		if !d.First(a) {
			continue
		}
		ss := sections.Divide(headings.Read(blocks.Split(a.HTML), a.Title, a.Locations()...))
		for _, s := range ss {
			var texts []string
			for _, p := range s.Parts {
				if p.Role == headings.Text {
					texts = append(texts, p.Text)
				}
			}
			ix.Add(boilerplate.SourceOf(a), s.Kind, texts...)
		}
		all = append(all, divided{a, ss})
	}
	return all, ix, nil
}

// writeLines writes each of vs to path as a line of JSON.
func writeLines[T any](path string, vs []T) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, v := range vs {
		if err := enc.Encode(v); err != nil {
			return err
		}
	}
	return w.Flush()
}
