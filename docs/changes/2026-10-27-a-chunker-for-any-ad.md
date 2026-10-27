# A chunker for any ad

**Status:** shipped, unreleased

**Origin:** roadmap 18

**Started:** 2026-10-27, in 4b2db1d612d6ac23271a366060c1752d39c0bb5f

**Finished:** 2026-10-27

## The plan

**Serves:** structure, embeddings

**From:** jdc-lab's session, 2026-10-27 (jdc-lab 4b2db1d612d6ac23271a366060c1752d39c0bb5f)
**Waiting on:** nothing

**What:** a program that chunks one ad, as an embedding job would ask
for it: given an ad's ID in the samples, or an HTML file, it prints the
ad's chunks, each with its kind, its heading, its text, whether it's
worth embedding, about how many tokens it is, and what placed it there.

**Why:** the pipeline chunks every ad, but only shows it as figures and
tables for improving it. A chunker anyone can run on any ad is what
shows chunking is in hand, and is the thing each later item adds to.

**What's known:** blocks, headings and sections give an ad's sections;
the run's `boilerplate/repeated.tsv` holds the verdict on every
repeated text, so a chunker can mark boilerplate without indexing every
ad again. A section's kind comes from its heading's phrase, from
"why" or "about" at its start, from a field's label, or by default.

**Steps:**

1. `internal/chunk`: an ad's sections as chunks. A chunk of the
   summary, responsibilities, requirements, desirable or an unmapped
   section is embedded, without its boilerplate; benefits, details,
   company, apply and legal aren't; fields are metadata. Each says what
   placed it: the heading and the phrase in it, the rule, or the
   default.
2. `internal/sections` says which phrase in a heading gave its kind.
3. `cmd/chunk`: an ID, an HTML file or standard input; readable text by
   default, JSON with `-json`; boilerplate from a run's
   `repeated.tsv`.
4. PIPELINE.md and the README describe it.

**Done when** `chunk` prints any sample ad's chunks, readably and as
JSON, each with what placed it; tests cover the chunking; and the docs
say how to run it.

## What happened

- **Outcome:** done. `go run ./cmd/chunk ID` prints an ad's chunks in
  half a second, as text or, with `-json`, a line of JSON an embedding
  job can take; it chunks a file of HTML too. Each chunk says what
  placed it, down to the phrase: on the first ad tried, "Part-Time
  Teachers Wanted for Day-to-Day Supply!" is responsibilities by the
  phrase "day to day", and "Please Apply If You…", over a list of
  requirements, is apply by "apply". Faults like these, seen before
  only in the figures, now point at the line of `headings.txt` to
  change.
- **Changes from the plan:** `sections.Explain` names the phrase that
  won and those of other kinds it outranked, as well as the rule; the
  plan had a phrase only. Tokens are estimated at four for every three
  words, rounded, as the chunker runs without a model.
- **Afterwards,** still unreleased: `-yaml` writes each ad as a YAML
  document, for people, each chunk's text a block of its lines, where
  JSON is for programs. The lab writes its own YAML, as the usual Go
  library (`go.yaml.in/yaml/v3`) escapes emoji past the first 65,536
  characters, as ads use; that library stays, in a test that reads the
  YAML back, tricky strings and all, and finds it as written.
- **Boilerplate** comes from the run's `repeated.tsv`, by each text's
  own key; a text the run found by its templated key or as alike to
  another counts too, as the table lists every repeated text by its own
  key, with the verdict of the first ad it was seen in.
