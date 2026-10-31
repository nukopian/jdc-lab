# To do

## Check grammar's fewer allocations before it's released

**From:** grammar's session, 2026-10-31 (grammar 312894f9cd68206515f759a6123952602a50c502)
**Waiting on:** nothing

**What:** check grammar's roadmap 9 against the lab before it's
released: grammar 312894f9cd68206515f759a6123952602a50c502, on its
`develop`, which the lab's `go.work` picks up from `../grammar`. It
answers the task the lab gave grammar on 2026-10-31 (jdc-lab
8c9df97c5d9ab378637b3c95a2bb75b2d8193823): `once` now keeps its results
in slots on the match state, and `seq` steps over literals, classes and
anchors in a loop, both behind grammar's `fastPaths`.

**Why:** grammar won't be released until the lab finds every output
unchanged and Iain gives the go-ahead. The lab's ads aren't public, so
only the lab can show what the change does to its facts.

**What's known:** on grammar's own `BenchmarkFindLines`, which searches
each line of a posting with `FindAll(line, Outermost())`, as the lab's
facts do, the change is 40% faster with 66% fewer allocations; a pay
parse goes from 10.5 to 8.3 µs. grammar's tests pass with the fast paths
on and off, including the Raku comparisons, and the lab's
`go test ./...` passes against it. grammar's record of the change, with
every measurement, is `docs/changes/2026-10-31-r9-fewer-closures-in-the-matcher.md`
there. The baseline, at grammar v0.4.2, median of five, `-benchtime 3x`:

| Benchmark | Time | Allocated | Allocations |
|---|---:|---:|---:|
| `internal/facts` `BenchmarkFindSamples` | 1.76 s | 1.76 GB | 64.8 million |
| `internal/chunk` `BenchmarkAdOfSamples` | 2.16 s | 2.20 GB | 74.6 million |

**Steps:**

1. With `../grammar` at 312894f9cd68206515f759a6123952602a50c502, run
   `scripts/test-all.sh`, and `go test ./...` with `GOWORK=off`, against
   v0.4.2, for comparison.
2. `go run ./cmd/pipeline run claude/before` with `GOWORK=off` (grammar
   v0.4.2), and `go run ./cmd/pipeline run claude/after` with the
   `go.work`, then `pipeline compare`, showing every stage unchanged.
3. Every ad of the run through `cmd/chunk -json`, once each way, and the
   outputs identical, byte for byte: every fact the same.
4. `go test ./internal/facts ./internal/chunk -run '^$' -bench 'Samples$'
   -benchtime 3x -count 5`, each way, and the time to chunk every ad.
5. Report the figures to Iain, who decides whether grammar is released,
   and add them to grammar's `TODO.md` task, so its record can carry
   them.

**Done when** every output is unchanged, the figures are with Iain and
in grammar's `TODO.md`, and this task is removed.
