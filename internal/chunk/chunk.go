// Package chunk turns an ad's sections into chunks, as an embedding job
// would ask for them: each with its kind and heading, its text, what to
// do with it (embed it, keep it as metadata, or leave it out), about how
// many tokens it is, and what placed it in its section.
//
// The sections that say what the job is are embedded: the summary, the
// role, responsibilities, requirements and desirable. Details are metadata;
// benefits, company, apply and legal are left out. An unmapped section,
// whose heading has no known kind, is placed by a guess (see guess): by
// its lines' wording, or as the summary; or, when it's boilerplate, the
// advertiser's or after the apply text, its text goes to the boilerplate
// chunk, under its heading, and its fields to a details chunk.
//
// Text found as boilerplate leaves the embedded chunks for
// one chunk of its own, at the end, left out but there to see, each line
// with the section it came from: the verdict catches advertisers'
// footers, but also job lines common to many ads ("high levels of
// accuracy and attention to detail"), and this is where they show. Other
// chunks keep theirs, as part of what they say: an advertiser's
// standard perks are still a list of benefits.
//
// The chunks come with what says what the job is first: the summary, the
// role, responsibilities and requirements, in that order; then the rest
// in the ad's order, the chunks of lines placed by their wording and of
// lines of facts after its own; and the boilerplate last.
package chunk

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/nukopian/jdc-lab/internal/facts"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"github.com/nukopian/jdc-lab/internal/wording"
)

// Boilerplate is the kind of the chunk of an ad's boilerplate.
const Boilerplate sections.Kind = "boilerplate"

// What to do with a chunk.
const (
	Embed    = "embed"
	Metadata = "metadata"
	LeaveOut = "leave out"
)

// A Field is a label and its value: "salary", "£30,000".
type Field struct {
	Label string `json:"label" yaml:"label"`
	Value string `json:"value" yaml:"value"`
}

// A Chunk is one section of an ad, ready for what's done with it.
type Chunk struct {
	Kind    sections.Kind `json:"kind" yaml:"kind"`
	Heading string        `json:"heading,omitempty" yaml:"heading,omitempty"`
	// PlacedBy says what gave the chunk its kind.
	PlacedBy string `json:"placed_by" yaml:"placed_by"`
	// Use is Embed, Metadata or LeaveOut.
	Use string `json:"use" yaml:"use"`
	// Text is the chunk's heading and text, a line each, a list's items
	// as "- " entries.
	Text   string  `json:"text" yaml:"text"`
	Fields []Field `json:"fields,omitempty" yaml:"fields,omitempty"`
	// Facts are the facts read in a details chunk's lines (see
	// internal/facts): pay, hours, work pattern, schedule, location.
	Facts []facts.Fact `json:"facts,omitempty" yaml:"facts,omitempty"`
	// Unmapped is how many of the boilerplate chunk's tokens are text of
	// unmapped sections: what's left to map.
	Unmapped int `json:"unmapped_tokens,omitempty" yaml:"unmapped_tokens,omitempty"`
	// Tokens is about how many tokens Text is: four for every three
	// words, as models' tokenizers make of English.
	Tokens int `json:"tokens" yaml:"tokens"`
}

