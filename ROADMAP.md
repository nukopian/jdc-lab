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

Each item has an id, its year and a serial, such as 2026-0007. The year
is when the item was added, and the serial counts from 0001, starting
again each year. An id is given when the item is added, whichever group
it goes in, and is kept as the item moves, so that 2026-0007 means the
same item in a commit message years later. It names the item's change
record, its line in the CHANGELOG ("(roadmap 2026-0007)") and any branch
for it. A `TODO.md` task with a record takes an id from the same
sequence, though it never appears here. Items numbered 1 to 53 before
2026-11-09 kept their numbers as serials, so "roadmap 7" in older
commits is 2026-0007. Last id: 2026-0059.

Each item has a size too, a guess at how much work it is, as a T-shirt
size and its story points: XS 1, S 2, M 3 and L 5. An item bigger than
L is split into items that aren't, each saying what it was split from
("Split from:"). An item that can't be done before another says so
("Depends on:"), and is never in a group ahead of it.

An item whose questions aren't all settled says so, and its record
lists them. Work on it starts only once each is settled, or moved to
another item or task.

When an item ships, it leaves here for the CHANGELOG, whose line for it
ends with its id and links to its record. An item that's dropped is
deleted, and the commit that drops it names its id. Unless a plan says
otherwise, its figures are from the first sample, `ads-20261024` (1,890
distinct ads), and its commands run from the lab's folder, with `RUN`
standing for a run's name (see [Before and after](docs/PIPELINE.md#before-and-after)).

## Now

Nothing at the moment.

## Next

### 2026-0050. An employer's own account of the role isn't boilerplate

**Serves:** structure

**Size:** M (3)

**What:** keep as a template for the job an employer's own account of a
role, repeated across its ads, where boilerplate's verdict now takes it
as a footer or as found elsewhere.

