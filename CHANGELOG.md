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

- Setup installs Inline::Python 0.59 straight from CPAN: it carries the
  macOS fix that 0.58 had to be patched with by hand, so one `cpanm`
  command now installs all the lab's Perl modules.

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
