package boilerplate

import (
	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"github.com/nukopian/jdc-lab/internal/walls"
)

// A Divided is an ad, with its sections.
type Divided struct {
	Ad       ads.Ad
	Sections []sections.Section
}

// Read reads the distinct ads in a file of them, divides each into
// sections, and indexes the sections' text, and where in each ad it
// stands, ready to ask what's boilerplate. dups is how many repeats of an
// ad it skipped.
func Read(path string) (all []Divided, ix Index, dups int, err error) {
	var d ads.Distinct
	for a, err := range ads.File(path) {
		if err != nil {
			return nil, ix, 0, err
		}
		if !d.First(a) {
			continue
		}
		bs := blocks.Split(a.HTML)
		ss := sections.Divide(headings.Read(walls.Split(bs), a.Title, a.Locations()...))
		src := SourceOf(a)
		for _, s := range ss {
			ix.Add(src, s.Kind, Texts([]sections.Section{s})...)
		}
		addWalls(&ix, src, bs, ss)
		all = append(all, Divided{a, ss})
	}
	// Where a text stands is known only once every ad is in.
	for _, a := range all {
		ix.Ends(SourceOf(a.Ad), Texts(a.Sections)...)
	}
	return all, ix, d.Duplicates, nil
}

// addWalls records an ad's walls whole, with the lines each was split
// into, each in the section its first line is in (see Index.Wall).
func addWalls(ix *Index, src Source, bs []blocks.Block, ss []sections.Section) {
	kindOf := map[string]sections.Kind{}
	for _, s := range ss {
		for _, p := range s.Parts {
			if _, ok := kindOf[p.Text]; !ok {
				kindOf[p.Text] = s.Kind
			}
		}
	}
	for _, b := range bs {
		if !walls.Is(b) {
			continue
		}
		var lines []string
		kind, found := sections.Summary, false
		for _, l := range walls.Split([]blocks.Block{b}) {
			lines = append(lines, l.Text)
			if k, ok := kindOf[l.Text]; ok && !found {
				kind, found = k, true
			}
		}
		ix.Wall(src, kind, b.Text, lines...)
	}
}

// Texts returns the text parts of sections, in order.
func Texts(ss []sections.Section) []string {
	var texts []string
	for _, s := range ss {
		for _, p := range s.Parts {
			if p.Role == headings.Text {
				texts = append(texts, p.Text)
			}
		}
	}
	return texts
}
