# JDC Lab

A lab for one question: how can search over job descriptions get
better, while spending less on embeddings?

A whole job description makes a poor single embedding. One vector has
to stand for everything in the advert at once: the role, the
requirements, the salary, the benefits, the company blurb and the legal
boilerplate.

## The idea

Not all of a job description deserves an embedding. The lab tests a
pipeline that finds the advert's structure first, then gives each part
the representation that suits it:

- **Metadata** for facts with a definite value, such as salary,
  location, contract type and working pattern.
- **Taxonomy links** for skills and occupations.
- **Embeddings** only for the prose that carries meaning for search:
  the summary, the responsibilities and the requirements.

The rule throughout is to use the cheapest technique that is accurate
enough. Rules and grammars come first, then taxonomies, then
embeddings, and a large language model only where those fall short.

## Where it stands

It's early days.

- A first comparison of two small local embedding models,
  `BAAI/bge-base-en-v1.5` and `intfloat/e5-small-v2`, at matching
  section headings such as "Responsibilities" to section types. BGE
  separated the types more clearly, but the test was too small to
  settle anything, and it has to be rerun fairly for E5.
- ONNX Runtime works from both Python and Go, so a model chosen in
  Python can also run from Go.
- The next step is a sample of real job descriptions, to count which
  section headings occur and how many map to a section type by rule
  alone.

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
| `scripts/` | Scripts for the repository itself: running every language's tests, and building the PDFs |
| `.github/workflows/` | The secret scan that runs on every push |

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
