// Package walls splits walls of text into sentences: paragraphs and lines
// too long to be one thing, written as one block, with headings and
// standard text run into them. Each sentence becomes a Line, as a <br>
// would make it, so that the stages after read a wall as they read lines,
// and find the headings and boilerplate in it.
//
// It's a stage between internal/blocks and internal/headings: blocks only
// records what the markup says, and headings can't read the heading
// phrases internal/sections keeps.
package walls

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"github.com/nukopian/jdc-lab/internal/sentences"
)

// MinChars is how long a paragraph or line must be, in characters, to be
// a wall: longer than this. Over the samples, it takes in all the walls
// with a heading run in (the shortest is 466), and those mixing sentences
// repeated across ads with their own.
const MinChars = 400

// MaxRunInWords is the most words a heading run into a wall has.
const MaxRunInWords = 6

// Split returns bs with each wall split into its sentences (see
// internal/sentences), each a Line. A heading run into a sentence's start
// is a Line of its own, before it, with the cue RunIn: a known heading
// phrase ending with a colon or dash, or followed by a word that opens a
// sentence ("The Role Reporting to …"; see phrased), or a title followed
// by a capital "The", "A" or "An" ("Required Attributes and Experience
// The ideal …"). A wall's bold goes with every line, its other
// cues with the first. Lists, and paragraphs and lines no longer than
// MinChars, are left as they are.
func Split(bs []blocks.Block) []blocks.Block {
	var out []blocks.Block
	for _, b := range bs {
		if !Is(b) {
			out = append(out, b)
			continue
		}
		out = append(out, split(b)...)
	}
	return out
}

// Is reports whether b is a wall: a paragraph or line longer than MinChars.
func Is(b blocks.Block) bool {
	return b.Kind != blocks.List && utf8.RuneCountInString(b.Text) > MinChars
}

// split splits one wall.
func split(wall blocks.Block) []blocks.Block {
	var out []blocks.Block
	for _, s := range sentences.Split(wall.Text) {
		if head, rest, ok := runIn(s); ok {
			out = append(out, blocks.Block{Kind: blocks.Line, Text: head, Cues: blocks.Cues{RunIn: true}})
			s = rest
		}
		out = append(out, blocks.Block{Kind: blocks.Line, Text: s})
	}
	for i := range out {
		out[i].Cues.Bold = wall.Cues.Bold
	}
	first := &out[0].Cues
	first.Heading, first.Added, first.Bullet, first.Gap = wall.Cues.Heading, wall.Cues.Added, wall.Cues.Bullet, wall.Cues.Gap
	// A bold lead stays only where it's still the start of the first line,
	// and less than all of it; all of it is bold.
	switch lead := wall.Cues.BoldLead; {
	case lead == "":
	case lead == out[0].Text:
		first.Bold = true
	case strings.HasPrefix(out[0].Text, lead):
		first.BoldLead = lead
	}
	return out
}

// runIn reports whether a heading is run into the start of sentence s,
// and if so returns it, and the rest of s. Where both signs find one, the
// longer heading wins: "Required" is a phrase, but "Required Attributes
// and Experience" is the title before "The".
func runIn(s string) (head, rest string, ok bool) {
	words := strings.Split(s, " ")
	most := min(len(words)-1, MaxRunInWords) // something must follow the heading
	n := max(phrased(words, most), title(words, most))
	if n == 0 {
		return "", "", false
	}
	return strings.Join(words[:n], " "), strings.Join(words[n:], " "), true
}

// phrased returns how many of words, at most most, are the longest known
// heading phrase they start with, run into what follows it; 0 if none is.
// It's run in when it ends with a colon or dash ("The Legal Bit: Acme
// is …"), or the word after it opens a sentence ("The Role Reporting to
// …"): a capital alone may go on with the phrase ("Privacy Shield",
// "The Team Leader"), or the phrase open a sentence ("You will have CSA
// accreditations"). A field's label stays with its value: "Salary:
// Depending on experience" is a field.
func phrased(words []string, most int) int {
	for n := most; n >= 1; n-- {
		head := strings.Join(words[:n], " ")
		key := headings.Key(head)
		if _, ok := sections.Named(key); !ok || headings.IsLabel(key) {
			continue
		}
		if !upper(words[n]) {
			continue
		}
		if strings.ContainsAny(head[len(head)-1:], ":-") || opensSentence[strings.ToLower(words[n])] {
			return n
		}
	}
	return 0
}

// opensSentence are words that open a sentence after a heading run into
// it, more often than they go on with the heading.
var opensSentence = map[string]bool{
	"the": true, "a": true, "an": true, "we": true, "we're": true, "we've": true, "you": true,
	"you'll": true, "you're": true, "you've": true, "our": true, "your": true, "as": true,
	"this": true, "these": true, "if": true, "please": true, "it": true, "it's": true, "they": true,
	"there": true, "there's": true, "working": true, "reporting": true, "join": true, "do": true,
	"are": true, "is": true, "i": true, "i'm": true, "my": true, "here": true, "based": true,
}

// title returns how many of words, at most most and at least two, are a
// title followed by a capital article; 0 if they start with none.
func title(words []string, most int) int {
	if !titled(words[0], true) {
		return 0
	}
	for n := 2; n <= most; n++ {
		if !titled(words[n-1], false) {
			return 0
		}
		if articles[words[n]] {
			return n
		}
	}
	return 0
}

// articles are the words a sentence starts with, after a title run into it.
var articles = map[string]bool{"The": true, "A": true, "An": true}

// titled reports whether word can be part of a title: capitalised, a
// figure, or a small word titles keep in lower case ("and", "of"). A
// title's first word can't be a small one, nor a word that opens a
// sentence more often than a title ("As A Driver you will …").
func titled(word string, first bool) bool {
	lower := strings.ToLower(word)
	if first {
		return upper(word) && !opener[lower]
	}
	if small[lower] && !notInTitles[lower] {
		return true
	}
	r, _ := utf8.DecodeRuneInString(word)
	return (unicode.IsUpper(r) || unicode.IsDigit(r)) && !notInTitles[lower]
}

// upper reports whether word starts with a capital letter.
func upper(word string) bool {
	r, _ := utf8.DecodeRuneInString(word)
	return unicode.IsUpper(r)
}

// small are the words a title keeps in lower case.
var small = map[string]bool{
	"and": true, "or": true, "of": true, "the": true, "a": true, "an": true, "to": true, "for": true,
	"in": true, "on": true, "at": true, "with": true, "&": true, "-": true, "/": true, "|": true,
}

// opener are words that open a sentence, before a capital, more often than
// they open a title: "As A Driver you will …", "At The Heart of …".
var opener = map[string]bool{
	"as": true, "at": true, "in": true, "on": true, "for": true, "with": true, "by": true, "from": true,
	"to": true, "if": true, "when": true, "while": true, "after": true, "before": true, "and": true,
	"or": true, "but": true, "so": true, "we": true, "you": true, "our": true, "your": true,
	"this": true, "that": true, "take": true, "join": true, "please": true,
}

// notInTitles are words that make a run a sentence, not a title.
var notInTitles = map[string]bool{
	"is": true, "are": true, "was": true, "were": true, "will": true, "be": true, "have": true,
	"has": true, "we": true, "you": true, "i": true,
}
