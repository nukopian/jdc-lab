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
	"strings"
	"unicode"

	"github.com/nukopian/jdc-lab/internal/blocks"
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
		"reports to",
		"sector", "industry", "department", "employer", "client", "consultant",
		"contact", "email", "keywords",
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
	var parts []Part
	for _, b := range bs {
		if b.Kind == blocks.List {
			for _, it := range b.Items {
				p := item(it.Text)
				p.Item = true
				parts = append(parts, p)
			}
			continue
		}
		parts = append(parts, block(b.Text, b.Cues, repeats)...)
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
// first colon or bar.
var labelled = regexp.MustCompile(`^([^:|]{1,40})[:|]\s*(\S.*)$`)

// field reads text as a field, if its label is one of fieldLabels and it
// has a value.
func field(text string) (Part, bool) {
	m := labelled.FindStringSubmatch(text)
	if m == nil {
		return Part{}, false
	}
	key := Key(m[1])
	if !fieldLabels[key] {
		return Part{}, false
	}
	return Part{Role: Field, Text: text, Key: key, Value: strings.TrimSpace(m[2])}, true
}

// money matches what headings don't have, money and long figures:
// "£30,000 per annum" in bold is pay, and "Ref 225682916" a reference.
var money = regexp.MustCompile(`[£$€]|\d{3,}`)

// times matches a time of day, which a heading doesn't have either: "- 2pm
// to 10pm" or "Monday to Friday, 8:00am - 5:00pm" in bold is when the work
// is. A figure with a point is a time only with am or pm, or in a range, as
// "8.30-3.30": "37.50 hours" is a figure.
var times = regexp.MustCompile(`(?i)\b\d{1,2}([:.]\d{2})? ?[ap]m\b|\b\d{1,2}:\d{2}\b|\b\d{1,2}\.\d{2} *(-|–|—|to) *\d{1,2}\.\d{2}\b`)

// days matches a range of days: "Monday to Friday", "Mon–Thu", "M-F".
var days = regexp.MustCompile(`(?i)\b(mon|tue|wed|thu|fri|sat|sun)[a-z]*\.? *(-|–|—|to) *(mon|tue|wed|thu|fri|sat|sun)[a-z]*\b|\bM *[-–] *F\b`)

// hours matches a bare range of hours, "8-5" or "7.30-4", which is a time
// only beside a range of days: alone, "Ages 5-11" is no time. "2-3 days"
// or "10-12 hr shifts" is a count, whatever's beside it.
var hours = regexp.MustCompile(`(?i)\b([01]?\d|2[0-3])([:.]\d{2})? *(-|–|—|to) *([01]?\d|2[0-3])([:.]\d{2})?\b( *(hours?|hrs?|days?|nights?|weeks?|months?|years?|yrs|sessions?|mins?|minutes|%))?`)

// hasTimes reports whether text holds a time of day: one of times, or a
// bare range of hours beside a range of days, as in "Monday - Friday 8-5"
// or "M-F 8-4".
func hasTimes(text string) bool {
	if times.MatchString(text) {
		return true
	}
	if !days.MatchString(text) {
		return false
	}
	for _, m := range hours.FindAllStringSubmatch(text, -1) {
		if m[6] == "" { // not followed by a unit, so not a count
			return true
		}
	}
	return false
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
// money, long figures or times, and not ending as a sentence does.
func headingShaped(text, key string, most int) bool {
	if !strings.ContainsFunc(key, unicode.IsLetter) || onlyConjunctions(key) {
		return false
	}
	if len(strings.Fields(key)) > most || money.MatchString(text) || hasTimes(text) {
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
