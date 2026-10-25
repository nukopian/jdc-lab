# TODO

Planned work, most useful first. It lives on `develop` and never reaches
`main`: before a release, each task is done or moved to a GitHub issue,
and the file deleted with the last. Most tasks come from a known gap in
[docs/PIPELINE.md](docs/PIPELINE.md#known-gaps); when one is done, take
it out of here and its row out of the gaps table, in the same commit,
and update PIPELINE.md's figures. Others come from another repository's
session, as their From line says. Figures below are from the v0.2.0
sample of 1,890 distinct ads.

## Find near-duplicate boilerplate

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** find boilerplate that's the same standard text with a few
words changed, which `internal/boilerplate` misses because it matches
word for word.

**Why:** most of it is likely in summaries, which are 28% of the text
but measured only 6% boilerplate. It's probably the largest amount of
boilerplate still hidden, and what isn't found gets embedded.

**What's known:** texts of eight words or more whose first 50
characters recur in five ads or more number 178, against 155 whole
texts: about 23 groups of variants. The commonest change is likely the
ad's own title or place written into the standard text ("we're
recruiting a … in …").

**Steps:**

1. Cheapest first: before keying a text, replace the words of the ad's
   own title and locations in it with one placeholder, so that
   templated text keys the same in every ad. Measure what that finds.
2. If that leaves much, compare texts by shingles (overlapping runs of
   five words) and their overlap, or MinHash for speed, grouping texts
   above a threshold; tune it on the variant groups above.
3. Keep the title-groups rule: variants count as one text across
   different jobs.

**Done when** boilerplate's share of summaries and of the whole has been
remeasured, a sample of the newly found texts has been checked by eye
to be standard text and not job text, tests cover the matching, and
PIPELINE.md describes it.

## Catch footers filed under requirements

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** treat repeated text that comes after an ad's job content, at
its end, as boilerplate, even when it's repeated under only two to four
title groups.

**Why:** footers (an agency's notices, equal opportunities statements)
often have no heading, so they're filed under the last one, usually
requirements. Under five title groups, they're then counted as
templates for the job, and kept for embedding.

**What's known:** the largest templates in
`data/runs/RUN/ads-20261024/boilerplate/repeated.tsv` are almost all such footers. Text under
five groups or more is already boilerplate wherever it is.

**Steps:**

1. Record where each text comes in its ad: whether it's in the run of
   repeated texts that ends the ad.
2. Count a repeated text found mostly in such end runs as boilerplate,
   whatever its section.
3. Check the templates under three or more groups (the command in
   PIPELINE.md's Improving it) before and after.

**Done when** those footers are boilerplate, no line that describes a
job has become boilerplate (check the newly flipped texts by eye), a
test covers it, and PIPELINE.md's figures are updated.

## Split walls of text into sentences

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** split long paragraphs into sentences, and find the headings
and bullets run into them.

**Why:** some ads are a single wall of text, with headings run into the
prose ("… every day. What you'll be doing You'll be …") and bullets
flattened into capitalised phrases with no stops. Blocks keeps a wall as
one paragraph, so nothing in it is found as a heading, and boilerplate
sentences inside it aren't found either.

**What's known:** 80 ads have a paragraph over 1,500 characters. A
heading run into text follows the end of a sentence and comes before a
capital letter. Flattened bullets are harder: capitalised phrases with
no full stop between them.

**Steps:**

1. Agree the names first: a package for sentences, and what its pieces
   are called.
2. Split paragraphs into sentences with rules (stops, question marks,
   and the capital that follows), keeping each sentence's text as
   written.
3. Find headings inside walls: a known heading phrase (from
   `headings.txt`) between a sentence's end and a capital letter.
4. Let boilerplate match sentences as well as blocks.

**Done when** the walls are measured before and after (headings found in
them, boilerplate found in them), tests cover the splitting with
made-up text in the shapes seen, and PIPELINE.md has the new stage.

## Stop shift times passing as headings

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** don't take a bold time range, such as "2:00pm–10:00pm", for a
heading.

**Why:** it starts a section of no known kind, splitting the section
it's in.

**What's known:** `headingShaped` refuses money and runs of three
digits or more, but a time has only pairs. The headings with figures in
are listed by a command in PIPELINE.md's Improving it.

**Steps:** in `internal/headings`, refuse text whose key is mostly
times (digits with am or pm, or hh:mm); add cases to `TestRead`.

**Done when** no time range is in `data/runs/RUN/all/headings/headings.tsv`, and no
other heading was lost (compare the table before and after).

## Find headings written as plain lines

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** take more plain lines, with no bold and no colon, as headings
when what follows shows they are.

**Why:** a plain line starts a section only if it's exactly a kind's
phrase, so "Your role in the team" written plainly before a list is
missed.

**What's known:** 952 plain lines start a section by name. Short lines
followed by a list are the likeliest headings.

**Steps:** in `internal/headings`, consider a short plain line that's
followed by a list, and contains a kind's phrase, as a heading; check
`data/runs/RUN/all/sections/named.tsv` and the unmapped table before and after.

**Done when** more headings are found with none wrongly (check the new
ones by eye), and tests cover it.

## Check the rules on a second sample

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** work down the second sample's unmapped table, and change any
rule that fitted only the first.

**Why:** every rule, phrase and threshold was tuned on the first
sample, which leans towards education jobs. The second leans towards
vehicle jobs.

**What's known:** both samples are in `data/ingestion/`, and their
figures in PIPELINE.md's What it measured. The rules mostly hold. Where
they don't: heading uses of a known kind fall from 76% to 73%, unmapped
text rises from 14% to 16%, and a quarter of that is boilerplate. The
biggest unmapped heading is "and" (39 ads, 82,000 characters), with
"OR" beside it: one advertiser's template puts them alone in bold
between requirements (Known gaps: lone conjunctions). Then phrases
that look like kinds: "who thrives here", "what to expect", "the
business", "what is in it for you", "vacancy type", "working
arrangement", "important notice". "The business" also maps "The
business has a strong presence across:", so check each in context.

**Steps:**

1. `go run ./cmd/pipeline run before`.
2. Stop a lone conjunction passing as a heading, in
   `internal/headings`, with a case in `TestRead`.
3. Work down `data/runs/before/ads-20261025/sections/unmapped.tsv` as
   PIPELINE.md's Improving it says, comparing each change on every
   sample.

**Done when** the second sample's known headings are at least the
first's 76%, neither sample's figures got worse, the lone conjunctions
gap is out of PIPELINE.md, and its figures are updated.

