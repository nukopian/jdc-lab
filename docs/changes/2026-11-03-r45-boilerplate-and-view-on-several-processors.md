# Boilerplate and view on several processors

**Status:** shipped in v0.7.1

**Origin:** roadmap 45

**Started:** 2026-11-03, in aefbb608ab19c465f823e04568aaf65e5eab4836

**Finished:** 2026-11-03

## The plan

**Serves:** deployment

**What:** split the work boilerplate and view do over every ad across
the machine's processors, rather than one, keeping their verdicts the
same.

**Why:** run at once, the stages take as long as view, about 15 s, and
boilerplate and view each use about one processor, building one index
of every ad. A run is the lab's measure of every change, so each second
off it pays at every step; and work that splits across processors is
what running the pipeline at scale would need. The order texts are
added in decides how alike texts are grouped and named, so the split
must give the same groups, by the same names, whatever the order.

**What's known:** both programs call `boilerplate.Read`, then ask of
every text of every ad whether it's boilerplate. Timed by the clock on
all 25,954 distinct ads of the latest run, `Read` takes 9.6 s: decoding
the JSON 0.4 s, dividing each ad into sections 5.3 s, adding its texts
to the index 1.0 s (walls included), and `Ends`, which finds the runs
of repeated text that end each ad, 2.8 s, of which grouping the long
texts alike (`joinAlike`, done once, when first asked) is about 1.7 s.
Asking every text after takes 2.8 s more. The CPU profile put 3.6 s in
the read system call under the JSON decoder; the clock says decoding is
0.4 s, so on macOS the profile's samples land in the wrong place, and
the clock is to be trusted over it.

Dividing an ad is pure: grammar says a finished grammar may parse from
several goroutines, and the lab's own tables are filled at start-up and
only read after. Adding to the index is not, and its order decides the
groups' names, so it stays one ad at a time, in order. Whether two long
texts are alike is pure too; only joining them into groups depends on
order.

**Steps:**

1. Divide the ads into sections on every processor, keeping the order
   they're added to the index in, behind a switch, `parallelRead`. A
   test reads the ads both ways and compares every ad's sections, every
   answer cmd/boilerplate asks, and the groups of texts alike by their
   own names; a planted bug shows it would notice.
2. Find which long texts are alike on every processor, then join them in
   the order one processor would, so the groups and their names stay
   the same, behind `parallelAlike`, with the same test.
3. Time `Read` again by the clock, and if what's left is worth it, the
   questions asked after `Read` and the rest of `Ends`, whose answers
   depend on caches that are filled as they're asked, on every
   processor: only where the caches can be filled first, or safely.
4. Measure each step with `benchstat`, interleaved, and the whole with
   `pipeline run` both ways, against the run before: `pipeline compare`
   must show no change.

**Done when** both stages are profiled first, as on 2026-11-03
([record](2026-11-03-a-faster-pipeline.md)); the split is behind a
switch, with a differential test that catches a planted bug;
`pipeline compare` shows every stage's output the same; a run takes
under 10 s, or the record says what stood in the way; and PIPELINE.md
gives the new timings.

## What happened

- **Outcome:** boilerplate and view divide their ads, and compare long
  texts, on every processor, behind `parallelRead` and `parallelAlike`,
  and `pipeline compare` finds every stage's output the same, against
  `claude/before`, in each run below. Run alone (`pipeline run
  -serial`), boilerplate went from 12.7 s to 7.1 s and view from 13.8 s
  to 8.3 s, and a run from 40.6 s to 29.4 s. With the stages at once, a
  run went from 15.5 s to 12.1 s (two runs, both 12.1 s): not under
  10 s, for the reason under What's left.
- **Measurements,** on an Apple M2, four performance cores and four
  efficiency ones, `benchstat`, interleaved, 6 runs each, on 5,000 ads,
  each switch on and off in one binary: `Read` 1.52 s → 0.78 s (−49%)
  with `parallelRead`, the same allocations; `Read` and grouping alike
  0.78 s → 0.68 s (−13%) with `parallelAlike`. `BenchmarkSamples`,
  `develop`'s binary against this one's, turn about: 2.02 s → 1.13 s
  (−44%), the same memory. On all 25,954 distinct ads, by the clock,
  `Read` went from 9.6 s to 4.2 s: decoding 0.45 s, dividing 1.3 s (from
  5.3 s), adding to the index 0.8 s, grouping alike 1.0 s (from about
  1.7 s), and `Ends` 0.7 s.
- **Tests:** `TestParallelReadChangesNothing` reads 3,000 ads with both
  switches on and both off, and compares every ad's sections, all
  73,067 answers cmd/boilerplate asks, and every long text's group by
  its own name, not by its first key as `TestFastAlikeChangesNothing`
  does, as adding and joining in order should keep the names too. Both
  pass under the race detector. Planted bugs: an off-by-one leaving the
  first ad undivided (one ad, 19 answers and 4 groups differed); and
  pairs joined the other way round, `join(j, i)`, which changed groups'
  names and no answer, caught only by comparing names. A third, the
  joins for one text in reverse order, changed nothing, and rightly:
  each joins its groups under that text's root, whatever their order,
  so only the order between texts matters.
- **Changes from the plan:** step 3, the questions asked after `Read`
  and the rest of `Ends` on every processor, wasn't done. Asking takes
  2.8 s in each program, and would save about 2 s of each alone, but
  with the stages at once the run is bound by the processors' time, not
  by any stage's: the five use about 62 s of it in 12.1 s, five cores'
  worth on four fast and four slower ones. Spreading work wider doesn't
  shorten that, and spreading the questions would mean filling the
  index's caches first, or guarding them, in both programs. The worker
  pool is one function, `inParallel`, used by both steps.
- **What's left:** a run under 10 s needs less work, not more
  processors. Four of the five stages divide every ad again, and
  boilerplate and view each build the same index: roadmap idea 46,
  divide each ad once for every stage. The profile on macOS put 3.6 s
  in the read system call that the clock showed was 0.45 s of decoding:
  where they disagree, the clock is right.
