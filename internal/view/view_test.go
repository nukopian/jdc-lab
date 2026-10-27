package view

import (
	"slices"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

func TestPick(t *testing.T) {
	var cs []Candidate
	for _, id := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"} {
		cs = append(cs, Candidate{ID: id})
	}
	cs[6].NoHeading, cs[7].Unmapped, cs[8].MostlyBoilerplate, cs[9].Lost = true, true, true, true
	got := Pick(cs, 5)
	if len(got) != 5 || got[0] != 6 || got[1] != 7 || got[2] != 8 || got[3] != 9 {
		t.Fatalf("got %v, want 6, 7, 8 and 9 first, then one more", got)
	}
	// The same ads however the candidates come.
	rev := slices.Clone(cs)
	slices.Reverse(rev)
	var a, b []string
	for _, i := range got {
		a = append(a, cs[i].ID)
	}
	for _, i := range Pick(rev, 5) {
		b = append(b, rev[i].ID)
	}
	if !slices.Equal(a, b) {
		t.Errorf("reversed, got %v, want %v", b, a)
	}
	if n := len(Pick(cs, 50)); n != len(cs) {
		t.Errorf("asked for more than there are: got %d, want %d", n, len(cs))
	}
}

func TestRender(t *testing.T) {
	loc := "Leeds"
	ad := ads.Ad{ID: "7", Title: "Driver", Location: &loc, LocationCanonical: &loc}
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{
			{Role: headings.Repeat, Text: "Driver"},
			{Role: headings.Text, Text: "We deliver *fast*."},
		}},
		{Kind: sections.Details, Heading: "Details", Parts: []headings.Part{
			{Role: headings.Field, Text: "Salary: £30,000", Key: "salary", Value: "£30,000"},
			{Role: headings.Text, Text: "- Days only", Item: true},
			{Role: headings.Text, Text: "Weekends off", Item: true},
			{Role: headings.Text, Text: "Start soon."},
		}},
		{Kind: sections.Unmapped, Heading: "Our depots"},
		{Kind: sections.Apply, Heading: "How to apply", ByName: true, Parts: []headings.Part{
			{Role: headings.Text, Text: "We are an equal opportunities employer."},
			{Role: headings.Text, Text: "Apply by Friday."},
		}},
	}
	verdict := func(text string) Verdict {
		switch text {
		case "We are an equal opportunities employer.":
			return Boilerplate
		case "Apply by Friday.":
			return Template
		}
		return Own
	}
	var b strings.Builder
	if err := Render(&b, ad, ss, verdict, Lost{}); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	for _, want := range []string{
		"## Driver · 7\n\nLeeds\n\n",
		"### summary · no heading, placed by default\n\n*repeat* Driver\n\nWe deliver \\*fast\\*.\n\n",
		"### details · “Details”\n\n*field* Salary: £30,000\n\n- \\- Days only\n- Weekends off\n\nStart soon.\n\n",
		"### unmapped · “Our depots” · no kind known\n\n*(nothing under it)*\n\n",
		"### apply · “How to apply” · a plain line naming it\n\n",
		"> *boilerplate* We are an equal opportunities employer.\n\n",
		"> *template* Apply by Friday.\n\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Render's output lacks %q; it was:\n%s", want, got)
		}
	}
	if strings.Contains(got, "not in any section") {
		t.Error("nothing is missing, but Render says something is")
	}
	b.Reset()
	if err := Render(&b, ad, ss, verdict, Lost{Context: "…lost words…"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "### not in any section\n\n…lost words…\n\n") {
		t.Errorf("Render doesn't show what's missing:\n%s", b.String())
	}
	b.Reset()
	if err := Render(&b, ad, ss, verdict, Lost{Context: "…moved words…", Moved: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "### out of the ad's order\n\n…moved words…\n\n") {
		t.Errorf("Render doesn't show what moved:\n%s", b.String())
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
	if got := Missing("<p></p>", nil); got != (Lost{}) {
		t.Errorf("an empty ad loses nothing, but Missing gives %+v", got)
	}
	// Every letter there, but not in the ad's order.
	moved := []sections.Section{ss[1], ss[0]}
	if got := Missing(whole, moved); !got.Moved || !strings.Contains(got.Context, "We are hiring") {
		t.Errorf("Missing gives %+v, want the text around We are hiring, moved", got)
	}
}
