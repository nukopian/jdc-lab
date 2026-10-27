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
number: 21.

When an item ships, it moves from here to the CHANGELOG, whose line for
it ends with its number, such as "(roadmap 7)", and links to its record.
An idea that's dropped is deleted, and the commit that drops it names
its number. Unless a plan says otherwise, its figures are from the first
sample, `ads-20261024` (1,890 distinct ads), and its commands run from
the lab's folder, with `RUN` standing for a run's name (see [Before and
after](docs/PIPELINE.md#before-and-after)).

## Now

### 18. A chunker for any ad

**Serves:** structure, embeddings

The near-term aim: chunking shown working, with grammars. One ad in,
its chunks out, each with its kind, whether it's embedded, kept as
metadata or left out, its size, and what placed it there, down to the
grammar's phrase. In place: `chunk` (text, JSON or YAML); text cleaned,
structure in English not symbols; summary and role apart; sub-headings
before lists found; boilerplate and unmapped text gathered in a chunk
of their own; the view showing the same chunks; and requirements,
responsibilities and benefits placed by their wording, where no
heading placed them (was 19). Still to come, folded in from the
roadmap: the ad's title split off where it runs into a heading or a
line of facts (was idea 15); and pages of ads before and after, which
close it (was 20). Roadmap 7's
facts are in its chunks. The plan, and how it's going, are in
its [record](docs/changes/2026-10-27-a-chunker-for-any-ad.md).

## Next

### 4. Make boilerplate's verdict right

**Serves:** structure, embeddings

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c), widened 2026-10-27
**Waiting on:** nothing

**What:** make the verdict on repeated text right both ways: footers
filed under requirements counted as boilerplate, and common job wording
no longer counted as it.

**Why:** footers (an agency's notices, equal opportunities statements)
often have no heading, so they're filed under the last one, usually
requirements, and under five title groups they're counted as templates
for the job and embedded. The other way, text under five title groups
or more is boilerplate wherever it is, so job lines common to many ads
("High levels of accuracy and attention to detail") are taken out of
what's embedded: about half of a sample of what's taken from summaries
and requirements is about the job.

**What's known:** the largest templates in the first sample's
`boilerplate/repeated.tsv` are almost all such footers. `chunk`'s
boilerplate chunk lists, ad by ad, every line the verdict takes from an
embedded chunk, with the section it came from: where to check it.

**Steps:**

1. Record where each text comes in its ad: whether it's in the run of
   repeated texts that ends the ad; count a repeated text found mostly
   in such end runs as boilerplate, whatever its section.
2. Keep a widespread text that's found mostly in the sections that
   describe the job, and is phrased as a requirement or a duty, out of
   boilerplate.
3. Check the templates under three or more groups, and a sample of the
   boilerplate chunks, before and after.

**Done when** those footers are boilerplate, common job lines aren't, a
sample of the newly flipped texts is checked by eye both ways, tests
cover it, and PIPELINE.md's figures are updated.

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

## Later

Nothing at the moment.

## Ideas

An idea is a numbered heading, the date it was added, and a line or
two on what and why. Each wash-up prunes them: each is kept, promoted
to Later or Next with its number, or dropped.

### 8. Skills and occupations as taxonomy links

Added 2026-10-25, to Ideas 2026-10-27. Link the skills and occupations
an ad names to the existing taxonomies, so they're filtered on, not
embedded: its "Keywords:" lists, now filed as details, and its title
first.

### 9. A fair test of embedding models

Added 2026-10-25, to Ideas 2026-10-27. Rerun the first comparison of
`BAAI/bge-base-en-v1.5` and `intfloat/e5-small-v2` fairly, with E5's
prefixes and many sections, beside other small local models.

### 10. Does it make search better?

Added 2026-10-25, to Ideas 2026-10-27. Measure retrieval: a first slice,
with job titles as queries, is in its
[record](docs/changes/2026-10-27-a-first-retrieval-test.md); judged
queries, a check of boilerplate's verdict, metadata and a managed
service are still to come.

### 11. Embeddings from Go

Added 2026-10-25, to Ideas 2026-10-27. Run the chosen model from Go
with ONNX Runtime, so nothing in the pipeline needs a Python service.

### 12. Boilerplate thresholds scaled to the ads' number

Added 2026-10-27. Boilerplate's thresholds are counts of title groups,
which more ads pass more easily, so its share rises as each batch is
added (12.5% of text on 5,000 ads, 15.1% on 21,000, by word for word
matching): figures from before and after a new batch don't compare.

### 17. Keep a list item's own blocks in the ad's order

Added 2026-10-27. When a list item holds a list or paragraphs of its
own, `internal/blocks` puts its text after the blocks that follow it,
so it lands under the wrong heading: 20 ads of the three samples, which
`cmd/view` counts and shows.
