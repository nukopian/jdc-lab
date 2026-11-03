package boilerplate

import (
	"runtime"
	"sync"
	"sync/atomic"

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
	var distinct []ads.Ad
	for a, err := range ads.File(path) {
		if err != nil {
			return nil, ix, 0, err
		}
		if d.First(a) {
			distinct = append(distinct, a)
		}
	}
	// Dividing an ad needs nothing from the others, so it's done on every
	// processor; the index takes them in order, as the order decides
	// what groups of texts alike are named.
	parts := divideAll(distinct)
	all = make([]Divided, len(distinct))
	for i, a := range distinct {
		src := SourceOf(a)
		for _, s := range parts[i].sections {
			ix.Add(src, s.Kind, Texts([]sections.Section{s})...)
		}
		for _, w := range parts[i].walls {
			ix.Wall(src, w.kind, w.text, w.lines...)
		}
		all[i] = Divided{a, parts[i].sections}
	}
	// Where a text stands is known only once every ad is in.
	for _, a := range all {
		ix.Ends(SourceOf(a.Ad), Texts(a.Sections)...)
	}
	return all, ix, d.Duplicates, nil
}

// parallelRead turns on dividing ads into sections on every processor,
// rather than one. TestParallelReadChangesNothing compares the two.
var parallelRead = true

// divided is an ad's sections, and its walls.
type divided struct {
	sections []sections.Section
	walls    []wall
}

// wall is a wall of an ad, whole, with the lines it was split into, in
// the section its first line is in (see Index.Wall).
type wall struct {
	kind  sections.Kind
	text  string
	lines []string
}

// divideAll divides each ad into sections, and finds its walls, on every
// processor if parallelRead is on, and returns them in the ads' order.
func divideAll(as []ads.Ad) []divided {
	out := make([]divided, len(as))
	if !parallelRead {
		for i, a := range as {
			out[i] = divide(a)
		}
		return out
	}
	// Each call writes only its own slot of out, so they need no lock.
	inParallel(len(as), func() func(int) {
		return func(i int) { out[i] = divide(as[i]) }
	})
	return out
}

// inParallel calls do(i) for each i from 0 to n-1, on every processor,
// and returns once all have returned, their writes visible to the caller.
// Each worker calls work once for its do, which can keep state of its
// own. Calls that write to shared memory must each write their own part.
func inParallel(n int, work func() func(i int)) {
	var next atomic.Int64 // the next i not yet taken
	var wg sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), n) {
		wg.Go(func() {
			do := work()
			for {
				i := int(next.Add(1)) - 1
				if i >= n {
					return
				}
				do(i)
			}
		})
	}
	wg.Wait()
}

// divide divides an ad into sections, and finds its walls.
func divide(a ads.Ad) divided {
	bs := blocks.Split(a.HTML)
	ss := sections.Divide(headings.Read(walls.Split(bs), a.Title, a.Locations()...))
	return divided{ss, wallsOf(bs, ss)}
}

// wallsOf returns an ad's walls whole, with the lines each was split
// into, each in the section its first line is in.
func wallsOf(bs []blocks.Block, ss []sections.Section) []wall {
	kindOf := map[string]sections.Kind{}
	for _, s := range ss {
		for _, p := range s.Parts {
			if _, ok := kindOf[p.Text]; !ok {
				kindOf[p.Text] = s.Kind
			}
		}
	}
	var ws []wall
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
		ws = append(ws, wall{kind, b.Text, lines})
	}
	return ws
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
