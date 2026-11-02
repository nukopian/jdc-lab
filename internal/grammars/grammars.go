// Package grammars holds the grammars for the facts ads state, pay,
// hours, work patterns, schedules and postcodes, written in Raku's
// grammar syntax: internal/facts reads their values with them, and
// internal/headings keeps lines holding them out of headings, so that
// both go by the same rules. It imports neither.
package grammars

import (
	"slices"
	"strings"
	"unicode"

	"github.com/nukopian/grammar"
	"github.com/nukopian/scalar"
)

// Pay is a figure of money: with a currency sign, or a period, or both,
// as one or a range, "up to" it, or not. A figure with neither isn't
// pay ("5 years"), though facts.Find takes one after a pay label; nor is
// a bonus, a figure followed by "bonus", with at most a word between that
// isn't "plus", "with" or the like: "£30,000 plus bonus" is pay.
var Pay = grammar.MustCompile(`
grammar Pay {
    token TOP    { [ $<upto>=[ :i up \s+ to ] \s+ ]? <cur>? \s? <min=amount>
                   [ <.sep> <cur>? \s? <max=amount> ]? <period>? <?{ pay($/) }> <!before <.bonus>> }
    # A bonus isn't pay: "£30 bonus for cover", "£1,500 joining bonus";
    # but pay "plus bonus" or "with bonus" is pay, with a bonus besides.
    token bonus  { \s+ [ :i bonus >>
                       | <!before :i [ plus | with | and | or | including | incl | inc ] >> > <[a..zA..Z-]>+ \s+ :i bonus >> ] }
    token cur    { '£' | '$' | '€' }
    token amount { <num> [ $<k>=[ :i k ] >> ]? }
    token num    { \d+ [ ',' \d ** 3 ]* [ '.' \d+ ]? }
    token sep    { \s* [ '-' | :i to ] \s* }
    # The dotted forms end in a point, not a letter, so >> isn't for them.
    token period { \s* :i [ 'p.a.' | 'p.h.' | 'p.w.'
                          | [ 'p/a' | pa | per \s+ annum | per \s+ year | a \s+ year | annually
                            | per \s+ hour | per \s+ hr | an \s+ hour | 'p/hr' | 'p/h' | '/hr' | '/hour' | phr | ph | hourly
                            | per \s+ day | a \s+ day | daily
                            | pw | 'p/w' | per \s+ week | a \s+ week | weekly
                            | pcm | per \s+ month | a \s+ month | monthly ] >> ] }
}`, pay)

// pay says whether a match of Pay so far is pay: it has a currency sign
// or a period. cur is captured twice in the rule, before each amount, so
// it's an array, empty when there's none; period is one match, or missing.
var pay = grammar.Assertion("pay", func(m *scalar.SV) bool {
	return len(m.At("cur").AV().Items()) > 0 || m.At("period").Defined()
})

// Hours is hours of work: "40hrs pw", "37.5 hours per week", "12 hours
// per day", "40 hours". A bare "h" is hours only with a period after it,
// as "24h support" isn't.
var Hours = grammar.MustCompile(`
grammar Hours {
    token TOP  { <!after <[\d.]>> <num> \s* :i [ [ hours | hrs | hr ] >> [ \s* <per> ]? | h >> \s* <per> ] <!{ implausible($/) }> }
    token num  { \d+ [ '.' \d+ ]? }
    token per  { <week> | <day> }
    token week { :i [ per \s+ week | pw | 'p/w' | a \s+ week | weekly ] >> }
    token day  { :i [ per \s+ day | a \s+ day | daily | per \s+ shift ] >> }
}`, implausible)

// implausible says whether a match of Hours is more hours than its period
// holds: over 24 a day, over 168 a week, or, with no period, over 80, a
// long week's, as "272 hours" of annual leave, "2026 Hours" after a date
// or "0600hrs" are.
var implausible = grammar.Assertion("implausible", func(m *scalar.SV) bool {
	n := m.At("num.~").Num()
	switch {
	case m.At("per.day").Defined():
		return n > 24
	case m.At("per.week").Defined():
		return n > 168
	}
	return n > 80
})

// patternRules are the work patterns and arrangements, each rule its
// canonical value first, then what ads write for it.
const patternRules = `
permanent, perm, permanent position, permanent role
temporary, temp, temporary role, temporary position
temp to perm, temp-to-perm, temporary to permanent
contract, fixed term, fixed-term, fixed term contract, ftc
full time, full-time
part time, part-time
hybrid, hybrid working
remote, fully remote, work from home, wfh
office based, office-based, site based, site-based
`

var patterns = func() *grammar.WordList {
	l := grammar.NewWordList("patterns", grammar.FoldCase(), grammar.LooseSpacing(), grammar.WholeWords())
	if err := l.AddRules(patternRules); err != nil {
		panic(err)
	}
	return l
}()

// Pattern is a work pattern or arrangement, its canonical value as the
// match's made: "Perm" is permanent. "Contract" as a field's label
// states nothing: before a colon, or "type", "length" or "duration", as
// in "Contract: Permanent" or "Contract type: Temp to Perm". Other words
// before a colon are values: "Part-Time: 25 hours per week".
var Pattern = grammar.MustCompile(`grammar Pattern {
    token TOP { <!before :i contract \s* [ ':' | :i [ type | length | duration ] >> ]> @patterns }
}`, patterns)

