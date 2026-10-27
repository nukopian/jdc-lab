package view

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/chunk"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

func TestPick(t *testing.T) {
	var cs []Candidate
	for i := range 30 {
		cs = append(cs, Candidate{ID: fmt.Sprint(i)})
	}
	byRank := slices.Clone(cs)
	slices.SortFunc(byRank, func(a, b Candidate) int { return cmp.Compare(rank(a.ID), rank(b.ID)) })
	// The last of the 30 by rank has every property worth showing.
	last := slices.IndexFunc(cs, func(c Candidate) bool { return c.ID == byRank[29].ID })
	cs[last].NoHeading, cs[last].Unmapped, cs[last].MostlyBoilerplate, cs[last].Lost = true, true, true, true

	got := ids(cs, Pick(cs, 10))
	want := []string{}
	for _, c := range byRank[:6] { // 10 less the four purposes, always
		want = append(want, c.ID)
	}
	if !slices.Equal(got[:6], want) || !slices.Contains(got, byRank[29].ID) {
		t.Fatalf("got %v; want the six lowest by rank, %v, and %s", got, want, byRank[29].ID)
	}
	if len(got) != 10 {
		t.Fatalf("got %d ads, want 10", len(got))
	}
	// In order of rank, so that each ad keeps its place in the view.
	if !slices.IsSortedFunc(got, func(a, b string) int { return cmp.Compare(rank(a), rank(b)) }) {
		t.Errorf("got %v, not in order of rank", got)
	}
	// Changing what makes an ad worth showing doesn't move the fixed ones.
	cs[last].NoHeading = false
	again := ids(cs, Pick(cs, 10))
	for _, id := range want {
		if !slices.Contains(again, id) {
			t.Errorf("%s left the view when another ad's properties changed: %v", id, again)
		}
	}
	// The same ads however the candidates come.
	rev := slices.Clone(cs)
	slices.Reverse(rev)
	if r := ids(rev, Pick(rev, 10)); !slices.Equal(r, again) {
		t.Errorf("reversed, got %v, want %v", r, again)
	}
	if n := len(Pick(cs[:5], 50)); n != 5 {
		t.Errorf("asked for more than there are: got %d, want 5", n)
	}
}

// ids returns the IDs of the candidates at indexes.
func ids(cs []Candidate, indexes []int) []string {
	var out []string
	for _, i := range indexes {
		out = append(out, cs[i].ID)
	}
	return out
}

func TestRender(t *testing.T) {
	loc := "Leeds"
	ad := ads.Ad{ID: "7", Title: "Driver", Location: &loc, LocationCanonical: &loc}
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{
			{Role: headings.Repeat, Text: "Driver"},
			{Role: headings.Text, Text: "We deliver *fast*."},
			{Role: headings.Text, Text: "We are an equal opportunities employer."},
		}},
		{Kind: sections.Details, Heading: "Details", Parts: []headings.Part{
			{Role: headings.Field, Text: "Salary: £30,000", Key: "salary", Value: "£30,000"},
			{Role: headings.Text, Text: "- Days only", Item: true},
			{Role: headings.Text, Text: "Weekends off", Item: true},
			{Role: headings.Text, Text: "Start soon."},
		}},
	}
	boiler := func(s string) bool { return s == "We are an equal opportunities employer." }
	c := chunk.Totalled(chunk.Ad{ID: ad.ID, Title: ad.Title, Chunks: chunk.Of(ss, boiler)}, 30)
	var b strings.Builder
	if err := Render(&b, ad, c, Lost{}); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	for _, want := range []string{
		"## Driver · 7\n\nLeeds\n\n",
		"*about 30 tokens: 5 embedded (16%), 0 unclassified (0%)*\n\n",
		"### summary · embed · 5 tokens\n\n*placed by nothing before it: summary by default*\n\nDriver\n\nWe deliver \\*fast\\*.\n\n",
		"### details · “Details” · metadata · ",
		"Salary: £30,000\n\n- \\- Days only\n- Weekends off\n\nStart soon.\n\n",
		"### boilerplate · leave out · ",
		"- summary: We are an equal opportunities employer.\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Render's output lacks %q; it was:\n%s", want, got)
		}
	}
	if strings.Contains(got, "\nDetails\n") {
		t.Error("a chunk's heading is shown twice")
	}
	if strings.Contains(got, "not in any section") {
		t.Error("nothing is missing, but Render says something is")
	}
	for _, tt := range []struct {
		lost Lost
		want string
	}{
		{Lost{Context: "…lost words…"}, "### not in any section\n\n…lost words…\n\n"},
		{Lost{Context: "…moved words…", Moved: true}, "### out of the ad's order\n\n…moved words…\n\n"},
	} {
		b.Reset()
		if err := Render(&b, ad, c, tt.lost); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(b.String(), tt.want) {
			t.Errorf("Render lacks %q:\n%s", tt.want, b.String())
		}
	}
}

