// Package wording says what kind of section a line belongs in by how it
// opens, where no heading placed it: "Experience of…" is a requirement,
// "Conduct site surveys…" a responsibility, "Free parking" a benefit.
//
// The openings are data, in openings.txt, as headings.txt is, read by a
// grammar with the grammar package: a word list of the openings, found
// only at the start of a line.
package wording

import (
	_ "embed"
	"strings"
	"unicode"

	"github.com/nukopian/grammar"

	"github.com/nukopian/jdc-lab/internal/sections"
)

//go:embed openings.txt
var rules string

var (
	openings = grammar.NewWordList("openings",
		grammar.FoldCase(), grammar.LooseSpacing(), grammar.WholeWords(), grammar.MapOnly())
	// opening reads a line's opening: past any bullet or other mark, and
	// a figure ("33 days holiday"), the longest of the openings.
	opening *grammar.Grammar
)

func init() {
	if err := openings.AddRules(rules); err != nil {
		panic("wording: openings.txt: " + err.Error())
	}
	opening = grammar.MustCompile(`
grammar Opening {
    token TOP { ^ <-[\w]>* $<phrase>=[ [ \d+ \s+ ]? $<open>=@openings ] }
}`, openings)
}

// Of returns the kind of section a line belongs in by its opening, and the
// opening, as the line has it; ok is false for a line that opens with none
// of the openings, or that's a question.
func Of(line string) (kind sections.Kind, phrase string, ok bool) {
	if strings.HasSuffix(strings.TrimRightFunc(line, unicode.IsSpace), "?") {
		return "", "", false
	}
	m := opening.Find(line)
	if !m.Defined() {
		return "", "", false
	}
	// The list's own capture, open, has its canonical value as made; it
	// belongs to the rule, not to phrase, which captures text only.
	k := m.At("open.made").Str()
	if k == "" {
		return "", "", false
	}
	return sections.Kind(k), strings.TrimSpace(m.At("phrase.~").Str()), true
}
