// Command retrieval-set builds a test of retrieval from a sample's ads
// (see internal/retrieval): 5,000 of its distinct ads, chosen by a fixed
// hash of their IDs, with the texts each representation of them embeds,
// and 50 job titles as queries, with the ads relevant to each. Python
// embeds and scores them: python/bin/retrieval_eval.py.
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

	"github.com/nukopian/jdc-lab/internal/boilerplate"
	"github.com/nukopian/jdc-lab/internal/retrieval"
)

// The test's size: ads in the sample, queries, and the fewest relevant
// ads a query needs.
const (
	sampled = 5000
	queries = 50
	least   = 5
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: retrieval-set ADS.jsonl OUTDIR")
	}
	source, dir := os.Args[1], os.Args[2]
	all, ix, _, err := boilerplate.Read(source)
	if err != nil {
		log.Fatal(err)
	}

	byID := map[string]boilerplate.Divided{}
	var ids []string
	for _, d := range all {
		byID[d.Ad.ID] = d
		ids = append(ids, d.Ad.ID)
	}
	sample := retrieval.Sample(ids, sampled)
	titles := map[string]string{}
	var texts []retrieval.Texts
	for _, id := range sample {
		d := byID[id]
		src := boilerplate.SourceOf(d.Ad)
		titles[id] = d.Ad.Title
		texts = append(texts, retrieval.TextsOf(id, d.Sections, func(s string) bool { return ix.Boilerplate(src, s) }))
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
