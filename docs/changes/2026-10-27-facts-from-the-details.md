# Facts from the details, by grammar

**Status:** shipped, unreleased

**Origin:** roadmap 7

**Started:** 2026-10-27, in 4b2db1d612d6ac23271a366060c1752d39c0bb5f

**Finished:** 2026-10-27

## The plan

**Serves:** metadata, structure

**From:** jdc-lab's session, 2026-10-27 (jdc-lab 4b2db1d612d6ac23271a366060c1752d39c0bb5f)
**Waiting on:** roadmap 18, whose chunks show the facts

**What:** grammars, with the lab's grammar package, for the facts ads
state in lines of their own: pay, seniority, location, work pattern
(full-time, permanent, temporary) and schedule (days, hours, rotas). A
line that's a fact comes out as details, with its values, wherever it
was, so a summary no longer holds them.

**Why:** reading the views, pay, seniority, location, work patterns and
schedules sit in summaries, embedded as text, when they're facts to
filter on. And it's where grammars show what they're for: rules that
read like the text, and typed values out.

**What's known:** a third of ads state a working pattern; roadmap 2's
record has the cues for times and days ("M-F", a bare range of hours
beside a range of days, "9-5pm"). `internal/headings` has regular
expressions for times and patterns, to be replaced. A summary's text is
61% there by default.

**Steps:**

1. A grammar for each fact, each giving typed values: pay (amounts,
   range, currency, period), seniority, location, work pattern,
   schedule; tests from made-up lines.
2. A line that's mostly facts is a fact line: in `chunk`, it leaves its
   section for details, with its values.
3. Measure how much summary text moves, and check a sample by eye.
4. Where they do the same, the grammars replace `internal/headings`'
   regular expressions for times and patterns.

**Done when** `chunk` shows facts with their values, the moved share of
summary text is measured and a sample checked, and PIPELINE.md
describes the grammars.

## What happened

Work started 2026-10-27.

- **Folded in** (Iain, 2026-10-27), two ideas that are facts to read,
  their roadmap text kept as it was:

  > ### 14. Pay without a currency sign
  >
  > Added 2026-10-27. "Pay Rate – 13.10" in bold still passes as a heading
  > (25 ads), as the rule against money looks for £, $ or €.

  > ### 13. "Label – value" lines with a dash as fields
  >
  > Added 2026-10-27. "Working Hours - Monday to Friday, 08:00 - 16:30" is
  > a field, but fields split only at a colon or bar, so 39 such lines
  > became plain text when roadmap 2 stopped them passing as headings.

  Cleaning has since made every dash ASCII, so a label and " - " is the
  one shape to read.

- **Outcome:** done. `internal/facts` has a grammar each for pay,
  hours, work pattern (a word list), schedule and location (a UK
  postcode), each giving typed values; `chunk` takes lines of facts out
  of embedded chunks and unmapped sections into a details chunk of
  facts, and reads the facts in details chunks' own lines. On a project
  manager's ad, the opening's "Gateshead NE8 1AQ", "Monday to Friday
  …", "40hrs pw" and "50,000pa + excellent bonus" come out as location,
  schedule, hours and pay of 50,000 a year. On the three samples, lines
  of facts come out of 25% of ads: 9,277 lines, 0.5% of tokens, with
  5,930 pay facts, 3,131 schedules, 2,623 work patterns, 709 hours and
  261 postcodes.
- **Checked by eye:** two samples of 30 lines taken out. The first
  found five wrong ("Canteen on site", a sentence with "temporary" in
  it, "Contract administration…", a sentence of hours, a list of
  benefits opening with "Monday to Friday"), so a line of facts may not
  be a sentence, work patterns alone must be nearly all of a line, and
  "on site" left the word list. The second sample was right but for a
  pay line of a bonus's terms.
- **Changes from the plan:** seniority is left to the title, roadmap
  18's increment for it, as in lines "senior" is more often about
  stakeholders than the job. A pay figure needs a currency or a period,
  or a pay label; location is a postcode, as place names need a
  gazetteer. `internal/headings`' regular expressions for times and
  patterns stay: they keep facts out of headings, and the grammars read
  the facts; making one serve both is for later. Idea 13 is done by
  reading a known label and " - " as a field: 1,267 more fields on the
  three samples, 286 fewer heading uses ("Pay Rate - 13.10", "Location
  - Plymouth"), and idea 14 by the pay label.
- **Grammar:** written as the package's HOWTO shows, with no friction
  to report: a capture named `from` is refused, clearly, as it's the
  match's own; and `cur`, captured twice in a rule, is an array, as the
  reference says.
- **What's left:** a list of benefits that opens with a schedule is
  taken as facts; place names without a postcode aren't read.
