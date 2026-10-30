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
        │  internal/view           all of it → ads to read       (cmd/view)
        ▼
data/runs/RUN/  ads.jsonl (every sample)  blocks/  headings/  sections/  boilerplate/  view/
```

Each stage is a package in `internal/`, with tests and examples beside
it, and each has a program in `cmd/` that runs it over every ad, writes
what it found to a folder, and prints a summary. The programs don't
read each other's output: each reads a sample and runs the stages before
its own, which are fast. So any folder in `data/` except `ingestion/`
can be deleted and rebuilt.

A **sample** is one export's ads, converted: a `.jsonl` file in
`data/ingestion/`, named by the export's date. `cmd/pipeline` runs
every stage on all the samples together, and keeps each **run** under a
name of its own in `data/runs/`: `latest` for where the pipeline
stands, and runs in `claude/`, such as `claude/before` and
`claude/after`, for comparing a change. A run lists the samples it read
in `samples.txt`. To check a rule on one sample alone, run a stage's
program on it by hand (below).

The ad is kept as it was received. Blocks, the first stage, **cleans**
its text for the stages after it, and for search: structure in English
words, not symbols (see [Blocks](#2-blocks-what-a-reader-sees)). After
that, every stage only adds to what the one before it found, and a
separate **key** (see [Keys](#keys)) is normalised further for
matching.

## Running it

From the lab's folder. Convert each new export into a sample, named by
its date:

```
go run ./cmd/xml-to-jsonl data/ingestion/extract-20261025.xml data/ingestion/ads-20261025.jsonl
```

Then run every stage on the samples together, as the run `latest`:

```
go run ./cmd/pipeline run latest
```

It takes under a minute: `data/runs/latest/sections/` holds the
sections stage's tables, and `data/runs/latest/sections.txt` its
summary. A run of the same name is replaced.

It builds the stages first, so the code as it stands is what runs, and
ends with how each stage ran, also kept in
`data/runs/latest/timings.tsv`:

```
sample  stage               ads      wall       cpu    ads/s      peak
all     sections          28000     8.03s     8.56s     3487    206 MB
all     boilerplate       28000    12.16s    13.02s     2302    676 MB
```

`wall` is the time on the clock, `cpu` the time on the processors (more
than `wall` when a stage uses several), and `peak` the most memory the
stage held at once (on macOS and Linux; 0 elsewhere). `compare` leaves
the timings out, as they move a little on every run: compare two runs'
`timings.tsv` by hand.

Each stage's program can also be run alone, on one sample, into any
folder:

```
go run ./cmd/blocks         data/ingestion/ads-20261025.jsonl data/blocks
go run ./cmd/heading-counts data/ingestion/ads-20261025.jsonl data/headings
go run ./cmd/sections       data/ingestion/ads-20261025.jsonl data/sections
go run ./cmd/boilerplate    data/ingestion/ads-20261025.jsonl data/boilerplate
go run ./cmd/view           data/ingestion/ads-20261025.jsonl data/view
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
Control characters such as U+0002, which XML forbids but MySQL writes
as they are, come through unchanged: the reader stands each in for a
private use character while it parses, and refuses an export already
holding one.
Duplicates stay in: later stages leave them out of their counts.

## 2. Blocks: what a reader sees

**Package** `internal/blocks`, **program** `cmd/blocks`.

`blocks.Split` reads the HTML with Go's HTML tokenizer and returns the
blocks a reader would see, in order:

| Kind | What it is |
|---|---|
| `paragraph` | a block of text on its own: a `<p>`, a `<div>`, a heading tag, or loose text |
| `line` | one of several lines a `<br>` or a newline split a block into |
| `list` | consecutive list items, each an item of its own |

Each block has its text as a browser shows it (entities decoded, runs
of spaces made one), **cleaned** (`internal/clean`), and its **cues**:
what the markup says that the text doesn't.

Cleaning keeps what means something to search, and makes structure
English:

