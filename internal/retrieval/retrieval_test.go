package retrieval

import (
	"slices"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

func text(s string) headings.Part { return headings.Part{Role: headings.Text, Text: s} }

func TestTextsOf(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{{Role: headings.Repeat, Text: "Driver"}, text("We deliver.")}},
		{Kind: sections.Responsibilities, Heading: "Duties", Parts: []headings.Part{text("Drive the van.")}},
		{Kind: sections.Requirements, Heading: "You'll need", Parts: []headings.Part{text("A licence.")}},
		{Kind: sections.Company, Heading: "About us", Parts: []headings.Part{text("We are an agency.")}},
	}
	boiler := func(s string) bool { return s == "We are an agency." }
	got := TextsOf("7", ss, boiler)
	want := Texts{
		ID:       "7",
		Whole:    "Driver\nWe deliver.\nDuties\nDrive the van.\nYou'll need\nA licence.\nAbout us\nWe are an agency.",
		Own:      "Driver\nWe deliver.\nDuties\nDrive the van.\nYou'll need\nA licence.\nAbout us",
		Core:     "Duties\nDrive the van.\nYou'll need\nA licence.",
		Summary:  "Driver\nWe deliver.",
		Sections: []string{"Driver\nWe deliver.", "Duties\nDrive the van.", "You'll need\nA licence."},
	}
	if got.Whole != want.Whole || got.Own != want.Own || got.Core != want.Core || got.Summary != want.Summary ||
		!slices.Equal(got.Sections, want.Sections) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

func TestQueries(t *testing.T) {
	titles := map[string]string{
		"1": "Support Worker", "2": "Support Worker", "3": "Residential Support Worker",
		"4": "Chef", "5": "Chef", "6": "Head Chef",
		"7": "Crane Operator",
	}
	qs := Queries(titles, 5, 2)
	if len(qs) != 2 {
		t.Fatalf("got %d queries, want 2 (the crane operator has too few ads): %+v", len(qs), qs)
	}
	// The commonest groups first, each as its commonest spelling, with
	// every ad whose title is similar to it, in order of ID.
	if qs[0].Query != "Chef" && qs[0].Query != "Support Worker" {
		t.Errorf("first query %q", qs[0].Query)
	}
	for _, q := range qs {
		want := map[string][]string{"Support Worker": {"1", "2", "3"}, "Chef": {"4", "5", "6"}}[q.Query]
		if !slices.Equal(q.Relevant, want) {
			t.Errorf("%s: relevant %v, want %v", q.Query, q.Relevant, want)
		}
	}
	if strings.Contains(qs[0].Query+qs[1].Query, "Head") {
		t.Errorf("a less common spelling was chosen: %+v", qs)
	}
}

func TestSample(t *testing.T) {
	ids := []string{"a", "b", "c", "d", "e"}
	got := Sample(ids, 3)
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
	slices.Reverse(ids)
	if again := Sample(ids, 3); !slices.Equal(got, again) {
		t.Errorf("not the same however the IDs come: %v, %v", got, again)
	}
}

// A query comes once, and needs its least relevant ads by its own title,
// not only its group's size.
func TestQueriesOnceEachWithEnoughRelevant(t *testing.T) {
	// One group of seven, joined through "Chef Manager Assistant Cook":
	// but neither common spelling has five ads similar to it.
	titles := map[string]string{
		"1": "Chef", "2": "Chef", "3": "Chef", "4": "Chef Manager Assistant Cook",
		"5": "Manager Assistant Cook Baker", "6": "Manager Assistant Cook Baker", "7": "Manager Assistant Cook Baker",
	}
	if qs := Queries(titles, 5, 5); len(qs) != 0 {
		t.Errorf("got %+v; want none, as no title has five ads similar to it", qs)
	}
	// Two groups can share their commonest title, as grouping goes in
	// order: the first "Administrator" starts a group of its own, then
	// "Payroll Manager"'s group takes in a title like it, and the later
	// ones join that. The query comes once.
	twice := map[string]string{
		"1": "Payroll Manager", "2": "Administrator", "3": "Payroll Administrator Manager",
		"4": "Administrator", "5": "Administrator",
	}
	n := 0
	for _, q := range Queries(twice, 5, 1) {
		if q.Query == "Administrator" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("Administrator is a query %d times, want once", n)
	}
}
