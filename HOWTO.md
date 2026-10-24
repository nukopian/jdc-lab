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
| Python | 3.12.12 | experiments, ONNX export | see [Python](#python) |
| gvm | any recent | Go versions, side by side | see [Go](#go) |
| Go | 1.26.10 or newer | running models | `gvm install go1.26.10 -B` |
| ONNX Runtime | 1.31.0 | running ONNX models from Go | `brew install onnxruntime` |
| perlbrew | any recent | Perl versions and module libraries | see [Perl](#perl) |
| Perl | 5.36.3 | text work, calling Python | `perlbrew install perl-5.36.3` |
| Inline::Python | 0.58, patched | calling Python from Perl | see [Perl](#perl) |
| gitleaks | 8.30.1 | scanning history for secrets | `brew install gitleaks` |
| pandoc, Prince | any recent | PDFs of the documents | `brew install pandoc` and `brew install --cask prince` |
| direnv | any recent | sets up the environment on `cd` | `brew install direnv` |

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

Build Python with `--enable-shared`, which pyenv doesn't do by default.
Inline::Python links against Python's shared library, so the Perl side
needs it.

```
cd jdc-lab
PYTHON_CONFIGURE_OPTS=--enable-shared pyenv install --skip-existing "$(cat .python-version)"
"$(pyenv prefix)/bin/python" -m venv .venv
.venv/bin/pip install -r requirements.txt
.venv/bin/pip check
```

If that Python was already installed, check how it was built: this
prints `1` for a shared build. Otherwise, reinstall it with `--force`
added to `pyenv install`, and rebuild the venv.

```
"$(pyenv prefix)/bin/python" -c 'import sysconfig; print(sysconfig.get_config_var("Py_ENABLE_SHARED"))'
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

### gvm

[gvm](https://github.com/moovweb/gvm) keeps several Go versions side by
side and switches between them, for the whole machine or for one
shell. That makes it easy to try the lab on another Go, or to build
grammar and scalar on the oldest Go they support. It needs the Xcode
Command Line Tools, which provide git, make, gcc and bison:

```
xcode-select --install
```

Install gvm with its installer script. It clones gvm into `~/.gvm` and
adds a line to `~/.zshrc` (or `~/.bashrc`) that loads gvm in each new
shell. Then load it into the current shell:

```
zsh < <(curl -sSL https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source ~/.gvm/scripts/gvm
```

Install the lab's Go and make it the default. `-B` installs Go's
published binary. Without `-B`, gvm builds Go from source, which needs
an older Go to build it.

```
gvm install go1.26.10 -B
gvm use go1.26.10 --default
go version
```

`gvm list` shows the installed versions. `gvm use go1.25.14` switches
the current shell only, and a new shell starts on the default again.
To move to a newer Go, install it and make it the default the same
way. The `go` line in `go.mod` is the oldest Go the lab accepts, so it
only changes when the lab needs something newer.

### The module

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

## Perl

Perl is for text work, and through Inline::Python it can call the lab's
Python, ONNX Runtime included.

### perlbrew

[perlbrew](https://perlbrew.pl) keeps Perl versions apart from the
system's, and gives each project its own library of modules. Install
it, then load it in every new shell by adding the line it prints to
`~/.zshrc`. It's usually this one:

```
curl -L https://install.perlbrew.pl | bash
echo 'source ~/perl5/perlbrew/etc/bashrc' >> ~/.zshrc
source ~/perl5/perlbrew/etc/bashrc
```

Then install the lab's Perl, cpanm, and the lab's library. Building
Perl takes a while: adding `--notest` to `perlbrew install` skips Perl's
own test suite to save time.

```
perlbrew install perl-5.36.3
perlbrew install-cpanm
perlbrew lib create perl-5.36.3@jdc-lab
```

The lab's `.envrc` switches to `perl-5.36.3@jdc-lab` on entering the
folder (see [direnv](#direnv)), so modules installed with `cpanm` there
go into the lab's library.

### Inline::Python

Inline::Python 0.58 doesn't link Python's library on macOS unless
Python is Apple's own, so it fails all its tests and won't install.
Until a fixed release is out ([TODO.md](TODO.md) has the details), install
it with a one-line fix applied. In the lab's folder, with direnv on:

```
lab=$PWD
cd "$(mktemp -d)"
curl -sSL https://cpan.metacpan.org/authors/id/N/NI/NINE/Inline-Python-0.58.tar.gz | tar -xz
cd Inline-Python-0.58
perl -pi -e 's/\(\?:\\\.\\Q\$Config\{dlext\}\\E\|/(?:\\.\\Q\$Config{so}\\E|\\.\\Q\$Config{dlext}\\E|/' Makefile.PL
grep -c 'Config{so}' Makefile.PL
INLINE_PYTHON_EXECUTABLE="$lab/.venv/bin/python3" cpanm -v .
cd "$lab"
```

`grep` prints `1` when the fix is in. `INLINE_PYTHON_EXECUTABLE` builds
it against the venv's Python, without asking which one to use. The
build links that Python's shared library, so after changing the lab's
Python, install Inline::Python again.

To check it works, run the demo. It builds a small ONNX model in Python,
runs it with ONNX Runtime, and passes Perl arrays in and out:

```
perl perl/onnx-demo.pl
```

It prints the ONNX Runtime and Perl versions, then a table of how alike
three vectors are, with `1.000` along the diagonal.

## direnv

The lab's [`.envrc`](.envrc) sets up its environment whenever you
enter the folder:

- **Python:** it activates the venv.
- **Perl:** it switches to `perl-5.36.3@jdc-lab`, and sets `PYTHONPATH`
  so that Inline::Python sees the venv's packages.
- **Go:** it puts gvm's Go at the version in `go.mod` first on the
  `PATH`, whatever `gvm use` says elsewhere. Raising the version in
  `go.mod` moves the lab with it, once gvm has that version installed.
- **`.env`:** it reads it, if there is one.

It skips any of the three languages, with a note, when its setup isn't
there yet. Without gvm, the lab uses whichever `go` is on the `PATH`.
Any Go from 1.21 on will then download the version `go.mod` asks for.

Hook direnv into the shell once, by adding this line to `~/.zshrc`:

```
eval "$(direnv hook zsh)"
```

direnv only runs an `.envrc` you've allowed. Allow the lab's after
cloning, and again whenever it changes:

```
direnv allow
```

Values that are yours alone, such as tokens, go in `.env`. It's
git-ignored, and holds lines of the form `NAME=value`. direnv reads it
as data, without running it as shell code, and watches it: creating or
editing `.env` reloads the environment at the next prompt, with no
`direnv allow` needed.

`.envrc` is bash, with direnv's standard library loaded first.
`log_status` and `dotenv_if_exists` come from there. `man direnv-stdlib`
documents the library, and `direnv stdlib` prints its source.

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
