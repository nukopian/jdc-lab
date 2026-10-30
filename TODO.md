# To do

## Test grammar's fix for a word list's value beside a lookahead

**From:** grammar's session, 2026-10-30 (grammar fa339e2a, work uncommitted)
**Waiting on:** nothing; grammar's fix is uncommitted in `../grammar`,
which this repository's `go.work` already joins

**What:** check grammar's uncommitted fix against this repository, then
tell Iain whether it holds, so that grammar can commit and release it.
Iain asked for the fix not to be committed until jdc-lab has tested it.

**Why:** this repository's task in grammar (grammar fa339e2a) found that
`token TOP { <!before …> @patterns }` left `made` empty, so `Pattern` in
`internal/grammars/grammars.go` gained a rule of its own for the list,
`token word { @patterns }`, read as `word.made`. Iain chose to make the
first form work.

**What's known:**

- In grammar's working copy, a rule (or `$<k>=[ … ]`) whose only part
  that matches text of its own is one word list, beside lookarounds,
  anchors and a `rule`'s whitespace, gets the list's canonical value as
  `made`. It's looked up in what the list alone matched, so
  `rule TOP { @patterns }` works too, though its match takes in the
  space after. Anything else that matches text (`@patterns \s*`), or a
  second list, still gives no value.
- Matching itself is unchanged: grammar's Raku differential tests pass,
  with three new cases for these shapes.
- Checked in grammar with `Pattern`'s own lookahead and the list
  inline, `token TOP { <!before :i contract \s* [ ':' | :i [ type |
  length | duration ] >> ]> @patterns }`: "Contract: Permanent" gives
  "permanent", "Contract type: Temp to Perm" "temp to perm",
  "Part-Time: 25 hours" "part time", and "Contract" as a label is still
  skipped. `CheckPath("made")` accepts it.
- The `word` rule keeps working as it is, so the test is first that
  nothing moves, then that the simpler form gives the same output.
- `m.At("word.made")` is read in `internal/facts/facts.go` (line 63),
  and the path is checked in `internal/facts/paths_test.go` (line 24).

**Steps:**

1. With `go.work` as it is, run `go test ./...` and the programs whose
   output you compare, unchanged: the output should be the same as
   before, since `word` still works.
2. Try the simpler form: `token TOP { <!before …> @patterns }` with no
   `word` rule, reading `made` in `facts.go` and checking `made` in
   `paths_test.go`. The output should be unchanged again.
3. Tell Iain the result. Don't commit step 2: it needs a grammar
   release first. Keep the diff, or note it here, for after the
   release.

**Done when** Iain knows whether the fix holds in jdc-lab, with the
output compared both ways. A task to take up the release, and perhaps
the simpler form, follows once grammar is released.
