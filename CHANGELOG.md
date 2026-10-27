# Changes

What changed in each release of JDC Lab, newest first. Each release's
entry is also its signed tag's message. Versions follow
[semantic versioning](https://semver.org), loosely, as befits a lab:
before v1, a release that changes how the lab is set up or worked on
bumps the minor version.

Work collects under Unreleased on `develop`, and becomes a release's
entry when it's tagged ([docs/MAINTAINING.md](docs/MAINTAINING.md#releases)
has the steps).

## Unreleased

- Boilerplate is found when it isn't word for word the same: with the
  ad's own title or place written in, as placeholders, or, in texts of
  20 words or more, with a few words changed, by their runs of five
  words. On the three samples together it's 17.6% of the text, from
  15.8%, and 10% of summaries, from 7%; short lines are matched only
  as written, as alike they're more often about the job. The view's
  ads keep their places: sixteen are always the first by the hash
  (roadmap 3;
  [record](docs/changes/2026-10-27-find-near-duplicate-boilerplate.md)).
- `pipeline run` runs every stage once, on all the samples together,
  in place of each sample and `all`: a run's tables are in
  `data/runs/RUN/STAGE/`, with the samples it read in `samples.txt`. A
  run's name can have a folder, so runs keep to two groups: `latest`,
  for where the pipeline stands, and `claude/NAME`, such as
  `claude/before`, for comparing a change. `compare` warns when two
  runs read different samples. A stage's program still runs on one
  sample by hand.
- Each run writes, for each sample, 20 of its ads in their chunked
  form, to read by eye: `view/ads.md`, with each section's kind and
  heading, fields, repeats, and boilerplate and template text marked,
  and the text no heading placed. The ads are the same in every run,
  so a diff of two runs' views shows what a change did, ad by ad. It
  also checks every ad's text is in its sections, in the ad's order:
  none is lost, but 20 ads of the three samples have text out of order
  (roadmap 16;
  [record](docs/changes/2026-10-27-a-readable-view-of-chunked-ads.md)).

## v0.4.0 (2026-10-27)

Times and shift patterns kept out of headings

Bold times of day and patterns of work, which are facts about the
job, are no longer taken for headings, and headings about shifts start
details sections. The lab gains a third sample, of 21,000 ads, on
which the rules hold, and `pipeline run` times each stage. The
CHANGELOG and change records no longer get PDFs, a change to how the
lab is worked on, so a minor release.

- The conversion reads exports holding control characters. XML 1.0
  forbids characters such as U+0002, but MySQL's `--xml` writes them
  as they are, and Go's decoder refused the whole export. They now
  come through to the JSON Lines unchanged; earlier exports convert
  exactly as before.
- `pipeline run` ends with how each stage ran on each sample: its time
  on the clock and on the processors, ads per second, and peak memory,
  kept too in the run's `timings.tsv`. It builds the stages once,
  instead of with `go run` each time, so the timings leave out
  compiling.
- Bold times of day, such as "- 2pm to 10pm" or "Monday to Friday,
  8:00am - 5:00pm", are no longer headings: 500 of 95,174 heading uses
  across the three samples, each of which had started a section of its
  own, most of them unmapped. A bare range of hours, such as "8-5",
  counts as a time beside a range of days, M-F among them, and
  "9-5pm" implies am (roadmap 2;
  [record](docs/changes/2026-10-25-stop-shift-times-passing-as-headings.md)).
