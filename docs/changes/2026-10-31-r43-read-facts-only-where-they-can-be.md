# Read facts only where they can be

**Status:** in progress

**Origin:** roadmap 43

**Started:** 2026-10-31, in 8c9df97c5d9ab378637b3c95a2bb75b2d8193823

## The plan

**Serves:** deployment

**Size:** S (2)

**What:** make reading facts cheaper without changing a single one:
(a) read each line's facts once per ad, where the chunker now reads some
lines two or three times; (b) skip a grammar for a line that can't hold
what it matches, as a line with no digit can hold no pay.

**Why:** chunking runs at about 200 ads a second, where everything before
it runs at over 1,500, and five times as many ads may come. Profiled,
79% of chunking is `chunk.AdOf`, nearly all of it `facts.Find`, which
runs all five fact grammars over almost every line of an ad: every line
of the details, of every other section but the employer's own (so the
details list every fact wherever it's stated), and of an unmapped
section while its kind is guessed. Most lines are prose about the job,
which no fact grammar can match. Reading facts only in the sections
likely to hold them would be cheaper still, but would lose pay stated in
the benefits, or a pattern in the role.

**What's known:** a line can hold a fact only if it has:

| Grammar | Needs |
|---|---|
| Pay, Hours, Postcode, and a labelled pay | a digit |
| Schedule | a digit, a day's name or its first three letters, or "M - F" |
| Pattern | the first word of one of its word list's phrases, in any case |

Grammar's own matcher is to get faster separately (grammar's roadmap 9,
in its `TODO.md`), and waits on this.

**Steps:**

1. On a branch from a clean `develop`, add a benchmark of chunking and
   of `facts.Find` over the samples' ads, skipped where `data/` isn't,
   and record the baseline; save a run, `claude/before`, and every ad's
   chunks as JSON.
2. (b) In `internal/grammars`, beside each grammar, a check of whether a
   line could hold a match, the work pattern's stems taken from its
   word list; `facts.Find` skips a grammar whose check fails, behind a
   switch, `skipLines`. A test that, over every line the tests know,
   each grammar finds nothing the check would have skipped.
3. (a) In `chunk`, each line's facts read once per ad and handed to each
   caller as a copy, since `placeless` changes what it's given.
4. Measure each against the baseline; keep what pays, and record what
   doesn't, with its numbers.
5. Check every ad's chunks are the same, byte for byte, and `pipeline
   compare` shows nothing moved.

**Done when** chunking is measurably faster, with the figures here and
in PIPELINE.md; every ad's chunks and every stage's output are
unchanged; the switch's test passes; and the lab's facts benchmark is
ready as the baseline for grammar's roadmap 9.

## What happened
