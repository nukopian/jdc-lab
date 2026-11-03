package clean

import (
	"testing"
	"unicode/utf8"

	"github.com/nukopian/jdc-lab/internal/ads"
)

// textTexts are texts whose cleaning is easy to get wrong: spaces at
// either end and in runs, folds that hold spaces or make more than one
// character, symbols and controls dropped between words, full-width
// forms, and broken UTF-8.
var textTexts = []string{
	"", " ", "   ", "a", " a ", "a  b", "\ta\nb\r", "a\rb", "a\u0085b", "What You’ll Be Doing:",
	"✅ Benefits ✅", "a ✅ b", "a✅b", "a\u200bb", "a\u00a0b", "\u00a0a\u00a0", "rest—then", "—a—",
	"a — b", "Mon–Fri…", "…", "“quoted”", "it`s", "it´s", "30°C", "£30,000", "Ｆｕｌｌ\u3000ｔｉｍｅ",
	"^caret^", "\xff", "a\xffb", "a\xe2\x80", "\x00a\x7f", "\uf0b7 Word's bullet", "a\ufe0fb", "1\u20e3",
}

// Text's fast path cleans every text as slowText does: the texts above,
// every pair of ASCII characters, and the titles and HTML of 2,000 ads
// of the latest run, where there's one.
func TestFastTextChangesNothing(t *testing.T) {
	texts := append([]string(nil), textTexts...)
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
	}
	for _, s := range texts {
		if got, want := Text(s), slowText(s); got != want {
			t.Errorf("Text(%q) = %q, slowText gives %q", s, got, want)
		}
	}
	t.Logf("%d texts, %d ads", len(texts), n)
}

func FuzzText(f *testing.F) {
	for _, s := range textTexts {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if got, want := Text(s), slowText(s); got != want {
			t.Errorf("Text(%q) = %q, slowText gives %q", s, got, want)
		}
	})
}
