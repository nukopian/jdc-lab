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
number: 42.

Each item has a size too, a guess at how much work it is, as a T-shirt
size and its story points: XS 1, S 2, M 3 and L 5. An item bigger than
L is broken into items that aren't, each saying what it was split from.
Items that can't be done one without another say so ("Moves with:"),
and move between groups together.

When an item ships, it moves from here to the CHANGELOG, whose line for
it ends with its number, such as "(roadmap 7)", and links to its record.
An idea that's dropped is deleted, and the commit that drops it names
its number. Unless a plan says otherwise, its figures are from the first
sample, `ads-20261024` (1,890 distinct ads), and its commands run from
the lab's folder, with `RUN` standing for a run's name (see [Before and
after](docs/PIPELINE.md#before-and-after)).

## Now

Nothing at the moment.

## Next

### 28. Bold labels on benefit lines aren't headings

**Serves:** structure

**Size:** M (3)

**What:** read a bold label that opens a line of a list set as lines
("**Retail discounts** - Save at 200+ brands"), and a bold line among
such lines, as lines of the list, not headings.

**Why:** each label is read as a heading of no known kind, starting a
section of one line, which the guess places as summary; and since an
ad's chunks of each kind are one, those lines are embedded with the
summary, benefits taken for what the job is. 6 headings in the samples,
from lists roadmap 24 split out of their items.

**Done when** the labels in the samples' such lists are lines of their
lists, with no heading of their own; tests cover the shapes with
made-up text; a run against the one before shows nothing else moved;
and PIPELINE.md says how headings reads them.

## Later

Nothing at the moment.

## Ideas

An idea is a numbered heading, the date it was added, its size, and a
line or two on what and why. Each wash-up prunes them: each is kept,
promoted to Later or Next with its number, or dropped.

### 9. A fair test of embedding models

Added 2026-10-25, to Ideas 2026-10-27. **Size:** L (5). Rerun the first
comparison of `BAAI/bge-base-en-v1.5` and `intfloat/e5-small-v2`
fairly, with E5's prefixes and many sections, beside other small local
models.

### 12. Boilerplate thresholds scaled to the ads' number

Added 2026-10-27. **Size:** M (3). Boilerplate's thresholds are counts
of title groups, which more ads pass more easily, so its share rises as
each batch is added (12.5% of text on 5,000 ads, 15.1% on 21,000, by
word for word matching): figures from before and after a new batch
don't compare.

### 23. Split a list with no heading by its wording

Added 2026-10-28. **Size:** M (3). A list that follows a headed one,
with no heading of its own, lands under the heading before it, even
when its lines read as another kind: requirements under
"Responsibilities:", embedded as duties. The openings that already
place a line's wording as duties or requirements could split such a
list off, as a section of its own.

### 25. Split inline lists into items

Added 2026-10-30. **Size:** S (2). Some ads run a list along one line,
its items between bullet symbols ("Free parking • Pension • 25 days'
holiday"). Cleaning drops the symbols, so the list is embedded as one
line of text: the symbols could split it into items instead. 49 ads in
the samples.

### 26. Italic as a cue, like bold

Added 2026-10-30. **Size:** XS (1). Blocks records a paragraph all in
bold, but not one all in `<em>` or `<i>`, though ads set headings
("You'll also need to:") and notices in italic too. 470 ads in the
samples have one.

### 27. Keep what label emoji mean, wherever they are

Added 2026-10-30. **Size:** M (3). Cleaning writes a label emoji as its
English label only when it opens a line ("📍 Leeds" is "Location:
Leeds"), and drops it anywhere else. So a line of facts set between
emoji ("📍 Bristol | 💰 £26,000-£28,000") keeps the location's label but
loses the pay's, a label emoji after a bullet gets none, and only one
of the 24 clock faces is mapped (about 56 lines in the samples open
with another). The other way, a mapped emoji used as decoration gets a
label it shouldn't ("📞 Participate in an on-call rota" would be a
contact): a label could wait for text that reads as a value of it.

### 29. Keywords and titles read as fields

Added 2026-10-30, split from 8 (skills and occupations as taxonomy
links, added 2026-10-25). **Size:** S (2). Read an ad's "Keywords:"
lists, now filed as details, and its title as fields of their own, the
first step to linking them to the existing taxonomies, so that they're
filtered on, not embedded.

### 30. Skills linked to a skills taxonomy

Added 2026-10-30, split from 8. **Size:** L (5). Match the skills an ad
names, in its keywords and its requirements, to an existing skills
taxonomy, so they're filtered on, not embedded. Needs the taxonomy's
data in the lab.

### 31. An occupation coded from the title

Added 2026-10-30, split from 8. **Size:** L (5). Code each ad's
occupation from its title against an existing occupation taxonomy, to
filter on, as titles embed well but don't group by occupation.

### 32. A set of judged queries

Added 2026-10-30, split from 10 (does it make search better?, added
2026-10-25; a first slice, with job titles as queries, is in its
[record](docs/changes/2026-10-27-r10-a-first-retrieval-test.md)).
**Size:** L (5). Queries with the ads people judge right for them, so
retrieval is measured by more than titles. Needs people's judgements.

### 33. Retrieval with boilerplate's verdict checked

Added 2026-10-30, split from 10. **Size:** M (3). Measure whether
leaving out what boilerplate's verdict takes costs retrieval anything,
before anything is dropped.

### 34. Metadata as filters in the retrieval test

Added 2026-10-30, split from 10. **Size:** M (3). Use the facts read as
metadata (pay, hours, work pattern, place) as filters beside the
embeddings, and measure what they add.

### 35. Compared with a managed embedding service

Added 2026-10-30, split from 10. **Size:** L (5). Run the retrieval test
against a managed service's embeddings too, to see whether local models
are competitive. Needs access to one.

### 36. The model's tokenizer in Go

Added 2026-10-30, split from 11 (embeddings from Go, added 2026-10-25).
**Size:** L (5). **Moves with:** 37, 38 and 39. Tokenize as the chosen
model does, in Go, matching Python's tokenizer token for token, so
nothing in the pipeline needs a Python service.

### 37. The model run from Go

Added 2026-10-30, split from 11. **Size:** M (3). **Moves with:** 36, 38
and 39. Run the chosen model with ONNX Runtime from Go, pooling (CLS or
mean, as the model expects) and normalising its output.

### 38. Go's vectors held to Python's

Added 2026-10-30, split from 11. **Size:** S (2). **Moves with:** 36, 37
and 39. Tests that Go's vectors match Python's for the same text, within
a tolerance, as the ONNX demo's scores are held in all three languages.

### 39. Vectors ready for a search engine's dense field

Added 2026-10-30, split from 11. **Size:** S (2). **Moves with:** 36, 37
and 38. Vectors normalised for an Elasticsearch `dense_vector` field
with `dot_product`, and quantised (`int8_hnsw`) to cut its memory.

### 40. Sections the rules can't place, classified by their text

Added 2026-10-30. **Size:** L (5). A small classifier, trained on the
sections the rules already place, reads the text under a heading of no
known kind, and the lines of a list with no heading, and says which
kind it is, where it's confident; the guess still decides the rest. It
has to beat the guess's 39 to 43 in 50, on text the rules couldn't
place. Run from Go, it needs 36 to 38; used offline, to suggest phrases
for `headings.txt`, it needs none of them.

The offline use has had a first try, `python/bin/heading_suggest.py`:
on the three samples together, `BAAI/bge-small-en-v1.5` embeds 13,303
mapped headings and 9,638 unmapped ones in under a minute, and each
unmapped heading's ten nearest mapped ones vote for its kind. The vote
gives back the rules' own kind for 85% of mapped headings, and 99% of
those with nine-tenths of the vote or more, but that flatters it, as their neighbours
share the phrases the rules matched. Read by eye, 32 of 40 unmapped
headings with nine-tenths of the vote were given the kind a reader would:
missing phrases ("what do you need", "to be considered you must",
"caring and investing in you") and pay, places and hours read as
headings. The misses are names of advertisers, slogans, job titles and
codes, given a kind with as much confidence, after a mapped heading
with the same name and a phrase ("why" and the name). Its suggestions
are worth reading; applied as they stand, they'd be no better than the
guess. Next: leave out headings that are mostly a name, and let a weak
nearest neighbour say no kind. `BAAI/bge-base-en-v1.5` took three times
as long and agreed on 71% of unmapped headings, for no clear gain. A
section's text, not its heading, is still the part to classify.

### 41. Cheap local ways to place a section, beside a model

Added 2026-10-30. **Size:** L (5), for choosing among them; each taken
up becomes an item of its own. Ways to place what the rules can't, each
learning from the rules' own labels, most promising first:

- **Openings mined by kind:** count the first two to four words of the
  lines in placed sections, and rank each opening by how often it
  occurs and how consistently it belongs to one kind, to suggest
  entries for `openings.txt`, as `heading_suggest.py` does for
  `headings.txt`.
- **Word counts, not embeddings:** a naive Bayes or logistic regression
  classifier on words and pairs of words (TF-IDF), in Go: the baseline
  any model has to beat (40).
- **Place in the ad:** sections come in a fairly fixed order, so a
  sequence model (an HMM, decoded with Viterbi) can place an unmapped
  section by the kinds around it.
- **Static embeddings:** a model distilled into a table of token
  vectors (Model2Vec's `minishlab/potion-base-8M`), embedding by lookup
  and average, without ONNX Runtime; one for the fair test (9).
- **Names found before embedding:** a capitalised run found in mapped
  headings such as "Why X" and "About X", or across an advertiser's
  ads, left out or stripped, to stop `heading_suggest.py` giving names
  a kind.
- **Taxonomy hits as evidence:** many skills in a section point to
  requirements (30), benefit words to benefits.
- **A judged set of sections:** sections judged by hand, chosen where
  the ways above disagree most, to measure any of them on text the
  rules couldn't place.
