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
  descriptions go into `data/ingestion/`. Each arrives there as a MySQL
  `--xml` export, named by its date (`extract-20261025.xml`), which is
  never edited. `go run ./cmd/xml-to-jsonl` converts it once to a
  sample of the same date beside it (`ads-20261025.jsonl`), JSON Lines
  with one ad per line (`id`, `title`, `location`,
  `location_canonical`, `html`): the lab's code reads that, not the
  XML. Samples are kept apart, so a rule tuned on one can be checked on
  the others; `cmd/pipeline` runs them all into `data/runs/`.
  Everything else in `data/` is derived from `data/ingestion/`, each
  step in a folder of its own, so any of it can be deleted and rebuilt.

Python and Go meet only through `models/` and `data/`: each reads what
the other writes there, and neither calls the other.

A program stays thin: the work it does goes in its language's library
folder, where the tests can reach it. The ONNX demo shows the pattern in
all three: `jdclab.onnx_demo`, `internal/onnxdemo` and
`JDCLab::OnnxDemo` hold the model and its use, their tests check the
same scores to within a millionth, and the three demo programs only
print the table.

## Branches

- `develop` is where work happens. Every commit goes there, and CI runs
  on every push to it.
- `main` is the latest release, and nothing else. It moves only when a
  release is made, by a fast-forward to `develop`, and the release's tag
  is on it. Visitors to GitHub see `main`, so its README and docs
  describe the lab as released. Releases come often, at the close of
  each run of work, so `main` is never far behind.
- There are no feature branches, for now: changes are made one at a
  time.
