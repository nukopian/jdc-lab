# Facts from the details, by grammar

**Status:** in progress

**Origin:** roadmap 7

**Started:** 2026-10-27, in 4b2db1d612d6ac23271a366060c1752d39c0bb5f

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