// Schedule is days and times of work: a range of days, with times or
// not, a range of times on its own, or a rota. A bare range of hours,
// "8-4", is a time only after a range of days.
//
// Its other rules are for telling a line that states a schedule, as
// internal/headings does, rather than reading one: clock, a time of day;
// points, a range of times with points ("8.30-3.30"); span, a bare range
// of hours, with a unit after it if it's a count ("2-3 days"); and shift,
// days or nights, with any figure before them. A time starts at a figure
// that isn't inside another, so "Friday7:00am" is a time and "745am" none.
var Schedule = grammar.MustCompile(`
grammar Schedule {
    token TOP   { <days> [ \s* ','? \s* <times> ]? | <times=clocks> | <rota> }
    token day   { << :i [ monday | tuesday | wednesday | thursday | friday | saturday | sunday
                        | mon | tues | tue | weds | wed | thurs | thur | thu | fri | sat | sun ] >> '.'? }
    token to    { \s* [ '-' | :i to | :i till? | :i until ] \s* }
    token days  { <first=day> <.to> <last=day> | << $<mf>=[ M \s* '-' \s* F ] >> }
    token times { <a=time> <.to> <b=time> }
    token time  { \d ** 1..2 [ <[:.]> \d ** 2 ]? [ \s? :i [ am | pm ] >> ]? }
    token clock { <!after \d> [ \d ** 1..2 [ <[:.]> \d ** 2 ]? \s? :i [ am | pm ] >> | \d ** 1..2 ':' \d ** 2 >> ] }
    token clocks { <a=clock> <.to> <b=time> }
    token rota  { <!after \d> \d+ \s* [ :i [ days? | nights? ] \s* ]? :i on [ \s* <[,/&-]> ]? \s*
                  \d+ \s* [ :i [ days? | nights? ] \s* ]? :i off >> }
    token points { <!after \d> \d ** 1..2 '.' \d ** 2 <.to> \d ** 1..2 '.' \d ** 2 >> }
    token hour  { $<h>=[ 2 <[0..3]> | <[01]>? \d ] [ <[:.]> \d ** 2 ]? }
    token span  { <!after \d> <a=hour> <.to> <b=hour> >>
                  [ \s* $<unit>=[ :i [ hours? | hrs? | days? | nights? | weeks? | months? | years? | yrs
                                    | sessions? | mins? | minutes ] >> ] | $<unit>='%' ]? }
    token shift { [ $<count>=[ \d \s* ] ]? << :i [ days | nights ] >> }
}`)

// Figure is a long figure, three digits or more, as in a reference
// ("Ref 225682916") or a salary: not a fact itself, but nothing a heading
// has either.
var Figure = grammar.MustCompile(`grammar Figure { token TOP { \d ** 3 \d* } }`)

// Postcode is a UK postcode, as a location: "NE8 1AQ".
var Postcode = grammar.MustCompile(`
grammar Postcode {
    token TOP { << <[A..Z]> ** 1..2 \d [ <[A..Z]> | \d ]? \s* \d <[A..Z]> ** 2 >> }
}`)

// MayHold reports whether line could hold a match of g, one of the fact
// grammars here, by what every match of it has, so that a line that can't
// is skipped without running g: most of an ad's lines are prose with no
// fact in them. It errs towards true, and true for a grammar it doesn't
// know. Each check is kept in step with its grammar, and a test checks it
// never says false for a line g finds a match in:
//
//   - Pay, Hours, Postcode and Figure: a digit, as \d is, any of Unicode's.
//   - Schedule: a digit, a day's first three letters, in any case, or
//     "M" and "F" with a "-", for "M - F".
//   - Pattern: the first word of one of its phrases, in any case.
func MayHold(g *grammar.Grammar, line string) bool {
	switch g {
	case Pay, Hours, Postcode, Figure:
		return hasDigit(line)
	case Schedule:
		if hasDigit(line) || strings.Contains(line, "M") && strings.Contains(line, "F") && strings.Contains(line, "-") {
			return true
		}
		return containsAny(folded(line), days)
	case Pattern:
		return containsAny(folded(line), patternStems)
	}
	return true
}

func hasDigit(s string) bool { return strings.ContainsFunc(s, unicode.IsDigit) }

// days are the first three letters of the days' names, which every name
// Schedule's day rule knows starts with.
var days = []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

// patternStems are the first words of patternRules' phrases, each the
// letters before its first space or hyphen: "temp" for "temp-to-perm".
// LooseSpacing can change what's between a phrase's words, but not its
// first word.
var patternStems = func() []string {
	var stems []string
	for _, phrase := range strings.FieldsFunc(patternRules, func(r rune) bool { return r == ',' || r == '\n' }) {
		stem, _, _ := strings.Cut(strings.TrimSpace(phrase), " ")
		stem, _, _ = strings.Cut(stem, "-")
		if stem != "" && !slices.Contains(stems, stem) {
			stems = append(stems, stem)
		}
	}
	return stems
}()

func containsAny(s string, subs []string) bool {
	return slices.ContainsFunc(subs, func(sub string) bool { return strings.Contains(s, sub) })
}

// folded returns s with each letter that grammar matches to an ASCII
// letter made that letter in lower case. A word list's FoldCase lowers a
// letter ("İ" is "i"), and :i folds it as Unicode's simple case folding
// does (the long s is "s"); the Kelvin sign is "k" either way. It errs
// towards true: the long s matches :i's days, but not the word list.
func folded(s string) string {
	return strings.Map(func(r rune) rune {
		if l := unicode.ToLower(r); l < unicode.MaxASCII {
			return l
		}
		for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
			if f < unicode.MaxASCII {
				return unicode.ToLower(f)
			}
		}
		return r
	}, s)
}