| What | Becomes | Example |
|---|---|---|
| A symbol that opens a line, stands for one thing, and is followed by a value of it | its English label, unless the line has one, is bold, or ends as a heading does (`:` `?` `!`) | "📍 Leeds" → "Location: Leeds", a field; also 💷 💰 Pay, ⏰ 🕒 ⏱ Hours, 📞 Contact, 📧 Email; but "💷 What We Offer" keeps no label |
| Lines in a row that open with the same symbol, not ending with it | a list, each an item | "✅ Weekly pay" and "✅ Free parking" on lines of their own → two items |
| A line that opens with `*`, or with `-` before a space or a letter | the same: the mark goes, and lines in a row with it are a list | "*Home based" → "Home based"; "-5 degrees" keeps its minus |
| Other emoji and pictographs, and symbols used as bullets | nothing: when in doubt, a symbol goes | "✅ Long term work" → "Long term work"; "📢 Hiring now 📢" → "Hiring now" |
| Invisible and control characters, Word's bullet, the replacement character | nothing | |
| Odd spaces, such as the no-break space | a plain space | |
| Typographic quotes, dashes and ellipses, and full-width forms | plain ASCII; an em dash ` -- ` | "Mon–Fri…" → "Mon-Fri...", "rest—then" → "rest -- then" |
| Letters, accents, figures, currency signs | kept | "Café", "£30,000" |

Of the three samples' 864,828 lines, 1.1% open with a symbol; the label
symbols are about 1,150 of them.

| Cue | Example |
|---|---|
| `heading` | the text was in `<h1>` to `<h6>` |
| `bold` | all of it was in `<strong>` or `<b>` |
| `bold_lead` | it starts with bold text: `Salary:` in "**Salary:** £30,000" |
| `added` | the English label cleaning added: `Location: ` for "📍 Leeds" |
| `gap` | a blank line came before it: two `<br>`s with nothing between, or a blank line in text laid out by its newlines (not the space every paragraph has); over the three samples, 69,126 lines and paragraphs. Nothing reads it yet |

**Lists.** A list item is one piece of text, and a sublist's items join
its list. But ads run headings, fields and footers into their items,
the last item most of all, and some put paragraphs, headings and further
lists inside one, up to the whole ad. So an item ends at a `<br>`, `<p>`
or `<div>` after its text, and the text after it in the `<li>` is lines
or paragraphs of their own; so is text outside any item of its own. Each
stands where it is, and ends the list before it, so that nothing comes
out of the ad's order. An item that really is two lines (`Days<br>Nights`)
is split too: the price of finding what's run into the others.

**Tags left open.** Ads leave out end tags, so an opening tag is taken
to end what's open before it, as a browser takes it: a `<p>`, a list or
a heading tag ends a `<p>` or a heading tag still open, keeping the
heading's cue, and a `<p>` or `<br>` after an item's text ends the item
(see Lists), so a list never closed doesn't swallow the rest of the ad.
Bold opened in a paragraph ends with it, closed or not: in the samples,
each of the 5 paragraphs a `<strong>` left open ran into is a footer
after a forgotten `</strong>`, not text meant bold. Bold around
paragraphs goes on. HTML's other block-level tags, such as `<aside>`, `<center>`
and `<figure>`, end blocks too, though none is in the samples.

It decides nothing. Whether a bold block is a heading is the next stage's
call, which keeps the HTML parsing apart from the judgement, so either
can change without the other.

**Newlines.** Many ads are laid out by their newlines as much as by
their tags: plain text, with no tags that lay out blocks, and ads with
tags round only part of their text, such as a footer, or with bare lines
between their tags. So a newline ends a line, and a blank line a block,
outside list items and heading tags, though a browser would show a
space. In an ad with tags, only a newline with text either side, and no
tag between, counts: one beside a tag is how the source is formatted,
often a tag to a line. And there a newline before a lower-case letter,
or after a comma, "&" or a word such as "to" that a line can't end with
("Welcome to" before a name), is a line wrapped in the source, so it
stays a space. Over the three samples, that keeps 2,480 more lines than
reading every newline in a tagged ad as a space. A **wall** of text with
no newlines at all stays one paragraph: splitting it needs sentences, a
later stage.

