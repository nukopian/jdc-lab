# Changes

What changed in each release of JDC Lab, newest first. Each release's
entry is also its signed tag's message. Versions follow
[semantic versioning](https://semver.org), loosely, as befits a lab:
before v1, a release that changes how the lab is set up or worked on
bumps the minor version.

Work collects under Unreleased on `main`, and becomes a release's entry
when it's tagged ([docs/MAINTAINING.md](docs/MAINTAINING.md#releases)
has the steps).

## Unreleased

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
