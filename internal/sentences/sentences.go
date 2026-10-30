// Package sentences splits text into its sentences, by rules: a sentence
// ends at a full stop, question mark or exclamation mark followed by a
// space and a capital letter, as written.
//
// It's for walls: paragraphs too long to be one thing, which later stages
// read sentence by sentence (see internal/walls). It knows nothing of job
// ads, only of English text cleaned as internal/clean cleans it: an
// ellipsis is "...", quotes are plain.
package sentences

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Split returns text's sentences in order, each as written, without the
// spaces between them: joined by single spaces they give the text back.
// Text with no sentence end is one sentence, and empty text none.
//
// A stop followed by a lower-case letter or a figure doesn't end one
// ("example. com", "Grade 3. 2 posts"), nor does one after an
// abbreviation or an initial ("e.g. Sage", "J. Smith"). Further stops and
// a closing quote or bracket after it stay with its sentence ("more...",
// "(Full training given.)"), and an opening one may come before the
// capital.
func Split(text string) []string {
	var out []string
	start := 0
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		stop := i
		i += size
		if r != '.' && r != '!' && r != '?' {
			continue
		}
		end := i
		for end < len(text) && strings.IndexByte(`.!?"')]`, text[end]) >= 0 {
			end++
		}
		next := end
		for next < len(text) && text[next] == ' ' {
			next++
		}
		if next == end || !capital(text[next:]) || (r == '.' && abbreviated(text[start:stop])) {
			i = end
			continue
		}
		out = append(out, text[start:end])
		start, i = next, next
	}
	if start < len(text) {
		out = append(out, text[start:])
	}
	return out
}

// capital reports whether s starts with a capital letter, after an opening
// quote or bracket if it has one.
func capital(s string) bool {
	if s != "" && strings.IndexByte(`"'(`, s[0]) >= 0 {
		s = s[1:]
	}
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(r)
}

// abbreviated reports whether before, the text up to a full stop, ends
// with an abbreviation or an initial, whose stop ends no sentence.
func abbreviated(before string) bool {
	word := before[strings.LastIndexByte(before, ' ')+1:]
	word = strings.TrimLeft(word, `"'(`)
	if r, size := utf8.DecodeRuneInString(word); size == len(word) && unicode.IsUpper(r) {
		return true // an initial: "J. Smith"
	}
	return abbreviations[strings.ToLower(word)]
}

// abbreviations are the words, without their last stop, that a capital
// often follows in a sentence: "e.g. Sage", "Mr. Patel".
var abbreviations = map[string]bool{
	"e.g": true, "i.e": true, "mr": true, "mrs": true, "ms": true, "dr": true, "prof": true,
	"st": true, "no": true, "nos": true, "vs": true, "approx": true, "ref": true, "tel": true,
	"mob": true, "dept": true,
}
