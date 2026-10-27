# How the pipeline works

How the lab turns a job ad into its parts: which text is a heading, a
field, a section, or boilerplate. This is the working answer to the
lab's first questions: what sections ads have, how their headings vary,
how many can be recognised without AI, and how much of an ad is worth
embedding at all.

It covers each stage in order, how to run it, what it decided and why,
what it measured, and what it doesn't do yet, with an item in the
[roadmap](../ROADMAP.md) for each. Keep it up to date: a
change to a stage changes this document in the same commit
([MAINTAINING.md](MAINTAINING.md#before-committing)).

## The flow

```
data/ingestion/extract-DATE.xml   an export, as received
        │  cmd/xml-to-jsonl        internal/mysqlxml
        ▼
data/ingestion/ads-DATE.jsonl     a sample: one ad per line: id, title, location, html
        │
        │  internal/blocks         HTML → blocks, with cues      (cmd/blocks)
        │  internal/headings       blocks → parts, with roles    (cmd/heading-counts)
        │  internal/sections       parts → sections, by kind     (cmd/sections)
        │  internal/boilerplate    sections → boilerplate        (cmd/boilerplate)
        ▼
data/runs/RUN/SAMPLE/  blocks/  headings/  sections/  boilerplate/
```

Each stage is a package in `internal/`, with tests and examples beside
it, and each has a program in `cmd/` that runs it over every ad, writes
what it found to a folder, and prints a summary. The programs don't
read each other's output: each reads a sample and runs the stages before
its own, which are fast. So any folder in `data/` except `ingestion/`
can be deleted and rebuilt.

A **sample** is one export's ads, converted: a `.jsonl` file in
`data/ingestion/`, named by the export's date. Samples are kept apart,
not merged, so that a rule tuned on one can be checked on another.
`cmd/pipeline` runs every stage on every sample, and on all of them
together as the sample `all`, and keeps each **run** under a name of its
own in `data/runs/`.

Every stage only adds to what the one before it found, and never changes
the ad's text. Text is kept as written, and only a separate **key** (see
[Keys](#keys)) is normalised for matching.

## Running it

From the lab's folder. Convert each new export into a sample, named by
its date:

```
go run ./cmd/xml-to-jsonl data/ingestion/extract-20261025.xml data/ingestion/ads-20261025.jsonl
```

Then run every stage on every sample, as a run called `now`:

```
go run ./cmd/pipeline run now
```

It takes a few seconds, and prints where each sample's output went:
`data/runs/now/ads-20261025/sections/` holds the sections stage's tables
for that sample, and `data/runs/now/ads-20261025/sections.txt` its
summary. A run of the same name is replaced, and
`data/runs/now/all/` has the samples together.

Each stage's program can also be run alone, on one sample, into any
folder:

```
go run ./cmd/blocks         data/ingestion/ads-20261025.jsonl data/blocks
go run ./cmd/heading-counts data/ingestion/ads-20261025.jsonl data/headings
go run ./cmd/sections       data/ingestion/ads-20261025.jsonl data/sections
go run ./cmd/boilerplate    data/ingestion/ads-20261025.jsonl data/boilerplate
```

`cmd/heading-uses` is for looking at one heading across a sample's ads,
when improving the stages (see [Improving it](#improving-it)).
`go install ./cmd/...` puts the programs in `bin/`, on the lab's `PATH`,
so they can be run by name. Ads whose HTML is the same as an earlier
one's are counted once by every stage after blocks.

## 1. Ingestion: the export to JSON Lines

**Package** `internal/mysqlxml`, **program** `cmd/xml-to-jsonl`.

The ads arrive as a MySQL `--xml` export: a `<row>` per ad, a `<field>`
per column. XML is awkward to work with, so it's converted once to JSON
Lines, one ad per line, with the lab's own field names (`internal/ads`
holds the `Ad` type, and reads the file for a `range` loop):

| Field | What it is |
|---|---|
| `id` | the ad's reference |
| `title` | its job title |
| `location`, `location_canonical` | where it is, as written and as a place name; `null` if missing |
| `html` | the description, as HTML |

Nothing else changes, except that XML parsing turns each CRLF into LF.
Duplicates stay in: later stages leave them out of their counts.

## 2. Blocks: what a reader sees

**Package** `internal/blocks`, **program** `cmd/blocks`.

`blocks.Split` reads the HTML with Go's HTML tokenizer and returns the
blocks a reader would see, in order:

| Kind | What it is |
|---|---|
| `paragraph` | a block of text on its own: a `<p>`, a `<div>`, a heading tag, or loose text |
| `line` | one of several lines a `<br>` split a block into |
| `list` | consecutive list items, each an item of its own |

Each block keeps its text as a browser shows it (entities decoded, runs
of spaces made one), and its **cues**: what the markup says that the text
doesn't.

| Cue | Example |
|---|---|
| `heading` | the text was in `<h1>` to `<h6>` |
| `bold` | all of it was in `<strong>` or `<b>` |
| `bold_lead` | it starts with bold text: `Salary:` in "**Salary:** £30,000" |

It decides nothing. Whether a bold block is a heading is the next stage's
call, which keeps the HTML parsing apart from the judgement, so either
can change without the other.

**Plain text.** Some ads have no tags that lay out blocks. There, a
newline ends a line and a blank line ends a block. A **wall** of text
with no newlines at all stays one paragraph: splitting it needs
sentences, a later stage.

On the sample, no ad uses heading tags: headings are bold paragraphs, or
lines ending with a colon. 80 ads have a paragraph over 1,500 characters.

## 3. Headings: what each block is

**Package** `internal/headings`, **program** `cmd/heading-counts`.

`headings.Read` gives each block, and each list item, a **role**, and
returns them as **parts**:

| Role | How it's recognised | Example |
|---|---|---|
| `heading` | short (`MaxWords`, 10, or `MaxIntroWords`, 20, with a cue and ending in `:`), with a cue (heading tag, all bold) or ending in `:` or `;`; a word that isn't a conjunction; no money or long numbers; not ending in `.` | "**Key Responsibilities**" |
| `field` | "Label: value" or "Label \| value", where the label is a known one | "Hours: 40 per week" |
| `repeat` | only the ad's title or location again | "**Delivery Driver**" at the top |
| `text` | everything else | |

A few rules refine this:

- **A bold lead that isn't a field** is a heading run into its text:
  "**Key Responsibilities:** You will…" is a heading, then text.
- **A field's label written as a heading,** with a short value (up to
  `MaxValueWords`, 15) on the next line, is joined into one field:
  "**Working Hours**", then "Monday to Friday".
- **A long line in bold that ends with a colon** introduces what follows,
  as a heading does, so it can have up to `MaxIntroWords` (20) words:
  "**As a Mobile Mechanic, you'll get benefits that go the extra mile:**".
  Without the colon, or the bold, it's a sentence.
- **Conjunctions alone aren't headings.** "**OR**", "**and**",
  "**Either:**" or "**Plus:**" on a line of their own join what's before
  them to what's after, so the section carries on.
- **List items are never headings.** A bold item is far more often a
  point made strongly than a heading.
- **Field labels are a list** (`fieldLabels`), so that a labelled bullet
  such as "Financial reporting: Oversee the accounts" isn't taken for
  data.

Each rule errs towards text: a heading missed is only counted as text,
but text taken for a heading would start a section that isn't there.

`cmd/heading-counts` writes `headings.tsv` and `fields.tsv`, how many ads
use each heading and field label, by key, with its commonest spelling.

### Keys

`headings.Key` is the form of text used for matching and counting: lower
case; typographic quotes, dashes and odd spaces folded to plain ones;
bullets, emoji and invisible characters dropped; punctuation other than
`' & /` made a space. "What You’ll Be Doing:" and "what you'll be doing"
have the same key. The table of characters it folds includes those these
ads use, such as Word's bullet and the emoji variation selector.

## 4. Sections: what each heading starts

**Package** `internal/sections`, **program** `cmd/sections`.

`sections.Divide` divides an ad's parts into **sections**. What comes
before the first heading is the summary, and each heading starts a
section of its **kind**:

| Kind | Holds | Example headings |
|---|---|---|
| `summary` | what comes first, and role overviews | "The role", "The opportunity" |
| `responsibilities` | what the job involves | "What you'll be doing", "Main duties" |
| `requirements` | what the job needs | "About you", "Essential criteria" |
| `desirable` | what would help | "Desirable", "Nice to have" |
| `benefits` | what the job offers | "Why join us?", "What's on offer" |
| `company` | who the employer is | "About us", "About the company" |
| `apply` | how to apply | "How to apply", "Interested?" |
| `legal` | notices | "Equal opportunities", "Safeguarding" |
| `details` | the job's terms | "Job details", "Working Hours" |
| `unmapped` | under a heading of no known kind | anything not yet recognised |

**The phrases are data.**
[`internal/sections/headings.txt`](../internal/sections/headings.txt)
holds synonym rules, in the format Elasticsearch and Solr use: each the
kind first, then the phrases that mean it.

```
benefits, benefit, perk, offer, offering, what's on offer, on offer,
	what we offer, in return, what's in it for you, ...
```

A `grammar` word list (github.com/nukopian/grammar) finds the phrases
anywhere in a heading's key, as whole words, in any case, and in the
plural ("duty" finds "Main Duties"). So adding a synonym means adding a
phrase to the file, not changing code.

**A kind's name is a phrase too,** as "benefits" is in the rule above,
unless its rule is written the other way round, phrases first and
`=> kind` at the end:

```
equal opportunity, equal opportunities, diversity, ...,
	legal information, the legal bit => legal
```

Legal's rule is, because "legal" in a heading more often names a job's
field than a notice: "The successful Legal Secretary will have:" isn't
one. `internal/sections` asks grammar to read `=>` so, with its
`MapOnly` option: by default, as in Elasticsearch, the name would be a
phrase too.

**"Why" first.** A heading starting "why" is benefits, whatever phrases
are in it: it's a pitch to the reader, as in "Why join us?", and also
"Why you'll love this role" or "Why should you apply?", whose phrases
would otherwise make them responsibilities or apply.

**Choosing between kinds.** Where a heading has phrases of two kinds, the
kind whose rule comes first in the file wins. The file's order is the
ranking: desirable, benefits, legal, requirements, responsibilities,
company, apply, details, summary. So "Desirable skills" is desirable, not
requirements, and "What our client offers" benefits, not company.

**Phrases that overlap all count.** In "To succeed in this role, you will
need:", both "in this role you will" (responsibilities) and "you will
need" (requirements) are found, so requirements wins; read only left to
right, the first would have hidden the second. A phrase wholly inside a
longer one doesn't count: "Company Profile" is company, though
"profile" alone is requirements. grammar's `Outermost` option finds
them so.

**With no phrase**, a heading starting "about" is company ("About" and a
company's name), and one that's a field's label is details. Any other starts an `unmapped` section: so the text
under it counts as unmapped, rather than hiding in the section before,
and unmapped's share of the text measures what's left to map.

**A plain line** with no cue starts a section only if it's exactly a
kind's phrase: `<p>Responsibilities</p>`, not "Our responsibilities to
you".

`cmd/sections` writes each ad's sections to `ads.jsonl` (a section
started by a plain line is marked `by_name`), and three tables
to work from: `unmapped.tsv` (headings of no known kind, with the text
under each), `disagree.tsv` (headings whose phrases point to two kinds,
and the kind that won), and `named.tsv` (plain lines taken by name).
[Improving it](#improving-it) has how to use them.

## 5. Boilerplate: what ads repeat whatever the job

**Package** `internal/boilerplate`, **program** `cmd/boilerplate`.

Boilerplate is text about the advertiser, not the job: its description of
itself, its standard benefits, its notices. The ads don't say who placed
them, so it's found by **repetition**: a text of `MinWords` (8) words or
more, word for word by its key, in ads for different jobs.

**Different jobs, not different ads.** An advertiser posting one job in
ten towns repeats every block of it, and none of that is boilerplate. So
what counts is how many **groups of similar titles** a text appears
under. Two titles are one group if one's words are all in the other, or
they share at least half their words: "Support Worker" and "Residential
Support Worker" are one group; "Crane Operator" and "Forklift Driver" are
two. A text under `MinGroups` (2) groups or more is **repeated**.

**Boilerplate or template.** Related jobs from one advertiser also share
lines that describe the job, such as "Excellent ICT skills", which are
worth keeping. So a repeated text is:

- a **template** for the job if it's found at least as often in the
  sections that describe the job (responsibilities, requirements,
  desirable) as outside them;
- **boilerplate** if it's found more often outside them, or if it's under
  `Widespread` (5) groups or more, wherever it is.

The verdict belongs to the text, over every ad it's in, so a heading
misfiled in one ad doesn't change it. The widespread rule catches
footers, which often come after an ad's last heading and so land under
its requirements, and lines such as "Able to work as part of a team",
which, though about the job, are in so many jobs that they say nothing of
any: as with a common word in a search index.

`cmd/boilerplate` writes `repeated.tsv`: every repeated text, with its
groups, ads, characters, the kinds of section it's in, and its verdict.

## Improving it

Each program leaves tables in `data/` that say where its stage falls
short. Improving the pipeline is a loop over them:

1. **Measure:** `go run ./cmd/pipeline run before`, every stage on
   every sample (see [Before and after](#before-and-after)).
2. **Pick** the biggest gap: the table row with the most ads or text.
   Work on one sample, and check the others as you go.
3. **Look** at it in the ads, in context, before deciding anything.
4. **Change one thing:** a phrase, a label, a rule, a threshold.
5. **Measure again,** as the run `after`, and compare. Keep the change
   only if it helped and broke nothing, on every sample: one that helps
   one sample and harms another fits that sample, not job ads.
6. **Pin it:** add a test case for what it settled, and update the
   figures in [What it measured](#what-it-measured).

### Questions, and where to look

All from the lab's folder, after a run (see
[Running it](#running-it)). The commands look at one sample of the
run `before`, which these names choose; set them first, to the sample
and run you want, and each command then runs as it stands:

```
sample=ads-20261025
r=data/runs/before/$sample
ads=data/ingestion/$sample.jsonl
```

For the samples together, `sample=all`, and their ads are in the run:
`ads=$r/ads.jsonl`.

**Which headings are worth working on?** The unmapped headings with the
most text under them:

```
sort -t$'\t' -k2,2nr $r/sections/unmapped.tsv | head -20
```

**What is this heading, really?** `cmd/heading-uses` shows every ad that
uses a heading: how it's spelt, what section comes before it, and the
first lines under it. Give it the heading's key (the third column of
`unmapped.tsv`), or the heading as written; this takes the biggest
from the table:

```
top=$(sort -t$'\t' -k2,2nr $r/sections/unmapped.tsv | head -1 | cut -f3)
go run ./cmd/heading-uses "$ads" "$top"
```

or one by name:

```
go run ./cmd/heading-uses "$ads" "please note"
```

**Did a heading go to the wrong kind?** `disagree.tsv` lists headings
whose phrases point to more than one kind, with the kind that won:

```
head -20 $r/sections/disagree.tsv
```

**Is a plain line being taken for a heading?** `named.tsv` lists the
lines without cues that started a section because they're exactly a
phrase:

```
head -20 $r/sections/named.tsv
```

**Is something taken for a heading that isn't one?** Headings with
figures in are often times, steps or references:

```
awk -F'\t' 'NR > 1 && $2 ~ /[0-9]/' $r/headings/headings.tsv | head -20
```

**Is a field label missing?** Short unmapped headings ending in a colon
are candidates; `fields.tsv` has the labels already known, by use:

```
awk -F'\t' 'NR > 1 && $4 ~ /:$/ && split($3, w, " ") <= 2' $r/sections/unmapped.tsv | head -20
head -20 $r/headings/fields.tsv
```

**Why did an ad come out as it did?** Take its id (here the first from
`heading-uses`), then see its blocks, and its HTML, to tell a block
split wrongly from a heading judged wrongly:

```
id=$(go run ./cmd/heading-uses "$ads" "please note" | awk '$1 ~ /^[0-9]+$/ {print $1; exit}')
jq -c --arg id "$id" 'select(.id == $id) | .blocks[]' $r/blocks/ads.jsonl | head -20
jq -r --arg id "$id" 'select(.id == $id) | .html' "$ads" | head -c 3000
```

**Is a boilerplate verdict wrong?** Templates under three or more title
groups are worth a look, as footers filed under requirements end up
there:

```
awk -F'\t' '$5 == "template" && $1 >= 3' $r/boilerplate/repeated.tsv | cut -f1-4,6 | head -20
```

**Is boilerplate hiding in near-duplicates?** Texts that share their
opening but aren't word for word the same are variants of one text.
Compare how many openings and how many whole texts recur in five ads or
more; the difference is the near-duplicates:

```
jq -r '.sections[].parts[]? | select(.role == "text") | .text' $r/sections/ads.jsonl |
  awk 'NF >= 8' | cut -c1-50 | sort | uniq -c | sort -rn | awk '$1 >= 5' | head -20
```

### What to change

| What you found | What to change | Where |
|---|---|---|
| A heading whose words mean a kind | Add the phrase to that kind's rule; add the heading to `TestOf` | `internal/sections/headings.txt` |
| The wrong kind won | Move the winning kind's rule below the right one's, or make the phrase longer | `headings.txt` |
| A plain line wrongly taken by name | Make the phrase longer, or drop it | `headings.txt` |
| A field's label not known | Add it to `fieldLabels`; add a case to `TestRead` | `internal/headings/headings.go` |
| Text taken for a heading, or a heading missed | Change the rule (`headingShaped`, `MaxWords`, `MaxIntroWords`, `conjunctions`, `money`); add a case to `TestRead` | `internal/headings/headings.go` |
| A block split wrongly | Change the splitter; add a case to `TestSplit`, with made-up HTML in the same shape | `internal/blocks/blocks.go` |
| A boilerplate verdict wrong | Change a threshold (`MinWords`, `MinGroups`, `Widespread`) or the rule | `internal/boilerplate/boilerplate.go` |
| An advertiser's name as a heading | Nothing: it stays unmapped, and its text is found as boilerplate | |
| A heading that introduces mixed things | Nothing: unmapped is its right place | |

"Please note" is an example of the last: `heading-uses` shows it before
notices about applications, about sites and travel, and about working
hours, so no one kind fits it.

### Before and after

Keep a run from before a change:

```
go run ./cmd/pipeline run before
```

Make the change, run again under another name, and compare:

```
go run ./cmd/pipeline run after
go run ./cmd/pipeline compare before after
```

For each sample, and all of them together, `compare` shows the summary
lines whose figures moved, `-` before and `+` after, and the headings the
change mapped and those it unmapped, with the most text first. A sample
in only one of the runs, such as a new export's, is named, with nothing
to compare: run `before` again to include it. Newly unmapped headings
should be none, and newly mapped ones should all be right: a phrase can match more than was meant, as "the business" would
take "The business has a strong presence across:" to be company. Any
other table can be compared by hand, as in:

```
diff data/runs/before/all/boilerplate/repeated.tsv data/runs/after/all/boilerplate/repeated.tsv | head -20
```

Then run the tests (`test-all.sh`). For the next change, `after` can
become the new `before`: `go run ./cmd/pipeline run before` again.

## What it measured

On two samples, measured on 2026-10-26 with
`go run ./cmd/pipeline run now`, which gives current figures:

| Sample | Ads | Distinct | Leans towards |
|---|---:|---:|---|
| `ads-20261024` | 2,000 | 1,890 | education: teaching, support staff |
| `ads-20261025` | 5,000 | 4,296 | vehicles: mobile technicians, mechanics |
| `all` | 7,000 | 6,167 | both; 19 ads' text is in both samples |

The rules were tuned on the first, then checked on the second
([record](changes/2026-10-25-check-the-rules-on-a-second-sample.md)),
and hold on both: how the text divides into sections moves by two
points at most, except for company. How much of each section is
boilerplate moves more, as each slice has advertisers of its own.

**Headings.** 89% of ads have at least one heading, in both samples.
Headings of a known kind are 77% of heading uses in the first, 76% in
the second; ads with every heading known, 52% and 51%; ads with no
section but the summary, 5% and 4%.

**Sections and boilerplate**, by share of all the ads' text:

| Section | Of the text | | | Of it, boilerplate | | |
|---|---:|---:|---:|---:|---:|---:|
| | **1024** | **1025** | **all** | **1024** | **1025** | **all** |
| summary | 28% | 28% | 28% | 6% | 7% | 7% |
| responsibilities | 16% | 17% | 16% | 0% | 1% | 1% |
| requirements | 19% | 20% | 20% | 7% | 7% | 8% |
| desirable | 2% | 2% | 2% | 3% | 4% | 4% |
| benefits | 13% | 12% | 12% | 32% | 30% | 33% |
| company | 5% | 3% | 3% | 46% | 28% | 39% |
| apply | 3% | 3% | 3% | 37% | 42% | 44% |
| legal | 1% | 1% | 1% | 77% | 69% | 75% |
| details | 2% | 2% | 2% | 9% | 15% | 14% |
| unmapped | 13% | 13% | 13% | 17% | 19% | 20% |

Boilerplate is 13.2% of the text in the first sample, 12.5% in the
second, and 14.0% in both together, as more ads show more of what's
repeated; template for the job is 2.1%, 3.3% and 2.9%. So boilerplate,
with the company, apply and legal text that isn't, is roughly 15–20% of
an ad that probably needn't be embedded. That's a first answer to
whether retrieval can improve while embedding less.

What the second sample showed: its employers say less about themselves
under a heading of their own (company 3% of the text, not 5%), and its
headings were worded differently enough to need phrases of their own
("Who thrives here?", "What you can expect", "The business"), labels
("Vacancy type", "Reports to"), and two rule changes: conjunctions alone
in bold aren't headings, and a long bold line ending in a colon is. Its
boilerplate fell from 13.3% to 12.5% with them: one advertiser's job
text had been counted as boilerplate only because it was in unmapped
sections, and is now template, as it should be; but some of the same
advertiser's benefits and application text is now filed under
requirements, as footers are, and counted as template too (see
[Known gaps](#known-gaps)).

## Decisions, and why

| Decision | Why |
|---|---|
| Rules before models | The lab's principle: the cheapest accurate technique first. Every stage so far is rules and data. |
| Go for the stages | Go is the lab's main language; Python is kept for what needs its machine-learning tools. |
| Evidence first, decisions later | Blocks record cues; headings decide. Each decision can be measured and changed alone. |
| Text kept, keys normalised | Grammars later need the text as written (`£`, `–`); matching needs one form. |
| Phrases as data, canonical first | Synonyms are edited, not coded, in the format search engines use. |
| Legal's name isn't a phrase | In headings, "legal" is more often a job's field ("Legal Secretary") than a notice. |
| Rule order as ranking | Which kind wins is data too, changed by moving a line. |
| "Why" headings are benefits | They pitch the job to the reader, whatever words follow "why". |
| Unmapped and details sections | So what's left to map shows, and the job's terms are kept together for grammars. |
| "The role" is summary | What follows is usually an overview; to revisit with more evidence. |
| Title groups for advertisers | The ads don't say who placed them; different jobs is what makes text boilerplate. |
| Ties go to template | Keeping job text in an embedding is the safer mistake. |

## Known gaps

What the pipeline doesn't do yet. Each gap has an item in the
[roadmap](../ROADMAP.md), with a plan written to be picked up cold:
there, or, once work on it starts, in its
[change record](changes/README.md). When a gap is closed, its row here
and its item there go, in the same commit that finishes its record,
and the figures in [What it measured](#what-it-measured) are updated.

| Gap | Effect | Likely fix |
|---|---|---|
| [Walls of text](../ROADMAP.md#5-split-walls-of-text-into-sentences) | Headings run into the text, and bullets flattened into it, go unseen | Split into sentences, and find headings inside them |
| [Near-duplicate boilerplate](../ROADMAP.md#3-find-near-duplicate-boilerplate) | Standard text with a job title or place swapped in doesn't match word for word, which likely hides boilerplate in summaries | Fuzzy matching, such as shingles or MinHash |
| [Footers under few title groups](../ROADMAP.md#4-catch-footers-filed-under-requirements) | Filed under the last heading, usually requirements, they count as template | Their place at the end of the ad as a cue |
| [Plain lines without cues](../ROADMAP.md#6-find-headings-written-as-plain-lines) | Only exact phrases are taken as headings | More evidence from the following text |
| [Shift times as headings](../ROADMAP.md#2-stop-shift-times-passing-as-headings) | "2:00pm–10:00pm" in bold passes as a heading | A rule for times |

## Words used here

| Word | Meaning |
|---|---|
| block | a paragraph, line or list, as a reader sees it |
| cue | what the markup says about text: heading tag, bold, bold lead |
| part | a block or list item with its role |
| role | heading, field, repeat or text |
| key | the normalised form of text, for matching only |
| section | a heading and what follows it, up to the next |
| kind | what a section is: summary, requirements and the rest |
| phrase | a synonym in `headings.txt` |
| sample | one export's ads, as a `.jsonl` file in `data/ingestion/`; `all` is every sample together |
| run | every stage on every sample, kept in `data/runs/` under a name |
| title group | titles similar enough to be the same job |
| boilerplate | text repeated across different jobs, not about the job |
| template | text repeated across related jobs that does describe the job |
