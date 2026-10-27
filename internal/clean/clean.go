// Package clean normalises an ad's text for the stages that read it, and
// for search: structure in English words, not symbols.
//
// Emoji, pictographs and the symbols ads use as bullets go, as do
// invisible and control characters and the replacement character left by
// text mis-encoded; odd spaces become plain ones, and typographic quotes,
// dashes and ellipses plain ASCII. Letters, figures and currency signs
// stay, accented letters included: "Café" and "£30,000" are words. A
// symbol that opens a line and stands for one thing becomes the English
// label it stands for, so "📍 Leeds" is "Location: Leeds", a field.
package clean

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// labels are the symbols that, opening a line, stand for one thing, and
// the English label each becomes: those ads use most, with one meaning.
var labels = map[rune]string{
	0x1F4CD: "Location", // round pushpin
	0x1F4B7: "Pay",      // banknote with pound sign
	0x1F4B0: "Pay",      // money bag
	0x23F0:  "Hours",    // alarm clock
	0x1F552: "Hours",    // clock face, three o'clock
	0x23F1:  "Hours",    // stopwatch
	0x1F4DE: "Contact",  // telephone receiver
	0x1F4E7: "Email",    // e-mail symbol
}

// fold are the punctuation marks that become plain ASCII: quotes and
// primes, hyphens and dashes, and the ellipsis. An em dash, a break in a
// sentence, becomes " -- ", an en dash or hyphen "-". Full-width forms
// become what they're the width of (see Text).
var fold = map[rune]string{
	0x2018: "'", 0x2019: "'", 0x201A: "'", 0x201B: "'", 0x2032: "'", 0x2039: "'", 0x203A: "'",
	0x201C: `"`, 0x201D: `"`, 0x201E: `"`, 0x201F: `"`, 0x2033: `"`, 0x00AB: `"`, 0x00BB: `"`,
	0x2010: "-", 0x2011: "-", 0x2012: "-", 0x2013: "-", 0x2212: "-", 0xFE63: "-",
	0x2014: " -- ", 0x2015: " -- ",
	0x2026: "...",
}

// gone reports whether r goes: a symbol other than a currency sign, a
// bullet, or a character no one sees.
func gone(r rune) bool {
	switch {
	case unicode.Is(unicode.So, r), unicode.Is(unicode.Sk, r), unicode.Is(unicode.Co, r), unicode.Is(unicode.Cf, r):
		return true // emoji, pictographs, modifiers, private use (Word's bullet), zero-width
	case unicode.IsControl(r) && r != '\n' && r != '\t':
		return true
	case variation(r):
		return true
	case r == 0xFFFD: // the replacement character, from mis-encoded text
		return true
	}
	switch r {
	case 0x2022, 0x2023, 0x2043, 0x00B7, 0x2219, 0x25CF, 0x25CB, 0x25AA, 0x25AB, 0x25A0, 0x25A1,
		0x25C6, 0x25C7, 0x2756, 0x27A2, 0x27A4, 0x2794: // bullets that aren't symbols by category
		return true
	}
	return false
}

// Text returns s normalised: see the package's doc. Spaces are collapsed,
// and trimmed from both ends.
func Text(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case gone(r):
			continue
		case unicode.IsSpace(r):
			b.WriteByte(' ')
		case r >= 0xFF01 && r <= 0xFF5E: // full-width forms of ASCII
			b.WriteRune(r - 0xFEE0)
		default:
			if f, ok := fold[r]; ok {
				b.WriteString(f)
			} else {
				b.WriteRune(r)
			}
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// Line returns a line of text normalised, as Text does, with a label
// symbol that opens it written as its English label, unless the line has
// a label of its own, or ends as a heading does, with ":", "?" or "!":
// "📍 Leeds" is "Location: Leeds", but "📍 Location: Leeds" and "💰 Why
// apply?" stay as they are. added is what was added, for comparing the
// line with what the ad had. A line in bold, a heading's, takes Text.
func Line(s string) (text, added string) {
	text = Text(s)
	label := opening(s)
	if label == "" || text == "" || labelled(text) || strings.ContainsAny(text[len(text)-1:], ":?!") || !fits(label, text) {
		return text, ""
	}
	added = label + ": "
	return added + text, added
}

// values say what a label's value has in it, so that a heading after a
// symbol, such as "💷 What We Offer", isn't taken for one. A location
// can be any words.
var values = map[string][]string{
	"Pay":     {"£", "$", "€", "competitive", "negotiable", "doe", "per hour", "per annum", "ph", "k"},
	"Hours":   {"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday", "mon", "fri", "full time", "full-time", "part time", "part-time", "flexible", "shift", "days", "nights", "hours", "am", "pm"},
	"Contact": {"call", "phone", "removed", "text", "whatsapp"},
	"Email":   {"@", "removed", "email", "e-mail"},
}

// fits reports whether text could be a value of label: a location can be
// anything, and the others need a figure or one of their values' words.
func fits(label, text string) bool {
	if label == "Location" {
		return true
	}
	if strings.ContainsFunc(text, unicode.IsDigit) {
		return true
	}
	lower := strings.ToLower(text)
	for _, v := range values[label] {
		if strings.ContainsAny(v, "£$€@") && strings.Contains(lower, v) || word(lower, v) {
			return true
		}
	}
	return false
}

// word reports whether w is in s as a word, with no letter either side.
func word(s, w string) bool {
	for i := 0; ; {
		j := strings.Index(s[i:], w)
		if j < 0 {
			return false
		}
		start, end := i+j, i+j+len(w)
		before, _ := utf8.DecodeLastRuneInString(s[:start])
		after, _ := utf8.DecodeRuneInString(s[end:])
		if !unicode.IsLetter(before) && !unicode.IsLetter(after) {
			return true
		}
		i = start + 1
	}
}

// opening returns the English label of the label symbol that opens s,
// past any spaces and other symbols before it, or "".
func opening(s string) string {
	for _, r := range s {
		if l, ok := labels[r]; ok {
			return l
		}
		if !gone(r) && !unicode.IsSpace(r) {
			return ""
		}
	}
	return ""
}

// Bullet returns the symbol a line opens with, past any spaces and
// invisible characters, when it
// stands as a bullet: an emoji or bullet symbol, with text after it, and
// not the symbol the line ends with, as "📢 Hiring now 📢" is decoration.
// Lines in a row that open with the same bullet are a list. It's "" for
// a line that opens with none.
func Bullet(s string) string {
	s = strings.TrimLeftFunc(s, func(r rune) bool { return unicode.IsSpace(r) || unicode.Is(unicode.Cf, r) })
	r, size := utf8.DecodeRuneInString(s)
	if size == 0 || !gone(r) || variation(r) {
		return ""
	}
	if Text(s[size:]) == "" {
		return ""
	}
	end := strings.TrimRightFunc(s, func(r rune) bool { return unicode.IsSpace(r) || variation(r) })
	if last, _ := utf8.DecodeLastRuneInString(end); last == r {
		return ""
	}
	return string(r)
}

// variation reports whether r is a variation selector, or the keycap
// that follows a figure: marks on the symbol before them.
func variation(r rune) bool {
	return r >= 0xFE00 && r <= 0xFE0F || r == 0x20E3
}

// labelled reports whether text starts with a label of its own: a few
// words, then a colon, or a dash between spaces.
func labelled(text string) bool {
	i := strings.IndexAny(text, ":")
	if j := strings.Index(text, " - "); j >= 0 && (i < 0 || j < i) {
		i = j
	}
	return i > 0 && len(strings.Fields(text[:i])) <= 4
}
