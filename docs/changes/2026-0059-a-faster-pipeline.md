# 2026-0059. A faster pipeline

**Status:** shipped in v0.7.0

**Origin:** a session looking for what grammar 2026-0025 (a faster
matcher) and scalar's values in one allocation teach about the lab's own
Go code

**Started:** 2026-11-03, in 3baebf12fcae06bacd5e2e8a39a1624f196e2b48

**Finished:** 2026-11-03

## The plan

Look through the lab's Go code for what makes the pipeline slow, by the
practice's rule for speculative speed-ups: branch from a clean tree,
save a run and baseline benchmarks, profile first, put each change
behind a switch with a differential test and a planted bug to show the
test would notice, keep what pays, revert what doesn't, and record what
was reverted and its numbers. Every stage's output must stay the same,
byte for byte, as `pipeline compare` shows.

The baseline run, `claude/perf-before`, took 128 s, and boilerplate (45
s, 1.4 GB at its peak) and view (50 s, 1.4 GB) were three quarters of
it. Their profiles have the same shape, as view reads boilerplate's
verdicts: of boilerplate's 55 s on the processors, finding long texts
alike (`findAlike`) was 24%, most of it comparing their runs of words as
maps; `headings.Key` 21%, a map lookup for every character, ASCII
included; and the garbage collector and the memory it asks for about
20%. grammar worked out the starts of `Schedule` afresh at every search
(3% of the time, and 1.2 GB of garbage), as it keeps one rule's starts
and the lab searches it for six rules in turn.

The candidates, by their share of the profile:

1. `findAlike`: runs of words as sorted slices, compared by merging, not
   maps; and the groups worked out once, not again after `Ends`.
2. `headings.Key`: ASCII from a table, and spaces collapsed as it goes.
3. Work done again for every question: `template` keying an ad's title
   and places each time, word counts through `strings.Fields`.
4. grammar keeping each rule's starts: grammar's change to make, in its
   working copy, with a task in its `TODO.md`.
5. The garbage collector's settings for the batch programs, last, once
   the rest has changed what they hold.

New benchmarks to measure them: `headings.BenchmarkKeySamples` and
`BenchmarkReadSamples` (2,000 ads of the latest run), and
`boilerplate.BenchmarkSamples` (5,000 ads, read and asked what
cmd/boilerplate asks).

## What happened

- **Outcome:** a run of the pipeline takes 67 s rather than 128 s
  (`claude/perf-before` against `claude/perf-lab`, built with
  `GOWORK=off`, so against grammar v0.5.2), and `pipeline compare`
  finds every stage's output unchanged:

  | stage | before | after | peak memory |
  |---|---:|---:|---:|
  | blocks | 5.56 s | 3.42 s | 14 → 15 MB |
  | heading-counts | 9.76 s | 8.73 s | 199 → 197 MB |
  | sections | 18.30 s | 12.75 s | 205 → 206 MB |
  | boilerplate | 44.53 s | 20.16 s | 1400 → 1050 MB |
  | view | 49.83 s | 21.96 s | 1429 → 1144 MB |

  With grammar's fix below in its working copy, through `go.work`, a
  run takes 56.7 s (heading-counts 5.62 s, sections 9.90 s, boilerplate
  17.71 s, view 19.58 s), its output unchanged too. Single runs, on a
  machine doing other work, so each stage's figure moves by a second
  or so from run to run.
