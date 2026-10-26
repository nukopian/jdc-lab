# TODO

## Take up grammar v0.3.0 and scalar v0.6.0

**From:** grammar's session, 2026-10-26 (grammar 62136b8d57a65a8358f23887a01daec4d8cec929)
**Waiting on:** nothing

**What:** move the lab from grammar v0.2.4 to v0.3.0, which brings
scalar from v0.5.9 to v0.6.0, and say so under Unreleased in the
CHANGELOG, as grammar v0.2.2's line did.

**Why:** grammar v0.3.0 changes nothing in grammar's matching or API;
it needs scalar v0.6.0, whose
[notes](https://github.com/nukopian/scalar/releases/tag/v0.6.0) say what
changes: a number read as text keeps its digits when a `float64` can't
hold them, a whole number beyond `int64` prints as an integer, `%` and
`++` follow Perl further (`"007"++` is `"008"`), and `To` and `Decode`
fill `*big.Rat` and the like. grammar calls it a minor version because
scalar does. job-meta has both, and its programs print the same
before and after.

**Files:** `go.mod`, `go.sum`, `CHANGELOG.md`; anything in the pipeline
that prints a long number or a long decimal, or uses `%` or `++` on
strings of digits.

**How to check it's done:** save a run on today's versions, bump, run
again and compare, all without the workspace, so that the check is
against the published modules:

```
cd ~/Git/github.com/nukopian/jdc-lab
GOWORK=off go run ./cmd/pipeline run before
GOWORK=off go get github.com/nukopian/grammar@v0.3.0
GOWORK=off go mod tidy
grep -E 'nukopian/(grammar|scalar)' go.mod
GOWORK=off go test ./...
GOWORK=off go run ./cmd/pipeline run after
GOWORK=off go run ./cmd/pipeline compare before after
```

`go.mod` names grammar v0.3.0 and scalar v0.6.0, and the compare shows
every stage unchanged, or the commit message says what changed and why.
