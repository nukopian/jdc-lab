# One set of grammars for facts and headings

**Status:** in progress

**Origin:** roadmap 22

**Started:** 2026-10-28, in 7dbc459e5efbfa5891397f2b2510ee61bd316c67

## The plan

**Serves:** structure, metadata

**From:** jdc-lab's session, 2026-10-28 (jdc-lab 7dbc459e5efbfa5891397f2b2510ee61bd316c67)
**Waiting on:** nothing

**What:** move the grammars for facts out of `internal/facts` into a
package that both `internal/facts` and `internal/headings` import.
Then `headings` decides whether a line holds pay, a time or a pattern
of work with those grammars, in place of its own regexes.

**Why:** at the moment a line is a time, or a rota, by two sets of
rules: `headings`' regexes, which keep facts out of headings, and
`facts`' grammars, which read their values. The two can drift apart,
and a line one calls a time the other may not read. `facts` imports
`headings` (for `Key`, to read a pay label), so `headings` can't import
`facts` back. A package below both is the way round that. 21 builds on
this: a heading made only of facts, such as "Full-Time | Permanent", can
then be told by the same grammars.

**What's known:** the regexes are at the end of
`internal/headings/headings.go`, and `headingShaped` is their only
user:

- `money` matches any currency sign, or any run of three digits or
  more. It catches pay, and long figures such as a reference number.
  `facts`' pay grammar needs a currency or a period, so the long
  figures need a rule of their own.
- `times` matches a clock time, or a range with points, as "8.30-3.30".
- `days` matches a range of days, `hours` a bare range of hours, and
  `hasTimes` takes hours only beside days, and not when a unit makes
  them a count ("2-3 days"), unless they run round the clock ("7-3
  Days"). `facts`' schedule grammar has the same idea, a bare range only
  after days, but no units or round the clock.
- `rota` is the same in both.
- `shiftWord` (days or nights, with no figure before them) has no
  grammar yet.

`labelled` reads a label, not a fact, so it stays. `headings_test.go`
pins the cases each regex was written for.

**Steps:**

1. A new package, for the grammars both use, which imports neither
   `facts` nor `headings`. Move pay, hours, pattern, schedule and
   postcode into it, with `facts` reading them from there, and its
   output unchanged.
2. In `headings`, put the grammars in place of the regexes one at a
   time, adding rules to the shared grammars where `headings` needs
   more than `facts` does (long figures, counts, `shiftWord`), so that
   `facts` reads what `headings` keeps out. After each, the tests
   pass and a compare shows what moved.
3. The new package's `paths_test.go`, as MAINTAINING asks.
4. PIPELINE.md says where the grammars live, and what keeps facts out
   of headings.

**Done when** `internal/headings` has no regex for a fact; `facts` and
`headings` use the same grammars; `go test ./...` passes; and the
compare shows no change, or each change is read and explained here.

## What happened
