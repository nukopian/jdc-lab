// Package chunk turns an ad's sections into chunks, as an embedding job
// would ask for them: each with its kind and heading, its text, what to
// do with it (embed it, keep it as metadata, or leave it out), about how
// many tokens it is, and what placed it in its section.
//
// The sections that say what the job is are embedded: the summary, the
// role, responsibilities, requirements and desirable. Details are metadata;
// benefits, company, apply and legal are left out; and unmapped ones,
// whose kind isn't known, aren't embedded: they're what's left to map. An unmapped section's text goes to the boilerplate chunk,
// under its heading, and its fields to a details chunk.
//
// Text found as boilerplate leaves the embedded chunks for
// one chunk of its own, at the end, left out but there to see, each line
// with the section it came from: the verdict catches advertisers'
// footers, but also job lines common to many ads ("high levels of
// accuracy and attention to detail"), and this is where they show. Other
// chunks keep theirs, as part of what they say: an advertiser's
// standard perks are still a list of benefits.
package chunk

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// Boilerplate is the kind of the chunk of an ad's boilerplate.
const Boilerplate sections.Kind = "boilerplate"

// What to do with a chunk.
const (
	Embed    = "embed"
	Metadata = "metadata"
	LeaveOut = "leave out"
)

// A Field is a label and its value: "salary", "£30,000".
type Field struct {
	Label string `json:"label" yaml:"label"`
	Value string `json:"value" yaml:"value"`
}

// A Chunk is one section of an ad, ready for what's done with it.
type Chunk struct {
	Kind    sections.Kind `json:"kind" yaml:"kind"`
	Heading string        `json:"heading,omitempty" yaml:"heading,omitempty"`
	// PlacedBy says what gave the chunk its kind.
	PlacedBy string `json:"placed_by" yaml:"placed_by"`
	// Use is Embed, Metadata or LeaveOut.
	Use string `json:"use" yaml:"use"`
	// Text is the chunk's heading and text, a line each, a list's items
	// as "- " entries.
	Text   string  `json:"text" yaml:"text"`
	Fields []Field `json:"fields,omitempty" yaml:"fields,omitempty"`
	// Unmapped is how many of the boilerplate chunk's tokens are text of
	// unmapped sections: what's left to map.
	Unmapped int `json:"unmapped_tokens,omitempty" yaml:"unmapped_tokens,omitempty"`
	// Tokens is about how many tokens Text is: four for every three
	// words, as models' tokenizers make of English.
	Tokens int `json:"tokens" yaml:"tokens"`
}

// Of returns an ad's sections as chunks; boilerplate says whether a text
// part was found to be boilerplate.
func Of(ss []sections.Section, boilerplate func(string) bool) []Chunk {
	var cs []Chunk
	var gathered []string // the boilerplate of the embedded chunks, and the unmapped text
	unmapped := 0         // the unmapped text's tokens
	for _, s := range ss {
		if s.Kind == sections.Unmapped {
			// Unmapped text goes to the boilerplate, under its heading, and
			// its fields stay, as details.
			d := Chunk{Kind: sections.Details, Use: Metadata, Heading: s.Heading,
				PlacedBy: fmt.Sprintf("the fields under the unmapped heading %q", s.Heading)}
			var lines []string
			for _, p := range s.Parts {
				if p.Role == headings.Field {
					d.Fields = append(d.Fields, Field{Label: p.Key, Value: p.Value})
					lines = append(lines, line(p))
					continue
				}
				gathered = append(gathered, fmt.Sprintf("unmapped %q: %s", s.Heading, line(p)))
				unmapped += Tokens(p.Text)
			}
			if len(d.Fields) > 0 {
				d.Text = strings.Join(lines, "\n")
				d.Tokens = Tokens(d.Text)
				cs = append(cs, d)
			}
			continue
		}
		c := Chunk{Kind: s.Kind, Heading: s.Heading, PlacedBy: placedBy(s), Use: use(s.Kind)}
		var lines []string
		if s.Heading != "" {
			lines = append(lines, s.Heading)
		}
		for _, p := range s.Parts {
			if c.Use == Embed && p.Role == headings.Text && boilerplate(p.Text) {
				gathered = append(gathered, string(s.Kind)+": "+line(p))
				continue
			}
			if p.Role == headings.Field {
				c.Fields = append(c.Fields, Field{Label: p.Key, Value: p.Value})
			}
			lines = append(lines, line(p))
		}
		c.Text = strings.Join(lines, "\n")
		c.Tokens = Tokens(c.Text)
		cs = append(cs, c)
	}
	if len(gathered) > 0 {
		text := strings.Join(gathered, "\n")
		cs = append(cs, Chunk{
			Kind: Boilerplate, Use: LeaveOut, Text: text, Tokens: Tokens(text), Unmapped: unmapped,
			PlacedBy: "found as boilerplate in the chunks above (in ads for five groups of titles or more, or more often outside the sections about the job), or under a heading of no known kind",
		})
	}
	return cs
}

// line returns a part's text as a line of a chunk: an item of a list as
// an entry, "- " first.
func line(p headings.Part) string {
	if p.Item {
		return "- " + p.Text
	}
	return p.Text
}

// Tokens returns about how many tokens text is: four for every three
// words, a word being something with a letter or figure in it, so that a
// list's "- " isn't one.
func Tokens(text string) int {
	n := 0
	for _, w := range strings.Fields(text) {
		if strings.ContainsFunc(w, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) {
			n++
		}
	}
	return (n*4 + 1) / 3 // to the nearest
}

// use says what's done with a section of kind.
func use(kind sections.Kind) string {
	switch kind {
	case sections.Summary, sections.Role, sections.Responsibilities, sections.Requirements, sections.Desirable:
		return Embed
	case sections.Details:
		return Metadata
	}
	return LeaveOut
}

// placedBy says what gave a section its kind.
func placedBy(s sections.Section) string {
	switch {
	case s.Heading == "":
		return "nothing before it: summary by default"
	case s.ByName:
		return fmt.Sprintf("the line %q: exactly a section's name", s.Heading)
	}
	return fmt.Sprintf("the heading %q: %s", s.Heading, sections.Explain(headings.Key(s.Heading)))
}
