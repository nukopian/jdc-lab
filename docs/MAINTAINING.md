# Maintaining JDC Lab

How to work on the lab once it's set up ([SETUP.md](SETUP.md)): where
code goes, how to test and style it, how to keep the dependencies tidy,
and what to check before committing. Run every command here from the
lab's folder, with direnv on.

## Where things go

Each language has its own place, with programs kept apart from
reusable code. The README's [Layout](../README.md#layout) lists every
entry at the root.

- **Python** lives in `python/`: programs in `python/bin`, the lab's own
  packages in `python/lib`, notebooks in `python/notebooks`, tests in
  `python/tests`. `.envrc` puts `python/lib` on `PYTHONPATH`, so
  programs and notebooks import those packages without installing them.
  Tool settings are in `python/pyproject.toml`.
- **Go** owns the root, where its tools expect it, but the root holds no
  Go code, only `go.mod` and `go.sum`. Programs go in `cmd/`, one folder
  each, and the packages they share in `internal/`, each with its tests
  and examples beside it.
- **Perl**, which is optional, lives in `perl/`: programs in `perl/bin`,
  the lab's own modules in `perl/lib` (on `PERL5LIB`), tests in
  `perl/t`. Nothing else in the lab may depend on it.
- **`scripts/`** is only for the repository itself, such as building the
  PDFs, never for any language's code.
- **`models/`** and **`data/`** are git-ignored. ONNX exports go into
  `models/`, rebuilt by the programs in `python/bin`, never copied; job
  descriptions go into `data/`.

Python and Go meet only through `models/` and `data/`: Python writes,
Go reads, and neither calls the other.

`python/lib`, `perl/lib`, `perl/t`, `python/tests` and `internal/` are
committed while empty, each holding only a `.gitkeep`. Delete a folder's
`.gitkeep` in the commit that adds its first real file.

## Python

