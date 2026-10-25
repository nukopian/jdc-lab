// Command boilerplate finds the text job ads repeat whatever the job (see
// internal/boilerplate), and measures how much of the ads it is: as a
// whole, and in each kind of section. Ads whose HTML is the same as an
// earlier one's are counted once.
//
// It writes OUTDIR/repeated.tsv: every repeated text, with the groups of
// similar titles and the ads it's in, its characters over all of them, the
// kinds of section it's in, and whether it's boilerplate or a template for
// the job, by where most of it is; most characters first. And it prints a summary.
//
// From the lab's folder:
//
//	go run ./cmd/boilerplate data/ingestion/ads.jsonl data/boilerplate
//
// or, once installed with `go install ./cmd/...`, just boilerplate.
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
	"github.com/nukopian/jdc-lab/internal/boilerplate"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// repeat is what's known of one repeated text.
type repeat struct {
	text        string
	ads         int
	chars       int
	groups      int
	kinds       map[sections.Kind]bool
	boilerplate bool
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: boilerplate ADS.jsonl OUTDIR")
	}
	source, dir := os.Args[1], os.Args[2]

	all, ix, dups, err := read(source)
	if err != nil {
		log.Fatal(err)
	}
	t := measure(all, &ix)
	if err := t.write(filepath.Join(dir, "repeated.tsv")); err != nil {
		log.Fatal(err)
	}
	t.print(len(all), dups, dir)
}

// read divides every ad into sections, and indexes their text. It keeps
// each ad's sections for the second pass.
func read(source string) (all [][]sections.Section, ix boilerplate.Index, dups int, err error) {
	var d ads.Distinct
	for a, err := range ads.File(source) {
		if err != nil {
			return nil, ix, 0, err
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
			ix.Add(a.Title, s.Kind, texts...)
		}
		all = append(all, ss)
	}
	return all, ix, d.Duplicates, nil
}

// tally is what's measured.
type tally struct {
	chars, repeated, boiler, template int
	kindChars, kindBoiler             map[sections.Kind]int
	mostly                            int // ads that are more than half boilerplate
	repeats                           map[string]*repeat
}

func measure(all [][]sections.Section, ix *boilerplate.Index) tally {
	t := tally{kindChars: map[sections.Kind]int{}, kindBoiler: map[sections.Kind]int{}, repeats: map[string]*repeat{}}
	for _, a := range all {
		chars, boiler := 0, 0
		seen := map[string]bool{}
		for _, s := range a {
			for _, p := range s.Parts {
				if p.Role != headings.Text {
					continue
				}
				n := len(p.Text)
				chars += n
				t.kindChars[s.Kind] += n
				if !ix.Repeated(p.Text) {
					continue
				}
				t.repeated += n
				if ix.Boilerplate(p.Text) {
					boiler += n
					t.kindBoiler[s.Kind] += n
				} else {
					t.template += n
				}
				k := headings.Key(p.Text)
				r := t.repeats[k]
				if r == nil {
					r = &repeat{text: p.Text, groups: ix.Groups(p.Text), kinds: map[sections.Kind]bool{},
						boilerplate: ix.Boilerplate(p.Text)}
					t.repeats[k] = r
				}
				if !seen[k] {
					r.ads++
					seen[k] = true
				}
				r.chars += n
				r.kinds[s.Kind] = true
			}
		}
		t.chars += chars
		t.boiler += boiler
		if 2*boiler > chars {
			t.mostly++
		}
	}
	return t
}

func (t tally) write(path string) (err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	rs := make([]*repeat, 0, len(t.repeats))
	for _, r := range t.repeats {
		rs = append(rs, r)
	}
	slices.SortFunc(rs, func(a, b *repeat) int {
		return cmp.Or(cmp.Compare(b.chars, a.chars), cmp.Compare(a.text, b.text))
	})
	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "groups\tads\tchars\tkinds\tverdict\ttext")
	for _, r := range rs {
		var kinds []string
		for _, k := range sections.Kinds {
			if r.kinds[k] {
				kinds = append(kinds, string(k))
			}
		}
		verdict := "template"
		if r.boilerplate {
			verdict = "boilerplate"
		}
		fmt.Fprintf(w, "%d\t%d\t%d\t%s\t%s\t%s\n", r.groups, r.ads, r.chars, strings.Join(kinds, ","), verdict, r.text)
	}
	return w.Flush()
}

func (t tally) print(ads, dups int, dir string) {
	fmt.Printf("%d ads (%d duplicates left out); repeated texts written to %s\n\n", ads, dups, dir)
	fmt.Printf("text repeated across jobs:   %4.1f%%\n", pct(t.repeated, t.chars))
	fmt.Printf("  boilerplate:               %4.1f%%\n", pct(t.boiler, t.chars))
	fmt.Printf("  template for the job:      %4.1f%%\n", pct(t.template, t.chars))
	fmt.Printf("ads over half boilerplate:   %d (%.0f%%)\n\n", t.mostly, pct(t.mostly, ads))
	fmt.Printf("%-17s %8s %12s\n", "section", "of text", "boilerplate")
	for _, k := range sections.Kinds {
		fmt.Printf("%-17s %7.0f%% %11.0f%%\n", k, pct(t.kindChars[k], t.chars), pct(t.kindBoiler[k], t.kindChars[k]))
	}
}

func pct(n, of int) float64 {
	if of == 0 {
		return 0
	}
	return 100 * float64(n) / float64(of)
}