- Bold patterns of work are no longer headings either: a rota ("4 on 4
  off") or days or nights beside a range of days ("Monday - Friday
  (Days)"), unless it ends in a colon and heads what follows. Taken
  for headings, they started unmapped sections that swallowed the
  location, the pay or the opening summary under them. Headings about
  shifts ("Shifts Available:", "Night Shift", "Rotation:") now start
  details sections, not unmapped ones. Across the three samples,
  headings of a known kind rise from 76% to 77% of uses, and ads with
  every heading known from 13,672 to 13,882. A heading with both a
  shift word and a summary phrase now goes to details, as in
  "Production Supervisor – Nights The Role", the ad's title run into
  a heading: a handful of uses.
- The CHANGELOG and change records no longer get PDFs: they're read
  as Markdown, on GitHub or in an editor. `readme-pdf.sh` builds the
  other five documents' PDFs, and refuses these.
- A wash-up before the release found two faults in the new rules and
  fixed them. "Shift" alone among the details phrases took job titles
  for details headings ("Shift Quality Inspector", "HGV Technician
  Shift Supervisor – Leeds"): 59 headings, back to unmapped now that
  the phrases name shift headings ("day shift", "shift times"). And
  "Mon - Fri 7-3 Days", "Mon-Fri 8 till 5" and "4ON4OFF" still passed
  as headings, while "Monmouth to Sunderland" was read as a range of
  days. Ad counts in `pipeline run` no longer read each sample whole,
  or miss a last line without a newline.

## v0.3.5 (2026-10-27)

The rules checked on a second sample

The lab's rules for headings and sections, tuned on its first sample,
now hold on the second too, and the lab takes up three releases of
grammar, two of them with features the lab's own use asked for. No
change to how the lab is set up or worked on, so a patch release.

- The rules hold on the second sample. Headings of a known kind there
  rose from 73% of uses to 76%, as on the first (now 77%), and unmapped
  text fell from 16% to 13%, with new phrases and field labels and two
  rule changes. "OR", "and", "Either:" or "Plus:" alone in bold no
  longer start a section. A bold line ending in a colon can have up to
  20 words, as such lines introduce lists. Boilerplate on the second
  sample fell from 13.3% to 12.5%: some job text had been counted as
  boilerplate only because it sat in unmapped sections, but some
  benefits and application text is now filed under requirements. And
  "legal" alone is no longer a phrase, so a heading such as "The
  successful Legal Secretary will have:" isn't taken for a legal notice:
  a rule written `phrases => kind` leaves the kind's name out
  (roadmap 1; [record](docs/changes/2026-10-25-check-the-rules-on-a-second-sample.md)).
- grammar v0.3.2, with `Overlap` and `Outermost`, both suggested from
  the lab's use: phrases in a heading that overlap now all count, so
  "To succeed in this role, you will need:" is requirements rather than
  responsibilities (six headings, 22 uses), while one wholly inside a
  longer one still doesn't, so "Company Profile" stays company
  ([record](docs/changes/2026-10-26-try-overlap-and-take-up-grammar-v0.3.2.md)).
- grammar v0.3.1, whose `MapOnly` option reads a rule written
  `phrases => kind` without making the kind a phrase, as the lab asked
  for legal's rule. `internal/sections` now uses it, in place of reading
  its rules twice; the pipeline's output is unchanged
  ([record](docs/changes/2026-10-26-take-up-grammar-v0.3.1-and-read-headings-with-maponly.md)).
- grammar v0.3.0, which changes nothing in grammar itself but brings
  scalar from v0.5.9 to v0.6.0 (long numbers keep their digits, and `%`
  and `++` follow Perl further). The pipeline's output is unchanged
  ([record](docs/changes/2026-10-26-take-up-grammar-v0.3.0-and-scalar-v0.6.0.md)).
- A wash-up before the release found only small things, none changing
  the pipeline's output: a test that never reached the code it named,
  a check made twice, and a stale comment and note.

## v0.3.4 (2026-10-26)

A numbered roadmap, and a record of each change

The roadmap numbers its items and keeps ideas, each planned change
keeps a record of how it went, and the lab takes up grammar v0.2.4.
These change how the lab is worked on, which would usually make this
v0.4.0, but they change documents and none of the lab's code, so it's
a patch release, as v0.3.1 was.

- The roadmap numbers its items, and has a fourth group, Ideas, after
  Later, for what might be done, or never, kept so it isn't lost. A
  number stays with its item as it moves between groups and is never
  used again, so an item can be named by number, in pruning and in the
  CHANGELOG ("(roadmap 7)"). Later now means meant to be done, but not
  yet planned. Wash-ups prune the ideas.
- Each planned change keeps a record in
  [docs/changes/](docs/changes/README.md), opened when work starts: the
  plan as it was written, then what happened, with changes from the
  plan, measurements, and what was tried and reverted. The roadmap's
  Now items keep a paragraph and a link to theirs, and the CHANGELOG
  links to it when the change ships.
- grammar v0.2.4, whose limit errors from `ParseErr` say where they
  stopped; the lab doesn't call it, and the pipeline's output is
  unchanged. It still needs scalar v0.5.9.
- A wash-up before the release found PIPELINE.md still saying each
  gap's plan was in the roadmap, after two had moved into change
  records; it now says where to find each.

## v0.3.3 (2026-10-26)

Inline::Python from CPAN

The optional Perl setup gets shorter, and the lab takes up grammar
v0.2.2. A change to how the lab is set up would usually make this
v0.4.0, but this one only shortens an optional step and changes no
code, so it's a patch release.

- Setup installs Inline::Python 0.59 straight from CPAN: it carries the
  macOS fix that 0.58 had to be patched with by hand, so one `cpanm`
  command now installs all the lab's Perl modules.
- grammar v0.2.2, which changes nothing in grammar itself but brings
  scalar from v0.5.7 to v0.5.9 (faster substitution with replacement
  strings). The pipeline's output is unchanged.
- A wash-up before the release found MAINTAINING.md re-running
  `cpanm --installdeps` without `INLINE_PYTHON_EXECUTABLE`. That command
  now upgrades a 0.58, and with direnv off it would have built
  Inline::Python against Homebrew's Python rather than the venv's. It
  now uses SETUP.md's command.

## v0.3.2 (2026-10-25)

Now holds three at most

The roadmap's Now is capped, as grammar's and scalar's now are, and the
lab follows grammar's documentation-only patch release.

- The roadmap's Now holds three items at most, so it stays a queue to
  pull from.
- grammar v0.2.1, which changes only its documentation. It still asks
  for scalar v0.5.7, so jdc-lab stays on that; scalar v0.5.8 is
  documentation only too. The pipeline's output is unchanged.
- A wash-up before the release found only a sentence in
  docs/MAINTAINING.md left broken mid-line by an edit; it's mended.

## v0.3.1 (2026-10-25)

A roadmap

The lab's plans are now public, in a roadmap beside the CHANGELOG and
the README. It changes how the lab is worked on, which would usually
make this v0.4.0, but it adds and moves documents and changes no code,
so it's a patch release.

- A roadmap. ROADMAP.md says what the lab means to do, and why, in Now,
  Next and Later, each item naming the question it serves: structure,
  metadata, taxonomies, embeddings, search or deployment. It's the
  future to go with the CHANGELOG's past and the README's present. The
  plans for the pipeline's known gaps move into it word for word, and
  PIPELINE.md's Known gaps link to them. TODO.md stays for the next
  session's tasks, which come first, and never reaches `main`.
- A wash-up before the release, of documents alone, as no code changed.
  The roadmap's opening called every item "next", Later's ideas
  included, and gave readers a rule meant for whoever maintains it;
  it's rewritten, and every relative link and anchor in the documents
  was checked after so many headings moved.

## v0.3.0 (2026-10-25)

Samples, runs, and releases from develop

The pipeline's figures now come from two samples of ads, and a tool
runs every stage on them and shows what a change moved, so the
pipeline can be improved without losing ground on either. The lab now
works as grammar and scalar do, with work on `develop` and releases on
`main`.

- grammar v0.2.0 (and scalar v0.5.7 with it). `internal/sections`
  takes each kind's rank from the phrase list's `Canonicals`, rather
  than reading `headings.txt` a second time with its own idea of the
  format, and `All` takes its kinds with `ToAll`, no longer reaching
  into scalar. Its sections are the same as before.
- Samples and runs. Each export is a sample of its own,
  `data/ingestion/ads-DATE.jsonl`, named by its date. `cmd/pipeline run
  NAME` runs every stage on every sample, and on all of them together,
  into `data/runs/NAME/`; `cmd/pipeline compare BEFORE AFTER` shows how
  each sample's figures moved, and which headings a change mapped or
  unmapped. docs/PIPELINE.md's way of improving the pipeline uses them,
  and its figures are for both samples: a second, of 5,000 ads leaning
  towards vehicle jobs, joins the first.
- Branches and releases, as in grammar and scalar. Work happens on
  `develop`, where CI runs on every push, and `main` is the latest
  release, moved only by a fast-forward at a release, and protected.
  `TODO.md` lives on `develop` and never reaches `main`, and the plans
  for the pipeline's known gaps, which were in it, are now in
  docs/PIPELINE.md, word for word, beside the gaps. CI now runs the
  Go checks and tests (gofmt, vet, the race detector, govulncheck) as
  well as the secret scan. docs/MAINTAINING.md has the branches and the
  release steps.
- The README says why the lab is public: it's for learning as much as
  research, so its code, method and findings are open, while the job
  ads it reads, which aren't ours to publish, and the detailed results
  drawn from them are kept privately. `data` can be a link to such a
  private store. Where it stands now describes the pipeline and its
  figures, and is kept current with every change.
- A wash-up before the release. It found and fixed
  `cmd/pipeline compare` stopping at the first sample in only one of
  the two runs, which the first new export would have hit, so no sample
  after it was compared, and skipping without a word any sample gone
  from the second run; and 15 of 717 headings starting "why" filed by a
  phrase inside them, such as "Why you'll love this role" as
  responsibilities and "Why should you apply?" as apply, when they're
  benefits. Only the second sample's figures moved, and only apply's
  boilerplate, from 40% to 41%. Smaller: `cmd/heading-counts` ended a
  summary line with spaces.

## v0.2.0 (2026-10-25)

From job ads to sections and boilerplate

The first work on real job ads: a pipeline, in Go, that splits each ad
into blocks, finds its headings and fields, divides it into sections,
and finds the boilerplate ads repeat whatever the job, all measured on
a sample of ads and described in docs/PIPELINE.md.

- Job ads in, as JSON Lines. `cmd/xml-to-jsonl` converts a MySQL
  `--xml` export of job ads into one JSON object per ad, with the lab's
  own field names, for every language to read.
- Job ads split into blocks. `cmd/blocks` splits each ad's HTML into
  the paragraphs, lines and lists a reader sees, keeping their text as
  written, with cues from the markup, such as text that is all bold.
  Ads in plain text are laid out by their newlines. Deciding which
  blocks are headings is left to the next step.
- Headings and fields found. `internal/headings` gives each block a
  role: a heading, a field such as "Salary: £30,000", a repeat of the
  ad's title or location, or text. `cmd/heading-counts` counts how many
  ads use each heading and field label.
- Sections. `internal/sections` divides an ad into its summary,
  responsibilities, requirements, desirable, benefits, company, apply,
  legal and details sections, by the phrases in its headings. A heading
  of no known kind starts an unmapped section, so what's left to map
  shows. The phrases are data, synonym rules in the format
  Elasticsearch and Solr use, read with a `grammar` word list.
  `cmd/sections` measures how well they cover the ads, and lists the
  headings they miss.
- Boilerplate. `internal/boilerplate` finds text that ads repeat
  whatever the job, word for word in ads for different jobs, counting
  similar titles as one job so that a job posted in many places isn't
  taken for boilerplate. Repeated text mostly in the sections that
  describe the job is a template for it instead, unless it's so
  widespread that it says nothing about any job. `cmd/boilerplate`
  measures how much of each kind of section is boilerplate.
- How it works, written down. `docs/PIPELINE.md` describes the flow
  from export to boilerplate: each stage, how to run it, what it
  decides and why, what it measured, and what it doesn't do yet, and
  how to improve each stage from the tables it writes. `cmd/heading-uses`
  shows how the ads use any one heading.
- A wash-up before the release. Among smaller things, it found and
  fixed a crash in `cmd/heading-counts` on a file with no ads; blocks
  put out of order by a stray `</li>` or `</ul>`, or by a heading tag
  inside a list item; a title with no words being grouped with every
  other when finding boilerplate; and sections started by a plain line
  going uncounted when the same ad had the same heading in bold. Only
  the last changed a figure on the sample: 952 lines start a section by
  name, not 950. The programs now share one way of reading ads, and
  their errors say which ad.
- Go does most of the work. Python stays for what needs its
  machine-learning tools: evaluating models, exporting them to ONNX,
  and notebooks.
- Type checking. ty, pinned in the venv, checks the Python code's
  annotations, and `test-all.sh` fails on a type error as on a failing
  test.

## v0.1.0 (2026-10-24)

The lab's foundations

One repository for Python and Go, with Perl as an optional extra, ready
for the retrieval experiments to come.

- Three languages, one environment. Each language has its own folders
  for programs, shared code and tests. A direnv `.envrc` sets up the
  venv, perlbrew's Perl and gvm's Go on entering the folder.
- ONNX Runtime from every language. A small cosine-similarity model,
  built in Python, runs from Python, from Go through `onnxruntime_go`,
  and from Perl through Inline::Python. The tests hold all three to the
  same scores.
- Tests and style. pytest and ruff for Python; `go test`, `go vet` and
  `gofmt` for Go; Test2 with `yath`, Perl::Critic and Perl::Tidy for
  Perl. `scripts/test-all.sh` runs every language that's set up.
- Docs. The README, `docs/SETUP.md` (from nothing to a working lab) and
  `docs/MAINTAINING.md` (working on it).
- Hygiene. Signed commits, gitleaks on every push, and nothing private
  in the repository.
