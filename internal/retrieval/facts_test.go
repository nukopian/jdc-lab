package retrieval

import (
	"slices"
	"testing"

	"github.com/nukopian/jdc-lab/internal/chunk"
	"github.com/nukopian/jdc-lab/internal/facts"
)

func TestListingOf(t *testing.T) {
	pattern := func(v string) facts.Fact { return facts.Fact{Kind: "pattern", Values: map[string]string{"value": v}} }
	a := chunk.Ad{Chunks: []chunk.Chunk{
		{Facts: []facts.Fact{pattern("part time"), {Kind: "pay", Values: map[string]string{"min": "12"}}}},
		{Facts: []facts.Fact{pattern("permanent"), pattern("part time")}},
	}}
	got := ListingOf(Texts{ID: "7", Whole: "w", Own: "o"}, "Chef", "Leeds", a)
	want := Listing{ID: "7", Title: "Chef", Whole: "w", Own: "o", Place: "Leeds", Patterns: []string{"part time", "permanent"}}
	if got.ID != want.ID || got.Title != want.Title || got.Whole != want.Whole || got.Own != want.Own ||
		got.Place != want.Place || !slices.Equal(got.Patterns, want.Patterns) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
	// No facts: an empty list, not null, for Python.
	if got := ListingOf(Texts{ID: "8"}, "Chef", "", chunk.Ad{}); got.Patterns == nil {
		t.Error("Patterns is nil")
	}
}

// listings makes listings from (title, place, patterns…) rows, IDs in
// order from "01".
func listings(rows ...[]string) []Listing {
	var ls []Listing
	for i, r := range rows {
		id := string(rune('0'+(i+1)/10)) + string(rune('0'+(i+1)%10))
		ls = append(ls, Listing{ID: id, Title: r[0], Place: r[1], Patterns: r[2:]})
	}
	return ls
}

func TestPlaceQueries(t *testing.T) {
	ls := listings(
		[]string{"Chef", "Leeds"}, []string{"Chef", "Leeds"}, []string{"Head Chef", "Leeds"},
		[]string{"Chef", "York"}, []string{"Chef", "York"},
		[]string{"Support Worker", "Leeds"}, []string{"Support Worker", "Leeds"},
		[]string{"Crane Operator", "Hull"},
	)
	qs := PlaceQueries(ls, 10, 2, 1)
	// Chef in Leeds has most (3, Head Chef among them); Chef in York is
	// the second for Chef, over the most of one a title; Support Worker
	// in Leeds the second for Leeds; Crane Operator in Hull too few.
	if len(qs) != 1 {
		t.Fatalf("got %d queries, want 1: %+v", len(qs), qs)
	}
	q := qs[0]
	if q.Query != "Chef in Leeds" || q.Kind != PlaceFact || q.Value != "Leeds" {
		t.Errorf("got %+v", q)
	}
	if !slices.Equal(q.Relevant, []string{"01", "02", "03"}) || !slices.Equal(q.Candidates, []string{"01", "02", "03", "04", "05"}) {
		t.Errorf("relevant %v, candidates %v", q.Relevant, q.Candidates)
	}
	if qs := PlaceQueries(ls, 10, 2, 2); len(qs) != 3 {
		t.Errorf("with two a title or place, got %d queries, want 3: %+v", len(qs), qs)
	}
}

func TestPatternQueries(t *testing.T) {
	ls := listings(
		[]string{"Chef", "", "part time"}, []string{"Chef", "", "part time", "permanent"}, []string{"Chef", ""},
		[]string{"Support Worker", "", "part time"}, []string{"Support Worker", ""},
		[]string{"Crane Operator", "", "contract"}, []string{"Crane Operator", "", "contract"},
	)
	qs := PatternQueries(ls, []string{"part time", "contract", "remote"}, 2, 2)
	var got []string
	for _, q := range qs {
		got = append(got, q.Query)
		if len(q.Relevant) != 0 || q.Kind != PatternFact {
			t.Errorf("%s: kind %s, relevant %v", q.Query, q.Kind, q.Relevant)
		}
	}
	// Support Worker has one part time ad, too few; nothing is remote.
	if want := []string{"part time Chef", "contract Crane Operator"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if !slices.Equal(qs[0].Candidates, []string{"01", "02", "03"}) {
		t.Errorf("candidates %v", qs[0].Candidates)
	}
}
