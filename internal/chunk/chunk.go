// Package chunk turns an ad's sections into chunks, as an embedding job
// would ask for them: each with its kind and heading, its text, what to
// do with it (embed it, keep it as metadata, or leave it out), about how
// many tokens it is, and what placed it in its section.
//
// The sections that say what the job is are embedded: the summary, the
// role, responsibilities, requirements and desirable; and the company's,
// the employer's own words, with the summary, where its standard text is
// found as boilerplate. Details are metadata; benefits, apply and legal are
// left out. An unmapped section,
// whose heading has no known kind, is placed by a guess (see guess): by
// its lines' wording, or as the summary; or, when it's boilerplate, the
// advertiser's or after the apply text, its text goes to the boilerplate
// chunk, under its heading, and its fields to a details chunk.
//
// Fields, and facts in prose, stay in their sections' chunks, but the details
// list them too, as metadata, without their text: every field, and every fact
// read in a section's lines, but those about the employer or applying.
//
// Text found as boilerplate leaves the embedded chunks for
// one chunk of its own, at the end, left out but there to see, each line
// with the section it came from: the verdict catches advertisers'
// footers, but also job lines common to many ads ("high levels of
// accuracy and attention to detail"), and this is where they show. Other
// chunks keep theirs, as part of what they say: an advertiser's
// standard perks are still a list of benefits.
//
// An ad's chunks of each kind are one chunk, the desirable joining the
// requirements, each section's heading a line over its own lines (see
// together).
//
// The chunks come with what says what the job is first: the title (see
// AdOf), the summary, the role, responsibilities and requirements, in that
// order; then the rest
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

// Title is the kind of the chunk of an ad's title: its title field.
const Title sections.Kind = "title"

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

// AdOf returns an ad's chunks, with its totals: its title, the ad's title
// field, as a chunk of its own, first and embedded, as titles embed well
// alone; then its sections' chunks (see Of). The title is counted in the
// ad's tokens. A title the ad writes as a field ("Job Title: …") leaves
// the embedded chunks, so as not to be embedded twice, and is listed in
// the details, as fields are.
func AdOf(id, title string, ss []sections.Section, boilerplate func(string) bool) Ad {
	var cs []Chunk
	if title != "" {
		cs = append(cs, Chunk{Kind: Title, Use: Embed, Text: title, Tokens: Tokens(title), PlacedBy: "the ad's title field"})
	}
	cs = append(cs, Of(ss, boilerplate)...)
	return Totalled(Ad{ID: id, Title: title, Chunks: cs}, Whole(ss)+Tokens(title))
}

