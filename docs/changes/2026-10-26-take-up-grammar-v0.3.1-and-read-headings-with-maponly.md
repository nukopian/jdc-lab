# Take up grammar v0.3.1, and read headings.txt with MapOnly

**Status:** shipped, unreleased

**Origin:** a TODO task from grammar's session (grammar 675df5ba00d3828d4b37bbb6e439ba30e864b7e8)

**Started:** 2026-10-26, in 7b52cb2681da5b187c2fe6828cb272651e434b07

**Finished:** 2026-10-26

## The plan

**From:** grammar's session, 2026-10-26 (grammar 675df5ba00d3828d4b37bbb6e439ba30e864b7e8)
**Waiting on:** nothing

**What:** move the lab from grammar v0.3.0 to
[v0.3.1](https://github.com/nukopian/grammar/releases/tag/v0.3.1) (scalar
stays at v0.6.0), and use its new `MapOnly()` list option in
`internal/sections`. Under it, a rule written `phrases => kind` adds the
phrases meaning the kind, as `AddAs` does, without making the kind a
phrase too, so `phrases.AddRules(rules)` builds the list directly. The
`unnamed` regex, the `read` list and the `AddAs` loop go, and rank is
read from `phrases.Canonicals()`, which keeps the rules' order:

```go
phrases = grammar.NewWordList("sections",
	grammar.FoldCase(), grammar.LooseSpacing(), grammar.Plurals(), grammar.WholeWords(), grammar.MapOnly())
```

```go
func init() {
	// MapOnly keeps the name of a kind written "phrases => kind" out of
	// its phrases.
	if err := phrases.AddRules(rules); err != nil {
		panic(fmt.Sprintf("sections: headings.txt: %v", err))
	}
	// Each rule's kind must be one of Kinds: ...
	for i, k := range phrases.Canonicals() {
```

and drop the `regexp` import. Then reword the paragraph in
`headings.txt` that says "This is the lab's reading: grammar, as
Elasticsearch, would take the name as a phrase too": the lab now asks
grammar for that reading with `MapOnly`. Say both under Unreleased in
the CHANGELOG.

**Why:** this was roadmap 12 in grammar, from the lab's "legal" section,
whose name in a heading usually names a job ("Legal Secretary"). grammar
always made a `=>` rule's value a phrase, so the lab read its rules
twice and rebuilt them with `AddAs`; `MapOnly` is that rebuild, built
in. grammar's
[change record](https://github.com/nukopian/grammar/blob/main/docs/changes/2026-10-26-rules-whose-canonical-isnt-a-phrase.md)
has the details. A rule with nothing before `=>` is now an error under
`MapOnly`, as it would add nothing; `headings.txt` has none.

**Files:** `go.mod`, `go.sum`, `internal/sections/sections.go`,
`internal/sections/headings.txt`, `CHANGELOG.md`.

**How to check it's done:** grammar's session tried this change against
its checkout, and then undid it here. The `MapOnly` list was the same as
the rebuilt one, compared by a temporary test: 239 phrases, the same
meanings for each, the kinds in the same order, and "legal" not a
phrase. `go test ./internal/...` passed, and
`go run ./cmd/sections data/ingestion/ads-20261024.jsonl` wrote the same
files as before. Do the same with `GOWORK=off`, after
`GOWORK=off go get github.com/nukopian/grammar@v0.3.1`: the tests pass,
and `cmd/sections`'s output for that file is unchanged, apart from the
folder it names.

## What happened

- **Outcome:** done. `go.mod` names grammar v0.3.1, scalar stays at
  v0.6.0, and `internal/sections` builds its list with `MapOnly` and
  `AddRules` alone: the regexp, the second list and the `AddAs` loop
  are gone. `GOWORK=off go test ./...` and `scripts/test-all.sh` pass.
- **Changes from the plan:** the check was a whole run rather than
  `cmd/sections` on one sample, and PIPELINE.md's paragraph on `=>`
  was reworded too, as it described the rebuild.
- **Measurements:** `GOWORK=off go run ./cmd/pipeline run after`,
  compared with `now` (the run of the commit before): every stage
  unchanged on both samples and `all`, and each `sections/ads.jsonl`
  the same byte for byte.
- **Tried and reverted:** nothing.
- **What's left:** nothing. This came from the lab's dogfooding of
  grammar: roadmap 12 there, suggested on 2026-10-26 and shipped the
  same day.
