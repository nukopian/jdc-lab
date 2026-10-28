# One set of grammars for facts and headings

**Status:** shipped, unreleased

**Origin:** roadmap 22

**Started:** 2026-10-28, in f448b6ad38e655513232d97d79c2f9b7be076f5a

**Finished:** 2026-10-28

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

- **Outcome:** done. `internal/grammars` holds pay, hours, pattern,
  schedule and postcode, imports neither `facts` nor `headings`, and
  both use it. `headings` keeps facts out of headings with the
  grammars' rules (`Pay`'s `cur`, `Schedule`'s `clock`, `points`,
  `days`, `span`, `rota` and `shift`, and a `Figure` grammar for long
  figures), and has no regex for a fact; `labelled`, which reads a
  label, stays. `go test ./...` passes, with and without the
  workspace, and the compare shows no change on the three samples.
- **Changes from the plan:**
  - Step 3: the paths tests stay with the packages that read the
    paths, as MAINTAINING says, rather than going in the new package,
    which reads none. `headings` has its own now (`shift`'s `count`,
    `span`'s `a.h.~`, `b.h.~` and `unit`). `internal/grammars` has a
    test of the rules `headings` uses, and an example.
  - "from" is reserved for a match's place, so `span`'s ends are `a`
    and `b`, as `times`' are.
  - Reading `a.h` rather than `a.h.~` gave the match itself, which
    printed as "HASH(…)" and read as 0, so every range ran round the
    clock and every count was a time. `TestRead` caught it at once;
    `CheckPath` can't, as `a.h` is a real path.
- **What `facts` reads now:** the shared rules are those `headings`
  needed, so `facts` reads with them too. A day is a whole word
  (`<<`, so "Salmon-Fri" is no range), "til" joins a range as "till"
  does, and "M - F" may have spaces. A time starts at a figure that
  isn't inside another (`<!after \d>`): the regexes' word boundary,
  tried first as `<<`, lost "Friday7:00am", where a day runs into the
  time. Of 103,015 facts on the three samples, 15 fewer are read, and
  the same 22,918 ads have one. The 17 lost were fragments read from
  inside a figure ("45am-5pm" from "745am-5pm", "30am - 16" from
  "0730am - 1630pm"), and one real time, "18:00 - 04:30", from a line
  where two bold times run together ("18:3018:00 - 04:30"). The two
  gained are a range joined by "til".
- **Measurements:** `pipeline compare claude/before claude/after`, on
  all three samples, after each step: unchanged. The facts were
  compared by dumping every fact found, from a worktree of the commit
  before and from this one.
- **Tried and reverted:** `<<` as the time rules' boundary, above.
- **What's left:** times written as four figures ("0730am",
  "1630pm") aren't read at all now, where before they gave fragments;
  and two times run together by markup lose the second.
