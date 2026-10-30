// Package view shows ads in their chunked form, for reading by eye, as the
// chunker makes them (internal/chunk): each chunk with its kind and
// heading, what's done with it, its size and what placed it, its lines,
// and the boilerplate gathered at the end. It also shows any of an ad's
// text that's in no section at all, or out of the ad's order.
//
// The ads shown are chosen by a fixed hash of their IDs, so that every
// run of the pipeline shows the same ones, and a diff of two runs' views
// shows what a change did to them.
package view

import (
	"cmp"
	"fmt"
	"hash/fnv"
	"io"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/net/html"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/chunk"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// A Candidate is an ad that could be shown, with what makes it worth
// showing on purpose.
type Candidate struct {
	ID string
	// NoHeading is true for an ad with no heading at all.
	NoHeading bool
	// Unmapped is true for an ad with a section of no known kind.
	Unmapped bool
	// MostlyBoilerplate is true for an ad over half of whose text is
	// boilerplate.
	MostlyBoilerplate bool
	// Lost is true for an ad with text in no section, or out of its
	// order (see Missing).
	Lost bool
}

// Pick returns the indexes of n candidates to show, in a fixed order: by
// the hash of their IDs. All but one for each kind worth showing on
// purpose are the first by that order, always, so a change to the
// pipeline never moves them; then, for each kind (no heading, an unmapped
// section, mostly boilerplate, text lost), the first with it, if none
// shown has it already, as a random draw would mostly show the easy
// cases; and the rest are the next by the order.
func Pick(cs []Candidate, n int) []int {
	order := make([]int, len(cs))
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int {
		return cmp.Or(cmp.Compare(rank(cs[a].ID), rank(cs[b].ID)), cmp.Compare(cs[a].ID, cs[b].ID))
	})
	wanted := []func(Candidate) bool{
		func(c Candidate) bool { return c.NoHeading },
		func(c Candidate) bool { return c.Unmapped },
		func(c Candidate) bool { return c.MostlyBoilerplate },
		func(c Candidate) bool { return c.Lost },
	}
	var picked []int
	taken := map[int]bool{}
	take := func(i int) {
		if len(picked) < n && !taken[i] {
			picked = append(picked, i)
			taken[i] = true
		}
	}
	for _, i := range order[:min(len(order), max(0, n-len(wanted)))] {
		take(i)
	}
	for _, is := range wanted {
		if slices.ContainsFunc(picked, func(i int) bool { return is(cs[i]) }) {
			continue
		}
		for _, i := range order {
			if is(cs[i]) && !taken[i] {
				take(i)
				break
			}
		}
	}
	for _, i := range order {
		take(i)
	}
	slices.SortFunc(picked, func(a, b int) int { return slices.Index(order, a) - slices.Index(order, b) })
	return picked
}

// rank is an ID's place in the fixed order: its FNV-1a hash.
func rank(id string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(id))
	return h.Sum64()
}

