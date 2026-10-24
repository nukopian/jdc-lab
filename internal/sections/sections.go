// Package sections divides a job ad into sections, by the headings
// internal/headings found: its summary, responsibilities, requirements and
// the rest.
//
// Which section a heading starts is data, not code: headings.txt holds
// synonym rules, each a section and the phrases that mean it, which a
// grammar word list (github.com/nukopian/grammar) finds in the heading's
// key. Adding a synonym means adding it there.
package sections

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/nukopian/grammar"
	"github.com/nukopian/jdc-lab/internal/headings"
)

// A Kind is a kind of section.
type Kind string

// The kinds of section. Summary is what comes before the first heading, as
// well as what headings such as "The role" start.
const (
	Summary          Kind = "summary"
	Responsibilities Kind = "responsibilities"
	Requirements     Kind = "requirements"
	Desirable        Kind = "desirable"
	Benefits         Kind = "benefits"
	Company          Kind = "company"
	Apply            Kind = "apply"
	Legal            Kind = "legal"
)

// Kinds are all the kinds, in the order ads tend to have them.
var Kinds = []Kind{Summary, Responsibilities, Requirements, Desirable, Benefits, Company, Apply, Legal}

//go:embed headings.txt
var rules string

// phrases is the word list built from rules, heading the grammar that
// finds its phrases in a heading, and rank each kind's place among the
// rules, which decides between kinds when a heading has phrases of more
// than one.
var (
	phrases = grammar.NewWordList("sections",
		grammar.FoldCase(), grammar.LooseSpacing(), grammar.Plurals(), grammar.WholeWords())
	heading *grammar.Grammar
	rank    = map[Kind]int{}
)

func init() {
	if err := phrases.AddRules(rules); err != nil {
		panic(fmt.Sprintf("sections: headings.txt: %v", err))
	}
	// Each rule's first phrase must be a kind: a misspelt one would
	// otherwise be a ninth kind of section, quietly.
	for _, p := range phrases.Phrases() {
		for _, k := range phrases.Canonical(p) {
			if !slices.Contains(Kinds, Kind(k)) {
				panic(fmt.Sprintf("sections: headings.txt: %q is no kind of section", k))
			}
		}
	}
	heading = grammar.MustCompile(`grammar Heading { token TOP { @sections } }`, phrases)

	// A rule starts at the beginning of a line, and carries on in indented
	// ones; its first phrase is its kind.
	for line := range strings.Lines(rules) {
		if line == "" || strings.ContainsRune(" \t\n#", rune(line[0])) {
			continue
		}
		kind, _, _ := strings.Cut(line, ",")
		if _, ok := rank[Kind(kind)]; !ok {
			rank[Kind(kind)] = len(rank)
		}
	}
	for _, k := range Kinds {
		if _, ok := rank[k]; !ok {
			panic(fmt.Sprintf("sections: headings.txt has no rule for %q", k))
		}
	}
}

// Of returns the kind of section a heading starts, given its key (see
// headings.Key), and whether it's known. Of the kinds its phrases point
// to, the one whose rule comes first in headings.txt wins. With none, a
// heading starting "about" is company ("About TeacherActive") and one
// starting "why" benefits ("Why Tradewind?").
func Of(key string) (Kind, bool) {
	if kinds := All(key); len(kinds) > 0 {
		return slices.MinFunc(kinds, func(a, b Kind) int { return rank[a] - rank[b] }), true
	}
	switch first, _, _ := strings.Cut(key, " "); first {
	case "about":
		return Company, true
	case "why":
		return Benefits, true
	}
	return "", false
}

// All returns every kind the phrases in a heading point to, in the order
// they come in it, each once: more than one means the heading's words
// disagree, and Of chose between them.
func All(key string) []Kind {
	var kinds []Kind
	for _, m := range heading.FindAll(key).Items() {
		if k := Kind(m.At("made").String()); !slices.Contains(kinds, k) {
			kinds = append(kinds, k)
		}
	}
	return kinds
}

// Named returns the kind of section text names, if the whole of it is one
// of the phrases: "Responsibilities", but not "Our responsibilities to
// you". It's for text with no cue that it's a heading, such as a plain
// <p>Responsibilities</p>, where only an exact match is safe.
func Named(key string) (Kind, bool) {
	if ks := phrases.Canonical(key); len(ks) > 0 {
		return Kind(ks[0]), true
	}
	return "", false
}

// A Section is a heading and what follows it, up to the next heading that
// starts a section.
type Section struct {
	Kind Kind `json:"kind"`
	// Heading is the heading that started it, as written; empty for the
	// summary before the first heading.
	Heading string `json:"heading,omitempty"`
	// Parts are what follows: text, list items and fields, and headings
	// that start no section of their own, such as "Software Development"
	// among the requirements.
	Parts []headings.Part `json:"parts"`
}

// Divide divides an ad's parts into sections. Parts before the first
// heading are the summary. A heading of no known kind stays among the
// parts of the section it's in, as a sub-heading; a line of text that is
// exactly a section's name, with no cue, starts that section.
func Divide(parts []headings.Part) []Section {
	sections := []Section{{Kind: Summary}}
	for _, p := range parts {
		if kind, ok := starts(p); ok {
			sections = append(sections, Section{Kind: kind, Heading: p.Text})
			continue
		}
		last := &sections[len(sections)-1]
		last.Parts = append(last.Parts, p)
	}
	if len(sections[0].Parts) == 0 {
		sections = sections[1:]
	}
	return sections
}

// starts reports whether p starts a section, and of what kind.
func starts(p headings.Part) (Kind, bool) {
	switch {
	case p.Role == headings.Heading:
		return Of(p.Key)
	case p.Role == headings.Text && !p.Item:
		return Named(headings.Key(p.Text))
	}
	return "", false
}
