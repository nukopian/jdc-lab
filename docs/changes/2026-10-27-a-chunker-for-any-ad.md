# A chunker for any ad

**Status:** in progress

**Origin:** roadmap 18

**Started:** 2026-10-27, in 4b2db1d612d6ac23271a366060c1752d39c0bb5f

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

- **The item stays open** (2026-10-27): it's the near-term aim,
  chunking shown working with grammars, and roadmap 7 (facts), 19
  (lines placed by their wording) and 20 (pages of ads before and
  after) build on it; it's done when 20's pages show it. What follows
  is its first increment.
- **First increment:** done. `go run ./cmd/chunk ID` prints an ad's chunks in
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
- **Second increment, text cleaned:** blocks now cleans an ad's text
  for every stage after it (`internal/clean`), the ad itself kept as
  received: structure in English, not symbols. A symbol that opens a
  line and stands for one thing becomes its label when what follows is
  a value of it ("📍 Leeds" is "Location: Leeds", a field), and goes
  otherwise; lines in a row opening with the same symbol are a list;
  other emoji, bullets and invisible characters go; quotes, dashes and
  ellipses become ASCII, an em dash " -- ". `headings.Key` now parts
  words at a typographic dash, as at a hyphen, the fault found in
  roadmap 10, so titles, uncleaned, still match their text. On the
  three samples together: 32,767 lines become items of 5,297 lists,
  most of them "•" bullets in ads laid out as plain text; 554 more
  fields; heading uses fall 251, as bulleted points that were taken for
  headings become items; ads with every heading known rise from 53% to
  54%; 66 ads lose a section of a known kind, almost all such points,
  and none checked lost a real heading. No text is lost (the view's
  check).
- **Unmapped chunks are unclassified,** not embedded (Iain, 2026-10-27):
  an unmapped heading is mapping still to do, not a sign its text is
  worth embedding. Each ad totals its unclassified tokens beside its
  embedded ones; on a driver's ad tried, 46% of tokens are embedded and
  28% unclassified.
- **Boilerplate is marked, not taken out** (Iain, 2026-10-27). Out of
  a chunk left out or kept as metadata, it broke lists apart: an
  advertiser's standard perks showed as one benefit and six lines of
  boilerplate. And out of embedded chunks it took job text: of what's
  found as boilerplate in summaries (9.7% of their text), requirements
  (11.7%), responsibilities and desirable, about half of a sample is
  about the job ("High levels of accuracy and attention to detail", an
  agency's template for a SEND tutor's duties), as the rule that text
  under five groups of titles is boilerplate also takes in common job
  wording. So every chunk keeps its text; an embedded one lists what's
  marked. Telling footers from common job wording is still to do.
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
