# Setting up JDC Lab

How to set the lab up from nothing on a Mac: every tool, where it comes
from, and how to check it works. Other systems need the same tools, from
their own package managers.

Keep this guide in step with the lab. When a tool, version or setup
step changes, the change and the guide go in the same commit.

## Tools

| Tool | Version | For | Install on macOS |
|------|---------|-----|------------------|
| Homebrew | any | installing the rest | [brew.sh](https://brew.sh) |
| git, gh | any recent | the repository; gh for GitHub | `brew install gh` |
| pyenv | any recent | the Python version | `brew install pyenv` |
| Python | 3.12.12 | experiments, ONNX export | `pyenv install 3.12.12` |
| Go | 1.26.10 or newer | running models | [go.dev/dl](https://go.dev/dl/) or a version manager |
| ONNX Runtime | 1.31.0 | running ONNX models from Go | `brew install onnxruntime` |
| gitleaks | 8.30.1 | scanning history for secrets | `brew install gitleaks` |
| pandoc, Prince | any recent | PDFs of the documents | `brew install pandoc` and `brew install --cask prince` |
| direnv | any recent | optional: activates the venv on `cd` | `brew install direnv` |

## Get the code

```
git clone git@github.com:nukopian/jdc-lab.git
```

The Go side can use two sibling projects,
[grammar](https://github.com/nukopian/grammar) and
[scalar](https://github.com/nukopian/scalar). Clone them beside the lab
if you want to work on them at the same time, so the three share a
parent folder:

```
git clone git@github.com:nukopian/grammar.git
git clone git@github.com:nukopian/scalar.git
```

## Python

`.python-version` names the Python, and pyenv picks it up in the lab's
folder. `requirements.txt` pins every package, including indirect ones,
for that version.

```
cd jdc-lab
pyenv install --skip-existing "$(cat .python-version)"
"$(pyenv prefix)/bin/python" -m venv .venv
.venv/bin/pip install -r requirements.txt
.venv/bin/pip check
```

To activate the venv on entering the folder, create `.envrc` (it's
git-ignored) and allow it:

```
echo 'source .venv/bin/activate' > .envrc
direnv allow
```

After installing or upgrading a package, refresh the pins with
`.venv/bin/pip freeze > requirements.txt` and commit them.

### Models

The embedding models come from Hugging Face. They're downloaded on
first use into Hugging Face's cache (`~/.cache/huggingface`), not into
the lab. ONNX exports go into `models/`, which is git-ignored, so they
are rebuilt from the scripts, not copied.

## ONNX Runtime

Two copies, one for each language:

- **Python** uses the `onnxruntime` package in the venv.
- **Go** uses the shared library from Homebrew,
  `/opt/homebrew/lib/libonnxruntime.dylib`, through
  [`github.com/yalue/onnxruntime_go`](https://github.com/yalue/onnxruntime_go).

Keep them on the same version, so that a model gives the same answers
from both languages. To compare them:

```
.venv/bin/python -c 'import onnxruntime; print(onnxruntime.__version__)'
brew list --versions onnxruntime
```

Exporting a model to ONNX needs `onnx` and `onnxscript`, which are in
`requirements.txt`.

## Go

The lab is one Go module, `github.com/nukopian/jdc-lab`, at the root of
the repository. Programs go in `cmd/`, and the packages they share in
`internal/`.

`go.mod` requires the published versions of grammar and scalar. To use
your sibling clones instead, without publishing a release, create a
workspace. It's git-ignored, as it depends on where the clones are:

```
go work init . ../grammar ../scalar
```

`go env GOWORK` shows whether a workspace is in use. Set `GOWORK=off` to
build against the published versions, as anyone else cloning the lab
would.

## Data

Job descriptions are not in the repository, and never go into it. Put
them in `data/`, which is git-ignored.

## Checks

Before committing notebooks, clear their outputs: they carry local
paths, and job description text once data is in use.

gitleaks scans the whole history for secrets on every push to GitHub.
Run it by hand before a tag, and after any history rewrite, before
pushing:

```
gitleaks git --redact .
```

Commits follow [Conventional Commits](https://www.conventionalcommits.org)
(`feat:`, `fix:`, `docs:`, `chore:`, `build(deps):` and so on), and are
signed. To sign with an SSH key, in the lab's folder:

```
git config gpg.format ssh
git config user.signingkey ~/.ssh/id_ed25519.pub
git config commit.gpgsign true
```

Use the public half of whichever key GitHub has as your signing key.

## Documents

`scripts/readme-pdf.sh` builds a PDF of each document, beside it, for
reading offline. The PDFs are git-ignored.

```
scripts/readme-pdf.sh
```