Tests use [pytest](https://docs.pytest.org), and
[ruff](https://docs.astral.sh/ruff/) checks and formats the code,
notebooks included. Both are in the venv, so their versions are pinned
in `requirements.txt` like everything else:

```
pytest python
ruff check python
ruff format python
```

pytest finds the lab's packages in `python/lib` by itself, so tests
also run where direnv doesn't, such as inside an editor. Until the
first test arrives, pytest reports "no tests ran" and exits with code
5; that's expected.

`ruff format` rewrites files in place; `ruff format --check python`
only reports. ruff may also be installed system-wide, as on a Mac with
Homebrew, but the venv's copy is the one whose version is pinned, and
the one `.envrc` puts first.

After installing or upgrading a package, refresh the pins and commit
them with the change that needed them. `--exclude-editable` keeps out
the lab's own packages, once they're installed into the venv, whose
lines would hold a local path:

```
.venv/bin/pip install some-package
.venv/bin/pip check
.venv/bin/pip freeze --exclude-editable > requirements.txt
```

## Go

```
go test ./...
go vet ./...
gofmt -l .
```

`gofmt -l` lists any file that isn't formatted; `gofmt -w .` formats
them. To build the programs, install them: `.envrc` points `GOBIN` at
the git-ignored `bin/` and puts it on the `PATH`, so they then run by
name from anywhere in the lab:

```
go install ./cmd/...
onnx-demo
```

`go run ./cmd/onnx-demo` builds and runs a program without keeping it.
Avoid a bare `go build` inside `cmd/<tool>/`: it leaves the binary
beside the code. `.gitignore` hides it, but it's clutter.

Dependencies come in with the code that imports them: `go get` adds a
module to `go.mod`, and `go mod tidy` drops what's no longer imported.
`go.mod` requires the published versions of grammar and scalar; with a
workspace (see [SETUP.md](SETUP.md#go)), `go env GOWORK` shows that it's
in use, and `GOWORK=off` builds against the published versions, as
anyone else cloning the lab would.

The `go` line in `go.mod` is the oldest Go the lab accepts, so it only
changes when the lab needs something newer. To move to a newer Go,
install it with gvm, make it the default, and raise the line; `.envrc`
follows `go.mod`.

## Perl

Tests use Test2 (`use Test2::V0;`) rather than Test::More, and run with
`yath` rather than `prove`. Perl::Critic reviews the code, with
[`.perlcriticrc`](../.perlcriticrc): severity 3 and up, from the core
policies, its comments saying why each excluded policy is excluded.
Perl::Tidy formats it, with [`.perltidyrc`](../.perltidyrc). Both
profiles are the same as DBIx-Squirrel's, so the Perl reads the same in
both.

```
yath test perl/t
perlcritic perl
for f in $(git ls-files 'perl/*.pl' 'perl/*.pm' 'perl/*.t'); do perltidy "$f" | diff -u "$f" -; done
```

Until the first test arrives, `yath` stops with "No tests were seen!"
and exits with an error; that's expected. The loop prints what
Perl::Tidy would change in each file, and nothing for a file that's
already tidy. `.perltidyrc` sends its output to the screen, for
editors, so to tidy a file in place, without leaving a backup:

```
perltidy -nst -b -bext='/' perl/bin/onnx-demo.pl
```

### The cpanfile

The [`cpanfile`](../cpanfile), at the root beside `requirements.txt` and
`go.mod`, lists the CPAN modules the lab's Perl uses. To keep it tidy:

- **What goes in:** every CPAN module that code in `perl/bin`,
  `perl/lib` or `perl/t` loads directly, with `use` or `require`, and
  nothing else. Leave out modules that come with Perl (`corelist Name`
  says whether one does), and other modules' own dependencies, which
  cpanm finds for itself. The one exception is the test toolchain,
  listed ahead of the first test so that it's in place when one
  arrives.
- **Versions:** give each module the version it was tested with, as its
  minimum. `perl -MName -e 'print $Name::VERSION'` shows the installed
  one.
- **Phases:** modules only the tests use go in an
  `on test => sub { ... }` block, and tools for working on the code,
  such as Perl::Critic, in `on develop => sub { ... }`.
- **Order and comments:** alphabetical within each block. Each tool in
  the `develop` block gets a comment naming the command it provides;
  other modules get one only when they need something unusual, as
  Inline::Python does.
- **When:** change the cpanfile in the same commit as the code that
  starts or stops using a module.

To check it, list the CPAN modules the code loads, compare that with
the cpanfile, then run `cpanm --installdeps --with-develop .` again.
`scan-perl-prereqs`, from Perl::PrereqScanner, reads the code properly,
leaving out modules that ship with Perl:

```
scan-perl-prereqs perl/
```

One line can load two modules: `use Inline Python => …` loads Inline,
which then loads Inline::Python. The scanner lists only Inline, so
Inline::Python is in the cpanfile by hand.

### Notes for later

Things learnt the hard way, kept here so they needn't be learnt again.

- **Inline::Python must stay patched** until a release fixes issue #44.
  On a fresh library, install it as [SETUP.md](SETUP.md#inlinepython)
  says before `cpanm --installdeps --with-develop .`, which would
  otherwise fetch the unpatched 0.58 and fail.
  `cpanm --reinstall Inline::Python` breaks it the same way.
- **Compiled modules may crash on macOS** since Xcode 27.1: a module
  with C or C++ code can build cleanly, then segfault. Rebuilding it
  with the `SDKROOT` and `MACOSX_DEPLOYMENT_TARGET` prefix from
  [SETUP.md](SETUP.md#if-perl-wont-build-on-macos) is the first thing
  to try, then the older Xcode through `DEVELOPER_DIR` described there.
  Neither is sure to help: some modules are simply buggy, as below.
- **Compiler::Lexer 0.23 crashes** whenever it's created in verbose
  mode, even on `use strict;`. It isn't the compiler: it crashes the
  same built with Xcode 26.3's clang 17 as with Xcode 27.1's clang 21.
  The module hasn't changed since 2020, and its issues list many open
  segfaults. So App::scan_prereqs_cpanfile and
  Perl::PrereqScanner::Lite, which use it, fail their tests and won't
  install. Perl::PrereqScanner is pure Perl, built on PPI, and works.
- **Net::SSLeay doesn't build against OpenSSL 4**, which is Homebrew's
  `openssl` now. Build it against `openssl@3` instead, before anything
  that needs it, such as IO::Socket::SSL:
  `OPENSSL_PREFIX=/opt/homebrew/opt/openssl@3 cpanm Net::SSLeay`.
  Net::SSLeay 1.96 builds, passes its tests and makes TLS connections
  that way. cpanm itself fetches from CPAN over plain HTTP, so it isn't
  affected.
- **Test2 under `yath` warns without UUID.** Its UUID support prefers
  the UUID module, and falls back to slower ones with a warning. UUID
  is in the cpanfile's `develop` block for that reason; it builds and
  passes its tests here, so `--notest` isn't needed.
- **`cpanm --test-only` still installs dependencies.** It leaves out
  only the module being tested, and installs whatever is needed to
  test it.
- **`cpanm --uninstall` removes only the modules named,** not what they
  brought in. To undo an install, save the library's list first, and
  compare afterwards:

  ```
  perlbrew list-modules > /tmp/before.txt
  cpanm Some::Module
  perlbrew list-modules | diff /tmp/before.txt -
  ```

## ONNX Runtime

Python's `onnxruntime` and Homebrew's library must stay on the same
version. `requirements.txt` pins Python's, and `brew pin` holds
Homebrew's, so neither moves by accident. To move both, unpin and
upgrade Homebrew's, install the same version in the venv, refreeze,
and pin again:

```
brew unpin onnxruntime && brew upgrade onnxruntime
.venv/bin/pip install "onnxruntime==$(brew list --versions onnxruntime | awk '{print $2}')"
.venv/bin/pip check
.venv/bin/pip freeze --exclude-editable > requirements.txt
brew pin onnxruntime
```

Then run the three demos ([SETUP.md](SETUP.md#check-it-works)): their
tables must still match. Homebrew keeps only the newest version of a
formula, so the pin is what keeps it on the lab's version; going back
once it has moved is awkward. If another formula needs a newer
`onnxruntime`, Homebrew says so, and that's the time to move both.

## Before committing

- **Notebooks:** clear their outputs. They carry local paths, and job
  description text once data is in use.
- **Style and tests:** the checks above, for each language touched.
- **Docs:** a change to tools, versions, dependencies or setup updates
  [SETUP.md](SETUP.md) in the same commit, and a change to how the lab
  is worked on updates this file. Rebuild the PDFs and read the pages
  that changed.
- **Messages:** commits follow
  [Conventional Commits](https://www.conventionalcommits.org) (`feat:`,
  `fix:`, `docs:`, `chore:`, `build(deps):` and so on), and are signed.
- **Secrets:** gitleaks scans the whole history for secrets on every
  push to GitHub. Run it by hand before a tag, and after any history
  rewrite, before pushing:

  ```
  gitleaks git --redact .
  ```

## Documents

`scripts/readme-pdf.sh` builds a PDF of each document, beside it, for
reading offline. The PDFs are git-ignored.

```
readme-pdf.sh
```
