// Package retrieval builds a test of retrieval from a sample's ads: the
// texts each representation of an ad would embed, and queries with the
// ads relevant to them. Python embeds and scores them (python/bin); this
// package decides what's in them, as the stages that know an ad's
// sections and boilerplate are in Go.
//
// The queries are job titles, and an ad is relevant to one if its title
// is similar to it (boilerplate.SimilarTitles). The title field is left
// out of every representation, so a query asks whether a representation
// of the description finds ads for the same job.
package retrieval

import (
	"cmp"
	"hash/fnv"
	"maps"
	"slices"
	"strings"

	"github.com/nukopian/jdc-lab/internal/boilerplate"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// Texts are what each representation of an ad embeds, as one text, or
// one for each section.
type Texts struct {
	ID string `json:"id"`
	// Whole is all the ad's text, headings with it, as a description is
	// embedded whole.
	Whole string `json:"whole"`
	// Own is Whole without the text found as boilerplate.
	Own string `json:"own"`
	// Core is the sections that describe the job: responsibilities,
	// requirements and desirable.
	Core string `json:"core"`
	// Summary is the summary sections.
	Summary string `json:"summary"`
	// Sections are each section's own text, without boilerplate, for a
	// vector each; a section with none of its own is left out.
	Sections []string `json:"sections"`
}

// TextsOf returns the texts of an ad's sections; boilerplate says
// whether a text part was found to be boilerplate.
func TextsOf(id string, ss []sections.Section, boilerplate func(string) bool) Texts {
	var whole, own, core, summary []string
	t := Texts{ID: id}
	for _, s := range ss {
		var all, mine []string
		if s.Heading != "" {
			all, mine = append(all, s.Heading), append(mine, s.Heading)
		}
		kept := 0
		for _, p := range s.Parts {
			all = append(all, p.Text)
			if p.Role == headings.Text && boilerplate(p.Text) {
				continue
			}
			mine = append(mine, p.Text)
			kept++
		}
		whole, own = append(whole, all...), append(own, mine...)
		switch s.Kind {
		case sections.Responsibilities, sections.Requirements, sections.Desirable:
			core = append(core, mine...)
		case sections.Summary, sections.Role: // as the summary was, before role had a kind of its own
			summary = append(summary, mine...)
		}
		if kept > 0 {
			t.Sections = append(t.Sections, strings.Join(mine, "\n"))
		}
	}
	t.Whole, t.Own = strings.Join(whole, "\n"), strings.Join(own, "\n")
	t.Core, t.Summary = strings.Join(core, "\n"), strings.Join(summary, "\n")
	return t
}

// Sample returns n of ids, the first by a fixed hash of them, in that
// order, so that every run of the test has the same ads.
func Sample(ids []string, n int) []string {
	out := slices.Clone(ids)
	slices.SortFunc(out, func(a, b string) int { return cmp.Or(cmp.Compare(hash(a), hash(b)), cmp.Compare(a, b)) })
	return out[:min(n, len(out))]
}

func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// A Query is a job title, with the IDs of the ads relevant to it.
type Query struct {
	Query    string   `json:"query"`
	Relevant []string `json:"relevant"`
}

// Queries returns up to n queries from titles, an ad's title by its ID:
// the groups of similar titles with the most ads, each written as its
// group's commonest title, with every ad whose title is similar to that
// title relevant, in order of ID. A query needs at least least relevant
// ads, and comes once, though two groups can share a commonest title, as
// grouping goes in order.
func Queries(titles map[string]string, n, least int) []Query {
	ids := slices.Sorted(maps.Keys(titles))
	var qs []Query
	used := map[string]bool{}
	for _, g := range titleGroups(titles) {
		if len(qs) == n || g.ads < least {
			break
		}
		q := g.commonest()
		if used[headings.Key(q)] {
			continue
		}
		used[headings.Key(q)] = true
		if rel := similar(q, ids, titles); len(rel) >= least {
			qs = append(qs, Query{Query: q, Relevant: rel})
		}
	}
	return qs
}

// A titleGroup is a group of similar titles: each spelling, with its
// ads, and its ads in all.
type titleGroup struct {
	spellings map[string]int
	ads       int
}

// titleGroups returns titles, an ad's title by its ID, in groups of
// similar titles, the most ads first. Each title joins the first group
// with a title similar to it, in order of ID, as boilerplate's groups do.
func titleGroups(titles map[string]string) []*titleGroup {
	var gs []*titleGroup
	for _, id := range slices.Sorted(maps.Keys(titles)) {
		t := titles[id]
		i := slices.IndexFunc(gs, func(g *titleGroup) bool {
			for s := range g.spellings {
				if boilerplate.SimilarTitles(s, t) {
					return true
				}
			}
			return false
		})
		if i < 0 {
			gs = append(gs, &titleGroup{spellings: map[string]int{}})
			i = len(gs) - 1
		}
		gs[i].spellings[t]++
		gs[i].ads++
	}
	slices.SortStableFunc(gs, func(a, b *titleGroup) int { return b.ads - a.ads })
	return gs
}

// commonest returns a group's commonest spelling, the first in sorting
// order of those as common.
func (g *titleGroup) commonest() string {
	spellings := slices.Sorted(maps.Keys(g.spellings))
	return slices.MaxFunc(spellings, func(a, b string) int {
		return cmp.Or(cmp.Compare(g.spellings[a], g.spellings[b]), cmp.Compare(b, a))
	})
}

// similar returns those of ids whose title is similar to q, in order.
func similar(q string, ids []string, titles map[string]string) []string {
	var out []string
	for _, id := range ids {
		if boilerplate.SimilarTitles(q, titles[id]) {
			out = append(out, id)
		}
	}
	return out
}
