// Command place-match finds the ads' canonical locations in a gazetteer
// that osm-gazetteer built, saves what each matched as SQLite, and
// prints how many names and ads ended in each status.
//
//	place-match [-gazetteer FILE] [-out FILE] ADS.jsonl...
//
// From the lab's folder:
//
//	go run ./cmd/place-match data/ingestion/ads-20261027.jsonl
//
// It reads local/osm/gazetteer.sqlite and writes data/isoc/locations.sqlite
// unless told otherwise. Ads with no canonical location are counted but
// not matched.
package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"text/tabwriter"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/gazetteer"
)

func main() {
	log.SetFlags(0)
	gaz := flag.String("gazetteer", "local/osm/gazetteer.sqlite", "the gazetteer to match against")
	out := flag.String("out", "data/isoc/locations.sqlite", "the SQLite file to write")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: place-match [-gazetteer FILE] [-out FILE] ADS.jsonl...")
	}
	ctx := context.Background()

	counts := make(map[string]int)
	missing := 0
	for _, path := range flag.Args() {
		for a, err := range ads.File(path) {
			if err != nil {
				log.Fatal(err)
			}
			if a.LocationCanonical == nil || *a.LocationCanonical == "" {
				missing++
				continue
			}
			counts[*a.LocationCanonical]++
		}
	}

	g, err := gazetteer.Load(ctx, *gaz)
	if err != nil {
		log.Fatalf("place-match: %v (build it with osm-gazetteer)", err)
	}
	var tallies []gazetteer.Tally
	for name, n := range counts {
		tallies = append(tallies, gazetteer.Tally{Match: g.Match(name), Count: n})
	}
	// Most-seen first, so the top of the table is what matters most.
	slices.SortFunc(tallies, func(a, b gazetteer.Tally) int {
		return cmp.Or(cmp.Compare(b.Count, a.Count), cmp.Compare(a.Name, b.Name))
	})

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := gazetteer.SaveMatches(ctx, *out, tallies); err != nil {
		log.Fatal(err)
	}
	report(tallies, missing)
	log.Printf("place-match: wrote %s", *out)
}

// report prints, for each status, how many names and ads ended in it, and
// how many of those have an outline.
func report(tallies []gazetteer.Tally, missing int) {
	type row struct{ names, ads, outlined int }
	rows := make(map[string]*row)
	var total row
	for _, t := range tallies {
		r := rows[t.Status]
		if r == nil {
			r = &row{}
			rows[t.Status] = r
		}
		outlined := (t.Place != nil && t.Place.Polygon != nil) || t.Area != nil
		for _, r := range []*row{r, &total} {
			r.names++
			r.ads += t.Count
			if outlined {
				r.outlined += t.Count
			}
		}
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "status\tnames\tads\tads outlined\t")
	for _, s := range []string{gazetteer.Matched, gazetteer.Chosen, gazetteer.OutOfArea, gazetteer.AreaOnly, gazetteer.Unmatched} {
		if r := rows[s]; r != nil {
			fmt.Fprintf(w, "%s\t%d\t%d\t%d\t\n", s, r.names, r.ads, r.outlined)
		}
	}
	fmt.Fprintf(w, "all\t%d\t%d\t%d\t\n", total.names, total.ads, total.outlined)
	if missing > 0 {
		fmt.Fprintf(w, "no canonical location\t\t%d\t\t\n", missing)
	}
	w.Flush()
}
