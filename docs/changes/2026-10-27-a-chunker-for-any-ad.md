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
- **Then gathered into a chunk of its own** (Iain, 2026-10-27): kept in
  embedded chunks, boilerplate cost tokens; taken out, it was lost from
  sight. So it leaves the embedded chunks for one `boilerplate` chunk at
  the end of the ad, left out, each line with the section it came from,
  where a job line the verdict took is plain to see. Other chunks keep
  theirs. On a support worker's ad, the advertiser's standard list of
  benefits, filed under the summary, is gathered there: 192 tokens no
  longer embedded.
- **Unmapped text goes to the boilerplate chunk too** (Iain,
  2026-10-27), under its heading, still counted as unclassified; an
  unmapped section's fields become a details chunk. On a driver's ad,
  the location and pay under its title stay as fields, and its
  introduction, under "Ongoing opportunity", goes to the boilerplate
  chunk until that heading is mapped.
- **Role is a kind of its own** (Iain, 2026-10-27): what the job is,
  under "About the role", "The role", "Job description" and the like,
  apart from the summary, the ad's opening and its introductions
  ("Overview", "The opportunity"). Both are embedded. On the three
  samples, role is in 40% of ads and 9% of the text, and the summary
  falls from 28% of the text to 19%. Role isn't yet among the sections
  boilerplate counts as describing the job, so its verdicts are as
  before.
- **List items** are "- " entries in a chunk's text.
- **Sub-headings before lists** (Iain, 2026-10-27): a short plain line,
  not a sentence, with a kind's phrase, before a list, starts that
  kind's section, so "Highly Desirable" and "Essential Experience &
  Qualifications" split the requirements, and "Key Skills" joins them.
  That's roadmap 6's plan. On the three samples, 4,111 more lines start
  sections; responsibilities are found in 60% of ads, from 53%; the
  summary falls to 17% of the text, and ads with only a summary to 4%.
  Of 30 of the rarer lines checked, one was wrong ("Duties required"
  went to requirements).
- **The view shows the chunker's chunks** (Iain, 2026-10-27): it had
  marked boilerplate in every section, benefits included, where `chunk`
  keeps it in sections not embedded and gathers the rest, so the two
  showed one ad two ways. `view.Render` now lays out `chunk.Of`'s chunks,
  with the ad's totals; the view keeps its choice of ads and its check
  for text lost or moved.
- **Chunks in a fixed order** (Iain, 2026-10-27): summary, role,
  responsibilities and requirements first, in that order, then the
  rest as the ad has them, and the boilerplate last.
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
- **Folded in** (Iain, 2026-10-27), as what's left to close 18, their
  roadmap text kept as it was:

  > ### 19. Requirements and responsibilities by their wording
  >
  > **Serves:** structure
  >
  > A grammar for the lines that say what they are ("Experience of…",
  > "You will…", "Must have…"), so that requirements and responsibilities
  > no heading placed come out in their sections, each saying which rule
  > placed it. Done when the share of summary text there by default, 61%
  > now, is remeasured and a sample checked by eye.

  > ### 15. The ad's title run into a heading
  >
  > Added 2026-10-27. About 30 bold paragraphs start with the ad's title
  > and run on into a heading ("General Foreman Responsibilities") or a
  > line of facts; split off as a repeat, the rest would read as it
  > should.

  > ### 20. Before and after pages
  >
  > **Serves:** structure, embeddings
  >
  > For a dozen ads chosen to show where chunking does well and badly,
  > the ad as posted beside its chunks: what's embedded, what's dropped as
  > boilerplate, what became metadata, and which rule placed each. In the
  > private data store, as the ads aren't ours to publish. Done when the
  > pages show the chunker's current output, from `chunk` itself.

- **Roadmap 6 done on the way:** the rule for a short line before a
  list is its plan, and meets its Done when (more headings found, a
  sample checked by eye, one of 30 wrong; tests cover it). It leaves
  the roadmap, with its row in PIPELINE.md's Known gaps.
