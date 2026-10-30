# JDC Lab

**Exploring better representations of job descriptions for search**

A lab for one question: how can search over job descriptions get
better, while spending less on embeddings?

A whole job description makes a poor single embedding. One vector has
to stand for everything in the advert at once: the role, the
requirements, the salary, the benefits, the company blurb and the legal
boilerplate.

## Why it's public

This is a learning lab as much as a research one. I'm using it to learn
Go, to get hands-on with embedding models and ONNX, and to work out how
search over text like this should be built, in the open, where anyone
can follow along and reuse what's useful.

So its code, its method and its findings are public; its data isn't.
The job ads it reads were written by the employers who placed them, and
aren't ours to publish, so the ads, and the detailed results drawn from
them, are kept privately. What's here is the code to run, how it works
([docs/PIPELINE.md](docs/PIPELINE.md)), and figures summed over many
ads, never an ad itself.

## The idea

Not all of a job description deserves an embedding. The lab tests a
pipeline that finds the advert's structure first, then gives each part
the representation that suits it:

- **Metadata** for facts with a definite value, such as salary,
  location, contract type and working pattern.
- **Taxonomy links** for skills and occupations.
- **Embeddings** only for the prose that carries meaning for search:
  the summary, the role, the responsibilities and the requirements.

The rule throughout is to use the cheapest technique that is accurate
enough. Rules and grammars come first, then taxonomies, then
embeddings, and a large language model only where those fall short.

## Where it stands

- **A pipeline, by rules alone.** It splits each ad into the blocks a
  reader sees, as its author laid them out, by newlines as well as tags,
  splits walls of text into sentences, finds its headings and fields,
  those run into walls too, divides it into sections
  (summary, role, responsibilities, requirements, benefits and the
  rest), and
  finds the boilerplate that ads repeat whatever the job. Grammars read
  the facts ads state, pay, hours, work pattern, schedule and postcode,
  as values. `shape` shows any ad at each stage, from its HTML to its
  chunks, to see where its shape changes. `chunk` shows any ad's chunks,
  saying what to embed, what
  becomes metadata, and what placed each, as text, data, or web pages
  of ads before and after, and each
  run shows 20 ads in their chunks, the same in every run, to read by
  eye. [docs/PIPELINE.md](docs/PIPELINE.md) says how each stage works,
  what it measured, and how to improve it.
- **Measured on two samples** of 2,000 and 5,000 ads, from different
  kinds of job, and checked on a third of 21,000. 88% of ads have
  headings, and about four in five heading uses are recognised by
  rule, on all three. About a fifth of an ad is boilerplate, found
  even with a few words changed, or company, application and legal
  text, that probably needn't be embedded: a first answer to whether
  search can improve while embedding less. A third of ads state a
  working pattern, such as "Monday to Friday, 9am–5pm": a fact for
  metadata, not embedding. Chunked, an ad's embedded text is 68% of its
  tokens, its title and the employer's own words included: the rest is
  facts, benefits, notices and boilerplate. Text under a heading of no known kind is placed by an
  educated guess, mostly as the summary, and right for about four
  sections in five.
- **Next:** the [roadmap](ROADMAP.md) says what the lab means to do,
  and why: nothing is under way, and its ideas are next to choose
  from.
- **Retrieval** has had a first test: with job titles as queries, an
  ad without its boilerplate is found as well as the whole ad, for 16%
  fewer tokens, and the summary alone nearly as well, for a quarter of
  them. Titles are a narrow test, and whether boilerplate's verdict is
  right is still to be checked, so nothing is dropped yet.
- **Embeddings** have had a first look: two small local models,
  `BAAI/bge-base-en-v1.5` and `intfloat/e5-small-v2`, compared at
  matching section headings to section types. BGE separated the types
  more clearly, but the test was too small to settle anything, and it
  has to be rerun fairly for E5. ONNX Runtime works from both Python and
  Go, so a model chosen in Python can also run from Go.

## Layout

Go does most of the work. Python is used where its machine-learning
tools are needed: evaluating models, exporting them to ONNX, and
notebooks. Perl is optional: a handy tool for working through
text along the way, able to call the lab's Python. Nothing in the lab
depends on it.

Each language has its own place, with programs and reusable code kept
apart. Everything in the repository, folder by folder:

| Path | What's there |
|------|--------------|
| `python/bin/`, `python/lib/` | Python programs, and the lab's Python package, `jdclab` |
| `python/notebooks/`, `python/tests/` | Jupyter notebooks for exploration, and tests |
| `python/pyproject.toml` | Settings for pytest and ruff |
| `cmd/`, `internal/` | Go programs and packages |
| `perl/bin/`, `perl/lib/`, `perl/t/` | Optional Perl programs, the lab's Perl modules, and their tests |
| `docs/` | [PIPELINE.md](docs/PIPELINE.md), [SETUP.md](docs/SETUP.md) and [MAINTAINING.md](docs/MAINTAINING.md), and the PDFs' style sheet |
| `docs/changes/` | A [record](docs/changes/README.md) of each planned change: the plan, and what happened |
| `scripts/` | Scripts for the repository itself: running every language's tests, and building the PDFs |
| `.github/workflows/` | CI on every push: the Go checks and tests, and the secret scan |

And file by file, at the root:

| File | What it's for |
|------|---------------|
| `requirements.txt` | Python packages, every one pinned |
| `.python-version` | The Python version, for pyenv |
| `go.mod`, `go.sum` | The Go module and its dependencies |
| `cpanfile` | Perl modules, with the tools for working on the code |
| `.perltidyrc`, `.perlcriticrc` | Perl::Tidy's and Perl::Critic's profiles |
| `.envrc` | Sets up Python, Perl and Go on entering the folder, with direnv |
| `.gitignore` | What stays out of git, by language |
| `CHANGELOG.md` | What changed in each release |
| `ROADMAP.md` | What the lab means to do, and why |

Kept out of git: `data/` (job descriptions), `models/` (models and ONNX
exports), `bin/` (Go programs once built), and `.venv/`. gitleaks scans
the whole history for secrets on every push, and GitHub's secret
scanning and push protection are on as well.

## Running it

The lab needs Python 3.12, Go 1.26 and ONNX Runtime, and optionally
Perl 5.36. [SETUP.md](docs/SETUP.md) sets everything up from nothing,
step by step, and [MAINTAINING.md](docs/MAINTAINING.md) covers working
on the lab: tests, style, dependencies and the checks before
committing. [PIPELINE.md](docs/PIPELINE.md) describes how the lab
turns a job ad into blocks, headings, sections and boilerplate, what
each stage decides, and what it has measured so far.
