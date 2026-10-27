package chunk

import (
	"reflect"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"go.yaml.in/yaml/v3"
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
	if len(cs) != 6 {
		t.Fatalf("got %d chunks, want 5 and one of boilerplate", len(cs))
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
	// Boilerplate leaves an embedded chunk for one of its own, at the end,
	// each line with the section it came from.
	b := cs[5]
	if b.Kind != Boilerplate || b.Use != LeaveOut || b.Text != "responsibilities: We are an agency." {
		t.Errorf("the boilerplate chunk: %+v", b)
	}
	if got := cs[2].Fields; len(got) != 1 || got[0].Label != "salary" || got[0].Value != "£30,000" {
		t.Errorf("the details' fields: %+v", got)
	}
	// Tokens: about four for every three words.
	if got := cs[0].Tokens; got != 13 {
		t.Errorf("tokens: got %d, want 13 (10 words)", got)
	}
}

// An unmapped section goes to the boilerplate chunk, each line with its
// heading, and counts as unclassified; its fields stay, as details.
func TestOfUnmapped(t *testing.T) {
	ss := []sections.Section{{Kind: sections.Unmapped, Heading: "Our depots", Parts: []headings.Part{
		text("Leeds and York."),
		{Role: headings.Field, Text: "Pay: £16.50 per hour", Key: "pay", Value: "£16.50 per hour", Added: "Pay: "},
	}}}
	cs := Of(ss, func(string) bool { return false })
	if len(cs) != 2 {
		t.Fatalf("got %d chunks, want details and boilerplate: %+v", len(cs), cs)
	}
	d, b := cs[0], cs[1]
	if d.Kind != sections.Details || d.Use != Metadata || len(d.Fields) != 1 || d.Fields[0].Value != "£16.50 per hour" ||
		!strings.Contains(d.PlacedBy, "unmapped") {
		t.Errorf("the fields: %+v; want a details chunk of the unmapped section's fields", d)
	}
	if b.Kind != Boilerplate || b.Text != `unmapped "Our depots": Leeds and York.` || b.Unmapped != Tokens("Leeds and York.") {
		t.Errorf("the boilerplate: %+v; want the unmapped text, with its heading, counted", b)
	}
	if a := Totalled(Ad{Chunks: cs}, 20); a.Unclassified != b.Unmapped || a.Embedded != 0 {
		t.Errorf("totals: %+v", a)
	}
}

// YAML writes an ad's chunks as YAML that reads back as written, a
// chunk's text a block of its lines, whatever the text holds.
func TestYAML(t *testing.T) {
	tricky := []string{
		"📢 Calling all teachers 📢\nnext line", "✏️ Pencils", "a: b", "# not a comment", "yes", "42",
		"it's \"quoted\"", " leading space\nsecond", "control\x02character", "trailing ", "-", "",
		"Pay: £130 - £170 per day", "tab\tin it",
	}
	a := Ad{ID: "225682916", Title: "Part-Time Teacher: Supply", Tokens: 525, Embedded: 143, Unclassified: 60}
	for i, s := range tricky {
		a.Chunks = append(a.Chunks, Chunk{
			Kind: sections.Summary, Heading: s, PlacedBy: s, Use: Embed, Text: s, Tokens: i,
			Fields: []Field{{Label: "pay", Value: s}},
		})
	}
	var b strings.Builder
	if err := a.WriteYAML(&b); err != nil {
		t.Fatal(err)
	}
	var back Ad
	if err := yaml.Unmarshal([]byte(b.String()), &back); err != nil {
		t.Fatalf("not YAML: %v\n%s", err, b.String())
	}
	if !reflect.DeepEqual(a, back) {
		t.Errorf("read back differs.\nwrote:\n%s\ngot %+v", b.String(), back)
	}
	for _, want := range []string{
		"id: '225682916'\n", "tokens: 525\nembedded_tokens: 143\nunclassified_tokens: 60\nchunks:\n",
		"    text: |-\n      📢 Calling all teachers 📢\n      next line\n", "    use: embed\n",
	} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("the YAML lacks %q; it was:\n%s", want, b.String())
		}
	}
}

// An ad's totals: all its tokens, those embedded, and those unclassified.
func TestTotals(t *testing.T) {
	cs := []Chunk{{Use: Embed, Tokens: 10}, {Kind: Boilerplate, Use: LeaveOut, Tokens: 9, Unmapped: 4}, {Use: LeaveOut, Tokens: 6}, {Use: Embed, Tokens: 1}}
	a := Totalled(Ad{Chunks: cs}, 30)
	if a.Tokens != 30 || a.Embedded != 11 || a.Unclassified != 4 {
		t.Errorf("got %d, %d, %d; want 30, 11, 4", a.Tokens, a.Embedded, a.Unclassified)
	}
}

// Boilerplate leaves embedded chunks for a chunk of its own; other
// chunks keep theirs, as part of a list such as an advertiser's perks.
func TestBoilerplateGathered(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Benefits, Heading: "Perks", Parts: []headings.Part{text("A virtual GP."), text("A pension plan.")}},
		{Kind: sections.Requirements, Heading: "You'll need", Parts: []headings.Part{text("A licence."), text("A pension plan.")}},
		{Kind: sections.Summary, Parts: []headings.Part{text("We hire drivers."), text("A pension plan.")}},
	}
	cs := Of(ss, func(s string) bool { return s == "A pension plan." })
	if len(cs) != 4 {
		t.Fatalf("got %d chunks, want 3 and one of boilerplate", len(cs))
	}
	if got := cs[0].Text; got != "Perks\nA virtual GP.\nA pension plan." {
		t.Errorf("benefits: %q; want every line kept", got)
	}
	if got := cs[1].Text; got != "You'll need\nA licence." {
		t.Errorf("requirements: %q; want the boilerplate gone", got)
	}
	if got := cs[3].Text; got != "requirements: A pension plan.\nsummary: A pension plan." {
		t.Errorf("boilerplate: %q; want each line with where it was", got)
	}
	// No boilerplate, no chunk of it.
	if got := Of(ss[:1], func(string) bool { return false }); len(got) != 1 {
		t.Errorf("got %d chunks, want 1", len(got))
	}
}
