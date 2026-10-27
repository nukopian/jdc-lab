// Command view writes a cross-section of a sample's ads in their chunked
// form, for reading by eye (see internal/view): about 20 ads, the same in
// every run, each with its sections, fields, repeats and boilerplate
// marked, and the text no heading placed. Ads whose HTML is the same as
// an earlier one's are left out.
//
// It writes OUTDIR/ads.md, and prints a summary: how the ads were chosen,
// and how many of the sample's ads have text in no section, or out of the
// ad's order, which should be none.
//
// From the lab's folder:
//
//	go run ./cmd/view data/ingestion/ads-20261024.jsonl data/view
//
// or, once installed with `go install ./cmd/...`, just view.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/boilerplate"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"github.com/nukopian/jdc-lab/internal/view"
)

// shown is how many ads the view has.
const shown = 20

// an ad, with its sections.
type chunked struct {
	ad ads.Ad
	ss []sections.Section
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: view ADS.jsonl OUTDIR")
	}
	source, dir := os.Args[1], os.Args[2]
	all, ix, dups, err := read(source)
	if err != nil {
		log.Fatal(err)
	}

	verdict := func(text string) view.Verdict {
		switch {
		case !ix.Repeated(text):
			return view.Own
		case ix.Boilerplate(text):
			return view.Boilerplate
		}
		return view.Template
	}
	cs := make([]view.Candidate, len(all))
	lost := make([]view.Lost, len(all))
	missing, moved := 0, 0
	for i, c := range all {
		cs[i] = candidate(c, verdict)
		switch lost[i] = view.Missing(c.ad.HTML, c.ss); {
		case lost[i].Moved:
			moved++
		case lost[i].Context != "":
			missing++
		}
		cs[i].Lost = lost[i].Context != ""
	}
	picked := view.Pick(cs, shown)

	path := filepath.Join(dir, "ads.md")
	if err := write(path, all, picked, verdict, lost); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d ads (%d duplicates left out); %d shown in %s\n\n", len(all), dups, len(picked), path)
	for _, why := range []struct {
		what string
		is   func(view.Candidate) bool
	}{
		{"no heading", func(c view.Candidate) bool { return c.NoHeading }},
		{"an unmapped section", func(c view.Candidate) bool { return c.Unmapped }},
		{"over half boilerplate", func(c view.Candidate) bool { return c.MostlyBoilerplate }},
		{"text lost or moved", func(c view.Candidate) bool { return c.Lost }},
	} {
		n := 0
		for _, i := range picked {
			if why.is(cs[i]) {
				n++
			}
		}
		fmt.Printf("shown with %-22s %d\n", why.what+":", n)
	}
	fmt.Printf("\nads with text in no section: %d\n", missing)
	fmt.Printf("ads with text out of order:  %d\n", moved)
}

// read reads the sample's distinct ads, divided into sections, and
// indexes their texts for boilerplate, as cmd/boilerplate does.
func read(source string) (all []chunked, ix boilerplate.Index, dups int, err error) {
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
		all = append(all, chunked{a, ss})
	}
	return all, ix, d.Duplicates, nil
}

// candidate says what makes an ad worth showing.
func candidate(c chunked, verdict func(string) view.Verdict) view.Candidate {
	can := view.Candidate{ID: c.ad.ID, NoHeading: true}
	chars, boiler := 0, 0
	for _, s := range c.ss {
		if s.Heading != "" {
			can.NoHeading = false
		}
		if s.Kind == sections.Unmapped {
			can.Unmapped = true
		}
		for _, p := range s.Parts {
			if p.Role != headings.Text {
				continue
			}
			chars += len(p.Text)
			if verdict(p.Text) == view.Boilerplate {
				boiler += len(p.Text)
			}
		}
	}
	can.MostlyBoilerplate = boiler*2 > chars
	return can
}

// write writes the picked ads to path, in the order picked.
func write(path string, all []chunked, picked []int, verdict func(string) view.Verdict, lost []view.Lost) (err error) {
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
	w := bufio.NewWriter(f)
	fmt.Fprintf(w, "# %d ads in their sections\n\n", len(picked))
	for _, i := range picked {
		if err := view.Render(w, all[i].ad, all[i].ss, verdict, lost[i]); err != nil {
			return err
		}
	}
	return w.Flush()
}