On the sample, no ad uses heading tags: headings are bold paragraphs, or
lines ending with a colon. 3 ads have a paragraph over 1,500 characters
(of the three samples' 28,000, 16).

## 3. Headings: what each block is

**Package** `internal/headings`, **program** `cmd/heading-counts`.

`headings.Read` gives each block, and each list item, a **role**, and
returns them as **parts**:

| Role | How it's recognised | Example |
|---|---|---|
| `heading` | short (`MaxWords`, 10, or `MaxIntroWords`, 20, with a cue and ending in `:`), with a cue (heading tag, all bold) or ending in `:` or `;`; a word that isn't a conjunction; no money, long numbers or times of day (a bare range of hours, such as 8-5, only beside a range of days); no pattern of work ("4 on 4 off", or days or nights beside a range of days), nor only facts ("Permanent | Full-time", "Monday - Friday"), unless it ends in `:`; no contact details an export took out ("(phone number removed)"); not ending in `.` | "**Key Responsibilities**" |
| `field` | "Label: value", "Label \| value" or "Label - value", where the label is a known one | "Hours: 40 per week" |
| `repeat` | only the ad's title or location again | "**Delivery Driver**" at the top |
| `text` | everything else | |

A few rules refine this:

- **A bold lead that isn't a field** is a heading run into its text:
  "**Key Responsibilities:** You will…" is a heading, then text.
- **The ad's title run into the rest of its line** is split off as a
  repeat, and the rest read as it would be alone, when the rest is a
  field, or a heading, the line having been one with the title too
  ("**General Foreman Responsibilities:**"), or, after a mark that ends
  the title (a dash, bar, colon, comma or opening bracket), a short
  line that isn't a sentence and doesn't open in lower case, as text
  ("Electrician (Permanent)", "Support Worker - Hull £12.93 ph"). A
  line that goes on as a sentence, or with the title made plural or
  longer ("Panel Beater / Panel Technician"), stays whole, as does a
  label cleaning added. After only a space, a heading of no known kind
  is the title made longer ("Head Chef Manager"): sections joins it
  back into one heading.
- **A field's label written as a heading,** with a short value (up to
  `MaxValueWords`, 15) on the next line, is joined into one field:
  "**Working Hours**", then "Monday to Friday".
- **A long line in bold that ends with a colon** introduces what follows,
  as a heading does, so it can have up to `MaxIntroWords` (20) words:
  "**As a Mobile Mechanic, you'll get benefits that go the extra mile:**".
  Without the colon, or without the bold or a heading tag, it's a sentence.
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
case; typographic quotes and odd spaces folded to plain ones; dashes,
typographic or not, and punctuation other than `' & /` made a space;
bullets, emoji and invisible characters dropped. It folds what cleaning
does as well, as titles and locations reach it uncleaned. "What You’ll Be Doing:" and "what you'll be doing"
have the same key. The table of characters it folds includes those these
ads use, such as Word's bullet and the emoji variation selector.

## 4. Sections: what each heading starts

**Package** `internal/sections`, **program** `cmd/sections`.

`sections.Divide` divides an ad's parts into **sections**. What comes
before the first heading is the summary, and each heading starts a
section of its **kind**:

| Kind | Holds | Example headings |
|---|---|---|
| `summary` | the ad's opening, and introductions | "Overview", "The opportunity" |
| `role` | what the job is | "About the role", "Job description" |
| `responsibilities` | what the job involves | "What you'll be doing", "Main duties" |
| `requirements` | what the job needs | "About you", "Essential criteria" |
| `desirable` | what would help | "Desirable", "Nice to have" |
| `benefits` | what the job offers | "Why join us?", "What's on offer" |
| `company` | who the employer is | "About us", "About the company" |
| `apply` | how to apply | "How to apply", "Interested?" |
| `legal` | notices | "Equal opportunities", "Safeguarding" |
| `details` | the job's terms | "Job details", "Working Hours" |
| `unmapped` | under a heading of no known kind | anything not yet recognised |

A heading of no known kind that the ad's title ran into is the ad's
**subtitle**, "Oldham" of "Head Chef - Oldham": text in the section
it's in, and not counted as a heading by `cmd/sections` (the headings
stage, which knows no kinds, still counts it). Before, the whole line
was an unmapped heading, and everything under it, sometimes the whole
ad, was unmapped too. A label after the title, ending in a colon
("You:") or over a list, is a heading still.

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
company's name), and one that's a field's label is details. Any other
starts an `unmapped` section: so the text under it counts as unmapped,
rather than hiding in the section before, and unmapped's share of the
text measures what's left to map.

