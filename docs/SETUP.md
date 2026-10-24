# Setting up JDC Lab

How to set the lab up from nothing on a Mac: every tool, where it comes
from, and how to check it works. Other systems need the same tools, from
their own package managers. [MAINTAINING.md](MAINTAINING.md) covers
working on the lab once it's set up.

Keep this guide in step with the lab. When a tool, version or setup
step changes, the change and the guide go in the same commit.

## Tools

| Tool | Version | For | Install on macOS |
|------|---------|-----|------------------|
| Homebrew | any | installing the rest | [brew.sh](https://brew.sh) |
| git, gh | any recent | the repository; gh for GitHub | `brew install gh` |
| direnv | any recent | sets up the environment on `cd` | `brew install direnv` |
| pyenv | any recent | the Python version | `brew install pyenv` |
| Python | 3.12.12 | experiments, ONNX export | see [Python](#python) |
| ONNX Runtime | 1.31.0 | running ONNX models from Go | `brew install onnxruntime` |
| gvm | any recent | Go versions, side by side | see [Go](#go) |
| Go | 1.26.10 or newer | running models | `gvm install go1.26.10 -B` |
| perlbrew | any recent | optional: Perl versions and module libraries | see [Perl](#perl) |
| Perl | 5.36.3 | optional: working through text | see [Perl](#perl) |
| gitleaks | 8.30.1 | scanning history for secrets | `brew install gitleaks` |
| pandoc, Prince | any recent | PDFs of the documents | `brew install pandoc` and `brew install --cask prince` |

## Get the code

```
git clone git@github.com:nukopian/jdc-lab.git
cd jdc-lab
```

The Go side can use two sibling projects,
[grammar](https://github.com/nukopian/grammar) and
[scalar](https://github.com/nukopian/scalar). Clone them beside the lab
if you want to work on them at the same time, so the three share a
parent folder:

```
git clone git@github.com:nukopian/grammar.git ../grammar
git clone git@github.com:nukopian/scalar.git ../scalar
```

Commits are signed. To sign with an SSH key, in the lab's folder, using
the public half of whichever key GitHub has as your signing key:

```
git config gpg.format ssh
git config user.signingkey ~/.ssh/id_ed25519.pub
git config commit.gpgsign true
```

## direnv

The lab's [`.envrc`](../.envrc) sets up its environment whenever you
enter the folder:

- **Python:** it activates the venv, and puts `python/lib` on
  `PYTHONPATH`.
- **Perl:** it switches to `perl-5.36.3@jdc-lab`, puts `perl/lib` ahead
  of that library on `PERL5LIB`, and adds the venv's packages to
  `PYTHONPATH`, after `python/lib`, so that Inline::Python sees them.
- **Go:** it puts gvm's Go at the version in `go.mod` first on the
  `PATH`, whatever `gvm use` says elsewhere. Raising the version in
  `go.mod` moves the lab with it, once gvm has that version installed.
  It also points `GOBIN` at `bin/`, for `go install`.
- **`bin/` and `scripts/`:** it puts both on the `PATH`, ahead of the
  system's folders, so the lab's programs and the repository's scripts
  run by name. `bin/` is added even before the first `go install`
  creates it, so it needs no reload afterwards.
- **`.env`:** it reads it, if there is one.

It skips any of the three languages, with a note, when its setup isn't
there yet, so it's safe to allow before the steps below. Without gvm,
the lab uses whichever `go` is on the `PATH`. Any Go from 1.21 on will
then download the version `go.mod` asks for.

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
`log_status`, `path_add`, `PATH_add` and `dotenv_if_exists` come from
there. `man direnv-stdlib` documents the library, and `direnv stdlib`
prints its source.

## Python

`.python-version` names the Python, and pyenv picks it up in the lab's
folder. `requirements.txt` pins every package, including indirect ones,
for that version, the lab's tools (pytest, ruff) included.

Build Python with `--enable-shared`, which pyenv doesn't do by default.
Only the optional Perl side needs it, because Inline::Python links
against Python's shared library. It does no harm otherwise, and saves
rebuilding Python if you want Perl later.

```
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

The embedding models come from Hugging Face. They're downloaded on
first use into Hugging Face's cache (`~/.cache/huggingface`), not into
the lab.

## ONNX Runtime

Two copies, one for each language:

- **Python** uses the `onnxruntime` package in the venv, installed
  above.
- **Go** uses the shared library from Homebrew,
  `/opt/homebrew/lib/libonnxruntime.dylib`, through
  [`github.com/yalue/onnxruntime_go`](https://github.com/yalue/onnxruntime_go).

```
brew install onnxruntime
```

Keep them on the same version, so that a model gives the same answers
from both languages. To compare them:

```
.venv/bin/python -c 'import onnxruntime; print(onnxruntime.__version__)'
brew list --versions onnxruntime
```

## Go

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

If you cloned grammar and scalar beside the lab, a workspace makes the
lab build against those clones rather than their published releases.
It's git-ignored, as it depends on where the clones are:

```
go work init . ../grammar ../scalar
```

## Perl

Perl is optional, and nothing else in the lab depends on it: skip this
section unless you want it. It's a handy tool for working through text
along the way, and through Inline::Python it can call the lab's Python,
ONNX Runtime included.

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
folder, so run `direnv allow` again now: modules installed with `cpanm`
in the lab then go into the lab's library.

#### If Perl won't build on macOS

Since Xcode 27.1 (Apple clang 21, building against the macOS 27.0 SDK),
`perlbrew install` fails on macOS 26.7 for every Perl version. The build
stops at "Failed to build miniperl", because miniperl crashes with a
segmentation fault, whether or not `--notest` is given. Building
against an older SDK works. The Command Line Tools keep several, and
`xcode-select --install` installs them:

```
ls /Library/Developer/CommandLineTools/SDKs/
SDKROOT=/Library/Developer/CommandLineTools/SDKs/MacOSX26.5.sdk MACOSX_DEPLOYMENT_TARGET=26.0 perlbrew install perl-5.36.3 --notest
```

Use the newest SDK older than the one that fails: `xcrun --show-sdk-version`
shows which one the compiler is using. The variable is `SDKROOT`, with
no underscore: clang ignores `SDK_ROOT` and keeps its default SDK.
Modules installed with `cpanm` afterwards, Inline::Python included,
build without either variable. Check whether this is still needed after
the next Xcode update.

### Inline::Python

Inline::Python 0.58 doesn't link Python's library on macOS unless
Python is Apple's own, so it fails all its tests and won't install
([issue #44](https://github.com/niner/inline-python-pm/issues/44) has
the cause and the fix). Until a fixed release is out, install it with
that one-line fix applied. When one is, replace these steps with plain
`cpanm Inline::Python`. In the lab's folder, with direnv on:

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

### The rest

Then install the rest of the lab's Perl modules, listed in the
[`cpanfile`](../cpanfile), with the tools for working on the code:

```
cpanm --installdeps --with-develop .
```

That installs the test toolchain (Test2::V0 and its plugins) and the
tools (`yath`, `perlcritic`, `perltidy` and `scan-perl-prereqs`). Leave
out `--with-develop` to skip the tools. It must come after
Inline::Python: on its own, it would fetch the unpatched 0.58 and fail.

## Check it works

Each language runs the same small demo: Python builds an ONNX model
that scores how alike some vectors are, saves it to
`models/onnx-demo.onnx`, and runs it; Go runs the same file; Perl, if
you set it up, runs it through Inline::Python. Run Python's first:

```
python python/bin/onnx_demo.py
go run ./cmd/onnx-demo
perl perl/bin/onnx-demo.pl
```

Each prints the ONNX Runtime version, then the same table, with `1.000`
along the diagonal. If the tables differ, the copies of ONNX Runtime
have drifted apart. The Go demo looks for the library in Homebrew's
place, unless `ONNXRUNTIME_LIB` names another.

## Data

Job descriptions are not in the repository, and never go into it. Put
them in `data/`, which is git-ignored.
