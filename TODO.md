
## Try grammar's Overlap in sections.All, before it's released

**From:** grammar's session, 2026-10-26 (grammar da7bc9b9cdc732bbed40f74bf56f312d654d3c02)
**Waiting on:** nothing: it's on grammar's `develop`, which the lab's `go.work` joins

**What:** grammar's new `Overlap()` option for `FindAll` (its roadmap
13, from the lab's idea) is Raku's `:ov`: a match from each position
where one starts, so "in this role you will need" gives both "in this
role you will" (responsibilities) and "you will need" (requirements).
Iain asked for the lab to try it before grammar releases it. In
`internal/sections`, `All` would pass it:
`finder.FindAll(key, grammar.Overlap())`.

**What's known:** grammar's session tried exactly that change, and then
undid it here. On `data/ingestion/ads-20261024.jsonl` (1,890 ads), the
tests passed, and:

- 9 more headings had phrases that disagree, in `disagree.tsv`, all
  but one with the same winner as before: "What you'll get in return"
  (benefits, responsibilities), "Why Apply?" (benefits, apply),
  "Client Details" (company, details), "Company Overview:" (company,
  summary), and others like them.
- 2 ads changed kind, both "Company Profile", from company to
  requirements. `:ov` also finds a phrase wholly inside a longer one,
  and "profile" (requirements) inside "company profile" outranks it.

So the lab probably wants the phrases that reach past another's end,
such as "you will need", but not those wholly inside one, such as
"profile": drop each match whose `from` and `to` lie within another's
(grammar's reference, under Overlapping matches, says the same). That
takes the match's positions, so it's done on the matches before
`ToAll`.

**Steps:**

1. Pass `Overlap()` in `All`, drop the matches inside another's, and
   add a test with "In this role you will need" (both kinds) and
   "Company Profile" (company only).
2. Run `go run ./cmd/sections data/ingestion/ads-20261024.jsonl` before
   and after, and read the differences.
3. Tell grammar: in grammar's `TODO.md`, in the task "Release Overlap
   once jdc-lab has tried it", set Waiting on to "nothing" and add what
   the lab found, especially anything grammar should change before
   the release; commit and push it there as a `docs:` commit. Keep the
   lab's change uncommitted, or on `develop` only with `go.work`, until
   grammar has released, then take up the release and say so in the
   CHANGELOG.

**How to check it's done:** grammar's task says the lab has tried it,
and the lab's sections use `Overlap` with nested matches dropped, its
output read and explained.
