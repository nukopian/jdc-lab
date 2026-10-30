# To do

## Take up grammar's release of a list's value beside a lookahead, and simplify Pattern

**From:** jdc-lab's session, 2026-10-30 (jdc-lab 7f1ea57f9ebf161589156821bd836efa7b82fd4e)
**Waiting on:** a grammar release with its fix for a word list's value
beside lookarounds (grammar's record
`docs/changes/2026-10-30-list-value-beside-lookarounds.md`)

**What:** require the grammar release that has the fix, and give
`Pattern` in `internal/grammars/grammars.go` back the one rule it was
first written as, reading the list's value as the match's `made`.

**Why:** a word list's canonical value was `made` only when the list was
a rule's whole body, so when `Pattern` gained a lookahead to skip
"Contract" as a label, it needed a rule of its own for the list
(`token word { @patterns }`, read as `word.made`). Grammar's fix gives a
rule whose only part that matches text is one list, beside lookarounds,
the list's value; Iain chose that over documenting a workaround.

**What's known:** grammar's fix was tested here before it was
committed, on 2026-10-30, with `go.work` joining the uncommitted
`../grammar` (at grammar fa339e22fa1456a041ff4c471dfc316876dc61ba plus
the fix):

- With jdc-lab as it is, the fix moves nothing: `go test ./...` passes
  with it and with `GOWORK=off`; `pipeline run` with it and with
  `GOWORK=off` gives every stage's summary unchanged and
  `blocks/ads.jsonl`, `headings/headings.tsv`, `headings/fields.tsv`,
  `sections/unmapped.tsv`, `boilerplate/repeated.tsv` and `view/ads.md`
  byte for byte the same; and the 100,501 facts read in every line of
  every ad of the three samples are the same.
- With the simpler form below and the fix, the same again: tests pass,
  outputs and facts identical. With the simpler form and the published
  grammar v0.4.0, `TestFind` in `internal/facts` fails, as the value is
  empty: so it waits on the release.

The simpler form, as a diff:

```diff
--- a/internal/grammars/grammars.go
+++ b/internal/grammars/grammars.go
 // Pattern is a work pattern or arrangement, its canonical value as the
-// made of its word: "Perm" is permanent. "Contract" as a field's label
+// match's made: "Perm" is permanent. "Contract" as a field's label
 var Pattern = grammar.MustCompile(`grammar Pattern {
-    token TOP  { <!before :i contract \s* [ ':' | :i [ type | length | duration ] >> ]> <word> }
-    token word { @patterns }
+    token TOP { <!before :i contract \s* [ ':' | :i [ type | length | duration ] >> ]> @patterns }
 }`, patterns)
--- a/internal/facts/facts.go
+++ b/internal/facts/facts.go
-		fs = append(fs, fact("pattern", m, map[string]string{"value": m.At("word.made").Str()}))
+		fs = append(fs, fact("pattern", m, map[string]string{"value": m.At("made").Str()}))
--- a/internal/facts/paths_test.go
+++ b/internal/facts/paths_test.go
-		{"pattern", grammars.Pattern, []string{"~", "from", "word.made"}},
+		{"pattern", grammars.Pattern, []string{"~", "from", "made"}},
```

**Steps:**

1. `GOWORK=off go get github.com/nukopian/grammar@` the new version, and
   `go mod tidy`; save a run first (`pipeline run claude/before`).
2. Apply the diff above.
3. `go test ./...` with `go.work` and with `GOWORK=off`, then
   `scripts/test-all.sh`; `pipeline run claude/after` and `pipeline
   compare claude/before claude/after`: nothing should move.
4. A CHANGELOG line under Unreleased for the grammar version taken up,
   and docs/SETUP.md if it names grammar's version.

**Done when** `go.mod` requires the release, `Pattern` is one rule read
as `made`, the tests pass both ways, and the run is unchanged.