func TestEscape(t *testing.T) {
	for in, want := range map[string]string{
		"- item":         `\- item`,
		"+ item":         `\+ item`,
		"1. First":       `1\. First`,
		"10 days":        "10 days",
		"<b>£30k</b> #1": `\<b\>£30k\</b\> \#1`,
	} {
		if got := escape(in); got != want {
			t.Errorf("escape(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMissing(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{{Role: headings.Text, Text: "We are hiring a Multi Skilled Operativeto join us."}}},
		{Kind: sections.Benefits, Heading: "Benefits", Parts: []headings.Part{{Role: headings.Text, Text: "Pension & parking"}}},
	}
	whole := `<p>We are hiring a <strong>Multi Skilled Operative</strong>to join us.</p>` +
		`<p><b>Benefits</b></p><p>Pension &amp; parking</p><style>p { color: red }</style>`
	if got := Missing(whole, ss); got != (Lost{}) {
		t.Errorf("every word is in a section, but Missing gives %+v", got)
	}
	lost := strings.Replace(whole, `<p><b>Benefits</b></p>`, `<p><b>Benefits</b></p><p>A company car</p>`, 1)
	if got := Missing(lost, ss); !strings.Contains(got.Context, "A company car") || got.Moved {
		t.Errorf("Missing gives %+v, want the text around A company car, missing", got)
	}
	// Labels cleaning added aren't the ad's, and aren't missing from it.
	located := []sections.Section{{Kind: sections.Details, Parts: []headings.Part{
		{Role: headings.Field, Text: "Location: Leeds", Key: "location", Value: "Leeds", Added: "Location: "}}}}
	if got := Missing("<p>📍 Leeds</p>", located); got != (Lost{}) {
		t.Errorf("an added label counted against the ad: %+v", got)
	}
	// Faults found in the wash-up: sections with more than the ad lose
	// nothing, and full-width letters are the letters they're the width of.
	more := []sections.Section{{Kind: sections.Summary, Parts: []headings.Part{{Role: headings.Text, Text: "Hello world again"}}}}
	if got := Missing("<p>Hello world</p>", more); got != (Lost{}) {
		t.Errorf("more in the sections than the ad: %+v", got)
	}
	wide := []sections.Section{{Kind: sections.Summary, Parts: []headings.Part{{Role: headings.Text, Text: "Full time role"}}}}
	if got := Missing("<p>\uff26ull time role</p>", wide); got != (Lost{}) {
		t.Errorf("full-width letters: %+v", got)
	}
	if got := Missing("<p></p>", nil); got != (Lost{}) {
		t.Errorf("an empty ad loses nothing, but Missing gives %+v", got)
	}
	// Every letter there, but not in the ad's order.
	moved := []sections.Section{ss[1], ss[0]}
	if got := Missing(whole, moved); !got.Moved || !strings.Contains(got.Context, "We are hiring") {
		t.Errorf("Missing gives %+v, want the text around We are hiring, moved", got)
	}
}

// A bold lead in emphasis marks, found in the wash-up: its line lost the
// opening mark as a bullet, so the lead no longer started it, and the
// line kept the lead's text as well, read twice.
func TestMissingBoldLead(t *testing.T) {
	for _, src := range []string{
		"<p><strong>*Recommend a friend*</strong> - Do you know someone looking for a position?</p>",
		"<p><strong>*Please note</strong>:- We reserve the right to close this vacancy early.</p>",
	} {
		ss := sections.Divide(headings.Read(blocks.Split(src), ""))
		if got := Missing(src, ss); got != (Lost{}) {
			t.Errorf("Missing(%q) = %+v, want none", src, got)
		}
	}
}
