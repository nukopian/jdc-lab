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
// well as what headings such as "The role" start. Details holds the job's
// terms, such as its hours and contract, under headings like "Job details"
// or a field's label. Unmapped is what a heading of no known kind starts:
// its share of the text is how much is left to find phrases for.
const (
	Summary          Kind = "summary"
	Responsibilities Kind = "responsibilities"
	Requirements     Kind = "requirements"
	Desirable        Kind = "desirable"
	Benefits         Kind = "benefits"
	Company          Kind = "company"
	Apply            Kind = "apply"
	Legal            Kind = "legal"
	Details          Kind = "details"
	Unmapped         Kind = "unmapped"
)

// Kinds are all the kinds, in the order ads tend to have them.
var Kinds = []Kind{
	Summary, Responsibilities, Requirements, Desirable, Benefits,
	Company, Apply, Legal, Details, Unmapped,
}

//go:embed headings.txt
var rules string

// phrases is the word list built from rules, finder the grammar that finds
// its phrases in a heading, and rank each kind's place among the rules,
// which decides between kinds when a heading has phrases of more than one.
var (
	phrases = grammar.NewWordList("sections",
		grammar.FoldCase(), grammar.LooseSpacing(), grammar.Plurals(), grammar.WholeWords(),
		grammar.MapOnly())
	finder *grammar.Grammar
	rank   = map[Kind]int{}
)

func init() {
	// MapOnly keeps the name of a kind written "phrases => kind" out of
	// its phrases.
	if err := phrases.AddRules(rules); err != nil {
		panic(fmt.Sprintf("sections: headings.txt: %v", err))
	}
	// Each rule's kind must be one of Kinds: a misspelt one would
	// otherwise be another kind of section, quietly. Its rank is where its
	// first rule comes.
	for i, k := range phrases.Canonicals() {
		if !slices.Contains(Kinds, Kind(k)) {
			panic(fmt.Sprintf("sections: headings.txt: %q is no kind of section", k))
		}
		rank[Kind(k)] = i
	}
	finder = grammar.MustCompile(`grammar Heading { token TOP { @sections } }`, phrases)

	for _, k := range Kinds {
		if _, ok := rank[k]; !ok && k != Unmapped {
			panic(fmt.Sprintf("sections: headings.txt has no rule for %q", k))
		}
	}
}

// Of returns the kind of section a heading starts, given its key (see
// headings.Key), and whether it's known. A heading starting "why" is
// benefits, whatever phrases are in it: it's a pitch to the reader ("Why
// Brightway?", "Why you'll love this role"). Otherwise, of the kinds its
// phrases point to, the one whose rule comes first in headings.txt wins.
// With none, a heading starting "about" is company ("About Brightway
// Recruitment"), and a field's label details ("Working Hours").
func Of(key string) (Kind, bool) {
	first, _, _ := strings.Cut(key, " ")
	if first == "why" {
		return Benefits, true
	}
	if kinds := All(key); len(kinds) > 0 {
		return slices.MinFunc(kinds, func(a, b Kind) int { return rank[a] - rank[b] }), true
	}
	if first == "about" {
		return Company, true
	}
	if headings.IsLabel(key) {
		return Details, true
	}
	return "", false
}

// All returns every kind the phrases in a heading point to, in the order
// they come in it, each once: more than one means the heading's words
// disagree, and Of chose between them.
func All(key string) []Kind {
	// With Outermost, a phrase that starts inside another is found too, as
	// "you will need" is in "in this role you will need"; one wholly
	// inside a longer one is left out, as "profile" is in "company
	// profile", or it would outrank the phrase it's part of. A phrase
	// list's canonical values are always strings, so ToAll can't fail.
	made, _ := grammar.ToAll[Kind](finder.FindAll(key, grammar.Outermost()))
	var kinds []Kind
	for _, k := range made {
		if !slices.Contains(kinds, k) {
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
	// ByName is true when Heading had no cue to be a heading, and started
	// the section by being exactly a kind's phrase.
	ByName bool `json:"by_name,omitempty"`
	// Parts are what follows: text, list items and fields. A heading
	// followed at once by another has none.
	Parts []headings.Part `json:"parts,omitempty"`
}

// Divide divides an ad's parts into sections. Parts before the first
// heading are the summary. A heading of no known kind starts an Unmapped
// section, even a sub-heading such as "Software Development" among the
// requirements, so that what's left to map shows; a line of text that is
// exactly a section's name, with no cue, starts that section.
func Divide(parts []headings.Part) []Section {
	sections := []Section{{Kind: Summary}}
	for _, p := range parts {
		if kind, ok := starts(p); ok {
			byName := p.Role != headings.Heading
			sections = append(sections, Section{Kind: kind, Heading: p.Text, ByName: byName})
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
		if kind, ok := Of(p.Key); ok {
			return kind, true
		}
		return Unmapped, true
	case p.Role == headings.Text && !p.Item:
		return Named(headings.Key(p.Text))
	}
	return "", false
}