**A plain line** with no cue starts a section if it's exactly a
kind's phrase, as `<p>Responsibilities</p>` is, or if it's short (up to
`MaxListWords`, 6 words), not a sentence, has a kind's phrase in it, and
a list follows it: a sub-heading such as "Highly Desirable" or
"Essential Experience & Qualifications" among the requirements, or "Key
Skills". "Our responsibilities to you" with no list after it stays text.
The rule took 4,111 more lines as headings on the three samples; of a
sample of 30 of the rarer ones, one was wrong ("Duties required",
requirements by "required", over "duty").

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
more, by its key, in ads for different jobs.

**Not always word for word.** Standard text often has the ad's own
title or place written in ("we're recruiting a Chef in Leeds"), or a
few words changed, such as a school's or an agency's name. So a text is
known by up to three keys, and counts as repeated, or boilerplate, if
any of them does:

- its own key;
- its key with the ad's title and places, where they stand as whole
  words, as placeholders, so "a Chef in Leeds" and "a Nurse in York"
  key alike;
- for a text of `LongWords` (20) words or more, its group of **texts
  alike**: long texts that share at least `Overlap` (60%) of their runs
  of five words, a group taking in any text alike to one of its own.

Short lines are matched only by the first two: phrased alike, as
"assist with goods in and goods out" and "assisting with goods in and
goods out processes" are, they're more often about the job than
standard text.

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
- **boilerplate** if it's found more often outside them; or if it's found
  mostly in the run of repeated texts that ends its ads, and talks as a
  footer does, whatever its section; or if it's under `Widespread` (5)
  groups or more;

