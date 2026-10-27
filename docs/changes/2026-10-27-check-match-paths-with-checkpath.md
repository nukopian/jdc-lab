# Check the paths read from grammar matches with CheckPath

**Status:** shipped, unreleased

**Origin:** a TODO task from grammar's session (grammar fe34207b5d6bfbd3b10ff6b2eea0416fb8a5cfec)

**Started:** 2026-10-27, in 0f607026478f77069b8fcd66889522aac4737993

**Finished:** 2026-10-27

## The plan

**From:** grammar's session, 2026-10-27 (grammar fe34207b5d6bfbd3b10ff6b2eea0416fb8a5cfec)
**Waiting on:** nothing: grammar v0.4.0 has `CheckPath` (grammar's
session, 2026-10-27, grammar d8892dc6722d60eacc0cdbf23b7d1cf74c5c3ea2)

**What grammar shipped:** v0.4.0 is a breaking release, but not for
jdc-lab: `$<k>=( … )`, `$<k>=<name>`, `$<k>=@list` and `$<k>=<@list>`
with a quantifier after them now capture as Raku does, each match
named, and jdc-lab's grammars use none of those (only `$<k>=[ … ]`, and
`$<open>=@openings` without a quantifier). Its tests passed against
grammar's checkout with the change. `Chunks` now gives the text
between captures inside another, as Raku does; jdc-lab doesn't call it.
v0.4.0 needs scalar v0.6.1, so take both:
`GOWORK=off go get github.com/nukopian/grammar@v0.4.0 github.com/nukopian/scalar@v0.6.1`,
then `GOWORK=off go mod tidy` and `GOWORK=off go test ./...`.

**What:** a test in `internal/facts` and one in `internal/wording`
that calls `CheckPath` on every path the package reads from a match,
grammar by grammar: `pay` (`cur`, `cur[0].~`, `min`, `max`, `min.k`,
`max.k`, `upto`, `period`, `period.~`, `min.num.~`, `max.num.~`, as
`amount` reads `num.~` and `k` from `min` and `max`), `hours` (`num.~`,
`per.week`, `per.day`), `pattern` (`made`), `schedule` (`days`,
`days.mf`, `days.first.~`, `days.last.~`, `times`, `times.a.~`,
`times.b.~`, `rota`, `rota.~`), `postcode`, and `opening` (`open.made`,
`phrase.~`), each with `~` and `from`, which `fact` reads.

**Why:** a misspelt capture reads undef, as one that didn't match does,
so it fails silently; `hours`' `per.weak` and `opening`'s
`phrase.open.made` were found that way, by hand. With the test, a typo
or a capture moved in a grammar fails `go test`.

**What's known:** grammar's session checked all 36 against its
checkout, through this repository's `go.work`: all pass, and
`hours.CheckPath("per.weak")` gives `rule per has no capture "weak" (it
has: day, week)`. The tests need to be inside the packages, as the
grammars aren't exported.

**Done when** `go test ./...` runs both tests against the released
grammar, with `GOWORK=off`, and a misspelt path in either list fails.

## What happened

- **Outcome:** done. `go.mod` names grammar v0.4.0 and scalar v0.6.1.
  `internal/facts/paths_test.go` and `internal/wording/paths_test.go`
  check every path in the plan's lists, and pass with `GOWORK=off`, as
  `scripts/test-all.sh` does. With `per.weak` put in for `per.week`,
  and `phrase.open.made` for `open.made`, both fail:
  `rule per has no capture "weak" (it has: day, week)`, and
  `phrase is text only: a capture inside $<phrase>=[ … ] is beside it,
  so read "open.made"`.
- **Changes from the plan:** wording's test checks `opening`, its one
  grammar, rather than a grammar per kind of fact, as facts' does.
  `~` and `from` are checked for facts' grammars only, as only `fact`
  reads them. MAINTAINING's Go section now says a new path goes in the
  list.
- **Measurements:** `pipeline compare claude/after claude/grammar040`,
  the second run built with `GOWORK=off` on v0.4.0: every stage
  unchanged, no heading newly mapped or unmapped.
- **Tried and reverted:** nothing.
- **What's left:** nothing.
- **Wash-up:** the plan missed `internal/sections`, whose `Explain`
  reads `~` from its finder's matches, so MAINTAINING's rule that every
  package reading a grammar's matches has a `paths_test.go` didn't hold.
  It has one now; a path of `phrase` fails it, `rule TOP has no capture
  "phrase"`.
