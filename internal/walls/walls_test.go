package walls

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/blocks"
)

// pad makes text a wall by adding sentences of its own after it.
func pad(text string) string {
	for len(text) <= MinChars {
		text += " We work across the region with a friendly team of engineers."
	}
	return text
}

// show writes blocks as one line each: the kind, the text, and run_in.
func show(bs []blocks.Block) []string {
	var out []string
	for _, b := range bs {
		s := fmt.Sprintf("%s %q", b.Kind, b.Text)
		if b.Cues.RunIn {
			s += " run_in"
		}
		if b.Cues.Bold {
			s += " bold"
		}
		if b.Cues.Gap {
			s += " gap"
		}
		out = append(out, s)
	}
	return out
}

func TestSplit(t *testing.T) {
	filler := "We work across the region with a friendly team of engineers."
	tests := []struct {
		name  string
		block blocks.Block
		want  []string
	}{
		{
			"a paragraph not over MinChars is left alone",
			blocks.Block{Kind: blocks.Paragraph, Text: "We're hiring. The Role You'll drive."},
			[]string{`paragraph "We're hiring. The Role You'll drive."`},
		},
		{
			"a wall becomes a line a sentence",
			blocks.Block{Kind: blocks.Paragraph, Text: pad("We're hiring a driver. You'll deliver parcels.")},
			slices.Concat([]string{`line "We're hiring a driver."`, `line "You'll deliver parcels."`},
				repeat(`line "`+filler+`"`, strings.Count(pad("We're hiring a driver. You'll deliver parcels."), filler))),
		},
		{
			"a known heading phrase at a sentence's start, before a capital, is a run-in heading",
			blocks.Block{Kind: blocks.Paragraph, Text: pad("We're a growing firm. The Role Reporting to the director, you'll lead a team.")},
			slices.Concat([]string{`line "We're a growing firm."`, `line "The Role" run_in`,
				`line "Reporting to the director, you'll lead a team."`},
				repeat(`line "`+filler+`"`, strings.Count(pad("We're a growing firm. The Role Reporting to the director, you'll lead a team."), filler))),
		},
		{
			"a title before a capital article is a run-in heading",
			blocks.Block{Kind: blocks.Paragraph, Text: pad("You'll run the accounts. Required Attributes and Experience The ideal person has ten years.")},
			slices.Concat([]string{`line "You'll run the accounts."`, `line "Required Attributes and Experience" run_in`,
				`line "The ideal person has ten years."`},
				repeat(`line "`+filler+`"`, strings.Count(pad("You'll run the accounts. Required Attributes and Experience The ideal person has ten years."), filler))),
		},
		{
			"no run-in heading: a lower-case word after the phrase, a name, or a run opening with As",
			blocks.Block{Kind: blocks.Paragraph, Text: pad("Benefits include a pension. We are The Best Agency in town. As A Driver you will travel.")},
			slices.Concat([]string{`line "Benefits include a pension."`, `line "We are The Best Agency in town."`,
				`line "As A Driver you will travel."`},
				repeat(`line "`+filler+`"`, strings.Count(pad("Benefits include a pension. We are The Best Agency in town. As A Driver you will travel."), filler))),
		},
		{
			"the wall's gap goes with its first line, its bold with every line",
			blocks.Block{Kind: blocks.Line, Text: pad("We're hiring a driver."), Cues: blocks.Cues{Bold: true, Gap: true}},
			slices.Concat([]string{`line "We're hiring a driver." bold gap`},
				repeat(`line "`+filler+`" bold`, strings.Count(pad("We're hiring a driver."), filler))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := show(Split([]blocks.Block{tt.block}))
			if !slices.Equal(got, tt.want) {
				t.Errorf("Split\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestRunIn(t *testing.T) {
	for s, want := range map[string]string{
		"The Role Reporting to the director, you'll lead.":        "The Role",
		"About the Role We are looking for a planner.":            "About the Role",
		"The Legal Bit: Acme is an equal opportunities employer.": "The Legal Bit:",
		"Required Attributes and Experience The ideal person.":    "Required Attributes and Experience",
		// The capital goes on with the phrase.
		"Privacy Shield.":                      "",
		"The Team Leader will act as a guide.": "",
		"Why Join Trade Mastermind?":           "",
		// The phrase opens a sentence.
		"You will have CSA accreditations in hand.": "",
		"In return, Acme offers a pension.":         "",
		// A field keeps its value.
		"Salary: Depending on Experience and skills.": "",
		"Hours: Monday to Friday, 8am to 4pm.":        "",
	} {
		head, _, ok := runIn(s)
		if !ok {
			head = ""
		}
		if head != want {
			t.Errorf("runIn(%q) = %q, want %q", s, head, want)
		}
	}
}

func TestSplitLeavesListsAlone(t *testing.T) {
	list := blocks.Block{Kind: blocks.List, Items: []blocks.Item{{Text: pad("One. Two.")}}}
	if got := Split([]blocks.Block{list}); len(got) != 1 || got[0].Kind != blocks.List {
		t.Errorf("Split of a list = %v, want the list as it was", got)
	}
}

func repeat(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}

func TestRunInNeedsACapital(t *testing.T) {
	for _, s := range []string{"Our client is a leading firm.", "We are seeking a planner.", "The role is to sell."} {
		if head, _, ok := runIn(s); ok {
			t.Errorf("runIn(%q) = %q, want none: a lower-case word goes on with the sentence", s, head)
		}
	}
}
