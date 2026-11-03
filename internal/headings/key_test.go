package headings

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
)

// keyTexts are texts whose keys are easy to get wrong: spaces at either
// end and in runs, folded characters that make a space, nothing or more
// than one character, symbols dropped between words, letters whose lower
// case is longer, and broken UTF-8.
var keyTexts = []string{
	"", " ", "   ", "a", " a ", "a  b", "\ta\nb\r", "What You’ll Be Doing:",
	"✅ Benefits ✅", "a ✅ b", "a✅b", "a\u200bb", "a\u00a0b", "\u00a0a\u00a0",
	"Mon–Fri…", "…", "a…b", "“quoted”", "Full-time / Part-time", "Skills & Experience",
	"£30,000 – £35,000", "^caret` and `tick^", "Ⱥ", "\u212a", "İstanbul", "ǅ",
	"\xff", "a\xffb", "a\xe2\x80", "\x00a\x7f", "\uf0b7 Word's bullet", "a\ufe0fb",
}

// Key's fast path gives every text the key slowKey does: the texts
// above, every pair of ASCII characters, and the titles, blocks and HTML
// of 2,000 ads of the latest run, where there's one.
func TestFastKeyChangesNothing(t *testing.T) {
	texts := append([]string(nil), keyTexts...)
	for a := range utf8.RuneSelf {
		for b := range utf8.RuneSelf {
			texts = append(texts, string([]byte{byte(a), byte(b)}))
		}
	}
	n := 0
	for a, err := range ads.File("../../data/runs/latest/ads.jsonl") {
		if err != nil || n == 2000 {
			break
		}
		n++
		texts = append(texts, a.Title, a.HTML)
		for _, b := range blocks.Split(a.HTML) {
			texts = append(texts, b.Text)
			for _, it := range b.Items {
				texts = append(texts, it.Text)
			}
		}
	}
	for _, s := range texts {
		if got, want := Key(s), slowKey(s); got != want {
			t.Errorf("Key(%q) = %q, slowKey gives %q", s, got, want)
		}
	}
	t.Logf("%d texts, %d ads", len(texts), n)
}

func FuzzKey(f *testing.F) {
	for _, s := range keyTexts {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if got, want := Key(s), slowKey(s); got != want {
			t.Errorf("Key(%q) = %q, slowKey gives %q", s, got, want)
		}
	})
}

// KeyWords counts a key's words as strings.Fields does.
func TestKeyWords(t *testing.T) {
	for _, s := range append(keyTexts, "Excellent ICT skills and a full UK driving licence") {
		k := Key(s)
		if got, want := KeyWords(k), len(strings.Fields(k)); got != want {
			t.Errorf("KeyWords(%q) = %d, want %d", k, got, want)
		}
	}
}