// Of returns an ad's sections as chunks; boilerplate says whether a text
// part was found to be boilerplate.
func Of(ss []sections.Section, boilerplate func(string) bool) []Chunk {
	var cs []Chunk
	var gathered []string                // the boilerplate of the embedded chunks, and the unmapped text
	unmapped := 0                        // the unmapped text's tokens
	var found Chunk                      // the lines of facts taken from the chunks, and every field and fact
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
		kind := s.Kind
		if kind == sections.Company {
			// The employer's own words go with the summary: not about the
			// job, but close enough to it to be worth embedding, and the
			// standard half of them is found as boilerplate there.
			kind, placed = sections.Summary, cmp.Or(placed, placedBy(s))+", the employer's own words, with the summary"
		}
		c := Chunk{Kind: kind, Heading: s.Heading, PlacedBy: cmp.Or(placed, placedBy(s)), Use: use(kind)}
		var lines []string
		if s.Heading != "" {
			lines = append(lines, s.Heading)
		}
		for _, p := range s.Parts {
			// The company's lines of facts are the employer's, and stay.
			if c.Use == Embed && ((!aboutTheEmployer(s.Kind) && takeFacts(p)) || takeWording(s, p)) {
				continue
			}
			if c.Use == Embed && p.Role == headings.Text && boilerplate(p.Text) {
				gathered = append(gathered, string(s.Kind)+": "+p.Text) // the gathered lines are a list
				continue
			}
			if p.Role == headings.Field {
				f := Field{Label: p.Key, Value: p.Value}
				// A field is metadata wherever it's written, so the details
				// list it too, without its text, which stays where it is;
				// but the title, which the title chunk embeds, isn't
				// embedded again (see AdOf).
				if s.Kind != sections.Details {
					found.Fields = append(found.Fields, f)
				}
				if c.Use == Embed && titleLabels[p.Key] {
					continue
				}
				c.Fields = append(c.Fields, f)
			}
			// So is a fact: a working pattern in the role, or salary in
			// the benefits, is listed in the details as well. Not from what
			// the employer says of itself or of applying, though: there
			// "temporary, contract and permanent roles" are an agency's, and
			// "within 48 hours" a reply's.
			if s.Kind != sections.Details && !aboutTheEmployer(s.Kind) {
				found.Facts = append(found.Facts, placeless(facts.Find(p.Text))...)
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
	found.Facts = once(found.Facts, cs)
	if found.Text != "" || len(found.Fields) > 0 || len(found.Facts) > 0 {
		found.Kind, found.Use = sections.Details, Metadata
		found.Text = strings.TrimPrefix(found.Text, "\n")
		found.Tokens = Tokens(found.Text)
		found.PlacedBy = "lines of facts, read by the facts grammars (pay, hours, work pattern, schedule, location), and the fields and facts of the chunks above, which keep them"
		cs = append(cs, found)
	}
	// What says what the job is comes first, in that order; the rest
	// keep the ad's order, and the boilerplate comes last.
	cs = together(cs)
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

// together returns cs with an ad's chunks of each kind made one, where
// the first of them stands: two summaries, the ad's own details and those
// gathered from its other sections, the requirements and the desirable
// (which join the requirements, so that "Essential" and "Desirable" are
// its sub-lists). Each chunk's heading is a line over its own lines,
// "Desirable" over desirable lines with none, what placed each is said in
// turn, and each fact is listed once.
func together(cs []Chunk) []Chunk {
	group := func(k sections.Kind) sections.Kind {
		if k == sections.Desirable {
			return sections.Requirements
		}
		return k
	}
	parts := map[sections.Kind][]Chunk{}
	var out []Chunk
	for _, c := range cs {
		g := group(c.Kind)
		if len(parts[g]) == 0 {
			out = append(out, Chunk{Kind: g}) // where the first of them stands
		}
		parts[g] = append(parts[g], c)
	}
	for i, c := range out {
		ps := parts[c.Kind]
		if len(ps) == 1 {
			out[i] = ps[0]
			continue
		}
		// The desirable come after the requirements, those placed by their
		// wording too.
		slices.SortStableFunc(ps, func(a, b Chunk) int {
			return cmp.Compare(boolInt(a.Kind == sections.Desirable), boolInt(b.Kind == sections.Desirable))
		})
		r := Chunk{Kind: c.Kind, Use: ps[0].Use}
		var texts, placed []string
		for _, p := range ps {
			text := p.Text
			if p.Heading == "" && p.Kind == sections.Desirable {
				text = "Desirable\n" + text
			}
			if text != "" {
				texts = append(texts, text)
			}
			placed = append(placed, p.PlacedBy)
			r.Fields = append(r.Fields, p.Fields...)
			r.Facts = append(r.Facts, p.Facts...)
			r.Unmapped += p.Unmapped
		}
		r.Text = strings.Join(texts, "\n")
		r.Tokens = Tokens(r.Text)
		r.PlacedBy = strings.Join(placed, "; ")
		r.Facts = once(r.Facts, nil)
		out[i] = r
	}
	return out
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// once returns fs with each fact listed once, by its kind and values, and
// none the details chunks in cs list already: an ad states its pay, or that
// the job is full time, in more than one place, and writes it differently
// ("Full Time", "full-time"), but the values are the same.
func once(fs []facts.Fact, cs []Chunk) []facts.Fact {
	key := func(f facts.Fact) string {
		ks := slices.Sorted(maps.Keys(f.Values))
		var b strings.Builder
		b.WriteString(f.Kind)
		for _, k := range ks {
			b.WriteString("\x00" + k + "=" + f.Values[k])
		}
		return b.String()
	}
	seen := map[string]bool{}
	for _, c := range cs {
		if c.Kind == sections.Details {
			for _, f := range c.Facts {
				seen[key(f)] = true
			}
		}
	}
	var out []facts.Fact
	for _, f := range fs {
		if k := key(f); !seen[k] {
			seen[k] = true
			out = append(out, f)
		}
	}
	return out
}

// titleLabels are the labels of fields that give the job's title.
var titleLabels = map[string]bool{"job title": true, "job role": true, "position": true, "role": true}

// aboutTheEmployer reports whether a kind of section is about the employer
// or applying, not the job: its facts are seldom the job's.
func aboutTheEmployer(k sections.Kind) bool {
	return k == sections.Company || k == sections.Apply || k == sections.Legal
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