- Plans for the pipeline's known gaps live in PIPELINE.md's
  [Known gaps](PIPELINE.md#known-gaps), one for each gap, where readers
  of the pipeline find them, and go when the gap is closed.
- `TODO.md`, when there is one, lists other work planned for a later
  session, such as tasks from grammar's and scalar's sessions. It's
  committed on `develop` and gone before a release, so it never reaches
  `main`: each task is done, or, if it's a gap in the pipeline, moved
  word for word into PIPELINE.md's plans (see
  [Before committing](#before-committing) and [Releases](#releases)).

`main` is protected on GitHub. Pushes to it need signed commits, a linear
history (no merge commits) and CI passing on every job, and it can't be
force-pushed or deleted. This holds for admins too. To see the settings:

```
gh api repos/nukopian/jdc-lab/branches/main/protection --jq '{signatures: .required_signatures.enabled, linear: .required_linear_history.enabled, checks: .required_status_checks.contexts, force_push: .allow_force_pushes.enabled, admins: .enforce_admins.enabled}'
```

For a fresh clone, or a checkout still on `main`, switch to `develop`:

```
git fetch origin
git switch develop
git pull --ff-only
```

## Running all the tests

`scripts/test-all.sh` runs every language's tests that can run here, and
skips, with a note, any language that isn't set up. It's on the `PATH`
in the lab, so from any folder in it:

```
test-all.sh
```

It ends with a line per language (pass, FAIL, or why it skipped), and
fails if any suite fails. The Go tests run the model the Python demo
saves, so it makes the model first if it's missing.

## Python

Tests use [pytest](https://docs.pytest.org),
[ruff](https://docs.astral.sh/ruff/) checks and formats the code,
notebooks included, and [ty](https://docs.astral.sh/ty/) checks the
type annotations. All three are in the venv, so their versions are
pinned in `requirements.txt` like everything else:

```
pytest python
ruff check python
ruff format python
ty check --project python
```

Annotate every function in `python/lib` in full: its arguments and what
it returns. Programs and tests use annotations where they help, and
notebooks needn't. ty reads its settings, the venv and `python/lib`,
from `pyproject.toml`. Unlike ruff, it looks for that file only from the
folder it's run in, hence `--project python` from the root; without it,
ty can't find the lab's packages. ty is young (version 0.0.x), and its
rules change between releases, so keep it pinned, and expect new
findings when upgrading it. `test-all.sh` runs it after pytest, and a
type error fails the run as a failing test does.

pytest finds the lab's packages in `python/lib` by itself, so tests
also run where direnv doesn't, such as inside an editor. ruff knows
`jdclab` is the lab's own package (`known-first-party` in
`pyproject.toml`), so it sorts its imports after third-party ones; add
any new package of the lab's there too.

`ruff format` rewrites files in place; `ruff format --check python`
only reports. ruff and ty may also be installed system-wide, as on a
Mac with Homebrew, but the venv's copies are the ones whose versions are
pinned, and the ones `.envrc` puts first.

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
them. A test that needs something the machine lacks, such as ONNX
Runtime's library or the model the Python demo saves, skips with a note
saying what to do, rather than failing; `go test -v ./...` shows the
notes. To build the programs, install them: `.envrc` points `GOBIN` at
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
perlcritic perl/lib perl/bin
for f in $(git ls-files 'perl/*.pl' 'perl/*.pm' 'perl/*.t'); do perltidy "$f" | diff -u "$f" -; done
```

The tests skip, rather than fail, where Inline::Python isn't installed,
as Perl is optional. Perl::Critic reviews the modules and programs, not
the tests: its module rules (a `package` line, ending in `1;`) don't fit
a test script. The loop prints what Perl::Tidy would change in each
file, and nothing for a file that's already tidy. `.perltidyrc` sends
its output to the screen, for editors, so to tidy a file in place,
without leaving a backup:

```
perltidy -nst -b -bext='/' perl/bin/onnx-demo.pl
```

Its trailing-comma rules can need a second pass to settle: run the loop
again afterwards, and tidy again if it still prints a change.

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

Then run `test-all.sh`: the tests in each language check the model's
scores against the same numbers, so they show at once if the two copies
have drifted apart. Homebrew keeps only the newest version of a
formula, so the pin is what keeps it on the lab's version; going back
once it has moved is awkward. If another formula needs a newer
`onnxruntime`, Homebrew says so, and that's the time to move both.

## Improving the pipeline

[PIPELINE.md](PIPELINE.md#improving-it) has how: the loop of measure,
look, change and measure again, the questions each stage's tables
answer, with commands, and what to change for each finding. In short,
headings' phrases are in `internal/sections/headings.txt`, field labels
in `internal/headings`, and each change keeps a test case and updates
PIPELINE.md's figures.

## Before committing

- **TODO.md:** on `develop`, `git pull --ff-only` and re-read it.
  Sessions working on grammar and scalar add tasks to it, and a commit that completes one
  removes it. Each task opens with **From** (the session that added it,
  the date, and the commit that prompted it, or the last commit and
  "work uncommitted") and **Waiting on** (what it needs first, or
  nothing). A task added to grammar's or scalar's TODO.md, also on
  their `develop`, is committed and pushed there at once, after the same
  pull, so TODO.md only ever changes through commits.
- **Notebooks:** clear their outputs. They carry local paths, and job
  description text once data is in use.
- **Style and tests:** the checks above, for each language touched.
- **Docs:** a change to tools, versions, dependencies or setup updates
  [SETUP.md](SETUP.md) in the same commit, and a change to how the lab
  is worked on updates this file. A change to a stage of the pipeline,
  what it decides, its thresholds or its programs, updates
  [PIPELINE.md](PIPELINE.md) in the same commit, with its figures if
  they moved. A change to what the lab can do, what it has measured or
  what comes next updates the README's
  [Where it stands](../README.md#where-it-stands) too, so that it's never
  behind. Rebuild the PDFs and read the pages that changed.
- **Messages:** commits follow
  [Conventional Commits](https://www.conventionalcommits.org) (`feat:`,
  `fix:`, `docs:`, `chore:`, `build(deps):` and so on), and are signed.
- **CI:** every push to `develop` runs the Go checks (gofmt, vet, the
  tests with the race detector, govulncheck) and gitleaks over the whole
  history. The ONNX demo's tests skip there, with no ONNX Runtime;
  `test-all.sh` runs them, and the Python and Perl tests, here.
- **Secrets:** gitleaks scans the whole history for secrets on every
  push to GitHub. Run it by hand before a tag, and after any history
  rewrite, before pushing:

  ```
  gitleaks git --redact .
  ```

## Releases

A release is a signed, annotated tag on `main`, with a GitHub release to
match. Before v1, a release that changes how the lab is set up or worked
on bumps the minor version (v0.2.0 to v0.3.0), and anything else the
patch (v0.2.0 to v0.2.1). Its notes live in
[CHANGELOG.md](../CHANGELOG.md): add to Unreleased as work lands, then
release from `develop` in these steps. Claude asks before tagging.

0. **Wash up:** review everything since the last tag together, code,
   tests, docs and data files such as `headings.txt`, for
   inconsistency, inefficiency, errors and slop, and fix what's found.
   Save a run first (`go run ./cmd/pipeline run before`): afterwards
   `go run ./cmd/pipeline compare before after` must show no change,
   unless a fix means to make one, and then say how. Reproduce each
   finding before fixing it, with a failing test where code is wrong,
   and note the worst in the release's entry. Check the README's Where
   it stands against the work since the last tag, as it's what
   visitors to `main` read first. A wash-up is also worth
   doing whenever a run of work comes to a natural close.

1. **Check `develop`:** CI is green on its last commit, there's no
   `TODO.md`, the tests pass against the grammar `go.mod` asks for, and
   the history has no secrets.

   ```
   git switch develop
   git pull --ff-only
   gh run list --branch develop --limit 2
   if [ -e TODO.md ]; then cat TODO.md; else echo "no TODO.md"; fi
   grep nukopian/grammar go.mod
   GOWORK=off go test ./...
   test-all.sh
   gitleaks git --redact --no-banner .
   ```

   If `TODO.md` is there, its tasks come first: do them, or move a gap
   in the pipeline word for word into PIPELINE.md's plans, with a row
   in Known gaps linking to it, and delete the file before going on. `go.mod` names a
   grammar tag, such as `v0.2.0`, never a pseudo-version or a `replace`
   pointing at `../grammar`.

2. **Turn Unreleased into the release** in CHANGELOG.md: rename the
   heading to the version and date, as in `## v0.3.0 (2026-10-26)`,
   start the entry with a one-line title (it becomes the tag's title),
   and add an Unreleased above it, saying `Nothing yet.` Commit
   (`docs: v0.3.0's entry in the CHANGELOG`), push, and wait for CI.

3. **Write the tag message from the entry,** and read it. The entry is
   also the GitHub release's notes, and the tag message is the same with
   the version in front. Set the version on the first line:

   ```
   v=v0.3.0
   awk -v h="## $v (" 'index($0, h) == 1 {on = 1; next} /^## / {on = 0} on' CHANGELOG.md | sed '/./,$!d' > "$TMPDIR/notes-$v.txt"
   sed "1s/^/$v: /" "$TMPDIR/notes-$v.txt" > "$TMPDIR/tag-$v.txt"
   cat "$TMPDIR/tag-$v.txt"
   ```

4. **Move `main` and tag it.** First check that every check run on the
   commit has completed successfully: `main`'s protection refuses the
   push until they have, and `gh run watch` can report a run as passed
   while a job is still in progress.

   ```
   gh api "repos/nukopian/jdc-lab/commits/$(git rev-parse develop)/check-runs" --jq '.check_runs[] | "\(.name) \(.status) \(.conclusion)"'
   ```

   Every line ends `completed success`. Then the merge, which is a
   fast-forward, so `main` gets exactly the commits CI tested on
   `develop`, with their signatures. Set the version on the first line:

   ```
   v=v0.3.0
   git switch main
   git pull --ff-only
   git merge --ff-only develop
   git tag -s "$v" -F "$TMPDIR/tag-$v.txt"
   git push origin main "$v"
   gh release create "$v" --verify-tag --title "$(git tag -l --format='%(contents:subject)' "$v")" --notes-file "$TMPDIR/notes-$v.txt"
   git switch develop
   ```

   If the merge says it can't fast-forward, `main` has a commit
   `develop` doesn't. Stop and find out why; don't make a merge commit.
   `--verify-tag` refuses to make a tag of its own if the push didn't
   happen.

5. **Tell the waiting tasks.** Look in grammar's and scalar's `TODO.md`,
   on their `develop`, for tasks waiting on this release, and say which
   can now be done. Nothing else tells their sessions it has landed:

   ```
   for repo in grammar scalar; do
     git -C ../$repo fetch -q
     git -C ../$repo show origin/develop:TODO.md 2>/dev/null | grep -n -B3 '^\*\*Waiting on:\*\*.*jdc-lab' | sed "s|^|$repo: |"
   done
   ```

A tag is never moved or deleted once pushed: a fix is a new release.

## Documents

`scripts/readme-pdf.sh` builds a PDF of each document, beside it, for
reading offline. They're git-ignored, and never published: readers
online have the Markdown, which GitHub shows.

```
readme-pdf.sh
```
