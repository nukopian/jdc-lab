package chunk

import (
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

func text(s string) headings.Part { return headings.Part{Role: headings.Text, Text: s} }

func TestOf(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{text("We deliver parcels across the north of England every day.")}},
		{Kind: sections.Responsibilities, Heading: "What you'll be doing", Parts: []headings.Part{text("Drive the van."), text("We are an agency.")}},
		{Kind: sections.Details, Heading: "Job details", Parts: []headings.Part{
			{Role: headings.Field, Text: "Salary: £30,000", Key: "salary", Value: "£30,000"},
		}},
		{Kind: sections.Company, Heading: "About Brightway", Parts: []headings.Part{text("We are a family firm.")}},
		{Kind: sections.Apply, Heading: "How to apply", ByName: true, Parts: []headings.Part{text("Apply today.")}},
	}
	boiler := func(s string) bool { return s == "We are an agency." }
	cs := Of(ss, boiler)
	if len(cs) != 5 {
		t.Fatalf("got %d chunks, want 5", len(cs))
	}
	tests := []struct {
		use, placed, text string
	}{
		{Embed, "nothing before it: summary by default", "We deliver parcels across the north of England every day."},
		{Embed, `the heading "What you'll be doing": "what you'll be doing" in headings.txt`, "What you'll be doing\nDrive the van."},
		{Metadata, `the heading "Job details": "job details" in headings.txt`, "Job details\nSalary: £30,000"},
		{LeaveOut, `the heading "About Brightway": it starts "about"`, "About Brightway\nWe are a family firm."},
		{LeaveOut, `the line "How to apply": exactly a section's name`, "How to apply\nApply today."},
	}
	for i, tt := range tests {
		c := cs[i]
		if c.Use != tt.use || c.PlacedBy != tt.placed || c.Text != tt.text {
			t.Errorf("chunk %d: got %s, %q, %q; want %s, %q, %q", i, c.Use, c.PlacedBy, c.Text, tt.use, tt.placed, tt.text)
		}
	}
	if got := cs[1].Boilerplate; len(got) != 1 || got[0] != "We are an agency." {
		t.Errorf("the responsibilities' boilerplate: %q", got)
	}
	if got := cs[2].Fields; len(got) != 1 || got[0].Label != "salary" || got[0].Value != "£30,000" {
		t.Errorf("the details' fields: %+v", got)
	}
	// Tokens: about four for every three words.
	if got := cs[0].Tokens; got != 13 {
		t.Errorf("tokens: got %d, want 13 (10 words)", got)
	}
}

func TestOfUnmapped(t *testing.T) {
	ss := []sections.Section{{Kind: sections.Unmapped, Heading: "Our depots", Parts: []headings.Part{text("Leeds and York.")}}}
	c := Of(ss, func(string) bool { return false })[0]
	if c.Use != Embed || !strings.Contains(c.PlacedBy, "no phrase in headings.txt") {
		t.Errorf("got %s, %q; want an unmapped section embedded, saying no phrase placed it", c.Use, c.PlacedBy)
	}
}
