# TODO

## readme-pdf.sh refuses a change record by its full path

**From:** scalar's session, 2026-10-27 (scalar 4ecbc6a70c815b55c4e97e74593d8cd9ac6f57d8)
**Waiting on:** nothing

**What:** `build` in `scripts/readme-pdf.sh` matches the file name as
typed, so `cd docs && ../scripts/readme-pdf.sh changes/<record>.md`
builds a PDF of a change record that should be refused.

**Why:** the refusal exists so that the CHANGELOG and change records
never get PDFs; a relative path slips past it. scalar's wash-up before
v0.6.1 found it in scalar's copy, which came from jdc-lab's.

**Steps:**

1. Match the full path, as scalar's fix (the commit above) does:

   ```sh
   	# By its full path, so that changes/x.md, given from docs/, is refused too.
   	case $(cd "$(dirname "$1")" && pwd)/$(basename "$1") in
   	*/CHANGELOG.md | */docs/changes/*)
   ```

**Done when** `readme-pdf.sh` refuses, with exit status 2, each of
`CHANGELOG.md`, `../CHANGELOG.md` from `docs/`, and `changes/` files
from `docs/`, and still builds the other documents.

## Check the paths read from grammar matches with CheckPath

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
