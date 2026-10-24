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

Python is used for experiments and for exporting models to ONNX, and
Go for running them.

| Path | What's there |
|------|--------------|
| `notebooks/` | Jupyter notebooks for exploration |
| `scripts/` | Python programs that produce models and datasets |
| `cmd/`, `internal/` | Go programs and packages |
| `data/`, `models/` | Job descriptions, models and ONNX exports, kept out of git |

Job descriptions and model files stay out of the repository. gitleaks
scans the whole history for secrets on every push, and GitHub's secret
scanning and push protection are on as well.

## Running it

Python 3.12 (see `.python-version`):

```
python -m venv .venv
.venv/bin/pip install -r requirements.txt
```

Go 1.26, and ONNX Runtime's shared library for the Go side. On macOS,
that is `brew install onnxruntime`.
