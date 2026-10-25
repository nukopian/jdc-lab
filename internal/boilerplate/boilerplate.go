// Package boilerplate finds the text job ads repeat whatever the job: an
// agency's description of itself, its standard benefits, its legal notes.
//
// The ads say nothing of who placed them, so boilerplate is found by
// repetition: a block of text, word for word (by its key, see
// headings.Key), in ads for different jobs. "Different jobs" is the point.
// An advertiser posting one job in ten towns repeats every block of it,
// and none of that is boilerplate. So what counts is how many groups of
// similar titles the text appears under, not how many ads: "Support
// Worker" and "Residential Support Worker" are one group, "Crane Operator"
// and "Forklift Driver" two.
//
// Repetition isn't enough on its own. Related jobs from one advertiser
// share lines that describe the job, such as "Excellent ICT skills", and
// those are worth keeping. So a repeated text found mostly in the sections
// that describe the job (responsibilities, requirements, desirable) is a
// template for the job, and only one found mostly elsewhere is
// boilerplate. The verdict is the text's, over all the ads it's in, so a
// heading filed under the wrong kind in one ad doesn't change it.
//
// Text under Widespread groups of titles or more is boilerplate wherever
// it's found. Footers, such as an agency's equal opportunities notice,
// often come after an ad's last heading, which is usually its
// requirements, and so are filed under it. And a line such as "Able to
// work as part of a team", though about the job, is in so many jobs that
// it says nothing of any of them: as with a common word in a search
// index, being everywhere is what makes it worth little.
package boilerplate

import (
	"slices"
	"strings"

	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// MinWords is the fewest words a repeated text has to count: shorter ones,
// such as "Pension", are common because they're short, not because
// they're boilerplate.
const MinWords = 8

// MinGroups is the fewest groups of similar titles a text must appear
// under to count as repeated.
const MinGroups = 2

// Widespread is the number of groups of similar titles a text must appear
// under to be boilerplate wherever it's found.
const Widespread = 5

// An Index knows, for each text, the titles of the ads it's in, and the
// kinds of section it's in there. Add every ad first, then ask about any
// text. Its zero value is ready to use.
type Index struct {
	titles map[string][]string // text key → distinct title keys
	where  map[string]*where   // text key → the kinds of section it's in
	groups map[string]int      // text key → groups of similar titles, once counted
}

// where counts a text's times in sections that describe the job, and in
// others.
type where struct{ describes, elsewhere int }

// Add records texts from a section of kind, in an ad whose title is title.
func (ix *Index) Add(title string, kind sections.Kind, texts ...string) {
	if ix.titles == nil {
		ix.titles, ix.where = map[string][]string{}, map[string]*where{}
	}
	t := headings.Key(title)
	for _, text := range texts {
		k := headings.Key(text)
		if len(strings.Fields(k)) < MinWords {
			continue
		}
		if !slices.Contains(ix.titles[k], t) {
			ix.titles[k] = append(ix.titles[k], t)
		}
		w := ix.where[k]
		if w == nil {
			w = &where{}
			ix.where[k] = w
		}
		if Describes(kind) {
			w.describes++
		} else {
			w.elsewhere++
		}
	}
	ix.groups = nil // counts made before are out of date
}

// Groups returns how many groups of similar titles text appears under: 0
// for a text shorter than MinWords, or never added.
func (ix *Index) Groups(text string) int {
	k := headings.Key(text)
	if n, ok := ix.groups[k]; ok {
		return n
	}
	n := groups(ix.titles[k])
	if ix.groups == nil {
		ix.groups = map[string]int{}
	}
	ix.groups[k] = n
	return n
}

// Repeated reports whether text appears under at least MinGroups groups
// of similar titles.
func (ix *Index) Repeated(text string) bool {
	return ix.Groups(text) >= MinGroups
}

// Boilerplate reports whether text is boilerplate: under Widespread groups
// of titles or more, or repeated and found more often outside the sections
// that describe the job than in them.
func (ix *Index) Boilerplate(text string) bool {
	if !ix.Repeated(text) {
		return false
	}
	if ix.Groups(text) >= Widespread {
		return true
	}
	w := ix.where[headings.Key(text)]
	return w.elsewhere > w.describes
}

// Template reports whether text is a template for the job: repeated, but
// not widespread, and found at least as often in the sections that
// describe the job as outside them. A tie goes this way, as keeping text
// about the job is the safer mistake.
func (ix *Index) Template(text string) bool {
	return ix.Repeated(text) && !ix.Boilerplate(text)
}

// Describes reports whether a section of kind describes the job itself,
// so that text repeated in it is a template for the job, not boilerplate.
func Describes(kind sections.Kind) bool {
	switch kind {
	case sections.Responsibilities, sections.Requirements, sections.Desirable:
		return true
	}
	return false
}

// groups counts the groups of similar titles among titles: each title
// joins the first group with a title similar to it.
func groups(titles []string) int {
	var gs [][]string
	for _, t := range titles {
		i := slices.IndexFunc(gs, func(g []string) bool {
			return slices.ContainsFunc(g, func(u string) bool { return similar(t, u) })
		})
		if i < 0 {
			gs = append(gs, []string{t})
		} else {
			gs[i] = append(gs[i], t)
		}
	}
	return len(gs)
}

// similar reports whether two title keys are titles of the same job: the
// words of one are all in the other, or they share at least half their
// words between them. A title with no words is like no other.
func similar(a, b string) bool {
	if a == b {
		return true
	}
	if a == "" || b == "" {
		return false
	}
	wa, wb := words(a), words(b)
	shared := 0
	for w := range wa {
		if wb[w] {
			shared++
		}
	}
	all := len(wa) + len(wb) - shared
	return shared == len(wa) || shared == len(wb) || 2*shared >= all
}

func words(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}
