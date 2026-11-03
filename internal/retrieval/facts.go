package retrieval

import (
	"cmp"
	"maps"
	"slices"

	"github.com/nukopian/jdc-lab/internal/chunk"
	"github.com/nukopian/jdc-lab/internal/headings"
)

// A Listing is an ad as the test of facts ranks it: its title, the texts
// it embeds, its place as the export gives it, and the work patterns its
// facts found.
type Listing struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Whole    string   `json:"whole"`
	Own      string   `json:"own"`
	Place    string   `json:"place"`
	Patterns []string `json:"patterns"`
}

// ListingOf returns an ad's listing, from its texts (see TextsOf), its
// place, and its chunks, whose details hold its facts.
func ListingOf(t Texts, title, place string, a chunk.Ad) Listing {
	l := Listing{ID: t.ID, Title: title, Whole: t.Whole, Own: t.Own, Place: place, Patterns: []string{}}
	for _, c := range a.Chunks {
		for _, f := range c.Facts {
			if v := f.Values["value"]; f.Kind == "pattern" && !slices.Contains(l.Patterns, v) {
				l.Patterns = append(l.Patterns, v)
			}
		}
	}
	slices.Sort(l.Patterns)
	return l
}

// A FactQuery is a job title with a fact: a place ("Care Assistant in
// Leeds") or a work pattern ("part time Receptionist"). Candidates are
// the ads whose title is similar to Title, in order of ID; Relevant are
// those of them for the fact, where the test knows: by the export's
// place for a place. A work pattern's are judged, so it has none.
type FactQuery struct {
	Query      string   `json:"query"`
	Title      string   `json:"title"`
	Kind       string   `json:"kind"`
	Value      string   `json:"value"`
	Candidates []string `json:"candidates"`
	Relevant   []string `json:"relevant"`
}

// The kinds of fact a query can name.
const (
	PlaceFact   = "place"
	PatternFact = "pattern"
)

// PlaceQueries returns up to n queries of a title and a place, from
// listings: the pairs of a group of similar titles, among the most
// common groups, and a place, with the most ads whose title is similar to
// the group's commonest and whose place is the place, at least least of
// them; no title or place in more than most queries.
func PlaceQueries(ls []Listing, n, least, most int) []FactQuery {
	titles, ids := titlesOf(ls)
	place := map[string]string{}
	for _, l := range ls {
		place[l.ID] = l.Place
	}
	type pair struct {
		title, place string
		candidates   []string
		relevant     []string
	}
	var pairs []pair
	for _, title := range commonTitles(titles, 4*n) {
		cands := similar(title, ids, titles)
		byPlace := map[string][]string{}
		for _, id := range cands {
			if p := place[id]; p != "" {
				byPlace[p] = append(byPlace[p], id)
			}
		}
		for p, rel := range byPlace {
			if len(rel) >= least {
				pairs = append(pairs, pair{title, p, cands, rel})
			}
		}
	}
	slices.SortFunc(pairs, func(a, b pair) int {
		return cmp.Or(cmp.Compare(len(b.relevant), len(a.relevant)), cmp.Compare(a.title, b.title), cmp.Compare(a.place, b.place))
	})
	var qs []FactQuery
	perTitle, perPlace := map[string]int{}, map[string]int{}
	for _, p := range pairs {
		if len(qs) == n {
			break
		}
		if perTitle[p.title] == most || perPlace[p.place] == most {
			continue
		}
		perTitle[p.title]++
		perPlace[p.place]++
		qs = append(qs, FactQuery{Query: p.title + " in " + p.place, Title: p.title, Kind: PlaceFact, Value: p.place,
			Candidates: p.candidates, Relevant: p.relevant})
	}
	return qs
}

// PatternQueries returns, for each of patterns, up to each queries of a
// title and that work pattern ("part time Receptionist"): the most common
// groups of similar titles with at least least ads whose facts found the
// pattern, so that there's something to find. Which ads are for the
// pattern is judged, so the queries have no relevant ads.
func PatternQueries(ls []Listing, patterns []string, each, least int) []FactQuery {
	titles, ids := titlesOf(ls)
	has := map[string][]string{}
	for _, l := range ls {
		has[l.ID] = l.Patterns
	}
	common := commonTitles(titles, 40)
	var qs []FactQuery
	for _, pattern := range patterns {
		n := 0
		for _, title := range common {
			if n == each {
				break
			}
			cands := similar(title, ids, titles)
			found := 0
			for _, id := range cands {
				if slices.Contains(has[id], pattern) {
					found++
				}
			}
			if found < least {
				continue
			}
			qs = append(qs, FactQuery{Query: pattern + " " + title, Title: title, Kind: PatternFact, Value: pattern,
				Candidates: cands, Relevant: []string{}})
			n++
		}
	}
	return qs
}

// titlesOf returns each listing's title by its ID, and the IDs in order.
func titlesOf(ls []Listing) (map[string]string, []string) {
	titles := map[string]string{}
	for _, l := range ls {
		titles[l.ID] = l.Title
	}
	return titles, slices.Sorted(maps.Keys(titles))
}

// commonTitles returns the commonest spelling of each of the n groups of
// similar titles with the most ads, each once.
func commonTitles(titles map[string]string, n int) []string {
	var out []string
	used := map[string]bool{}
	for _, g := range titleGroups(titles) {
		if len(out) == n {
			break
		}
		t := g.commonest()
		if !used[headings.Key(t)] {
			used[headings.Key(t)] = true
			out = append(out, t)
		}
	}
	return out
}
