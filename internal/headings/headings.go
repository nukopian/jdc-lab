// Package headings decides what each block of a job ad is: a heading, a
// field such as "Salary: £30,000", a repeat of the ad's title or location,
// or text.
//
// It works from what internal/blocks found, the text and the markup's cues,
// with rules, and nothing else: which section a heading starts is for
// internal/sections. Each rule errs towards text: a heading missed is
// counted later as text, but a sentence taken for a heading would start a
// section that isn't there.
package headings

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/nukopian/grammar"
	"github.com/nukopian/scalar"

	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/clean"
	"github.com/nukopian/jdc-lab/internal/grammars"
)

// A Role is what a Part is.
type Role int

const (
	// Text is anything that isn't one of the others.
	Text Role = iota
	// Heading starts a section of the ad.
	Heading
	// Field is a known label and its value: "Hours: 40 per week".
	Field
	// Repeat restates the ad's title or location, as many ads open with.
	Repeat
)

var roleNames = [...]string{Text: "text", Heading: "heading", Field: "field", Repeat: "repeat"}

// String returns the role's name: "heading".
func (r Role) String() string {
	if r < 0 || int(r) >= len(roleNames) {
		return "unknown"
	}
	return roleNames[r]
}

// MarshalText writes the role as its name, so JSON shows "heading", not 1.
func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// A Part is a block of an ad, or an item of a list, with its role. A block
// that starts with a bold heading, as "<b>The role</b> We're hiring" does,
// is two parts: the heading, then the text.
type Part struct {
	Role Role   `json:"role"`
	Text string `json:"text"` // as written
	// Key is the heading's or the field label's key (see Key); empty for
	// other roles.
	Key string `json:"key,omitempty"`
	// Value is a field's value: "40 per week".
	Value string `json:"value,omitempty"`
	// Item is true for an item of a list.
	Item bool `json:"item,omitempty"`
	// Added is the English label cleaning added at the start of Text,
	// for a symbol that stood for it (see blocks.Cues): not the ad's own.
	Added string `json:"added,omitempty"`
	// AfterTitle is true for what followed the ad's title in its block,
	// the title split off as a repeat: "Oldham" of "Head Chef - Oldham".
	AfterTitle bool `json:"after_title,omitempty"`
	// Whole is the block it was split from, when only a space ended the
	// title: "Head Chef Manager" of "Manager".
	Whole string `json:"whole,omitempty"`
}

// MaxWords is the most words a heading has. Longer bold text is a bold
// sentence, not a heading.
const MaxWords = 10

// MaxIntroWords is the most words a heading has when it's in bold, or a
// heading tag, and ends with a colon: it introduces what follows, as in
// "As a Mobile Mechanic, you'll get benefits that go the extra mile:".
const MaxIntroWords = 20

// fieldLabels are the labels of fields, by key. A label not here is no
// field: "Financial reporting: Oversee the accounts" is a list item with a
// bold lead, not data.
var fieldLabels = map[string]bool{}

func init() {
	for _, label := range []string{
		"salary", "pay", "rate", "rate of pay", "pay rate", "hourly rate", "wage", "wages",
		"salary range", "remuneration", "ir35",
		"location", "work location", "job location", "based", "base", "area",
		"hours", "working hours", "hours of work", "hours per week", "working pattern",
		"work pattern", "shift", "shifts", "shift pattern", "working days", "days", "schedule",
		"working arrangement", "working arrangements", "hybrid working",
		"contract", "contract type", "job type", "employment type", "employment", "type", "vacancy type",
		"contract length", "length of contract", "duration", "term",
		"start", "start date", "starting", "closing date", "deadline", "interview date",
		"job title", "job role", "position", "role", "reference", "ref", "job reference",
		"reports to", "reporting to", "responsible to", "reference number",
		"sector", "industry", "department", "employer", "client", "consultant",
		"contact", "email", "keywords", "key words",
	} {
		fieldLabels[label] = true
	}
}

