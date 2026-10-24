# How the pipeline works

How the lab turns a job ad into its parts: which text is a heading, a
field, a section, or boilerplate. This is the working answer to the
lab's first questions: what sections ads have, how their headings vary,
how many can be recognised without AI, and how much of an ad is worth
embedding at all.

It covers each stage in order, how to run it, what it decided and why,
what it measured, and what it doesn't do yet. Keep it up to date: a
change to a stage changes this document in the same commit
([MAINTAINING.md](MAINTAINING.md#before-committing)).

## The flow

```
data/ingestion/extract.xml        the export, as received
        │  cmd/xml-to-jsonl        internal/mysqlxml
        ▼
data/ingestion/ads.jsonl          one ad per line: id, title, location, html
        │
        │  internal/blocks         HTML → blocks, with cues      (cmd/blocks)
        │  internal/headings       blocks → parts, with roles    (cmd/heading-counts)
        │  internal/sections       parts → sections, by kind     (cmd/sections)
        │  internal/boilerplate    sections → boilerplate        (cmd/boilerplate)
        ▼
data/blocks/  data/headings/  data/sections/  data/boilerplate/
```

Each stage is a package in `internal/`, with tests and examples beside
it, and each has a program in `cmd/` that runs it over every ad, writes
what it found to its folder in `data/`, and prints a summary. The
programs don't read each other's output: each reads `ads.jsonl` and runs
the stages before its own, which are fast. So any folder in `data/`
except `ingestion/` can be deleted and rebuilt.

Every stage only adds to what the one before it found, and never changes
the ad's text. Text is kept as written, and only a separate **key** (see
[Keys](#keys)) is normalised for matching.

## Running it

From the lab's folder, in this order the first time:

```
go run ./cmd/xml-to-jsonl data/ingestion/extract.xml data/ingestion/ads.jsonl
go run ./cmd/blocks         data/ingestion/ads.jsonl data/blocks
go run ./cmd/heading-counts data/ingestion/ads.jsonl data/headings
go run ./cmd/sections       data/ingestion/ads.jsonl data/sections
go run ./cmd/boilerplate    data/ingestion/ads.jsonl data/boilerplate
```

After the first, only the stage being worked on needs running.
`go install ./cmd/...` puts the programs in `bin/`, on the lab's `PATH`,
so they can be run by name. Ads whose HTML is the same as an earlier
one's are counted once by the last three.

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
| `heading` | short (`MaxWords`, 10), with a cue (heading tag, all bold) or ending in `:` or `;`; no money or long numbers; not ending in `.` | "**Key Responsibilities**" |
| `field` | "Label: value" or "Label \| value", where the label is a known one | "Hours: 40 per week" |
| `repeat` | only the ad's title or location again | "**Delivery Driver**" at the top |
| `text` | everything else | |

A few rules refine this:

- **A bold lead that isn't a field** is a heading run into its text:
  "**Key Responsibilities:** You will…" is a heading, then text.
- **A field's label written as a heading,** with a short value (up to
  `MaxValueWords`, 15) on the next line, is joined into one field:
  "**Working Hours**", then "Monday to Friday".
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
have the same key. It borrows the idea of job-meta's `punct` package,
with the characters these ads use as well, such as Word's bullet.

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

**Choosing between kinds.** Where a heading has phrases of two kinds, the
kind whose rule comes first in the file wins. The file's order is the
ranking: desirable, benefits, legal, requirements, responsibilities,
company, apply, details, summary. So "Desirable skills" is desirable, not
requirements, and "What our client offers" benefits, not company.

**With no phrase**, a heading starting "about" is company ("About" and a
company's name), one starting "why" is benefits, and one that's a field's
label is details. Any other starts an `unmapped` section: so the text
under it counts as unmapped, rather than hiding in the section before,
and unmapped's share of the text measures what's left to map.

**A plain line** with no cue starts a section only if it's exactly a
kind's phrase: `<p>Responsibilities</p>`, not "Our responsibilities to
you".

`cmd/sections` writes each ad's sections to `ads.jsonl`, and three tables
to work from: `unmapped.tsv` (headings of no known kind, with the text
under each), `disagree.tsv` (headings whose phrases point to two kinds),
and `named.tsv` (plain lines taken by name).
[MAINTAINING.md](MAINTAINING.md#heading-phrases) has how to improve the
phrases with them.

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

## What it measured

On the sample of 1,890 distinct ads (2,000 less 110 duplicates), on
2026-10-25. Rerun the programs for current figures.

**Headings.** 89% of ads have at least one heading. 76% of heading uses
are of a known kind, and 51% of ads have every heading known. 5% of ads
have no section but the summary.

**Sections and boilerplate**, by share of all the ads' text:

| Section | Of the text | Of it, boilerplate |
|---|---:|---:|
| summary | 28% | 6% |
| responsibilities | 16% | 0% |
| requirements | 19% | 7% |
| desirable | 2% | 3% |
| benefits | 12% | 32% |
| company | 5% | 45% |
| apply | 3% | 36% |
| legal | 1% | 77% |
| details | 2% | 9% |
| unmapped | 14% | 17% |

13.2% of all the text is boilerplate, and 2.1% template for the job. So
boilerplate, with the company, apply and legal text that isn't, is
roughly 15–20% of an ad that probably needn't be embedded. That's a
first answer to whether retrieval can improve while embedding less.

## Decisions, and why

| Decision | Why |
|---|---|
| Rules before models | The lab's principle: the cheapest accurate technique first. Every stage so far is rules and data. |
| Go for the stages | Go is the lab's main language; Python is kept for what needs its machine-learning tools. |
| Evidence first, decisions later | Blocks record cues; headings decide. Each decision can be measured and changed alone. |
| Text kept, keys normalised | Grammars later need the text as written (`£`, `–`); matching needs one form. |
| Phrases as data, canonical first | Synonyms are edited, not coded, in the format search engines use. |
| Rule order as ranking | Which kind wins is data too, changed by moving a line. |
| Unmapped and details sections | So what's left to map shows, and the job's terms are kept together for grammars. |
| "The role" is summary | What follows is usually an overview; to revisit with more evidence. |
| Title groups for advertisers | The ads don't say who placed them; different jobs is what makes text boilerplate. |
| Ties go to template | Keeping job text in an embedding is the safer mistake. |

## Known gaps

| Gap | Effect | Likely fix |
|---|---|---|
| Walls of text | Headings run into the text, and bullets flattened into it, go unseen | Split into sentences, and find headings inside them |
| Near-duplicate boilerplate | Standard text with a job title or place swapped in doesn't match word for word, which likely hides boilerplate in summaries | Fuzzy matching, such as shingles or MinHash |
| Footers under few title groups | Filed under the last heading, usually requirements, they count as template | Their place at the end of the ad as a cue |
| Plain lines without cues | Only exact phrases are taken as headings | More evidence from the following text |
| Shift times as headings | "2:00pm–10:00pm" in bold passes as a heading | A rule for times |
| One sample | Every rule was tuned on one slice of ads, weighted to education | Check against a second sample |

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
| title group | titles similar enough to be the same job |
| boilerplate | text repeated across different jobs, not about the job |
| template | text repeated across related jobs that does describe the job |
