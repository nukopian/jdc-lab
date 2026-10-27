package chunk

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// An Ad is an ad's chunks, with its totals.
type Ad struct {
	ID    string `json:"id" yaml:"id"`
	Title string `json:"title" yaml:"title"`
	// Tokens is about how many tokens the whole ad is, and Embedded how
	// many of them its embedded chunks are.
	Tokens   int     `json:"tokens" yaml:"tokens"`
	Embedded int     `json:"embedded_tokens" yaml:"embedded_tokens"`
	Chunks   []Chunk `json:"chunks" yaml:"chunks"`
}

// WriteYAML writes the ad to w as a YAML document, for people to read:
// each chunk's text a block of its lines, and every other string as
// plainly as YAML allows. It writes its own YAML, as the usual library
// escapes emoji beyond the first 65,536 characters, which ads are full of.
func (a Ad) WriteYAML(w io.Writer) error {
	var b strings.Builder
	fmt.Fprintf(&b, "id: %s\ntitle: %s\ntokens: %d\nembedded_tokens: %d\nchunks:\n",
		scalar(a.ID), scalar(a.Title), a.Tokens, a.Embedded)
	for _, c := range a.Chunks {
		fmt.Fprintf(&b, "  - kind: %s\n", scalar(string(c.Kind)))
		if c.Heading != "" {
			fmt.Fprintf(&b, "    heading: %s\n", scalar(c.Heading))
		}
		fmt.Fprintf(&b, "    use: %s\n    tokens: %d\n    placed_by: %s\n", scalar(c.Use), c.Tokens, scalar(c.PlacedBy))
		fmt.Fprintf(&b, "    text: %s\n", block(c.Text, "      "))
		if len(c.Fields) > 0 {
			b.WriteString("    fields:\n")
			for _, f := range c.Fields {
				fmt.Fprintf(&b, "      - label: %s\n        value: %s\n", scalar(f.Label), scalar(f.Value))
			}
		}
		if len(c.Boilerplate) > 0 {
			b.WriteString("    boilerplate:\n")
			for _, s := range c.Boilerplate {
				fmt.Fprintf(&b, "      - %s\n", scalar(s))
			}
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// plain matches a string YAML reads as the same string when unquoted:
// words, figures and everyday punctuation, starting with a letter, and
// with nothing YAML takes as markup, such as ": " or " #".
var plain = regexp.MustCompile(`^\p{L}[\p{L}\p{N} ,.'()/&+£$€%-]*[\p{L}\p{N}.)%]$|^\p{L}$`)

// special are the plain words YAML reads as something else.
var special = map[string]bool{"yes": true, "no": true, "true": true, "false": true, "on": true, "off": true, "null": true, "y": true, "n": true}

// scalar returns s as YAML: plain if it can be, in single quotes if not,
// and in double quotes, with escapes, if it holds a character YAML
// allows only so.
func scalar(s string) string {
	switch {
	case strings.ContainsFunc(s, needsEscape):
		return strconv.Quote(s) // Go's escapes are all YAML's too
	case plain.MatchString(s) && !special[strings.ToLower(s)]:
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// block returns text as YAML, a literal block of its lines at indent
// when it has more than one, and as scalar does otherwise.
func block(text, indent string) string {
	inBlock := func(r rune) bool { return r != '\n' && needsEscape(r) }
	if !strings.Contains(text, "\n") || strings.ContainsFunc(text, inBlock) || strings.HasSuffix(text, "\n") {
		return scalar(text)
	}
	head := "|-"
	if strings.HasPrefix(text, " ") {
		head = "|" + strconv.Itoa(len(indent)-4) + "-" // the block's indent, as the line can't show it
	}
	var b strings.Builder
	b.WriteString(head)
	for _, line := range strings.Split(text, "\n") {
		b.WriteString("\n")
		if line != "" {
			b.WriteString(indent + line)
		}
	}
	return b.String()
}

// needsEscape reports whether YAML allows r only escaped in double
// quotes: a control character other than a tab, or a line break, which
// is fine only in a block.
func needsEscape(r rune) bool {
	return r != '\t' && (unicode.IsControl(r) || r == 0x2028 || r == 0x2029 || r == 0xFEFF)
}
