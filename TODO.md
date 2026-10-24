# TODO

## Make SETUP's Inline::Python patch command readable

**What:** in [docs/SETUP.md](docs/SETUP.md#inlinepython), the command
that applies the one-line fix to Inline::Python's `Makefile.PL` is a
`perl -pi -e` substitution with nearly every character escaped. Replace
it with one that passes the old and new text as plain strings.

**Why:** it's hard to read, to check, and to change. Anyone following
SETUP has to trust it rather than see what it does.

**What's known:** this does the same, tested on 2026-10-24 against a
fresh unpack of Inline-Python-0.58: the file it produces is identical to
the current command's. In single quotes the shell leaves both strings
alone, and Perl's `\Q...\E` matches the old text literally, so nothing
needs escaping:

```
old='(?:\.\Q$Config{dlext}\E|' new='(?:\.\Q$Config{so}\E|\.\Q$Config{dlext}\E|' perl -pi -e 's/\Q$ENV{old}\E/$ENV{new}/' Makefile.PL
```

The current command appears nowhere else in the repository.

**Steps:**

1. Replace the `perl -pi -e` line in SETUP's Inline::Python block with
   the one above, perhaps split over three lines for reading.
2. Run the whole block from a fresh unpack, as SETUP gives it, under
   zsh: `grep -c 'Config{so}' Makefile.PL` must print `1`, and
   `cpanm --test-only .` must pass. Use `--test-only` so the lab's
   installed Inline::Python isn't replaced.
3. Rebuild SETUP's PDF and read the page.

**Done when** SETUP shows the readable command, and it's been run as
written. Then remove this task, and this file with it.
