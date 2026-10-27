# Roadmap

What the lab means to do, and why. The [CHANGELOG](CHANGELOG.md) says
what it has done, and the README's
[Where it stands](README.md#where-it-stands) where it is now.

Each item is something the lab should be able to do, or find out, and
names the part of the lab's question it serves:

- **structure:** finding an ad's parts, its headings, sections and
  boilerplate;
- **metadata:** facts with a definite value, such as salary and hours;
- **taxonomies:** links to skills and occupations;
- **embeddings:** which models, for which parts;
- **search:** whether any of it makes search better, for less;
- **deployment:** running it all from Go.

Items come in four groups, the most useful first in each:

- **Now:** being worked on, each with a plan detailed enough to pick up
  cold, in its [change record](docs/changes/README.md); never more than
  three, so that it stays a queue;
- **Next:** decided, but not started;
- **Later:** meant to be done, but not yet planned;
- **Ideas:** might be done, or never; kept so they aren't lost, in a
  line or two, newest last.

Each item has a number, given when it's added, in whichever group. It
keeps it as it moves between groups, and a number is never used again,
so that 7 means the same item in a commit message months later. Next
number: 18.

When an item ships, it moves from here to the CHANGELOG, whose line for
it ends with its number, such as "(roadmap 7)", and links to its record.
An idea that's dropped is deleted, and the commit that drops it names
its number. Unless a plan says otherwise, its figures are from the first
sample, `ads-20261024` (1,890 distinct ads), and its commands run from
the lab's folder, with `RUN` standing for a run's name (see [Before and
after](docs/PIPELINE.md#before-and-after)).

## Now

Nothing at the moment: the next to start comes from Next.

## Next

### 10. Does it make search better?

**Serves:** search

Measure retrieval, not only structure: a set of queries with the ads
that should answer them, and a comparison of one embedding per ad with
embeddings of its chosen sections, with and without metadata to filter
on, and against a managed embedding service. This is what answers the
lab's question.

A first slice has shipped: job titles as queries, local
representations only ([record](docs/changes/2026-10-27-a-first-retrieval-test.md)).
The rest, judged queries beyond titles first, is still to come.

### 4. Catch footers filed under requirements

**Serves:** structure, embeddings

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** treat repeated text that comes after an ad's job content, at
its end, as boilerplate, even when it's repeated under only two to four
title groups.

**Why:** footers (an agency's notices, equal opportunities statements)
often have no heading, so they're filed under the last one, usually
requirements. Under five title groups, they're then counted as
templates for the job, and kept for embedding.

**What's known:** the largest templates in the first sample's
`boilerplate/repeated.tsv` (`cmd/boilerplate` run on `ads-20261024`)
are almost all such footers. Text under five groups or more is already boilerplate
wherever it is.

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

### 5. Split walls of text into sentences

**Serves:** structure

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

### 6. Find headings written as plain lines

**Serves:** structure

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
`data/runs/RUN/sections/named.tsv` and the unmapped table before and
after.

**Done when** more headings are found with none wrongly (check the new
ones by eye), and tests cover it.

## Later

### 7. Facts from the details

**Serves:** metadata

Grammars ([grammar](https://github.com/nukopian/grammar)) for the facts
in an ad's details sections and fields: salary and its period, hours,
working pattern, contract type and location, as metadata with definite
values, so that search can filter on them and they needn't be embedded.
A third of ads state a working pattern: a grammar for it would first
replace `internal/headings`' regular expressions for times and patterns
(their cues are in roadmap 2's
[record](docs/changes/2026-10-25-stop-shift-times-passing-as-headings.md)).

### 8. Skills and occupations as taxonomy links

**Serves:** taxonomies

Link the skills in requirements, and the job itself, to taxonomies of
skills and occupations, so that search can match on what a job needs
rather than on how its ad words it.

### 9. A fair test of embedding models

**Serves:** embeddings

Rerun the first comparison of `BAAI/bge-base-en-v1.5` and
`intfloat/e5-small-v2` fairly: with the `query: ` and `passage: `
prefixes E5 is trained with, on many headings and sections rather than
one, and with other small local models beside them. Separation between
kinds matters more than the scores themselves.

### 11. Embeddings from Go

**Serves:** deployment

Run the chosen model from Go with ONNX Runtime over the sections worth
embedding, as the ONNX demo does for its toy model, so that nothing in
the pipeline needs a Python service.

## Ideas

An idea is a numbered heading, the date it was added, and a line or
two on what and why. Each wash-up prunes them: each is kept, promoted
to Later or Next with its number, or dropped.

### 12. Boilerplate thresholds scaled to a sample's size

Added 2026-10-27. Boilerplate's thresholds are counts of title groups,
which a larger sample passes more easily (12.5% of text on 5,000 ads,
15.1% on 21,000), so its figures from samples of different sizes
don't compare.

### 13. "Label – value" lines with a dash as fields

Added 2026-10-27. "Working Hours - Monday to Friday, 08:00 - 16:30" is
a field, but fields split only at a colon or bar, so 39 such lines
became plain text when roadmap 2 stopped them passing as headings.

### 14. Pay without a currency sign

Added 2026-10-27. "Pay Rate – 13.10" in bold still passes as a heading
(25 ads), as the rule against money looks for £, $ or €.

### 15. The ad's title run into a heading

Added 2026-10-27. About 30 bold paragraphs start with the ad's title
and run on into a heading ("General Foreman Responsibilities") or a
line of facts; split off as a repeat, the rest would read as it
should.

### 17. Keep a list item's own blocks in the ad's order

Added 2026-10-27. When a list item holds a list or paragraphs of its
own, `internal/blocks` puts its text after the blocks that follow it,
so it lands under the wrong heading: 20 ads of the three samples, which
`cmd/view` counts and shows.
