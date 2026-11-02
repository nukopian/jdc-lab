# To do

## Take up grammar's faster matcher

**From:** grammar's session, 2026-11-02 (grammar f163c99fba29ca7e6e1fdc6176e2e83e27ef28ee)
**Waiting on:** a grammar release with roadmap 25 (after v0.5.1)

**What:** bump `github.com/nukopian/grammar` in `go.mod` to the release
with roadmap 25, a faster matcher, and check nothing changed but the
speed.

**Why:** checked against grammar's `develop` through `go.work`, the
`internal/facts` benchmark `FindSamples` took 1.11 s with v0.5.1 and
0.51 s with it (−54%, 84% fewer allocations), and `internal/chunk`'s
`AdOfSamples` 1.43 s and 0.67 s (−53%, 85% fewer). Every test in
`internal/...` passed, and `cmd/sections` on
`data/ingestion/ads-20261025.jsonl` wrote the same output (4,296 ads)
both ways.

**Steps:** `go get github.com/nukopian/grammar@` the release, then
`GOWORK=off go test ./...` and the two benchmarks with `GOWORK=off`,
so they run against the release, not the working copy. Run
`cmd/sections` on the same ads with the old and new builds and `diff -r`
the output folders; only the folder name in the summary should differ.
A call can count fewer steps against `grammar.StepLimit` than before, as
skipped alternation branches take none, so nothing that worked can hit
the limit now.

**Done when:** `go.mod` names the release, the tests pass with
`GOWORK=off`, and the sections output is unchanged.