// Render writes an ad's chunks to w as Markdown, as the chunker made them
// (internal/chunk): each with its kind, heading, use, tokens and what
// placed it, then its lines, a list's items as a list. lost is what
// Missing found, shown at the end when there is some.
func Render(w io.Writer, ad ads.Ad, c chunk.Ad, lost Lost) error {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s · %s\n\n", escape(ad.Title), ad.ID)
	if ls := slices.Compact(ad.Locations()); len(ls) > 0 {
		fmt.Fprintf(&b, "%s\n\n", escape(strings.Join(ls, " · ")))
	}
	share := func(n int) int {
		if c.Tokens == 0 {
			return 0
		}
		return 100 * n / c.Tokens
	}
	fmt.Fprintf(&b, "*about %d tokens: %d embedded (%d%%), %d unclassified (%d%%)*\n\n",
		c.Tokens, c.Embedded, share(c.Embedded), c.Unclassified, share(c.Unclassified))
	for _, ch := range c.Chunks {
		b.WriteString("### " + string(ch.Kind))
		if ch.Heading != "" {
			fmt.Fprintf(&b, " · “%s”", escape(ch.Heading))
		}
		fmt.Fprintf(&b, " · %s · %d tokens\n\n*placed by %s*\n\n", ch.Use, ch.Tokens, escape(ch.PlacedBy))
		lines := strings.Split(ch.Text, "\n")
		if ch.Heading != "" && len(lines) > 0 && lines[0] == ch.Heading {
			lines = lines[1:] // the heading is the chunk's own heading, above
		}
		if (len(lines) == 0 || lines[0] == "" && len(lines) == 1) && len(ch.Fields) == 0 {
			b.WriteString("*(nothing under it)*\n\n")
		}
		inList := false
		for _, line := range lines {
			// The boilerplate chunk's lines are a list of what was
			// gathered; elsewhere a list's items start "- ".
			item := ch.Kind == chunk.Boilerplate || strings.HasPrefix(line, "- ")
			if item {
				b.WriteString("- " + escape(strings.TrimPrefix(line, "- ")) + "\n")
				inList = true
				continue
			}
			if inList {
				b.WriteString("\n") // a blank line ends the list
				inList = false
			}
			b.WriteString(escape(line) + "\n\n")
		}
		if inList {
			b.WriteString("\n")
		}
		if len(ch.Fields) > 0 {
			b.WriteString("*fields:*\n\n")
			for _, f := range ch.Fields {
				fmt.Fprintf(&b, "- %s: %s\n", escape(f.Label), escape(f.Value))
			}
			b.WriteString("\n")
		}
	}
	switch {
	case lost.Moved:
		fmt.Fprintf(&b, "### out of the ad's order\n\n%s\n\n", escape(lost.Context))
	case lost.Context != "":
		fmt.Fprintf(&b, "### not in any section\n\n%s\n\n", escape(lost.Context))
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// escaper keeps an ad's own text from being read as Markdown.
var escaper = strings.NewReplacer(
	`\`, `\\`, "`", "\\`", "*", `\*`, "_", `\_`, "[", `\[`, "]", `\]`,
	"<", `\<`, ">", `\>`, "#", `\#`, "|", `\|`, "~", `\~`,
)

// escape returns s with what Markdown would read as markup escaped,
// including a leading "-", "+" or "1." that would start a list.
func escape(s string) string {
	s = escaper.Replace(s)
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		return `\` + s
	}
	if i := strings.IndexFunc(s, func(r rune) bool { return !unicode.IsDigit(r) }); i > 0 && s[i] == '.' {
		return s[:i] + `\` + s[i:]
	}
	return s
}

// Lost is what Missing found: nothing, when Context is empty.
type Lost struct {
	// Context is the ad's text around the first place its sections
	// part from it.
	Context string
	// Moved is true when the sections have all the ad's text, but not
	// in its order.
	Moved bool
}

// Missing compares an ad's sections with what its HTML shows, and finds
// where they first part: text that's in no section, or that's in one but
// out of the ad's order. Only letters and digits are compared, as the
// pipeline tidies spaces and leaves out what isn't visible; so an
// advertiser's "Operative</strong>to" is no loss, kept as written.
func Missing(src string, ss []sections.Section) Lost {
	shown, at := visibleText(src)
	var kept []rune
	add := func(s string) { kept = appendLetters(kept, s) }
	for _, s := range ss {
		add(s.Heading)
		for _, p := range s.Parts {
			add(strings.TrimPrefix(p.Text, p.Added)) // what cleaning added isn't the ad's
		}
	}
	wanted := appendLetters(nil, shown)
	if len(wanted) == 0 {
		return Lost{}
	}
	i := 0
	for i < len(wanted) && i < len(kept) && wanted[i] == kept[i] {
		i++
	}
	if i == len(wanted) { // all of the ad is there, and maybe more: none of it lost
		return Lost{}
	}
	// at maps each letter of wanted to its place in shown.
	start := max(0, at[i]-40)
	end := min(len(shown), at[i]+80)
	context := "…" + strings.Join(strings.Fields(shown[start:end]), " ") + "…"
	slices.Sort(wanted)
	slices.Sort(kept)
	return Lost{Context: context, Moved: slices.Equal(wanted, kept)}
}

// visibleText returns the text of an HTML fragment, leaving out scripts
// and styles, with the byte offset in it of each letter or digit, in
// order.
func visibleText(src string) (string, []int) {
	var b strings.Builder
	var at []int
	skip := 0
	z := html.NewTokenizer(strings.NewReader(src))
	for {
		switch z.Next() {
		case html.ErrorToken:
			return b.String(), at
		case html.StartTagToken:
			if name, _ := z.TagName(); string(name) == "script" || string(name) == "style" {
				skip++
			}
			b.WriteByte(' ')
		case html.EndTagToken:
			if name, _ := z.TagName(); (string(name) == "script" || string(name) == "style") && skip > 0 {
				skip--
			}
			b.WriteByte(' ')
		case html.SelfClosingTagToken:
			b.WriteByte(' ')
		case html.TextToken:
			if skip > 0 {
				continue
			}
			for _, r := range string(z.Text()) { // Text has decoded the entities
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					at = append(at, b.Len())
				}
				b.WriteRune(r)
			}
		}
	}
}

// appendLetters appends the letters and digits of s to rs, in lower case.
func appendLetters(rs []rune, s string) []rune {
	for _, r := range s {
		if r >= 0xFF01 && r <= 0xFF5E { // full-width forms, as cleaning folds them
			r -= 0xFEE0
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			rs = append(rs, unicode.ToLower(r))
		}
	}
	return rs
}
