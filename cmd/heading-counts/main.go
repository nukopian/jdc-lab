// Command heading-counts finds the headings and fields in every job ad (see
// internal/blocks and internal/headings), and counts them: how many ads use
// each heading, by its key, and each field label. Ads whose HTML is the same
// as an earlier one's are counted once.
//
// It writes two tables, tab-separated, most used first: headings.tsv and
// fields.tsv, with the number of ads, the key, and its commonest spelling.
// And it prints a summary.
//
// From the lab's folder:
//
//	go run ./cmd/heading-counts data/ingestion/ads-20261024.jsonl data/headings
//
// or, once installed with `go install ./cmd/...`, just heading-counts.
package main

import (
	"bufio"
	"cmp"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/walls"
)

// counter counts, for each key, the ads it's in and how it's spelt.
type counter map[string]*count

type count struct {
	ads      int
	spelling map[string]int
}

// add counts the keys of one ad's parts, each once however often it's used.
func (c counter) add(parts []headings.Part, role headings.Role) int {
	seen := map[string]bool{}
	for _, p := range parts {
		if p.Role != role {
			continue
		}
		k := c[p.Key]
		if k == nil {
			k = &count{spelling: map[string]int{}}
			c[p.Key] = k
		}
		if !seen[p.Key] {
			k.ads++
			seen[p.Key] = true
		}
		k.spelling[spelt(p, role)]++
	}
	return len(seen)
}

// spelt is how a part's heading or label is written: for a field, its text
// without its value.
func spelt(p headings.Part, role headings.Role) string {
	if role == headings.Field {
		return strings.TrimSpace(strings.TrimSuffix(p.Text, p.Value))
	}
	return p.Text
}

// row is one line of a table.
type row struct {
	ads           int
	key, spelling string
}

// rows returns the counter's rows, most ads first, then by key.
func (c counter) rows() []row {
	var rs []row
	for key, k := range c {
		var best string
		for s, n := range k.spelling {
			if n > k.spelling[best] || (n == k.spelling[best] && s < best) {
				best = s
			}
		}
		rs = append(rs, row{k.ads, key, best})
	}
	slices.SortFunc(rs, func(a, b row) int {
		return cmp.Or(cmp.Compare(b.ads, a.ads), cmp.Compare(a.key, b.key))
	})
	return rs
}

func (c counter) write(path string) (err error) {
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
	fmt.Fprintln(w, "ads\tkey\tspelling")
	for _, r := range c.rows() {
		fmt.Fprintf(w, "%d\t%s\t%s\n", r.ads, r.key, r.spelling)
	}
	return w.Flush()
}

// summary is what's printed at the end.
type summary struct {
	ads, duplicates, none int
	perAd                 []int // headings in each ad
	roles                 map[headings.Role]int
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: heading-counts ADS.jsonl OUTDIR")
	}
	source, dir := os.Args[1], os.Args[2]
	heads, fields, s, err := tally(source)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	for name, c := range map[string]counter{"headings.tsv": heads, "fields.tsv": fields} {
		if err := c.write(filepath.Join(dir, name)); err != nil {
			log.Fatal(err)
		}
	}
	s.print(heads, fields, dir)
}

func tally(source string) (heads, fields counter, s summary, err error) {
	heads, fields = counter{}, counter{}
	s.roles = map[headings.Role]int{}
	var d ads.Distinct
	for ad, err := range ads.File(source) {
		if err != nil {
			return nil, nil, s, err
		}
		if !d.First(ad) {
			continue
		}
		s.ads++

		parts := headings.Read(walls.Split(blocks.Split(ad.HTML)), ad.Title, ad.Locations()...)
		for _, p := range parts {
			s.roles[p.Role]++
		}
		n := heads.add(parts, headings.Heading)
		fields.add(parts, headings.Field)
		s.perAd = append(s.perAd, n)
		if n == 0 {
			s.none++
		}
	}
	s.duplicates = d.Duplicates
	return heads, fields, s, nil
}

func (s summary) print(heads, fields counter, dir string) {
	slices.Sort(s.perAd)
	fmt.Printf("%d ads (%d duplicates left out); tables written to %s\n\n", s.ads, s.duplicates, dir)
	var counts []string
	for _, r := range []headings.Role{headings.Text, headings.Heading, headings.Field, headings.Repeat} {
		counts = append(counts, fmt.Sprintf("%s %d", r, s.roles[r]))
	}
	fmt.Printf("parts:    %s\n\n", strings.Join(counts, "   "))
	fmt.Printf("ads with no heading:     %d (%.0f%%)\n", s.none, pct(s.none, s.ads))
	median := 0
	if len(s.perAd) > 0 {
		median = s.perAd[len(s.perAd)/2]
	}
	fmt.Printf("headings in an ad:       median %d\n", median)

	rows := heads.rows()
	total, once := 0, 0
	for _, r := range rows {
		total += r.ads
		if r.ads == 1 {
			once++
		}
	}
	fmt.Printf("distinct headings:       %d (%d in one ad only)\n", len(rows), once)
	for _, n := range []int{25, 50, 100, 200} {
		covered := 0
		for _, r := range rows[:min(n, len(rows))] {
			covered += r.ads
		}
		fmt.Printf("  the top %-3d cover      %.0f%% of uses\n", n, pct(covered, total))
	}
	fmt.Printf("distinct field labels:   %d\n", len(fields))
}

func pct(n, of int) float64 {
	if of == 0 {
		return 0
	}
	return 100 * float64(n) / float64(of)
}
