# To do

## Take up grammar v0.5.3: each rule's starts kept

**From:** grammar's session, 2026-11-03 (grammar 9b0b9cef9004a24f0a913690cbe31a5f775aea46)
**Waiting on:** nothing

**What:** bump `github.com/nukopian/grammar` in `go.mod` from v0.5.2 to
v0.5.3, which carries the fix this repository's session asked for at
3baebf12fcae06bacd5e2e8a39a1624f196e2b48: a grammar searched for
several rules in turn keeps each rule's starts, not only the last
one's.

**Why:** the headings stage searches its `Schedule` grammar for six
rules in turn, and `startsOf` was 38% of reading an ad's headings. With
the fix, through `go.work`, `headings.BenchmarkReadSamples` went from
412 ms to 195 ms and `boilerplate.BenchmarkSamples` from 3.82 s to
3.26 s. Until `go.mod` asks for v0.5.3, a build without `go.work` gets
the old code.

**What's known:** the code is the diff the task proposed, unchanged.
grammar's record is
[docs/changes/2026-11-03-keep-each-rules-starts.md](https://github.com/nukopian/grammar/blob/v0.5.3/docs/changes/2026-11-03-keep-each-rules-starts.md).
No API changed, and results are the same.

**Steps:**

1. `GOWORK=off go get github.com/nukopian/grammar@v0.5.3`, then
   `go mod tidy`.
2. `GOWORK=off go test ./...`, so the tests run against v0.5.3 rather
   than the local checkout.
3. Optionally, re-run `headings.BenchmarkReadSamples` with `GOWORK=off`
   to confirm the gain against the tagged version.

**Done when:** `go.mod` asks for grammar v0.5.3 and the tests pass with
`GOWORK=off`.