- **What was kept,** in the order it was done, each measured with
  `benchstat`, interleaved, 6 to 8 runs each, on an Apple M2, against
  the step before:
  - `headings.Key` reads ASCII from a table, `asciiKey`, worked out
    from the old code at start-up so the two can't disagree, and
    collapses spaces as it goes rather than through `strings.Fields`
    and `Join`. Behind `fastKey`. `BenchmarkKeySamples` 129 ms → 19.3
    ms (−85%), 84% fewer allocations; `BenchmarkReadSamples` −17%;
    `boilerplate.BenchmarkSamples` 7.20 s → 5.19 s (−28%).
  - `findAlike` keeps each long text's runs of five words as a sorted
    slice, hashed as before (FNV-1a over the words and spaces, without
    joining them), and compares two by merging, after checking that the
    shorter could share enough at all. Behind `fastAlike`. `Samples`
    5.19 s → 4.35 s (−16%).
  - `judged` asks whether a text is phrased as a job line (`JobLine`,
    a grammar search) only when a key is repeated enough for it to
    matter, and once. Most texts aren't. Equivalent by construction, as
    both checks are pure, so no switch. `Samples` −12%.
  - Texts alike are grouped once, while the long texts stay the same,
    rather than again after `Ends` counts the ending runs, which change
    only what's added up for each group; and `SourceOf` keys an ad's
    title and places once, rather than `template` keying them at every
    question. `Samples` −9.5%.
  - `clean.Text` reads ASCII from a table, `asciiText`, and collapses
    spaces as it goes, as `Key` does; a fold can hold spaces (an em
    dash is " -- "), so a small builder owes a space rather than writing
    it. Behind `fastText`. `blocks.BenchmarkSplitSamples` 374 ms → 148
    ms (−60%, but with a spread of ±22–25% on a busy machine), 37% fewer
    allocations.
  - `headings.KeyWords` counts a key's words by its spaces, which `Key`
    makes single, rather than through `strings.Fields`, in
    `boilerplate.Add` and `Ends`, `sections.beforeList` and
    `headings.headingShaped`. With `clean.Text`'s change, `Samples`
    −19%, 12% fewer allocations.

  In all, `boilerplate.BenchmarkSamples` went from 7.20 s, 2.28 GiB and
  30.1 million allocations to 2.46 s, 1.0 GiB and 9.4 million (the last
  with grammar's fix as well).
- **Tests:** `TestFastKeyChangesNothing` and `TestFastTextChangesNothing`
  compare each fast path with the old code on hand-picked edge cases,
  every pair of ASCII characters, and the titles, blocks and raw HTML of
  2,000 ads of the latest run; `FuzzKey` and `FuzzText` ran 60 s each
  clean, about 4.8 million inputs for `Key`. `TestFastAlikeChangesNothing`
  reads 3,000 ads both ways and compares every answer cmd/boilerplate
  asks of 73,067 texts, and the groups of 13,440 long texts (430
  joined to another), named by their first key, as the names a group
  is given depend on the order texts are joined. `TestKeyWords` checks
  the word count against `strings.Fields`. Each fast path was shown to
  catch a planted bug: a space dropped before a fold in `Key` (150
  texts failed), an em dash's leading space dropped in `Text` (55), and
  a merge advancing the wrong side in `alikeSorted`, which changed
  1,114 answers and the groups.
- **Found in grammar:** grammar keeps a grammar's starts for one rule,
  the last searched for, and `headings.hasTimes` searches `Schedule` for
  six rules in turn, so grammar worked the starts out afresh at almost
  every search: 38% of `BenchmarkReadSamples` once `Key` was fast, and
  1.2 GB of the boilerplate stage's garbage. Keeping each rule's starts,
  in grammar's working copy, made `BenchmarkReadSamples` 53% faster
  (412 ms → 195 ms, 68% fewer allocations) and `Samples` 15% faster;
  `facts.BenchmarkFindSamples` didn't move, as each fact grammar is
  searched for one rule. That's grammar's change to make, so it's left
  uncommitted there, with a task in grammar's `TODO.md` (grammar
  33ba54f295932260891feb4c1658264823e90c86), whose release is to leave
  the lab a task to take it up. grammar released it in v0.5.3, and the
  lab took it up the same day: run back to back with `GOWORK=off`, a
  run took 51.7 s with v0.5.2 and 41.8 s with v0.5.3 (heading-counts
  6.90 s → 4.25 s, sections 9.79 s → 8.14 s, boilerplate 15.49 s →
  12.71 s, view 16.67 s → 13.89 s), its output unchanged, on a quieter
  machine than the runs above.
- **Chunking every ad** (`xargs chunk -json`, 25,954 ads, `GOWORK=off`)
  gives the same chunks, byte for byte. It was faster after in each of
  12 interleaved pairs, by 4% to 39%, but the machine was busy
  throughout (load average 4 to 6), so no single figure was given then.
  Timed again on a quieter machine (load average about 2.5), on all
  28,000 ads, interleaved three times each, v0.6.6 took 25.6 s to
  26.1 s and `develop`, with grammar v0.5.3, 17.0 s, the same chunks,
  byte for byte: a third less.
- **At the release,** on a quiet machine, v0.6.6 took 113.6 s for a run
  (boilerplate 41.3 s, view 42.4 s) and v0.7.0 40.6 s with its stages
  one at a time (12.7 s and 13.8 s), every stage's output the same.
  128 s, above, was on a busier machine.
- **Measuring:** the profiles showed time in `madvise` and in the read
  system call under `json.Decoder` that reading alone didn't (0.5 s for
  all 28,000 ads). It's the cost of the heap's fresh memory, landing
  wherever it's first touched: with the collector off, `madvise` was
  75% of the samples and the time the same; with a 1 MB `bufio.Reader`
  under the decoder, the read time moved into `madvise`, the time the
  same (2.97 s against 3.01 s). `GOGC` at 100, 200 and 400 gave the same
  times, so plan step 5, the collector's settings, was dropped: fewer
  allocations is what reduces it.
- **Tried and reverted:** `collapse`, in blocks, writing the text
  between HTML's spaces whole, for valid UTF-8, rather than a character
  at a time: 3% fewer allocations, but 5% slower, measured in one
  binary with its switch on and off (123.5 ms against 117.2 ms), as
  checking UTF-8 and searching for spaces costs more than it saves on
  the short texts the tokenizer gives. Separate binaries had shown it
  3.5% slower too. Its differential test showed why one on `Split`
  alone isn't enough: a planted bug, doubled spaces, changed no block,
  as cleaning collapses spaces again, and was caught only by comparing
  `collapse` itself.
- **Not tried:** running the stages at once in `pipeline run`. By hand,
  the five stage programs run together took 23 s against 57 s one after
  another, but each stage's time rose 10% to 40% from sharing the
  machine, and `timings.tsv` is for comparing stages from run to run,
  so it's a decision about what the timings mean, left open.
- **What's left:** at full size, boilerplate's remaining time is spread
  thin: `findAlike` 14% (the merge itself 6%), `Key` 8%, `template` 5%,
  `groups` comparing titles 4%, and the heap's fresh memory about 15%.
  Each is a few percent at most, with less certain fixes. View and
  boilerplate each build the same index of every ad, by design, as no
  program reads another's output.
