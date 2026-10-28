// Package grammars holds the grammars for the facts ads state, pay,
// hours, work patterns, schedules and postcodes, written in Raku's
// grammar syntax: internal/facts reads their values with them, and
// internal/headings keeps lines holding them out of headings, so that
// both go by the same rules. It imports neither.
package grammars

import "github.com/nukopian/grammar"

// Pay is a figure of money: with a currency sign, or a period, or both,
// as one or a range, "up to" it, or not. A figure with neither isn't
// pay ("5 years"), though facts.Find takes one after a pay label.
var Pay = grammar.MustCompile(`
grammar Pay {
    token TOP    { [ $<upto>=[ :i up \s+ to ] \s+ ]? <cur>? \s? <min=amount>
                   [ <.sep> <cur>? \s? <max=amount> ]? <period>? }
    token cur    { '£' | '$' | '€' }
    token amount { <num> [ $<k>=[ :i k ] >> ]? }
    token num    { \d+ [ ',' \d ** 3 ]* [ '.' \d+ ]? }
    token sep    { \s* [ '-' | :i to ] \s* }
    token period { \s* :i [ 'p.a.' | 'p/a' | pa | per \s+ annum | per \s+ year | a \s+ year | annually
                          | per \s+ hour | per \s+ hr | an \s+ hour | 'p/hr' | 'p/h' | '/hr' | '/hour' | phr | ph | hourly
                          | per \s+ day | a \s+ day | daily
                          | per \s+ week | a \s+ week | weekly
                          | per \s+ month | a \s+ month | monthly ] >> }
}`)

// Hours is hours of work: "40hrs pw", "37.5 hours per week", "12 hours
// per day", "40 hours". A bare "h" is hours only with a period after it,
// as "24h support" isn't.
var Hours = grammar.MustCompile(`
grammar Hours {
    token TOP  { <num> \s* :i [ [ hours | hrs | hr ] >> [ \s* <per> ]? | h >> \s* <per> ] }
    token num  { \d+ [ '.' \d+ ]? }
    token per  { <week> | <day> }
    token week { :i [ per \s+ week | pw | 'p/w' | a \s+ week | weekly ] >> }
    token day  { :i [ per \s+ day | a \s+ day | daily | per \s+ shift ] >> }
}`)

// patterns are the work patterns and arrangements, each rule its
// canonical value first, then what ads write for it.
var patterns = func() *grammar.WordList {
	l := grammar.NewWordList("patterns", grammar.FoldCase(), grammar.LooseSpacing(), grammar.WholeWords())
	if err := l.AddRules(`
permanent, perm, permanent position, permanent role
temporary, temp, temporary role, temporary position
temp to perm, temp-to-perm, temporary to permanent
contract, fixed term, fixed-term, fixed term contract, ftc
full time, full-time
part time, part-time
hybrid, hybrid working
remote, fully remote, work from home, wfh
office based, office-based, site based, site-based
`); err != nil {
		panic(err)
	}
	return l
}()

// Pattern is a work pattern or arrangement, its canonical value as the
// match's made: "Perm" is permanent.
var Pattern = grammar.MustCompile(`grammar Pattern { token TOP { @patterns } }`, patterns)

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
