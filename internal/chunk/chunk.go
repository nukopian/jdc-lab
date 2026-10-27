// Package chunk turns an ad's sections into chunks, as an embedding job
// would ask for them: each with its kind and heading, its text, what to
// do with it (embed it, keep it as metadata, or leave it out), about how
// many tokens it is, and what placed it in its section.
//
// The sections that say what the job is are embedded: the summary,
// responsibilities, requirements, desirable, and unmapped ones, whose
// kind isn't known. Details are metadata, and benefits, company, apply
// and legal are left out. Text found as boilerplate is left out of a
// chunk's text wherever it is, and listed.
package chunk

import (
	"fmt"
	"strings"

	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

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
	// Text is the chunk's heading and text, a line each, without its
	// boilerplate.
	Text string `json:"text" yaml:"text"`
	// Boilerplate is the text left out of Text as boilerplate.
	Boilerplate []string `json:"boilerplate,omitempty" yaml:"boilerplate,omitempty"`
	Fields      []Field  `json:"fields,omitempty" yaml:"fields,omitempty"`
	// Tokens is about how many tokens Text is: four for every three
	// words, as models' tokenizers make of English.
	Tokens int `json:"tokens" yaml:"tokens"`
}

// Of returns an ad's sections as chunks; boilerplate says whether a text
// part was found to be boilerplate.
func Of(ss []sections.Section, boilerplate func(string) bool) []Chunk {
	var cs []Chunk
	for _, s := range ss {
		c := Chunk{Kind: s.Kind, Heading: s.Heading, PlacedBy: placedBy(s), Use: use(s.Kind)}
		var lines []string
		if s.Heading != "" {
			lines = append(lines, s.Heading)
		}
		for _, p := range s.Parts {
			if p.Role == headings.Text && boilerplate(p.Text) {
				c.Boilerplate = append(c.Boilerplate, p.Text)
				continue
			}
			if p.Role == headings.Field {
				c.Fields = append(c.Fields, Field{Label: p.Key, Value: p.Value})
			}
			lines = append(lines, p.Text)
		}
		c.Text = strings.Join(lines, "\n")
		c.Tokens = Tokens(c.Text)
		cs = append(cs, c)
	}
	return cs
}

// Tokens returns about how many tokens text is: four for every three
// words.
func Tokens(text string) int {
	return (len(strings.Fields(text))*4 + 1) / 3 // to the nearest
}

// use says what's done with a section of kind.
func use(kind sections.Kind) string {
	switch kind {
	case sections.Summary, sections.Responsibilities, sections.Requirements, sections.Desirable, sections.Unmapped:
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
