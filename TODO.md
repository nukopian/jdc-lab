# TODO

## Report the Inline::Python macOS link bug upstream

**What:** report a build bug in
[Inline::Python](https://metacpan.org/dist/Inline-Python) to its
maintainer, as a GitHub issue on
[niner/inline-python-pm](https://github.com/niner/inline-python-pm), with
the cause, a one-line fix, and what was tested. Report it as an issue,
not a pull request: the maintainer decides whether and how to fix it.

**Why:** on macOS, Inline::Python 0.58 (the latest release, and
`master` at `2b2e476`) builds without linking libpython when Python
comes from pyenv or anywhere other than Apple's system Python. Every test
then fails, so `cpanm Inline::Python` fails. The lab's Perl side needs
Inline::Python, and it's installed here from a locally patched copy until
a fixed release is out.

### What's known

Checked on 2026-10-24 on macOS 26.7.1 (arm64), Apple clang 21.0.0,
perlbrew Perl 5.36.3 (`darwin-thread-multi-2level`), with pyenv's
Python 3.12.12 and 3.14.8, both built with `--enable-shared`.

- **Symptom:** `Makefile.PL` prints
  `Unrecognized argument in LIBS ignored: 'libpython3.12.dylib'`. The
  bundle links without libpython, which `-undefined dynamic_lookup`
  hides until it is loaded. Then 37 of 38 test files fail with
  `symbol not found in flat namespace '_PyBool_Type'`. `otool -L` on
  `Python.bundle` lists no libpython.
- **Cause:** `Makefile.PL` line 90 turns the library's file name into a
  `-l` flag, but only if it ends in Perl's `$Config{dlext}` or
  `$Config{_a}`:

  ```perl
  $sel->{libpython} =~ s/lib(.*)(?:\.\Q$Config{dlext}\E|\Q$Config{_a}\E)/-l$1/;
  ```

  On macOS `dlext` is `bundle` (the extension of Perl's own XS modules),
  not `dylib`, so `libpython3.12.dylib` is passed through unchanged, and
  MakeMaker drops it. Perl's `$Config{so}` is the shared library
  extension: `dylib` on macOS, `so` on Linux.
- **When it started:** line 90 dates from 2008. It broke when
  `special_non_windows_check_shared_static_libs` started setting
  `libpython` to `LDLIBRARY` (the `.dylib` on macOS), in PR #35
  (commit `0a917ea`, merged December 2021). By date, that's first in
  0.57. On Linux `LDLIBRARY` ends in `.so`, which is also `dlext`, so
  Linux was unaffected.
- **Fix:** accept `$Config{so}` as well, keeping the other two, so no
  other platform changes:

  ```diff
  -$sel->{libpython} =~ s/lib(.*)(?:\.\Q$Config{dlext}\E|\Q$Config{_a}\E)/-l$1/;
  +$sel->{libpython} =~ s/lib(.*)(?:\.\Q$Config{so}\E|\.\Q$Config{dlext}\E|\Q$Config{_a}\E)/-l$1/;
  ```

- **Tested:** on `master` at `2b2e476` with the fix, `make test` passes
  (38 files, 837 tests) with both Pythons, and `Python Lib` shows
  `-lpython3.12` and `-lpython3.14`. Without the fix, both fail as
  above. Not tested: Linux, Windows, or Apple's system Python, which
  takes another path (`add_rpath_for_macos_system_python`) and may
  never reach the `.dylib` case.
- **Already reported?** Searches of the repository's issues and pull
  requests on 2026-10-24 for `dlext`, `dylib`, `PyBool_Type`, `darwin`,
  `macOS`, `symbol not found` and `libpython` found nothing about this.

### Steps

Emails with the draft to the current and original maintainers both
bounced on 2026-10-24, so the issue is the way to reach them, and
there's no reason to wait before filing it. It's also where others
hitting the bug will find it, so the draft ends with a workaround.

1. Search the issues again, open and closed, in case someone has
   reported it since.
2. Reproduce it yourself from a fresh clone of upstream, without and
   then with the fix. Keep the output for the report:

   ```
   git clone https://github.com/niner/inline-python-pm.git
   cd inline-python-pm
   INLINE_PYTHON_EXECUTABLE=~/.pyenv/versions/3.12.12/bin/python3 perl Makefile.PL
   make && make test
   ```

   Apply the one-line change, run `git clean -fdx -e Makefile.PL`, and
   repeat.
3. Write the issue in your own words, from the draft below, and check
   every claim against your own run.
4. File it at
   [github.com/niner/inline-python-pm/issues/new](https://github.com/niner/inline-python-pm/issues/new).
5. Follow up on any questions. If the maintainer asks for a pull
   request, make one then.

**Done when** the issue is filed. When a fixed release reaches CPAN,
reinstall Inline::Python with plain `cpanm Inline::Python` (see below),
and remove this task.

### Until it's fixed

[HOWTO.md](HOWTO.md#inlinepython) installs Inline::Python 0.58 with
the fix applied. When a fixed release is out, replace those steps there
with plain `cpanm Inline::Python`, run in the lab's folder so that it
installs into the lab's library.

### Draft issue

> **Title:** macOS: libpython isn't linked with a non-system Python
> (`$Config{dlext}` is `bundle`)
>
> On macOS with a pyenv-installed Python (3.12.12 and 3.14.8, both
> `--enable-shared`), 0.58 and current `master` build `Python.bundle`
> without linking libpython. `perl Makefile.PL` prints
> `Unrecognized argument in LIBS ignored: 'libpython3.12.dylib'`, the
> link succeeds because of `-undefined dynamic_lookup`, and then 37 of
> the 38 test files fail with
> `symbol not found in flat namespace '_PyBool_Type'`.
>
> The cause seems to be line 90 of `Makefile.PL`, which only turns the
> library name into a `-l` flag if it ends in `$Config{dlext}` or
> `$Config{_a}`. On macOS `dlext` is `bundle`, while the shared library
> extension (`$Config{so}`) is `dylib`. Since #35,
> `special_non_windows_check_shared_static_libs` sets `libpython` to
> `LDLIBRARY`, which is `libpython3.12.dylib` here, so it no longer
> matches. On Linux `LDLIBRARY` ends in `.so`, which is also `dlext`,
> so Linux isn't affected.
>
> Accepting `$Config{so}` as well fixes it for me:
>
> ```diff
> -$sel->{libpython} =~ s/lib(.*)(?:\.\Q$Config{dlext}\E|\Q$Config{_a}\E)/-l$1/;
> +$sel->{libpython} =~ s/lib(.*)(?:\.\Q$Config{so}\E|\.\Q$Config{dlext}\E|\Q$Config{_a}\E)/-l$1/;
> ```
>
> With it, `make test` passes (38 files, 837 tests) on both Pythons,
> and `Python Lib` shows `-lpython3.12` / `-lpython3.14`. I haven't
> tried Linux, Windows or Apple's system Python. The change keeps the
> existing two alternatives, so they should behave as before.
>
> Environment: macOS 26.7.1 arm64, Apple clang 21.0.0, Perl 5.36.3
> (perlbrew, `darwin-thread-multi-2level`), Inline::Python 0.58 and
> `master` at 2b2e476.
>
> **Workaround,** for anyone hitting this before a fix is released.
> Your Python must have been built with `--enable-shared` (pyenv:
> `PYTHON_CONFIGURE_OPTS=--enable-shared pyenv install …`). With
> `python3` on your `PATH` being the Python you want to embed:
>
> ```
> curl -sSL https://cpan.metacpan.org/authors/id/N/NI/NINE/Inline-Python-0.58.tar.gz | tar -xz
> cd Inline-Python-0.58
> perl -pi -e 's/\(\?:\\\.\\Q\$Config\{dlext\}\\E\|/(?:\\.\\Q\$Config{so}\\E|\\.\\Q\$Config{dlext}\\E|/' Makefile.PL
> INLINE_PYTHON_EXECUTABLE="$(python3 -c 'import sys; print(sys.executable)')" cpanm .
> ```
>
> The `perl -pi` line applies the change above. `INLINE_PYTHON_EXECUTABLE`
> picks the Python without `Makefile.PL` asking which to use, and
> works with pyenv's shims and with venvs.