// Read returns the parts of an ad's blocks. title and locations are the
// ad's own, for spotting a block that only repeats them.
func Read(bs []blocks.Block, title string, locations ...string) []Part {
	repeats := map[string]bool{}
	for _, s := range append([]string{title}, locations...) {
		if k := Key(s); k != "" {
			repeats[k] = true
		}
	}
	title = clean.Text(title) // as the blocks' text is
	var parts []Part
	for _, b := range bs {
		if b.Kind == blocks.List {
			for _, it := range b.Items {
				p := item(it.Text)
				p.Item = true
				parts = append(parts, added(p, it.Cues))
			}
			continue
		}
		ps, ok := afterTitle(b.Text, b.Cues, title, repeats)
		if !ok {
			ps = block(b.Text, b.Cues, repeats)
		}
		ps[0] = added(ps[0], b.Cues)
		parts = append(parts, ps...)
	}
	return joinFields(parts)
}

// MaxValueWords is the most words a field's value on a line of its own
// has. A longer line after a label is a paragraph about it, not its value.
const MaxValueWords = 15

// joinFields joins a field's label, written as a heading, and its value on
// the next line, as in "<b>Hours:</b><br>40 per week", into one Field.
func joinFields(parts []Part) []Part {
	var out []Part
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if p.Role == Heading && fieldLabels[p.Key] && i+1 < len(parts) {
			next := parts[i+1]
			if next.Role == Text && !next.Item && len(strings.Fields(next.Text)) <= MaxValueWords {
				out = append(out, Part{Role: Field, Text: p.Text + " " + next.Text, Key: p.Key, Value: next.Text})
				i++
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

// block reads a Paragraph or a Line.
func block(text string, cues blocks.Cues, repeats map[string]bool) []Part {
	key := Key(text)
	if repeats[key] {
		return []Part{{Role: Repeat, Text: text}}
	}
	if p, ok := field(text); ok {
		return []Part{p}
	}
	most := MaxWords
	if (cues.Heading || cues.Bold) && strings.HasSuffix(strings.TrimRightFunc(text, unicode.IsSpace), ":") {
		most = MaxIntroWords
	}
	if headingShaped(text, key, most) && (cues.Heading || cues.Bold || endsAsLabel(text)) {
		return []Part{{Role: Heading, Text: text, Key: key}}
	}
	// A bold lead that's a heading of its own, run into its text.
	if lead := cues.BoldLead; lead != "" {
		rest := strings.TrimLeftFunc(strings.TrimPrefix(text, lead), unicode.IsSpace)
		if k := Key(lead); rest != "" && headingShaped(lead, k, MaxWords) {
			return []Part{{Role: Heading, Text: lead, Key: k}, {Role: Text, Text: rest}}
		}
	}
	return []Part{{Role: Text, Text: text}}
}

// afterTitle splits the ad's title off a block that opens with it, as a
// repeat, and reads the rest as it would be alone, when the rest is:
// a field; a heading, the block having been one with the title too
// ("General Foreman Responsibilities:"); or, after a mark that ends the
// title, a short line that isn't a sentence ("General Foreman
// (Permanent)", "General Foreman - 12 Month Contract"), as text. A block
// that goes on as a sentence, or with the title made plural or longer
// ("General Foreman / Site Manager"), isn't split; nor is a label that
// cleaning added. With only a space after the title, the rest's parts
// keep the whole block, as a heading of no known kind there is the title
// made longer ("Head Chef Manager"), which sections reads whole.
func afterTitle(text string, cues blocks.Cues, title string, repeats map[string]bool) ([]Part, bool) {
	if title == "" || cues.Added != "" || len(text) <= len(title) || !strings.EqualFold(text[:len(title)], title) {
		return nil, false
	}
	head, after := text[:len(title)], text[len(title):]
	if next, _ := utf8.DecodeRuneInString(after); unicode.IsLetter(next) || unicode.IsDigit(next) {
		return nil, false // "General Foremans"
	}
	rest := strings.TrimLeft(after, " -|:,")
	if rest == "" {
		return nil, false
	}
	first, _ := utf8.DecodeRuneInString(rest)
	marked := first == '(' || strings.ContainsAny(after[:len(after)-len(rest)], "-|:,")
	// The rest has a bold lead only if it's left once the title is off.
	restCues := cues
	restCues.BoldLead = strings.TrimLeft(strings.TrimPrefix(cues.BoldLead, head), " -|:,")
	if !strings.HasPrefix(rest, restCues.BoldLead) {
		restCues.BoldLead = ""
	}
	ps := block(rest, restCues, repeats)
	if ps[0].Role == Heading && len(ps) == 1 && marked && (first == '(' || unicode.IsDigit(first)) {
		ps = []Part{{Role: Text, Text: rest}} // what the title is, in brackets or by figures
	}
	switch ps[0].Role {
	case Heading:
		whole := block(text, cues, repeats)
		if len(whole) != 1 || whole[0].Role != Heading || !unicode.IsLetter(first) {
			return nil, false
		}
	case Text:
		sentence := strings.Contains(rest, ". ") || strings.HasSuffix(rest, ".") || unicode.IsLower(first)
		if !marked || sentence || len(strings.Fields(rest)) > MaxWords {
			return nil, false
		}
	}
	for i := range ps {
		ps[i].AfterTitle = true
		if !marked {
			ps[i].Whole = text
		}
	}
	return append([]Part{{Role: Repeat, Text: head}}, ps...), true
}

// added records on p the label cleaning added, if p starts with it.
func added(p Part, cues blocks.Cues) Part {
	if cues.Added != "" && strings.HasPrefix(p.Text, cues.Added) {
		p.Added = cues.Added
	}
	return p
}

// item reads a list item: a field, or text. Items aren't taken as headings:
// a bold item is far more often a point made strongly than a heading.
func item(text string) Part {
	if p, ok := field(text); ok {
		return p
	}
	return Part{Role: Text, Text: text}
}

// IsLabel reports whether key (see Key) is a field's label, such as
// "salary" or "working hours".
func IsLabel(key string) bool {
	return fieldLabels[key]
}

// labelled finds "Label: value" or "Label | value", the label up to the
// first colon or bar; field also takes "Label - value".
var labelled = regexp.MustCompile(`^([^:|]{1,40})[:|]\s*(\S.*)$`)

// field reads text as a field, if its label is one of fieldLabels and it
// has a value.
func field(text string) (Part, bool) {
	if m := labelled.FindStringSubmatch(text); m != nil {
		if key := Key(m[1]); fieldLabels[key] {
			return Part{Role: Field, Text: text, Key: key, Value: strings.TrimSpace(m[2])}, true
		}
	}
	// "Working Hours - Monday to Friday": a known label, then a dash with
	// a space each side, as cleaning leaves every dash.
	if label, value, ok := strings.Cut(text, " - "); ok && len(label) <= 40 {
		if key := Key(label); fieldLabels[key] && strings.TrimSpace(value) != "" {
			return Part{Role: Field, Text: text, Key: key, Value: strings.TrimSpace(value)}, true
		}
	}
	return Part{}, false
}

// hasMoney reports whether text holds what headings don't have, money
// or a long figure: "£30,000 per annum" in bold is pay, and "Ref
// 225682916" a reference.
func hasMoney(text string) bool {
	return grammars.Pay.Find(text, grammar.Rule("cur")).Defined() || grammars.Figure.Find(text).Defined()
}

// hasPattern reports whether text states a pattern of work, which is a
// fact, as a time is: a rota, or days or nights beside a range of days, as
// in "Monday - Friday (Days)". "Night Shift" alone names a shift, and heads
// its hours, as a pattern ending in a colon heads what follows: "You will
// work a 4 on 4 off shift pattern:". "2-3 days" is a count, not the day
// shift.
func hasPattern(text string) bool {
	if endsAsLabel(text) {
		return false
	}
	if grammars.Schedule.Find(text, grammar.Rule("rota")).Defined() {
		return true
	}
	if !grammars.Schedule.Find(text, grammar.Rule("days")).Defined() {
		return false
	}
	for _, m := range grammars.Schedule.FindAll(text, grammar.Rule("shift")).Items() {
		if !m.At("count").Defined() {
			return true
		}
	}
	return false
}

// hasTimes reports whether text holds a time of day, which a heading
// doesn't have either: "- 2pm to 10pm" or "Monday to Friday, 8:00am -
// 5:00pm" in bold is when the work is. A figure with a point is a time
// only with am or pm, or in a range, as "8.30-3.30": "37.50 hours" is a
// figure. A bare range of hours, "8-5" or "7.30-4", is a time only beside
// a range of days, as in "Monday - Friday 8-5" or "M-F 8-4": alone,
// "Ages 5-11" is no time. "2-3 days" or "10-12 hr shifts" is a count,
// whatever's beside it, but a range that runs back round the clock, as
// "7-3 Days" does, is hours. Unlike a pattern, a time ending in a colon
// is still no heading.
func hasTimes(text string) bool {
	if grammars.Schedule.Find(text, grammar.Rule("clock")).Defined() ||
		grammars.Schedule.Find(text, grammar.Rule("points")).Defined() {
		return true
	}
	if !grammars.Schedule.Find(text, grammar.Rule("days")).Defined() {
		return false
	}
	for _, m := range grammars.Schedule.FindAll(text, grammar.Rule("span")).Items() {
		from, _ := strconv.Atoi(m.At("a.h.~").Str())
		to, _ := strconv.Atoi(m.At("b.h.~").Str())
		if !m.At("unit").Defined() || from >= to { // no unit, or round the clock: not a count
			return true
		}
	}
	return false
}

// removed are what an export puts in place of contact details it takes
// out: a heading has none, as it has no phone number or web address.
var removed = []string{"(phone number removed)", "(url removed)"}

// hasRemoved reports whether text holds one of removed, as "Call (phone
// number removed)" does.
func hasRemoved(text string) bool {
	text = strings.ToLower(text)
	return slices.ContainsFunc(removed, func(r string) bool { return strings.Contains(text, r) })
}

// onlyFacts reports whether text is nothing but work patterns,
// schedules, hours and pay, and what joins them: "Permanent | Full-time"
// or "Monday - Friday" in bold is a line of facts, not a heading. Ending
// in a colon, it heads what follows, as "Monday - Friday:" heads its
// hours, and a field's label ("Hybrid Working") heads its value. Pay
// counts with a currency or a period, as internal/facts reads it.
func onlyFacts(text string) bool {
	if endsAsLabel(text) || fieldLabels[Key(text)] {
		return false
	}
	left := []rune(text)
	cut := func(g *grammar.Grammar, keep func(m *scalar.SV) bool) {
		for _, m := range g.FindAll(text, grammar.Outermost()).Items() {
			if keep != nil && !keep(m) {
				continue
			}
			from, to := int(m.At("from").Int()), int(m.At("to").Int())
			for i := from; i < to && i < len(left); i++ {
				left[i] = ' '
			}
		}
	}
	cut(grammars.Pattern, nil)
	cut(grammars.Schedule, nil)
	cut(grammars.Hours, nil)
	cut(grammars.Pay, func(m *scalar.SV) bool {
		return len(m.At("cur").AV().Items()) > 0 || m.At("period").Defined()
	})
	if string(left) == text {
		return false // no facts at all
	}
	return !strings.ContainsFunc(string(left), func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) })
}

// conjunctions join what comes before them to what comes after: "OR" alone
// in bold, between two lists of requirements, is no heading.
var conjunctions = map[string]bool{"and": true, "or": true, "either": true, "nor": true, "but": true, "plus": true}

// onlyConjunctions reports whether every word of key is a conjunction, as
// in "or", "either" and "and/or".
func onlyConjunctions(key string) bool {
	words := strings.FieldsFunc(key, func(r rune) bool { return r == ' ' || r == '/' || r == '&' })
	for _, w := range words {
		if !conjunctions[w] {
			return false
		}
	}
	return len(words) > 0
}

// headingShaped reports whether text, with its key, could be a heading:
// some letters, a word that isn't a conjunction, at most most words, no
// money, long figures, times or patterns of work, and not ending as a
// sentence does.
func headingShaped(text, key string, most int) bool {
	if !strings.ContainsFunc(key, unicode.IsLetter) || onlyConjunctions(key) {
		return false
	}
	if len(strings.Fields(key)) > most || hasMoney(text) || hasTimes(text) || hasPattern(text) || onlyFacts(text) || hasRemoved(text) {
		return false
	}
	t := strings.TrimRightFunc(text, func(r rune) bool { return unicode.IsSpace(r) || unicode.Is(unicode.So, r) })
	return !strings.HasSuffix(t, ".") || strings.HasSuffix(t, "...")
}

// endsAsLabel reports whether text ends as a heading without a cue can:
// with a colon or semicolon, its text to follow.
func endsAsLabel(text string) bool {
	t := strings.TrimRightFunc(text, unicode.IsSpace)
	return strings.HasSuffix(t, ":") || strings.HasSuffix(t, ";")
}