but never by the last two if it's phrased as a job's requirement or
duty is, by its opening in `openings.txt` ("Ability to", "Experience
of", "Liaise with"), and, for the widespread rule, found mostly in the
sections that describe the job.

The verdict belongs to the text, over every ad it's in, so a heading
misfiled in one ad doesn't change it. Footers, an agency's notices or an
equal opportunities statement, often come after an ad's last heading,
so they land under its requirements and look like templates: where they
stand gives them away. The **ending run** of an ad is its texts from the
last back to the first that isn't repeated, passing over those too
short to count (`Ends`). A footer **talks as one** with a word of
applying, getting in touch, the agency or its notices, "we", "our" or
"us" (`FooterLine`): an advertiser's list of requirements can run to
the end of its ads too, but seldom has any. And "Ability to work
independently and as part of a team", in ads for 176 groups of titles,
is a job line still: it says little of any one job, but it is about the
job.

`cmd/boilerplate` writes `repeated.tsv`: every repeated text, with its
groups (by the key that finds the most), ads, characters, the kinds of
section it's in, and its verdict.

## 6. View: ads to read

**Package** `internal/view`, **program** `cmd/view`.

The figures say how much moved, not what a change did to real ads, and
reading ads is how most faults are found. So each run also writes
`view/ads.md`: 20 of the run's ads in their chunked form, in
Markdown, to read in an editor's preview or as text, exactly as `chunk`
makes them (see [Chunking an ad](#chunking-an-ad)): each chunk with its
kind, heading, use, size and what placed it, its lines, a list's items
as a list, and the boilerplate gathered at the end.

The ads are chosen by a fixed hash of their IDs, and written in its
order, so a diff of two runs' views shows what a change did, ad by ad.
Sixteen are always the first by the hash; up to four more are chosen on
purpose, as a random draw mostly shows the easy cases: an ad with no
heading, one with an unmapped section, one over half boilerplate, and
one with text in no section or out of order (below), each where none of
the others shown is one already. A change to the pipeline can change
those four, but never the sixteen, nor any ad's place among them.

It also checks every ad of the run, comparing the letters and digits
of what its HTML shows with those of its sections. Text in no section
should never happen, nor should text out of the ad's order, and neither
does on the three samples. The view shows where, for each ad it has,
and the summary counts both over the run.

## Chunking an ad

**Package** `internal/chunk`, **program** `cmd/chunk`.

Everything above, for one ad, as an embedding job would take it: its
sections as chunks, each with its kind and heading, its text (a
list's items as "- " entries), what to do with it, about how many tokens it is (four
for every three words), and what placed it there.

```
go run ./cmd/chunk 225682916
go run ./cmd/chunk -json 225682916
go run ./cmd/chunk -yaml 225682916
go run ./cmd/chunk -title "Driver" ad.html
go run ./cmd/chunk -pages data/pages/ads.tsv
```

The text is to read at a glance, JSON (`-json`, a line an ad) is for
programs, and YAML (`-yaml`, a document an ad) is for people who want
the structure too: each chunk's text is a block of its lines.

The summary, role, responsibilities, requirements and desirable chunks
are embedded; details are metadata, with their fields; benefits, company,
apply and legal are left out. Text found as boilerplate leaves the
embedded chunks for a **boilerplate** chunk of its own, at the end,
left out, each line with the section it came from: this is where to
see what the verdict takes, and whether it's right. Other chunks keep theirs, so an
advertiser's standard perks are still a list of benefits.

An **unmapped** section, under a heading of no known kind, is placed by
a **guess**, from what's already known about its text, and its chunk
says what placed it. The evidence is weighed in this order, the first
that holds placing it:

| Evidence | Placed as | Example |
|---|---|---|
| At least half its lines open as one kind's do (`openings.txt`), and fewer than half are boilerplate | that kind | "guessed requirements: 3 of its 6 lines open as requirements do" |
| At least half its lines are boilerplate | left out | an agency's standard footer |
| Its heading speaks for the advertiser: "we", "our", "us", or it opens "who are" or "join" | left out | "Who are Acme?" |
| It comes after the ad's apply or legal text | left out | |
| It has no lines of text, only fields or facts | left out (its fields and facts are details) | |
| A heading ending in a colon, over a line or two of at most 15 words | details | "IR35 Status:" over "Inside IR35" |
| Nothing else | the summary | a job title used as a heading, over the job's pitch |

Text left out goes to the boilerplate chunk, under its heading, with
why (`unmapped "Agency:" (after applying): …`); its fields become a
details chunk. The summary as the default is a choice: read by eye,
most unmapped text is about the job (its pitch, duties and
requirements in words `openings.txt` hasn't), and four samples of 50
guesses put 39 to 43 of each where a reader would; the misses are
mostly the employer's culture, perks and notices, under headings no
rule covers, now embedded. Each ad totals its tokens, its embedded
tokens and its unclassified ones (the unmapped text left out), so the
last shows what's left to map. The
chunks come in a fixed order: summary, role, responsibilities and
requirements first, then the rest in the ad's order, the chunks of
lines placed by their wording and of lines of facts after its own, and
the boilerplate last. Boilerplate is what the run's
`boilerplate/repeated.tsv` says, so a run comes first (`-run` chooses
which, `data/runs/latest` by default). That's quicker than reading
every ad again, but the table has only texts as written, so `chunk`
misses boilerplate the stage finds with the ad's own title or place
taken out, or as a near-copy; the view, which reads every ad, finds
those too, and so can gather more. What placed a chunk is the
heading and the phrase in it that won, with any phrases of other kinds
it outranked, or the rule (`it starts "why"`), or that nothing came
before it, so a wrong kind points at the line of `headings.txt` to
change.

### Lines placed by their wording

**Package** `internal/wording`.

A line no heading placed can say by its opening what it is: "Experience
of…" or "Must have…" is a requirement, "Conduct site surveys…" or
"Responsible for…" a responsibility, "Free parking" or "33 days
holiday" a benefit. The openings are data, in
`internal/wording/openings.txt`, in the rule format `headings.txt`
uses, read by a grammar: a word list found only at a line's start, the
longest opening winning, so "you'll have" is a requirement and "you'll
have the opportunity" a benefit. The chunker applies them to a
summary's lines and a role's list items, a role's own prose staying,
before boilerplate is gathered, and puts each line in a chunk of its
kind, saying which openings placed it. On the three samples, 8,524
lines move, in 13% of ads; of 40 checked by eye, one was wrong.

### Facts

**Package** `internal/facts`, its grammars in `internal/grammars`.

The facts ads state in lines of their own are read by grammars, with
[grammar](https://github.com/nukopian/grammar), one for each kind, each
in Raku's grammar syntax, so a rule reads much as the text it matches.
Headings use the same grammars to keep facts out (see the heading role
above), so that a line is a time, or a rota, by one set of rules:

| Fact | Reads | Values |
|---|---|---|
| pay | "£45k - £55,000 per annum", "£12.71ph", "50,000pa"; a bare figure after a pay label, "Pay Rate - 13.10" | `min`, `max`, `currency`, `period` |
| hours | "40hrs pw", "37.5 hours per week", "12 hours per day", "40 hours" | `per_week`, `per_day`, or `hours` with no period |
| pattern | "Full-time", "Permanent", "Temp to perm", "Hybrid", from a word list of canonical values | `value` |
| schedule | "Monday to Friday, 8am - 5pm", "Mon-Fri 07:00-15:30", "M-F 8-4", "4 on 4 off" | `days`, `times`, `rota` |
| location | a UK postcode, "NE8 1AQ" | `postcode` |

A figure is pay only with a currency sign or a period ("5 years" isn't),
and in a range, "£25-28k", the top figure's k is the bottom's too; a
bare "h" is hours only with a period after it ("24h support" isn't); a
bare range of hours is a time only after a range of days; and a time
starts at a figure that isn't inside another, so "Friday7:00am" is a
time, but "745am" none, rather than "45am". The chunker
takes a **line of facts** out of an embedded chunk, or an unmapped
section, into a details chunk of its own, with each fact and its
values: a line with a postcode in it, of at most 30 words; or one that
starts with a fact, in its first three words, of at most 25 words, not
a sentence or a question, and, if its only facts are work patterns,
which sentences use too, nothing else, or at most two other words set
apart by "|", "," or "/": "Leeds | Permanent" is a line of facts,
"Contract administration" isn't. A details chunk keeps its lines, and
gains the facts read from them. A known field label followed by " - "
is a field, as one followed by a colon is.

### Pages before and after

**Package** `internal/page`, **program** `cmd/chunk -pages`.

For showing what chunking does to an ad, to people who don't read YAML,
`chunk -pages FILE` writes web pages: each ad as posted beside its
chunks, each chunk coloured by what's done with it (embedded, metadata,
left out, or gathered as boilerplate), with its kind, size, what placed
it, and its fields and facts, under a bar of where the ad's tokens go;
and an index of them all. FILE lists the ads, an ID, a tab and why it's
there a line, and the pages go beside it. As `chunk` writes them, a
rerun shows the chunker as it stands. The ad as posted is cut down to
the tags that lay text out, with no attributes, scripts or styles, and
the pages are written with `html/template`, which escapes everything
else. They show ads, so they live in the private data store, in
`data/pages/`: a dozen chosen to show where chunking does well and
where badly.

## Testing retrieval

**Package** `internal/retrieval`, **programs** `cmd/retrieval-set`
and `python/bin/retrieval_eval.py`.

A first test of whether representing an ad by its parts finds it
better. `cmd/retrieval-set` takes 5,000 of a run's ads, by a fixed hash
of their IDs, and writes, for each, the text each representation would
embed: the whole ad; the ad without its boilerplate; the sections that
describe the job; the summary; and each section, a vector each. Its
queries are the 50 commonest groups of similar titles, each written as
its commonest title, and an ad is relevant to one if its title is
similar to it. The title field is left out of every representation.
The Python program embeds them with `BAAI/bge-base-en-v1.5`, ranks the
ads for each query, and writes `report.txt`:

```
go run ./cmd/retrieval-set data/runs/latest/ads.jsonl data/runs/latest/retrieval
python python/bin/retrieval_eval.py data/runs/latest/retrieval
```

Embedding takes about 20 minutes on a Mac's GPU. What it showed, and what it
can't, are in its
[record](changes/2026-10-27-r10-a-first-retrieval-test.md): a job title is
a narrow query.

## Improving it

Each program leaves tables in `data/` that say where its stage falls
short. Improving the pipeline is a loop over them:

1. **Measure:** `go run ./cmd/pipeline run claude/before`, every stage
   on the samples together (see [Before and after](#before-and-after)).
2. **Pick** the biggest gap: the table row with the most ads or text.
3. **Look** at it in the ads, in context, before deciding anything:
   `view/ads.md` has 20 of them, and `cmd/heading-uses` every use of a
   heading.
4. **Change one thing:** a phrase, a label, a rule, a threshold.
5. **Measure again,** as the run `claude/after`, and compare. Keep the
   change only if it helped and broke nothing. A change that fits one
   kind of job may harm another: to see a sample on its own, run the
   stage's program on it by hand.
6. **Pin it:** add a test case for what it settled, and update the
   figures in [What it measured](#what-it-measured).

### Questions, and where to look

All from the lab's folder, after a run (see
[Running it](#running-it)). The commands look at the run
`claude/before`, which these names choose; set them first, to the run
you want, and each command then runs as it stands:

```
r=data/runs/claude/before
ads=$r/ads.jsonl
```

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
| Text taken for a heading, or a heading missed | Change the rule (`headingShaped`, `MaxWords`, `MaxIntroWords`, `conjunctions`, `hasMoney`, `hasTimes`, `hasPattern`), or the grammars' rules they use (`clock`, `points`, `span`, `days`, `rota`, `shift`, `Figure`); add a case to `TestRead` | `internal/headings/headings.go`, `internal/grammars/grammars.go` |
| A block split wrongly | Change the splitter; add a case to `TestSplit`, with made-up HTML in the same shape | `internal/blocks/blocks.go` |
| A boilerplate verdict wrong | Change a threshold (`MinWords`, `MinGroups`, `Widespread`), `footerWords`, or the rule; a job line taken for a footer may want an opening | `internal/boilerplate/boilerplate.go`, `internal/wording/openings.txt` |
| An advertiser's name as a heading | Nothing: it stays unmapped, and the guess places its text: as boilerplate, mostly | |
| A heading that introduces mixed things | Nothing: unmapped is its right place, and the guess places its text | |
| An unmapped section guessed wrongly | Map its heading, if a phrase will reach others; or change the guess's evidence | `headings.txt`, `internal/chunk/chunk.go` (`guess`) |

"Please note" is an example of the last: `heading-uses` shows it before
notices about applications, about sites and travel, and about working
hours, so no one kind fits it.

### Before and after

Keep a run from before a change:

```
go run ./cmd/pipeline run claude/before
```

Make the change, run again under another name, and compare:

```
go run ./cmd/pipeline run claude/after
go run ./cmd/pipeline compare claude/before claude/after
```

`compare` shows the summary lines whose figures moved, `-` before and
`+` after, and the headings the change mapped and those it unmapped,
with the most text first. If the two runs read different samples, as
when a new export has come in between, it says so first, as figures
then move for that too: run `claude/before` again. Newly unmapped headings
should be none, and newly mapped ones should all be right: a phrase can match more than was meant, as "the business" would
take "The business has a strong presence across:" to be company. Any
other table can be compared by hand, as in:

```
diff data/runs/claude/before/boilerplate/repeated.tsv data/runs/claude/after/boilerplate/repeated.tsv | head -20
```

and the views show the change in the ads themselves:

```
diff data/runs/claude/before/view/ads.md data/runs/claude/after/view/ads.md
```

Then run the tests (`test-all.sh`), and `go run ./cmd/pipeline run
latest`, so that `latest` shows where the pipeline now stands. For the
next change, run `claude/before` again.

## What it measured

On the three samples together, measured on 2026-10-27 with
`go run ./cmd/pipeline run latest`:

| Sample | Ads | Leans towards |
|---|---:|---|
| `ads-20261024` | 2,000 | education: teaching, support staff |
| `ads-20261025` | 5,000 | vehicles: mobile technicians, mechanics |
| `ads-20261027` | 21,000 | not yet looked at by kind of job |
| together | 28,000 | 25,954 distinct |

The rules were tuned on the first sample, checked on the second
([record](changes/2026-10-25-r1-check-the-rules-on-a-second-sample.md)),
and hold on the third: how the text divides into sections moved by two
points at most from one sample to another, except for company.

**Headings.** 88% of ads have at least one heading. Headings of a known
kind are 81% of heading uses; ads with every heading known, 62%; ads
with no section but the summary, 4%.

**Sections and boilerplate**, by share of all the ads' text:

| Section | Of the text | Of it, boilerplate |
|---|---:|---:|
| summary | 18% | 11% |
| role | 8% | 8% |
| responsibilities | 18% | 2% |
| requirements | 21% | 14% |
| desirable | 2% | 8% |
| benefits | 13% | 38% |
| company | 3% | 50% |
| apply | 3% | 58% |
| legal | 1% | 84% |
| details | 2% | 18% |
| unmapped | 10% | 25% |

Boilerplate is 18.2% of the text, and template for the job 2.6%; with
the company, apply and legal text that isn't boilerplate, about a fifth
of an ad probably needn't be embedded. That's a first answer to whether
retrieval can improve while embedding less. Word for word matching
alone found 15.8%: standard text with the ad's own title or place
written in, or a few words changed, is most of the rest
([record](changes/2026-10-27-r3-find-near-duplicate-boilerplate.md)).

**Chunks.** Over the same ads, by the chunker (`chunk.Of`, boilerplate
by the index): 65.6% of an ad's tokens are embedded, 3.1% kept as
metadata, 21.2% left out (benefits, company, apply, legal), and 10.1%
gathered in the boilerplate chunk, 2.8 points of it text under
unmapped headings the guess left out. Of the unmapped text, 6.3 points
of an ad's tokens are placed by the guess, nearly all as the summary.
Facts are read in 35% of ads.

Boilerplate's thresholds are counts of title groups, which a larger
sample passes more easily, so boilerplate figures from samples of
different sizes don't compare: on its own, the first sample has 13.2%
by word for word matching, the third 15.1% (idea 12 on the roadmap).

## Decisions, and why

| Decision | Why |
|---|---|
| Rules before models | The lab's principle: the cheapest accurate technique first. Every stage so far is rules and data. |
| Go for the stages | Go is the lab's main language; Python is kept for what needs its machine-learning tools. |
| Evidence first, decisions later | Blocks record cues; headings decide. Each decision can be measured and changed alone. |
| The ad kept, its text cleaned, keys normalised | The ad as received is there to show and check against; the stages and search want text without symbols, structure in words; matching wants one form. |
| Phrases as data, canonical first | Synonyms are edited, not coded, in the format search engines use. |
| Legal's name isn't a phrase | In headings, "legal" is more often a job's field ("Legal Secretary") than a notice. |
| Rule order as ranking | Which kind wins is data too, changed by moving a line. |
| "Why" headings are benefits | They pitch the job to the reader, whatever words follow "why". |
| Unmapped and details sections | So what's left to map shows, and the job's terms are kept together for grammars. |
| Role apart from summary | What the job is, under "About the role", isn't the ad's opening or its introductions; both are embedded, but apart. |
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
| [Layout folded](../ROADMAP.md#24-keep-the-layout-an-ads-text-implies) | In an ad with any layout tag, newlines meant as layout are read as spaces, so headings and fields on them join the text around them; a `<br>` or `<p>` in a list item runs its lines together; tags left open can swallow or mislabel what follows | Read newlines as plain text's are, keep a list item's lines apart, and let an opening tag end what's left open |
| [Walls of text](../ROADMAP.md#5-split-walls-of-text-into-sentences) | Headings run into the text, and bullets flattened into it, go unseen | Split into sentences, and find headings inside them, once layout is kept |
| [Inline lists](../ROADMAP.md#25-split-inline-lists-into-items) | A list run along one line between bullet symbols is one line of text | Split it at the symbols, into items |
| [Italic not a cue](../ROADMAP.md#26-italic-as-a-cue-like-bold) | Headings and notices set in italic carry no cue, as bold ones do | Record italic as blocks records bold |
| [Label emoji lost](../ROADMAP.md#27-keep-what-label-emoji-mean-wherever-they-are) | A label emoji is written as its label only where it opens a line, and some clock faces aren't mapped, so facts set between emoji lose their labels; a decorative one can get a label it shouldn't | Label an emoji wherever it opens a value, map every clock face, and check the value fits the label |

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
| sample | one export's ads, as a `.jsonl` file in `data/ingestion/` |
| run | every stage on every sample together, kept in `data/runs/` under a name: `latest`, or one in `claude/` |
| title group | titles similar enough to be the same job |
| boilerplate | text repeated across different jobs, not about the job |
| template | text repeated across related jobs that does describe the job |