// Of returns an ad's sections as chunks; boilerplate says whether a text
// part was found to be boilerplate.
func Of(ss []sections.Section, boilerplate func(string) bool) []Chunk {
	var cs []Chunk
	var gathered []string                // the boilerplate of the embedded chunks, and the unmapped text
	unmapped := 0                        // the unmapped text's tokens
	var found Chunk                      // the lines of facts taken from the chunks
	worded := map[sections.Kind]*Chunk{} // lines placed by their wording, by kind
	var wordedKinds []sections.Kind      // in the order first found
	takeWording := func(s sections.Section, p headings.Part) bool {
		// Only a summary's lines, and a role's list items: a role's own
		// prose says what the job is, and stays.
		if p.Role != headings.Text || !(s.Kind == sections.Summary || s.Kind == sections.Role && p.Item) {
			return false
		}
		kind, phrase, ok := wording.Of(p.Text)
		if !ok {
			return false
		}
		w := worded[kind]
		if w == nil {
			w = &Chunk{Kind: kind, Use: use(kind)}
			worded[kind] = w
			wordedKinds = append(wordedKinds, kind)
		}
		w.Text += "\n" + line(p)
		if q := fmt.Sprintf("%q", phrase); !strings.Contains(w.PlacedBy, q) {
			w.PlacedBy += ", " + q
		}
		return true
	}
	takeFacts := func(p headings.Part) bool {
		if p.Role != headings.Text {
			return false
		}
		fs := facts.Find(p.Text)
		if !factLine(p.Text, fs) {
			return false
		}
		found.Text += "\n" + line(p)
		found.Facts = append(found.Facts, placeless(fs)...)
		return true
	}
	for i, s := range ss {
		placed := ""
		if s.Kind == sections.Unmapped {
			if kind, why := guess(ss, i, boilerplate); kind != "" {
				s.Kind, placed = kind, why
			} else {
				placed = why
			}
		}
		if s.Kind == sections.Unmapped {
			// Unmapped text the guess can't place goes to the boilerplate,
			// under its heading, with why it's left out; its fields stay, as
			// details.
			d := Chunk{Kind: sections.Details, Use: Metadata, Heading: s.Heading,
				PlacedBy: fmt.Sprintf("the fields under the unmapped heading %q", s.Heading)}
			var lines []string
			for _, p := range s.Parts {
				if p.Role == headings.Field {
					d.Fields = append(d.Fields, Field{Label: p.Key, Value: p.Value})
					lines = append(lines, line(p))
					continue
				}
				if takeFacts(p) {
					continue
				}
				gathered = append(gathered, fmt.Sprintf("unmapped %q (%s): %s", s.Heading, placed, p.Text))
				unmapped += Tokens(p.Text)
			}
			if len(d.Fields) > 0 {
				d.Text = strings.Join(lines, "\n")
				d.Tokens = Tokens(d.Text)
				cs = append(cs, d)
			}
			continue
		}
		c := Chunk{Kind: s.Kind, Heading: s.Heading, PlacedBy: cmp.Or(placed, placedBy(s)), Use: use(s.Kind)}
		var lines []string
		if s.Heading != "" {
			lines = append(lines, s.Heading)
		}
		for _, p := range s.Parts {
			if c.Use == Embed && (takeFacts(p) || takeWording(s, p)) {
				continue
			}
			if c.Use == Embed && p.Role == headings.Text && boilerplate(p.Text) {
				gathered = append(gathered, string(s.Kind)+": "+p.Text) // the gathered lines are a list
				continue
			}
			if p.Role == headings.Field {
				c.Fields = append(c.Fields, Field{Label: p.Key, Value: p.Value})
			}
			lines = append(lines, line(p))
		}
		c.Text = strings.Join(lines, "\n")
		c.Tokens = Tokens(c.Text)
		if c.Kind == sections.Details {
			for _, l := range lines {
				c.Facts = append(c.Facts, placeless(facts.Find(l))...)
			}
		}
		cs = append(cs, c)
	}
	for _, k := range wordedKinds {
		w := worded[k]
		w.Text = strings.TrimPrefix(w.Text, "\n")
		w.Tokens = Tokens(w.Text)
		w.PlacedBy = "lines placed by their wording (openings.txt): " + strings.TrimPrefix(w.PlacedBy, ", ")
		cs = append(cs, *w)
	}
	if found.Text != "" {
		found.Kind, found.Use = sections.Details, Metadata
		found.Text = strings.TrimPrefix(found.Text, "\n")
		found.Tokens = Tokens(found.Text)
		found.PlacedBy = "lines of facts, read by the facts grammars (pay, hours, work pattern, schedule, location), from the chunks above"
		cs = append(cs, found)
	}
	// What says what the job is comes first, in that order; the rest
	// keep the ad's order, and the boilerplate comes last.
	slices.SortStableFunc(cs, func(a, b Chunk) int { return first(a.Kind) - first(b.Kind) })
	if len(gathered) > 0 {
		text := strings.Join(gathered, "\n")
		cs = append(cs, Chunk{
			Kind: Boilerplate, Use: LeaveOut, Text: text, Tokens: Tokens(text), Unmapped: unmapped,
			PlacedBy: "found as boilerplate in the chunks above (more often outside the sections about the job, in the footers that end ads, or in ads for five groups of titles or more), or under a heading of no known kind the guess left out",
		})
	}
	return cs
}

