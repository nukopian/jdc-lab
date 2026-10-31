# Read facts only where they can be

**Status:** shipped in v0.6.1

**Origin:** roadmap 43

**Started:** 2026-10-31, in 8c9df97c5d9ab378637b3c95a2bb75b2d8193823

**Finished:** 2026-10-31

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

- **Outcome:** done against Done when. Chunking is 4.3 times faster on
  the benchmark, and chunking every ad of the samples 2.8 times faster
  (128 s to 45 s, about 575 ads a second), with every ad's chunks the
  same, byte for byte, and every stage's output unchanged. Both changes
  paid, so both stay, each behind its switch with its test.
- **Changes from the plan:**
  - The benchmarks and their tests are committed, in each package's
    `samples_test.go`, reading the latest run where there is one and
    skipped where there isn't, as in a fresh clone or CI, so the next
    change to facts, or grammar's roadmap 9, has a baseline to measure.
  - The labelled pay ("Pay Rate - 13.10") needs a digit too, so it's
    skipped with pay.
  - The whole-ad test of reading once (`TestReadOnceChangesNothing`)
    passed even with the copy taken out: in the chunker's order no line
    is read again after `placeless` has changed what it read. The copy
    is needed all the same, should that order change, so a test of the
    reader itself (`TestFactReaderCopies`) checks it, and fails without
    it.
- **Measurements:** `go test ./internal/facts ./internal/chunk -run '^$'
  -bench 'Samples$' -benchtime 3x -count 5`, on 2,000 ads of the
  latest run, on 2026-10-31, time per pass (the median of five):

  | | `BenchmarkFindSamples` | `BenchmarkAdOfSamples` |
  |---|---|---|
  | before | 5.54 s, 5.70 GB, 210.4 million allocations | 9.31 s, 9.84 GB, 358.1 million |
  | (a) reading each line once (skipping off) | | 5.35 s, 5.66 GB, 202.8 million |
  | (b) skipping lines | 1.78 s, 1.76 GB, 64.8 million | 3.28 s, 3.28 GB, 115.0 million |
  | both | 1.76 s, 1.76 GB, 64.8 million | 2.16 s, 2.20 GB, 74.6 million |

  So skipping makes `facts.Find` 3.1 times faster, and the two together
  make chunking 4.3 times faster, with a fifth of the allocations. Of
  64,273 lines in those ads, 39,608 (62%) can hold no fact at all.

  Every ad of the samples (25,954) chunked with `cmd/chunk -json`, run
  `claude/before` at 38ac2e6 and `claude/after` with the change: 127.6 s
  before, 45.2 s after, and the two outputs identical, byte for byte.
  `pipeline compare claude/before claude/after`: every stage unchanged.
  The pipeline's stage times didn't move, as only the view chunks ads,
  and only 20 of them; boilerplate and view, at about 58 s each, are now
  its slowest stages.
- **What's left:** the rest of reading facts is the grammars matching,
  which grammar's roadmap 9 is to make cheaper: its task in grammar's
  `TODO.md` now has these benchmarks as its baseline.
