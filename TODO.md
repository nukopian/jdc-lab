# To do

## Take up grammar v0.4.3

**From:** grammar's session, 2026-10-31 (grammar 03cad85b9fd9c2c60678e38258990efe03e3ef03, tagged v0.4.3)
**Waiting on:** nothing

**What:** move `go.mod` from grammar v0.4.2 to v0.4.3, the release of
the matcher change the lab checked on 2026-10-31 (grammar 312894f).

**Why:** the lab's `go.work` has been testing against grammar's
working copy; without the bump, a build with `GOWORK=off`, and CI, still
use v0.4.2 and its slower fact grammars.

**What's known:** v0.4.3 holds exactly the code the lab checked, plus
documents: every output unchanged, `BenchmarkFindSamples` 39% faster
with 74% fewer allocations. The check's figures are in grammar's record,
`docs/changes/2026-10-31-r9-fewer-closures-in-the-matcher.md`.

**Steps:**

1. `GOWORK=off go get github.com/nukopian/grammar@v0.4.3`, then
   `GOWORK=off go mod tidy`.
2. `scripts/test-all.sh`, and `GOWORK=off go test ./...`.
3. Commit `go.mod` and `go.sum` (`build:` or as the lab names
   dependency bumps), with this task removed, and the CHANGELOG if the
   lab records dependency bumps there.

**Done when** `go.mod` names grammar v0.4.3, the tests pass with
`GOWORK=off`, and this file is gone.
