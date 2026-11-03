package headings

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// fold maps characters that vary between ads, but not in meaning, to one
// form: typographic quotes and dashes to ASCII, unusual spaces to a space,
// and bullets, emoji modifiers and invisible characters to nothing.
// Runes are written as numbers so that none is invisible in the source.
var fold = map[rune]string{
	// Quotes and primes.
	0x2018: "'", 0x2019: "'", 0x201A: "'", 0x201B: "'", 0x2032: "'",
	0x201C: `"`, 0x201D: `"`, 0x201E: `"`, 0x201F: `"`, 0x2033: `"`,
	0x00AB: `"`, 0x00BB: `"`,
	// Dashes, hyphens and the minus sign: they part words, as a hyphen
	// does in a key.
	0x2010: " ", 0x2011: " ", 0x2012: " ", 0x2013: " ", 0x2014: " ",
	0x2015: " ", 0x2212: " ",
	// Bullets, including Word's (0xF0B7, in its Symbol font).
	0x2022: "", 0x2023: "", 0x2043: "", 0x25AA: "", 0x25CF: "", 0x00B7: "",
	0x2981: "", 0xF0B7: "",
	// Ellipsis.
	0x2026: "...",
	// Spaces: no-break, en, em, thin, hair, narrow no-break and the like.
	0x00A0: " ", 0x2000: " ", 0x2001: " ", 0x2002: " ", 0x2003: " ",
	0x2004: " ", 0x2005: " ", 0x2006: " ", 0x2007: " ", 0x2008: " ",
	0x2009: " ", 0x200A: " ", 0x202F: " ", 0x205F: " ", 0x3000: " ",
	// Invisible: zero-width space and joiners, word joiner, byte order mark,
	// soft hyphen, and the variation selector that makes ✔ into ✔️.
	0x200B: "", 0x200C: "", 0x200D: "", 0x2060: "", 0xFEFF: "", 0x00AD: "",
	0xFE0F: "", 0xFE0E: "",
}

// Key returns the form of text to match on: folded (see fold), lower case,
// with emoji and other symbols dropped, punctuation other than ' & / made a
// space, and spaces collapsed. "What You’ll Be Doing:" and "what you'll be
// doing" have the same key. The text itself is never changed: keys are
// only for comparing.
func Key(text string) string {
	if !fastKey {
		return slowKey(text)
	}
	var b strings.Builder
	b.Grow(len(text))
	gap := false // a space is due before whatever is written next
	for i := 0; i < len(text); {
		if c := text[i]; c < utf8.RuneSelf {
			i++
			switch k := asciiKey[c]; k {
			case 0:
			case ' ':
				gap = true
			default:
				if gap && b.Len() > 0 {
					b.WriteByte(' ')
				}
				gap = false
				b.WriteByte(k)
			}
			continue
		}
		r, n := utf8.DecodeRuneInString(text[i:])
		i += n
		if s, ok := fold[r]; ok {
			switch s {
			case "":
			case " ":
				gap = true
			default:
				if gap && b.Len() > 0 {
					b.WriteByte(' ')
				}
				gap = false
				b.WriteString(s)
			}
			continue
		}
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if gap && b.Len() > 0 {
				b.WriteByte(' ')
			}
			gap = false
			b.WriteRune(unicode.ToLower(r))
		case unicode.Is(unicode.So, r) || unicode.Is(unicode.Sk, r) || unicode.Is(unicode.Co, r):
		default:
			gap = true
		}
	}
	return b.String()
}

// KeyWords returns how many words a key has, as len(strings.Fields(key))
// would, without making them: Key parts a key's words with one space.
func KeyWords(key string) int {
	if key == "" {
		return 0
	}
	return strings.Count(key, " ") + 1
}

// fastKey turns on Key's fast path, which reads ASCII from asciiKey and
// collapses spaces as it goes, rather than splitting the key into words
// and joining them. TestFastKeyChangesNothing compares the two.
var fastKey = true

// asciiKey is what Key writes for each ASCII character: itself, lower
// case, a space, or nothing (0). It's worked out from slowKey, so the two
// can't disagree.
var asciiKey = func() (t [utf8.RuneSelf]byte) {
	for c := range t {
		switch k := slowKey(" " + string(rune(c)) + "x"); k {
		case "x":
			t[c] = ' ' // a space or nothing: told apart below
			if slowKey("x"+string(rune(c))+"x") == "xx" {
				t[c] = 0
			}
		default:
			t[c] = k[0]
		}
	}
	return t
}()

// slowKey is Key without its fast path.
func slowKey(text string) string {
	var b strings.Builder
	for _, r := range text {
		if s, ok := fold[r]; ok {
			b.WriteString(s)
			continue
		}
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		case r == '\'' || r == '&' || r == '/':
			b.WriteRune(r)
		case unicode.Is(unicode.So, r) || unicode.Is(unicode.Sk, r) || unicode.Is(unicode.Co, r):
			// emoji, other symbols, and private-use characters: nothing
		default:
			b.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