- **Lines placed by their wording** (was roadmap 19): `internal/wording`
  reads a line's opening, with a grammar over the openings in
  `openings.txt` ("experience of", "must have", "conduct", "free
  parking"), and the chunker moves a summary's lines and a role's list
  items that open so into a chunk of that kind, before boilerplate is
  gathered. On the three samples, 8,524 lines move, in 13% of ads:
  6,996 to responsibilities, 1,083 to requirements, 445 to benefits.
  Of tokens, the summary falls from 17.5% to 17.0% and the role from
  7.6% to 7.1%, and responsibilities rise from 15.8% to 16.6%; most of
  what moves is a role's lists of duties. Of 40 lines checked by eye,
  one was wrong ("You'll have the opportunity to…" as a requirement),
  so "you'll have the opportunity" became a benefit's opening. "You
  will" and "you'll" alone aren't openings, as they open requirements,
  duties and benefits alike, nor are "build" and "drive", which
  summaries use in pitches. 19's Done when asked for the share of
  summary text there by default, 61% when it was written: the opening
  before any heading is now 10.6% of all text, the summary as a whole
  17%.
- **Grammar:** one point of friction passed back, as grammar's idea 15:
  its reference doesn't say that a capture inside a text capture
  (`$<k>=[ … ]`) belongs to the enclosing rule, which took two tries
  to find.
- **Plain-text bullets** (Iain, 2026-10-27): a line that opens with `*`,
  or with `-` before a space or a letter, loses the mark, and lines in a
  row with it are a list, as with symbols. On the three samples, 29,666
  more list items, in 5,221 more lists; ads with only a summary fall
  from 1,048 to 1,011.
- **The title split off** (was idea 15): far more than its 30 bold
  paragraphs open with the ad's title, about 4,750 lines with six words
  or fewer after it, most of them a heading ("… Requirements:", "… - The
  Role"), a place or facts. The title is split off as a repeat when
  what follows is a field, a heading (the line having been one with the
  title too), or, after a dash, bar, colon, comma or bracket, a short
  line that isn't a sentence. Three tries: the first let a long bold
  line become a heading once the title was off it, and doubled the
  title where it was also the line's bold lead (117 ads with text in no
  section, found by the view's check); the second split longer titles,
  "Panel Beater / Panel Technician". A heading of no known kind after
  the title is the ad's subtitle ("Oldham", "Healthcare"), text in its
  section: before, it started an unmapped section, at times over the
  whole ad, which went to the boilerplate chunk. On the three samples,
  4,749 titles split off; distinct headings fall from 24,735 to 23,310;
  headings of a known kind rise from 77% of uses to 78%, ads with every
  heading known from 54% to 57%, and ads with an unmapped section fall
  from 46% to 43%; the summary rises from 17% of the text to 18%. Of
  30 splits checked by eye, all took the title off rightly. Among the
  headings that moved, "Cover Supervisor work is ideal for:" leaves
  "work is ideal for:", a weak heading, as the whole was.
- **Pages before and after** (was roadmap 20): `chunk -pages FILE`
  writes each ad FILE lists as a web page, the ad as posted beside its
  chunks, coloured by use, with what placed each and the facts read,
  and a bar of where its tokens go, with an index. The ad's HTML is cut
  down to the tags that lay text out, and everything else escaped by
  `html/template`. A dozen ads, picked from all three samples by what
  they show and then read, are in the private data store: six done
  well (headings to eight kinds; requirements split by sub-heading;
  duties and requirements placed by wording; the title split off; pay,
  hours, pattern and schedules read as facts) and six done badly (one
  heading swallowing the ad; no headings at all; headings still to
  map, leaving 88% of an ad unmapped; a "Why …?" heading taking in the
  pay; job lines taken as boilerplate, roadmap 4). The notes were
  checked against the chunks: six of the first drafts claimed more
  than the chunks showed, and were rewritten. 20's Done when, the
  pages showing the chunker's current output from `chunk` itself, is
  met: a rerun rebuilds them.

