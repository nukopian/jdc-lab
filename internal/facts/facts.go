// Package facts reads the facts ads state in lines of their own, with the
// lab's grammar package: pay, hours, work pattern, schedule and location,
// each with its values. "50,000pa + excellent bonus" is pay of 50,000 a
// year, "Mon-Fri 07:00-15:30" a schedule, "NE8 1AQ" a location.
//
// Each fact has a grammar of its own, in internal/grammars, written in
// Raku's grammar syntax, which reads much as the text it matches; Find
// runs them all over a line. A fact found means the line holds one, not that the line is only
// facts: whether it is, the chunker decides.
package facts

import (
	"cmp"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/nukopian/grammar"
	"github.com/nukopian/scalar"

	"github.com/nukopian/jdc-lab/internal/grammars"
	"github.com/nukopian/jdc-lab/internal/headings"
)

// A Fact is a fact found in a line: its kind, the text that states it,
// and its values, such as pay's min, max, currency and period.
type Fact struct {
	Kind   string            `json:"kind" yaml:"kind"`
	Text   string            `json:"text" yaml:"text"`
	Values map[string]string `json:"values" yaml:"values"`
	// From is where in the line the text starts, in characters.
	From int `json:"-" yaml:"-"`
}

// Find returns the facts in line, in the order they come in it.
func Find(line string) []Fact {
	var fs []Fact
	for _, m := range all(grammars.Pay, line) {
		if f, ok := payOf(m); ok {
			fs = append(fs, f)
		}
	}
	if !slices.ContainsFunc(fs, func(f Fact) bool { return f.Kind == "pay" }) {
		if f, ok := labelledPay(line); ok {
			fs = append(fs, f)
		}
	}
	for _, m := range all(grammars.Hours, line) {
		n := number(m.At("num.~").Str())
		v := map[string]string{"hours": n} // with no period, just hours
		switch {
		case m.At("per.week").Defined():
			v = map[string]string{"per_week": n}
		case m.At("per.day").Defined():
			v = map[string]string{"per_day": n}
		}
		fs = append(fs, fact("hours", m, v))
	}
	for _, m := range all(grammars.Pattern, line) {
		// The word list's canonical value is the match's made.
		fs = append(fs, fact("pattern", m, map[string]string{"value": m.At("made").Str()}))
	}
	for _, m := range all(grammars.Schedule, line) {
		fs = append(fs, fact("schedule", m, scheduleOf(m)))
	}
	for _, m := range all(grammars.Postcode, line) {
		fs = append(fs, fact("location", m, map[string]string{"postcode": m.At("~").Str()}))
	}
	slices.SortStableFunc(fs, func(a, b Fact) int { return cmp.Compare(a.From, b.From) })
	return fs
}

// all returns the matches of g in line, none inside another.
func all(g *grammar.Grammar, line string) []*scalar.SV {
	return g.FindAll(line, grammar.Outermost()).Items()
}

// fact returns a fact of kind, from the match m, with values.
func fact(kind string, m *scalar.SV, values map[string]string) Fact {
	return Fact{Kind: kind, Text: strings.TrimSpace(m.At("~").Str()), Values: values, From: int(m.At("from").Int())}
}

// payOf reads a pay match, if it's pay: it needs a currency or a period.
func payOf(m *scalar.SV) (Fact, bool) {
	// cur is captured twice in the rule, before each amount, so it's an
	// array, empty when there's none; period is one match, or missing.
	curs := m.At("cur").AV().Items()
	cur, period := len(curs) > 0, m.At("period").Defined()
	if !cur && !period {
		return Fact{}, false
	}
	v := map[string]string{"min": amount(m.At("min"))}
	if m.At("max").Defined() {
		v["max"] = amount(m.At("max"))
		// "£25-28k": the k of the top figure is the bottom's too.
		if m.At("max.k").Defined() && !m.At("min.k").Defined() {
			n, _ := strconv.ParseFloat(v["min"], 64)
			v["min"] = strconv.FormatFloat(n*1000, 'f', -1, 64)
		}
	}
	if m.At("upto").Defined() && !m.At("max").Defined() {
		v["max"], v["min"] = v["min"], "" // "up to £30k" is a most, not a least
		delete(v, "min")
	}
	if cur {
		v["currency"] = map[string]string{"£": "GBP", "$": "USD", "€": "EUR"}[curs[0].At("~").Str()]
	}
	if period {
		v["period"] = periodOf(m.At("period.~").Str())
	}
	return fact("pay", m, v), true
}

// amount returns an amount's figure, "45k" as 45000, commas gone.
func amount(m *scalar.SV) string {
	n, _ := strconv.ParseFloat(strings.ReplaceAll(m.At("num.~").Str(), ",", ""), 64)
	if m.At("k").Defined() {
		n *= 1000
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

// number returns a figure as written, as a plain number.
func number(s string) string {
	n, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
	if err != nil {
		return s
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

// periodOf names a pay period: year, hour, day, week or month.
func periodOf(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.Contains(s, "hour"), strings.HasSuffix(s, "hr"), s == "ph", s == "p/h", s == "p.h.":
		return "hour"
	case strings.Contains(s, "day") || strings.Contains(s, "daily"):
		return "day"
	case strings.Contains(s, "week"), s == "pw", s == "p/w", s == "p.w.":
		return "week"
	case strings.Contains(s, "month"), s == "pcm":
		return "month"
	}
	return "year"
}

// payLabels are the labels after which a bare figure is pay.
var payLabels = map[string]bool{
	"pay": true, "pay rate": true, "rate": true, "rate of pay": true, "hourly rate": true,
	"salary": true, "wage": true, "wages": true, "day rate": true,
}

// figure finds a figure in text.
var figure = regexp.MustCompile(`\d+(?:,\d{3})*(?:\.\d+)?`)

// labelledPay reads "Pay Rate - 13.10": a pay label, then " - " or ":",
// then a figure with no currency or period.
func labelledPay(line string) (Fact, bool) {
	label, rest, ok := strings.Cut(line, ":")
	if !ok {
		label, rest, ok = strings.Cut(line, " - ")
	}
	if !ok || !payLabels[headings.Key(label)] {
		return Fact{}, false
	}
	loc := figure.FindStringIndex(rest)
	if loc == nil {
		return Fact{}, false
	}
	text := rest[loc[0]:loc[1]]
	return Fact{Kind: "pay", Text: text, Values: map[string]string{"min": number(text)}, From: utf8.RuneCountInString(line[:len(line)-len(rest)+loc[0]])}, true
}

// scheduleOf reads a schedule match's days, times and rota.
func scheduleOf(m *scalar.SV) map[string]string {
	v := map[string]string{}
	trim := func(s string) string { return strings.Join(strings.Fields(s), "") }
	switch {
	case m.At("days.mf").Defined():
		v["days"] = "M-F"
	case m.At("days").Defined():
		v["days"] = strings.TrimSpace(m.At("days.first.~").Str()) + "-" + strings.TrimSpace(m.At("days.last.~").Str())
	}
	if m.At("times").Defined() {
		v["times"] = trim(m.At("times.a.~").Str()) + "-" + trim(m.At("times.b.~").Str())
	}
	if m.At("rota").Defined() {
		v["rota"] = strings.Join(strings.Fields(m.At("rota.~").Str()), " ")
	}
	return v
}
