# Change records

How each planned change to the lab went: the plan as it was written,
then what happened. The [CHANGELOG](../../CHANGELOG.md) says what
changed, and links here; a record keeps the thinking behind it, for
anyone looking back, including whoever plans the next change. In the
lab, a change is often an experiment as much as a feature, and a
record suits both: what was measured, and what it showed, matter as
much as what shipped.

A change gets a record when it has a plan: a roadmap item that reaches
Now, or a task in `TODO.md` with steps. A planned change that's dropped
keeps its record, saying why, as dead ends are the history most worth
having. Small fixes don't need one.

## How a record is kept

- **It's opened when work starts.** The plan moves here from the
  roadmap or `TODO.md`, and the roadmap's Now item shrinks to a
  paragraph and a link. `TODO.md` keeps only the next steps for a
  session.
- **The plan isn't edited once work starts.** What changed goes under
  What happened, so the record shows how the plan held up. A plan from
  the roadmap reads as the [roadmap](../../ROADMAP.md)'s opening says:
  its figures are from the first sample unless it says otherwise, and
  `RUN` stands for a run's name.
- **What happened is added as the work goes,** and finished when the
  change ships or is dropped, in the same commit. When it ships, the
  status says "shipped, unreleased", and the release names its version.
- **It's named for the day work started, "r" and its roadmap number,
  and the change,** such as `2026-10-25-r2-stop-shift-times-passing-as-headings.md`
  for roadmap 2: dates never collide between sessions and sort in order,
  and a search for `r2` finds the record, where a bare number would find
  every date and figure with a 2 in it. A record for a `TODO.md` task
  has no number, so it's named for the day and the change alone. The
  record gives its origin in its header too.
- **It's public,** on `main` and in every release, like the rest of
  `docs/`, so nothing personal goes in it, and nothing from job ads
  beyond the figures and short phrases, such as headings, that
  [PIPELINE.md](../PIPELINE.md) gives.

## The template

```markdown
# Title

**Status:** in progress | shipped, unreleased | shipped in v0.3.4 | dropped

**Origin:** roadmap 7 | a TODO task from grammar's session (grammar <commit>)

**Started:** 2026-10-26, in <commit>

**Finished:** 2026-10-27

## The plan

The plan as it was when work started, word for word: Serves, What,
Why, What's known, Steps, Done when.

## What happened

- **Outcome:** against Done when.
- **Changes from the plan,** and why.
- **Measurements,** and how they were taken: the runs compared, and
  what moved.
- **Tried and reverted,** with the numbers.
- **What's left,** linked to its gap or roadmap item.
```

## Records

Newest first.

- [Keep the layout an ad's text implies](2026-10-30-r24-keep-the-layout-an-ads-text-implies.md)
  (roadmap 24), in progress
- [Split walls of text into sentences](2026-10-30-r5-split-walls-of-text-into-sentences.md)
  (roadmap 5), in progress
- [Make boilerplate's verdict right](2026-10-28-r4-make-boilerplates-verdict-right.md)
  (roadmap 4), shipped in v0.5.2
- [Keep a list item's own blocks in the ad's order](2026-10-28-r17-a-list-items-own-blocks-in-order.md)
  (roadmap 17), shipped in v0.5.2
- [One set of grammars for facts and headings](2026-10-28-r22-one-set-of-grammars-for-facts-and-headings.md)
  (roadmap 22), shipped in v0.5.2
- [Map what's left](2026-10-28-r21-map-whats-left.md)
  (roadmap 21), shipped in v0.5.2
- [Check the paths read from grammar matches with CheckPath](2026-10-27-check-match-paths-with-checkpath.md),
  shipped in v0.5.1
- [A chunker for any ad](2026-10-27-r18-a-chunker-for-any-ad.md)
  (roadmap 18), shipped in v0.5.0
- [Facts from the details, by grammar](2026-10-27-r7-facts-from-the-details.md)
  (roadmap 7), shipped in v0.5.0
- [Does it make search better? A first test](2026-10-27-r10-a-first-retrieval-test.md)
  (roadmap 10), shipped in v0.5.0
- [Find near-duplicate boilerplate](2026-10-27-r3-find-near-duplicate-boilerplate.md)
  (roadmap 3), shipped in v0.5.0
- [A readable view of chunked ads](2026-10-27-r16-a-readable-view-of-chunked-ads.md)
  (roadmap 16), shipped in v0.5.0
- [Try grammar's Overlap in sections.All, then take up grammar v0.3.2](2026-10-26-try-overlap-and-take-up-grammar-v0.3.2.md),
  shipped in v0.3.5
- [Take up grammar v0.3.1, and read headings.txt with MapOnly](2026-10-26-take-up-grammar-v0.3.1-and-read-headings-with-maponly.md),
  shipped in v0.3.5
- [Take up grammar v0.3.0 and scalar v0.6.0](2026-10-26-take-up-grammar-v0.3.0-and-scalar-v0.6.0.md),
  shipped in v0.3.5
- [Check the rules on a second sample](2026-10-25-r1-check-the-rules-on-a-second-sample.md)
  (roadmap 1), shipped in v0.3.5
- [Stop shift times passing as headings](2026-10-25-r2-stop-shift-times-passing-as-headings.md)
  (roadmap 2), shipped in v0.4.0