// factLine reports whether a line, with the facts fs found in it, is a
// line of facts, to be details. One with a postcode in it, of at most 30
// words, is. Otherwise it must start with a fact, in its first three
// words, be at most 25 words, and not be a sentence or a question; and if
// its only facts are work patterns, words such as "contract" and
// "temporary" that sentences use too, they must be all of it, or set
// apart by "|", "," or "/" from at most two other words. "50,000pa +
// excellent bonus" and "Bromsgrove | Full-time | Permanent" are lines of
// facts; "This is a permanent role working with…" and "Contract
// administration" aren't.
func factLine(text string, fs []facts.Fact) bool {
	text = strings.TrimSpace(text)
	words := strings.Fields(text)
	if len(fs) == 0 || strings.HasSuffix(text, "?") {
		return false
	}
	for _, f := range fs {
		if f.Kind == "location" && len(words) <= 30 {
			return true
		}
	}
	sentence := strings.Contains(text, ". ") || strings.HasSuffix(text, ".") && len(words) > 10
	runes := []rune(text)
	before := len(strings.Fields(string(runes[:min(fs[0].From, len(runes))])))
	if sentence || before >= 3 || len(words) > 25 {
		return false
	}
	if slices.ContainsFunc(fs, func(f facts.Fact) bool { return f.Kind != "pattern" }) {
		return true
	}
	rest := text
	for _, f := range fs {
		rest = strings.Replace(rest, f.Text, " ", 1)
	}
	others := 0
	for _, w := range strings.Fields(rest) {
		if strings.ContainsFunc(w, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) {
			others++
		}
	}
	return others == 0 || others <= 2 && strings.ContainsAny(rest, "|,/")
}

// placeless returns facts without where they were in their line, which a
// chunk's reader has no use for.
func placeless(fs []facts.Fact) []facts.Fact {
	for i := range fs {
		fs[i].From = 0
	}
	return fs
}

// firsts are the kinds of chunk that come first, in this order.
var firsts = []sections.Kind{sections.Summary, sections.Role, sections.Responsibilities, sections.Requirements}

// first returns a chunk's place among the firsts, or after them.
func first(k sections.Kind) int {
	if i := slices.Index(firsts, k); i >= 0 {
		return i
	}
	return len(firsts)
}

// line returns a part's text as a line of a chunk: an item of a list as
// an entry, "- " first.
func line(p headings.Part) string {
	if p.Item {
		return "- " + p.Text
	}
	return p.Text
}

// Tokens returns about how many tokens text is: four for every three
// words, a word being something with a letter or figure in it, so that a
// list's "- " isn't one.
func Tokens(text string) int {
	n := 0
	for _, w := range strings.Fields(text) {
		if strings.ContainsFunc(w, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) {
			n++
		}
	}
	return (n*4 + 1) / 3 // to the nearest
}

// Whole returns about how many tokens an ad's sections are, all of them:
// the whole ad, as an embedding of it would take it.
func Whole(ss []sections.Section) int {
	n := 0
	for _, s := range ss {
		n += Tokens(s.Heading)
		for _, p := range s.Parts {
			n += Tokens(p.Text)
		}
	}
	return n
}

// use says what's done with a section of kind.
func use(kind sections.Kind) string {
	switch kind {
	case sections.Summary, sections.Role, sections.Responsibilities, sections.Requirements, sections.Desirable:
		return Embed
	case sections.Details:
		return Metadata
	}
	return LeaveOut
}

