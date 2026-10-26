# TODO

## Install Inline::Python 0.59 unpatched

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 43734d738ba3f92808e9e8d53b0ace6a39433b2e)
**Waiting on:** nothing

**What:** drop the patched install of Inline::Python 0.58 now that 0.59,
released on CPAN on 2026-10-25, carries the fix for
[issue #44](https://github.com/niner/inline-python-pm/issues/44)
(commit b52d63a876c9a1c9233a2e08e068819f70849523 upstream).

**Why:** setup has a download-and-patch step that a plain `cpanm` now
does, and `cpanm --installdeps` can't be run first, because it fetches
the broken 0.58.

**Steps:**

1. `cpanfile` line 14: require `Inline::Python` at `0.59`.
2. `docs/SETUP.md`, the Inline::Python section (around line 296):
   replace the patch steps with
   `INLINE_PYTHON_EXECUTABLE="$PWD/.venv/bin/python3" cpanm Inline::Python`,
   run in the lab's folder with direnv on. Keep the note that it links
   the venv's Python, so it must be reinstalled after changing the lab's
   Python. In The rest, drop "It must come after Inline::Python…" if
   the order no longer matters (`--installdeps` would build against the
   wrong Python without `INLINE_PYTHON_EXECUTABLE`, so check which order
   to recommend).
3. `docs/MAINTAINING.md`, Notes for later (around line 265): remove
   "Inline::Python must stay patched".
4. CHANGELOG, under Unreleased: setup installs Inline::Python 0.59 from
   CPAN, no patch.

**Check it's done:**

- In the lab's perlbrew library, with direnv on:
  `INLINE_PYTHON_EXECUTABLE="$PWD/.venv/bin/python3" cpanm --reinstall Inline::Python`
  installs 0.59 and its tests pass.
- `yath test perl/t` passes, with `onnx-demo.t` run, not skipped, and
  `scripts/test-all.sh` exits 0.
- A fresh clone set up by SETUP.md's steps alone, in a new perlbrew
  library, passes the same.
- `grep -rn '0\.58\|issue #44\|patched' cpanfile docs` finds nothing
  left over.
- The SETUP and MAINTAINING PDFs are rebuilt and the changed pages read.

## Take up grammar v0.2.2

**From:** grammar's session, 2026-10-26 (grammar 3a21142fa9cbe5a5c1c9c7e24c9cdf02345db631, tagged v0.2.2)
**Waiting on:** nothing

**What:** bump the lab's grammar from v0.2.1 to v0.2.2, which brings
scalar from v0.5.7 to v0.5.9 with it (the lab requires scalar only
indirectly):

```
cd ~/Git/github.com/nukopian/jdc-lab
go run ./cmd/pipeline run before
GOWORK=off go get github.com/nukopian/grammar@v0.2.2
GOWORK=off go mod tidy
GOWORK=off go test ./...
go run ./cmd/pipeline run after
go run ./cmd/pipeline compare before after
```

**Why:** Iain asked for jdc-lab to be on the latest grammar
(2026-10-26). v0.2.2 changes no behaviour or API in grammar; it takes up
scalar v0.5.9, whose `S` with a replacement string such as `"$1"` is
about 1.6 times as fast, adds `$+` and `$+{name}`, and gives nothing for
a reference to a group that doesn't exist, as Perl does. job-meta took
it up with its output unchanged.

**Files:** `go.mod`, `go.sum`, and `CHANGELOG.md` under Unreleased only
if the lab's users would notice, which they shouldn't.

**Done when:** the tests pass with `GOWORK=off`, `pipeline compare`
shows no change on either sample or both together, and the bump is
committed on `develop` as `build: grammar v0.2.2, on scalar v0.5.9`,
with this task removed in the same commit.
