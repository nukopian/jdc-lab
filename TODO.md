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
**Waiting on:** a grammar release with `CheckPath` (after v0.3.2)

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
