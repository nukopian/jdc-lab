# Boilerplate and view on several processors

**Status:** in progress

**Origin:** roadmap 45

**Started:** 2026-11-03, in aefbb608ab19c465f823e04568aaf65e5eab4836

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

Work in progress.
