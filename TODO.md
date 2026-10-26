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
