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

- Job ads in, as JSON Lines. `python/bin/xml_to_jsonl.py` converts a
  MySQL `--xml` export of job ads into one JSON object per ad, with the
  lab's own field names, for every language to read.
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