// guess places ss[i], an unmapped section, by what it says, rather than
// leaving all of it out. The evidence is weighed in this order. When at
// least half its lines open as one kind's do (see internal/wording), and
// fewer than half are boilerplate, it's that kind, and why says so:
// "guessed requirements: 5 of its 7 lines open as requirements do"; the
// first kind by name wins a tie. It's left out, kind "", when at least
// half its lines are boilerplate, its heading speaks for the advertiser
// ("we", "our"), or it comes after the ad's apply or legal text, or when
// it has no lines of text, only fields or facts; and why says which. A
// label over a short value is details. Otherwise it's the summary.
func guess(ss []sections.Section, i int, boilerplate func(string) bool) (kind sections.Kind, why string) {
	s := ss[i]
	counts := map[sections.Kind]int{}
	phrases := map[sections.Kind][]string{}
	lines, repeated := 0, 0
	for _, p := range s.Parts {
		if p.Role != headings.Text || factLine(p.Text, facts.Find(p.Text)) {
			continue
		}
		lines++
		if boilerplate(p.Text) {
			repeated++
		}
		if k, phrase, ok := wording.Of(p.Text); ok {
			counts[k]++
			if q := fmt.Sprintf("%q", phrase); !slices.Contains(phrases[k], q) && len(phrases[k]) < 3 {
				phrases[k] = append(phrases[k], q)
			}
		}
	}
	var best sections.Kind
	for _, k := range slices.Sorted(maps.Keys(counts)) {
		if counts[k] > counts[best] {
			best = k
		}
	}
	if lines > 0 && 2*counts[best] >= lines && 2*repeated < lines {
		return best, fmt.Sprintf("guessed %s: %d of its %d lines open as %s do (%s)",
			best, counts[best], lines, best, strings.Join(phrases[best], ", "))
	}
	switch {
	case lines > 0 && 2*repeated >= lines:
		return "", "mostly boilerplate"
	case speaksForAdvertiser(s.Heading):
		return "", "the advertiser's"
	case slices.ContainsFunc(ss[:i], func(t sections.Section) bool { return t.Kind == sections.Apply || t.Kind == sections.Legal }):
		return "", "after applying"
	case lines == 0:
		return "", "no text" // only fields or facts, which have places of their own
	case labelled(s):
		return sections.Details, "guessed details: a label, with a short value under it"
	}
	// Read by eye, most of what's left is about the job: its pitch,
	// duties and requirements in words openings.txt doesn't have.
	return sections.Summary, "guessed summary: nothing says it isn't about the job"
}

// advertiserWords are the words of a heading that speaks for the
// advertiser: "We're the AA", "Our Commitment", "Why choose us?";
// advertiserOpenings, how one opens: "Who are Acme?", "Join Acme".
var (
	advertiserWords    = map[string]bool{"we": true, "we're": true, "we've": true, "our": true, "us": true}
	advertiserOpenings = []string{"who are ", "join "}
)

func speaksForAdvertiser(heading string) bool {
	key := headings.Key(heading)
	return slices.ContainsFunc(strings.Fields(key), func(w string) bool { return advertiserWords[w] }) ||
		slices.ContainsFunc(advertiserOpenings, func(o string) bool { return strings.HasPrefix(key, o) })
}

// MaxLabelledWords is the most words under a heading ending in a colon
// for the section to be a label and its value: "IR35 Status:" over
// "Inside IR35".
const MaxLabelledWords = 15

// labelled reports whether s is a label and its value, split by the
// markup: a heading ending in a colon, over a line or two of at most
// MaxLabelledWords words in all.
func labelled(s sections.Section) bool {
	if !strings.HasSuffix(strings.TrimSpace(s.Heading), ":") || len(s.Parts) == 0 || len(s.Parts) > 2 {
		return false
	}
	words := 0
	for _, p := range s.Parts {
		words += len(strings.Fields(p.Text))
	}
	return words <= MaxLabelledWords
}

// placedBy says what gave a section its kind.
func placedBy(s sections.Section) string {
	switch {
	case s.Heading == "":
		return "nothing before it: summary by default"
	case s.ByName:
		if _, ok := sections.Named(headings.Key(s.Heading)); ok {
			return fmt.Sprintf("the line %q: exactly a section's name", s.Heading)
		}
		return fmt.Sprintf("the line %q, before a list: %s", s.Heading, sections.Explain(headings.Key(s.Heading)))
	}
	return fmt.Sprintf("the heading %q: %s", s.Heading, sections.Explain(headings.Key(s.Heading)))
}
