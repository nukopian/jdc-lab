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
number: 29.

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

Nothing at the moment.

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
[record](docs/changes/2026-10-27-r10-a-first-retrieval-test.md); judged
queries, a check of boilerplate's verdict, metadata and a managed
service are still to come.

### 11. Embeddings from Go

Added 2026-10-25, to Ideas 2026-10-27. Run the chosen model from Go
with ONNX Runtime, so nothing in the pipeline needs a Python service. Go
does the tokenizing, pooling (CLS or mean, as the model expects) and
normalising, so the vectors suit an Elasticsearch `dense_vector` field
with `dot_product`, and quantised (`int8_hnsw`) to cut its memory.

### 12. Boilerplate thresholds scaled to the ads' number

Added 2026-10-27. Boilerplate's thresholds are counts of title groups,
which more ads pass more easily, so its share rises as each batch is
added (12.5% of text on 5,000 ads, 15.1% on 21,000, by word for word
matching): figures from before and after a new batch don't compare.

### 23. Split a list with no heading by its wording

Added 2026-10-28. A list that follows a headed one, with no heading of
its own, lands under the heading before it, even when its lines read as
another kind: requirements under "Responsibilities:", embedded as
duties. The openings that already place a line's wording as duties or
requirements could split such a list off, as a section of its own.

### 25. Split inline lists into items

Added 2026-10-30. Some ads run a list along one line, its items between
bullet symbols ("Free parking • Pension • 25 days' holiday"). Cleaning
drops the symbols, so the list is embedded as one line of text: the
symbols could split it into items instead. 49 ads in the samples.

### 26. Italic as a cue, like bold

Added 2026-10-30. Blocks records a paragraph all in bold, but not one
all in `<em>` or `<i>`, though ads set headings ("You'll also need to:")
and notices in italic too. 470 ads in the samples have one.

### 27. Keep what label emoji mean, wherever they are

Added 2026-10-30. Cleaning writes a label emoji as its English label
only when it opens a line ("📍 Leeds" is "Location: Leeds"), and drops
it anywhere else. So a line of facts set between emoji ("📍 Bristol |
💰 £26,000-£28,000") keeps the location's label but loses the pay's, a
label emoji after a bullet gets none, and only one of the 24 clock faces
is mapped (about 56 lines in the samples open with another). The
other way, a mapped emoji used as decoration gets a label it shouldn't
("📞 Participate in an on-call rota" would be a contact): a label could
wait for text that reads as a value of it.

### 28. Bold labels on benefit lines aren't headings

Added 2026-10-30. A benefits list set as lines, each opening with a bold
label ("**Retail discounts** - Save at 200+ brands"), or one bold line
among them, has its labels read as headings of no known kind, each
starting a section of one line. Headings could read a bold lead
followed by " - ", amid lines like it, as a line of the list. 6 headings
in the samples, since roadmap 24 split such lists out of their items.