**Why:** roadmap 2026-0033's judged sample found a large employer's
account of the role ("As a Bank Care Assistant, you could be doing
anything from …", "As a Roadside Technician … you'll deliver expert
mechanical roadside repair") among the largest texts taken out of
search: its ads for one role carry titles in different groups, so the
role's own text counts as repeated across jobs. Of the text the verdict
takes out of search, about a third isn't boilerplate alone
([record](docs/changes/2026-0033-boilerplates-verdict-checked.md)).

**Done when** a text that names its ad's own title, or reads as the
role's duties or requirements, is a template, with tests in made-up
text; roadmap 2026-0033's judged texts of this kind come out of
boilerplate, and those judged boilerplate stay; a run against the one
before shows what else moved; and PIPELINE.md's Boilerplate says how.

### 2026-0051. Boilerplate judged sentence by sentence in mixed paragraphs

**Serves:** structure

**Size:** M (3)

**What:** judge a repeated paragraph that mixes standard text with the
job's own terms sentence by sentence, as walls over 400 characters
already are, so the job's sentences keep their own verdict.

**Why:** roadmap 2026-0033 found an agency's standard paragraph with the
job's terms in it ("All pay rates quoted will be inclusive of … holiday
pay. This advert is for a temporary position.") judged whole, and taken
out of search whole: mixed texts were about one in ten of those judged
([record](docs/changes/2026-0033-boilerplates-verdict-checked.md)).

**Done when** a paragraph's sentences are judged on their own where
they're repeated less than the paragraph, with tests in made-up text;
roadmap 2026-0033's mixed texts lose their job sentences from
boilerplate; a run against the one before shows what else moved; and
PIPELINE.md says how.

### 2026-0052. Facts of this job read in boilerplate

**Serves:** metadata

**Size:** S (2)

**What:** read into the details the facts of a sentence found as
boilerplate where it speaks of this job ("this role", "this position",
"this advert"), not where an agency describes itself.

**Why:** the chunker doesn't read the facts of a sentence found as
boilerplate in a section it embeds, so in 1,884 ads a fact never
reaches the details: 2,802 work patterns, 370 pay, 153 schedules. Most
are an agency's ("an employment agency for permanent recruitment and …
temporary workers"), rightly left out, but of 120 judged texts holding
a fact, 45 stated one of the job
([record](docs/changes/2026-0033-boilerplates-verdict-checked.md)).

**Done when** such facts reach the details, with tests in made-up text
for this job's and an agency's; the facts gained are counted, and a
sample of them judged right; and PIPELINE.md's Chunking says how.

### 2026-0053. The setting kept from the company's text

**Serves:** structure

**Size:** M (3)

**What:** keep in the embeddings an employer's account of what it
provides, the kind of school, care or site the job is in, where
boilerplate's verdict takes it from the company's section.

**Why:** roadmap 2026-0033 found the setting ("specialist education for
learners who need …") among the text taken out of search, judged about
the job: a job seeker searching for a special school, or a care home,
searches for it
([record](docs/changes/2026-0033-boilerplates-verdict-checked.md)).
Telling a setting from self-promotion is the hard part.

**Done when** a rule, or a cheap model, tells a setting from
self-promotion on roadmap 2026-0033's judged company texts, with tests;
those judged about the job stay embedded; a run against the one before
shows what moved; and PIPELINE.md says how.

### 2026-0047. Remote read only where the job is worked from home

**Serves:** metadata

**Size:** S (2)

**What:** read the work pattern "remote" only where the job is worked
from home, all or almost all the time, and a few days at home a week as
hybrid.

**Why:** roadmap 2026-0034 found filtering on a fact pays where the
fact is read right, and remote is read wrong: in negations ("This is
NOT a hybrid or remote role"), in other senses ("radio remote and
pendant controls", "24/7 remote GP service", "the remote location of the
site") and for part of the week ("2 days remote"). Of 21 ads whose
facts found it, 19 were judged not remote, so filtering on it found
nothing.

**Done when** those shapes, in made-up text, are tests; roadmap
2026-0034's judgements, scored again, show the facts' remote judged for
where it's found, and filtering on it at least as good as embedding; a
run against the one before shows what else moved; and PIPELINE.md's
Facts says how remote is read.

### 2026-0028. Bold labels on benefit lines aren't headings

**Serves:** structure

**Size:** M (3)

**What:** read a bold label that opens a line of a list set as lines
("**Retail discounts** - Save at 200+ brands"), and a bold line among
such lines, as lines of the list, not headings.

**Why:** each label is read as a heading of no known kind, starting a
section of one line, which the guess places as summary; and since an
ad's chunks of each kind are one, those lines are embedded with the
summary, benefits taken for what the job is. 6 headings in the samples,
from lists roadmap 2026-0024 split out of their items.

**Done when** the labels in the samples' such lists are lines of their
lists, with no heading of their own; tests cover the shapes with
made-up text; a run against the one before shows nothing else moved;
and PIPELINE.md says how headings reads them.

## Later

Nothing at the moment.

## Ideas

An idea is a heading with its id, the date it was added, its size, and
a line or two on what and why. Each wash-up prunes them: each is kept,
promoted to Later or Next with its id, or dropped.

### 2026-0009. A fair test of embedding models

**Added:** 2026-10-25; to Ideas 2026-10-27. **Size:** L (5)

Rerun the first comparison of `BAAI/bge-base-en-v1.5` and
`intfloat/e5-small-v2` fairly, with E5's prefixes and many sections,
beside other small local models.

### 2026-0023. Split a list with no heading by its wording

**Added:** 2026-10-28. **Size:** M (3)

A list that follows a headed one, with no heading of its own, lands
under the heading before it, even when its lines read as another kind:
requirements under "Responsibilities:", embedded as duties. The
openings that already place a line's wording as duties or requirements
could split such a list off, as a section of its own.

### 2026-0025. Split inline lists into items

**Added:** 2026-10-30. **Size:** S (2)

Some ads run a list along one line, its items between bullet symbols
("Free parking • Pension • 25 days' holiday"). Cleaning drops the
symbols, so the list is embedded as one line of text: the symbols could
split it into items instead. 49 ads in the samples.

### 2026-0027. Keep what label emoji mean, wherever they are

**Added:** 2026-10-30. **Size:** M (3)

Cleaning writes a label emoji as its English label only when it opens a
line ("📍 Leeds" is "Location: Leeds"), and drops it anywhere else. So
a line of facts set between emoji ("📍 Bristol | 💰 £26,000-£28,000")
keeps the location's label but loses the pay's, a label emoji after a
bullet gets none, and only one of the 24 clock faces is mapped (about
56 lines in the samples open with another). The other way, a mapped
emoji used as decoration gets a label it shouldn't ("📞 Participate in
an on-call rota" would be a contact): a label could wait for text that
reads as a value of it.

### 2026-0029. Keywords and titles read as fields

**Added:** 2026-10-30. **Size:** S (2). **Split from:** 2026-0008
(skills and occupations as taxonomy links, added 2026-10-25)

Read an ad's "Keywords:" lists, now filed as details, and its title as
fields of their own, the first step to linking them to the existing
taxonomies, so that they're filtered on, not embedded.

### 2026-0030. Skills linked to a skills taxonomy

**Added:** 2026-10-30. **Size:** L (5). **Split from:** 2026-0008

Match the skills an ad names, in its keywords and its requirements, to
an existing skills taxonomy, so they're filtered on, not embedded.
Needs the taxonomy's data in the lab.

### 2026-0031. An occupation coded from the title

**Added:** 2026-10-30. **Size:** L (5). **Split from:** 2026-0008

Code each ad's occupation from its title against an existing occupation
taxonomy, to filter on, as titles embed well but don't group by
occupation.

### 2026-0032. A set of judged queries

**Added:** 2026-10-30. **Size:** L (5). **Split from:** 2026-0010 (does
it make search better?, added 2026-10-25; a first slice, with job titles
as queries, is in its
[record](docs/changes/2026-0010-a-first-retrieval-test.md))

Queries with the ads people judge right for them, so retrieval is
measured by more than titles. Needs people's judgements.

### 2026-0035. Compared with a managed embedding service

**Added:** 2026-10-30. **Size:** L (5). **Split from:** 2026-0010

Run the retrieval test against a managed service's embeddings too, to
see whether local models are competitive. Needs access to one.

### 2026-0036. The model's tokenizer in Go

**Added:** 2026-10-30. **Size:** L (5). **Split from:** 2026-0011
(embeddings from Go, added 2026-10-25)

Tokenize as the chosen model does, in Go, matching Python's tokenizer
token for token, so nothing in the pipeline needs a Python service.

### 2026-0037. The model run from Go

**Added:** 2026-10-30. **Size:** M (3). **Split from:** 2026-0011.
**Depends on:** 2026-0036

Run the chosen model with ONNX Runtime from Go, pooling (CLS or mean, as
the model expects) and normalising its output.

### 2026-0038. Go's vectors held to Python's

**Added:** 2026-10-30. **Size:** S (2). **Split from:** 2026-0011.
**Depends on:** 2026-0036 and 2026-0037

Tests that Go's vectors match Python's for the same text, within a
tolerance, as the ONNX demo's scores are held in all three languages.

### 2026-0039. Vectors ready for a search engine's dense field

**Added:** 2026-10-30. **Size:** S (2). **Split from:** 2026-0011.
**Depends on:** 2026-0037

Vectors normalised for an Elasticsearch `dense_vector` field with
`dot_product`, and quantised (`int8_hnsw`) to cut its memory.

### 2026-0040. Sections the rules can't place, classified by their text

**Added:** 2026-10-30. **Size:** L (5)

A small classifier, trained on the sections the rules already place,
reads the text under a heading of no known kind, and the lines of a
list with no heading, and says which kind it is, where it's confident;
the guess still decides the rest. It has to beat the guess's 39 to 43
in 50, on text the rules couldn't place. Run from Go, it needs
2026-0036 to 2026-0038; used offline, to suggest phrases for
`headings.txt`, it needs none of them.

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

### 2026-0041. Cheap local ways to place a section, beside a model

**Added:** 2026-10-30. **Size:** L (5), for choosing among them; each
taken up becomes an item of its own

Ways to place what the rules can't, each learning from the rules' own
labels, most promising first:

- **Openings mined by kind:** count the first two to four words of the
  lines in placed sections, and rank each opening by how often it
  occurs and how consistently it belongs to one kind, to suggest
  entries for `openings.txt`, as `heading_suggest.py` does for
  `headings.txt`.
- **Word counts, not embeddings:** a naive Bayes or logistic regression
  classifier on words and pairs of words (TF-IDF), in Go: the baseline
  any model has to beat (2026-0040).
- **Place in the ad:** sections come in a fairly fixed order, so a
  sequence model (an HMM, decoded with Viterbi) can place an unmapped
  section by the kinds around it.
- **Static embeddings:** a model distilled into a table of token
  vectors (Model2Vec's `minishlab/potion-base-8M`), embedding by lookup
  and average, without ONNX Runtime; one for the fair test (2026-0009).
- **Names found before embedding:** a capitalised run found in mapped
  headings such as "Why X" and "About X", or across an advertiser's
  ads, left out or stripped, to stop `heading_suggest.py` giving names
  a kind.
- **Taxonomy hits as evidence:** many skills in a section point to
  requirements (2026-0030), benefit words to benefits.
- **Facts as evidence:** pay in a sentence, beside holidays or a
  pension, points to benefits, and perks are among the guess's
  misses; hours and patterns point to the role. Pay alone is
  weak, as ads open with it too, so first count, in placed sections,
  which kind each fact in a sentence sits in, and how many unmapped
  sections have one. A fact that is mostly one kind's becomes evidence
  in the guess, as openings are; one that isn't is a clue for place in
  the ad.
- **A judged set of sections:** sections judged by hand, chosen where
  the ways above disagree most, to measure any of them on text the
  rules couldn't place.

### 2026-0042. Codes and straplines aren't headings

**Added:** 2026-10-31. **Size:** S (2)

Job-board codes ("INDPSAL", "#LI-KC1") on a line of their own, in bold,
are read as headings, in about 450 ads in the samples, and so are
straplines of terms between bars ("Permanent | Full-time | 37.5 hours
per week"), in about 400, some of them facts. Most are of no known
kind, starting a section the guess places; one with a kind's word in it
("Application Support | SQL") starts that kind's section, taking the
ad's opening with it. Read a line of codes as text, and one of three or
more terms between bars as text, or as facts. Found by roadmap
2026-0026, which reads those set in italic alone, of no known kind, as
text.

### 2026-0044. Hours per month and per fortnight

**Added:** 2026-11-02. **Size:** XS (1)

Hours knows only a day and a week, so "168 hours per month" and "84
Hours per Fortnight" are read as bare hours, and, over 80, dropped as
implausible. Give Hours a month and a fortnight, with limits of their
own, about 744 and 336, so such hours are read with their period.

### 2026-0046. Divide each ad once for every stage

**Added:** 2026-11-03. **Size:** M (3)

With the stages run at once, a run keeps the machine's processors busy,
about 60 s of their time in 12 s, so it's faster only for less work.
Four of the five stages divide every ad into blocks, headings and
sections again, and boilerplate and view each build the same index of
every ad. Divide each ad once, for every stage, either by a stage whose
output the others read, which the stages don't yet do, or by one
program for the stages that share the most. Found by roadmap 2026-0045.

### 2026-0048. Pay and hours as filters in the retrieval test

**Added:** 2026-11-03. **Size:** M (3)

Roadmap 2026-0034 tested place and work pattern as filters, and both
paid where the fact is read right. Pay and hours want queries with a
range ("Warehouse Operative over £30,000", "under 20 hours a week"), and
pay a period to compare across ("£12 an hour" against "£25,000 a
year"). Judged as roadmap 2026-0034's work patterns were, or by the
fact's own text where it's plain.

### 2026-0049. Retrieval with boilerplate left out, on judged queries

**Added:** 2026-11-03. **Size:** M (3). **Split from:** 2026-0033.
**Depends on:** 2026-0032

Whether leaving what boilerplate's verdict takes out of the embeddings
costs retrieval, for queries a title can't answer: skills, duties,
conditions. The first retrieval test found nothing lost for job titles,
which can't see it. Waits on idea 2026-0032's judged queries.
