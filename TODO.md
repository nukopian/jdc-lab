# To do

## Take up grammar v0.4.4, and try FindErr where facts go missing

**From:** grammar's session, 2026-11-01 (grammar a562f5652968156c7133760b86d7b8bef6f8d169, work for v0.4.4 on develop)
**Waiting on:** grammar v0.4.4

**What:** move `go.mod` to grammar v0.4.4, and use its `FindErr` where
the lab asks why a fact wasn't found.

**Why:** v0.4.4 adds `FindErr` and `FindAllErr`, which say why a search
found nothing: of the attempts at each position, the one that got
furthest past where it started, and what it expected there, such as
`expected <amount>, found "TBC…" (trying from line 1, column 9)`. Until
now the only way was to cut out the likely span and run `ParseErr` on
it. It also skips positions for grammars that start with `Re`, which the
lab's grammars don't, and costs searches with other grammars about 3%
on long texts; the lab's line-by-line searches measured no change in
grammar's `BenchmarkFindLines`.

**What's known:**

- Pay rarely fails outright: its rule takes any figure, and `payOf`
  decides in Go what's pay. So `FindErr` helps most where a test expects
  a fact and gets none: `facts_test.go`'s `TestFind` and its neighbours
  could print `FindErr`'s report for the grammar that should have
  matched, so a failure says why.
- grammar's roadmap idea 23 (rejecting a text that can't hold a match)
  came from `grammars.MayHold`: it would make that check automatic for
  Pay, Hours, Postcode and Figure. It isn't planned yet; when it ships,
  grammar's session will hand the lab a task to retire `MayHold` for
  those grammars.

**Steps:**

1. `GOWORK=off go get github.com/nukopian/grammar@v0.4.4`, then
   `GOWORK=off go mod tidy`.
2. `scripts/test-all.sh`, and `GOWORK=off go test ./...`; every fact and
   chunk unchanged (`pipeline compare`), and `BenchmarkFindSamples`
   against its last figure.
3. Where a fact test fails for a missing fact, report `FindErr`'s error
   for the grammar expected, if it reads well; otherwise note why not as
   friction for grammar.
4. Commit, with this task removed.

**Done when** `go.mod` names grammar v0.4.4, the tests pass with
`GOWORK=off`, outputs are unchanged, and this file is gone.
