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

- **It's opened when its item first has a question to settle, or when
  work starts.** When work starts, the plan moves here from the
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
- **It's named for its id, then the change,** such as
  `2026-0002-stop-shift-times-passing-as-headings.md` for roadmap
  2026-0002. A roadmap item's id is its own, kept as its name in commits
  and the CHANGELOG; a `TODO.md` task with steps takes the next id from
  the roadmap's opening, in the same sequence, though it never appears
  there. Ids sort in the order items were added; the **Started:** line
  gives the date. Records opened before 2026-11-09 were named for the
  day work started and "r" and the item's number, and were renamed to
  their ids, a task's taking the next id in the order they were opened.
- **Its questions are settled before work starts,** numbered D1, D2
  and on under Decisions: unchecked while open, checked once settled,
  with the answer, the date and why, and summed up in a **Decisions:**
  line under the status.
- **It's public,** on `main` and in every release, like the rest of
  `docs/`, so nothing personal goes in it, and nothing from job ads
  beyond the figures and short phrases, such as headings, that
  [PIPELINE.md](../PIPELINE.md) gives.

## The template

```markdown
# 2026-0007. Title

**Status:** open | in progress | shipped, unreleased | shipped in v0.3.4 | dropped

**Decisions:** D1 settled: the answer, in a few words. D2 open: the question, in a few words. | None.

**Origin:** roadmap 2026-0007 | a TODO task from grammar's session (grammar <commit>)

**Depends on:** 2026-0003, if anything

**Started:** 2026-10-26, in <commit>

**Finished:** 2026-10-27

## Decisions

- [x] **D1.** The question? The answer, on 2026-10-26, and why.
- [ ] **D2.** A question still open? The options, and the
  recommendation.

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

- [Boilerplate's verdict checked](2026-0033-boilerplates-verdict-checked.md)
  (roadmap 2026-0033), shipped in v0.8.0
- [Metadata as filters in the retrieval test](2026-0034-metadata-as-filters-in-the-retrieval-test.md)
  (roadmap 2026-0034), shipped in v0.7.1
- [Boilerplate thresholds scaled to the ads' number](2026-0012-boilerplate-thresholds-scaled-to-the-ads-number.md)
  (roadmap 2026-0012), dropped
- [Boilerplate and view on several processors](2026-0045-boilerplate-and-view-on-several-processors.md)
  (roadmap 2026-0045), shipped in v0.7.1
- [A faster pipeline](2026-0059-a-faster-pipeline.md), shipped
  in v0.7.0
- [Checks in the fact grammars, with code assertions](2026-0058-checks-in-the-fact-grammars.md)
  (two TODO tasks), shipped in v0.6.4
- [Read facts only where they can be](2026-0043-read-facts-only-where-they-can-be.md)
  (roadmap 2026-0043), shipped in v0.6.1
- [Italic as a cue, like bold](2026-0026-italic-as-a-cue-like-bold.md)
  (roadmap 2026-0026), shipped in v0.6.1
- [Keep the layout an ad's text implies](2026-0024-keep-the-layout-an-ads-text-implies.md)
  (roadmap 2026-0024), shipped in v0.6.0
- [Split walls of text into sentences](2026-0005-split-walls-of-text-into-sentences.md)
  (roadmap 2026-0005), shipped in v0.6.0
- [Make boilerplate's verdict right](2026-0004-make-boilerplates-verdict-right.md)
  (roadmap 2026-0004), shipped in v0.5.2
- [Keep a list item's own blocks in the ad's order](2026-0017-a-list-items-own-blocks-in-order.md)
  (roadmap 2026-0017), shipped in v0.5.2
- [One set of grammars for facts and headings](2026-0022-one-set-of-grammars-for-facts-and-headings.md)
  (roadmap 2026-0022), shipped in v0.5.2
- [Map what's left](2026-0021-map-whats-left.md)
  (roadmap 2026-0021), shipped in v0.5.2
- [Check the paths read from grammar matches with CheckPath](2026-0057-check-match-paths-with-checkpath.md),
  shipped in v0.5.1
- [A chunker for any ad](2026-0018-a-chunker-for-any-ad.md)
  (roadmap 2026-0018), shipped in v0.5.0
- [Facts from the details, by grammar](2026-0007-facts-from-the-details.md)
  (roadmap 2026-0007), shipped in v0.5.0
- [Does it make search better? A first test](2026-0010-a-first-retrieval-test.md)
  (roadmap 2026-0010), shipped in v0.5.0
- [Find near-duplicate boilerplate](2026-0003-find-near-duplicate-boilerplate.md)
  (roadmap 2026-0003), shipped in v0.5.0
- [A readable view of chunked ads](2026-0016-a-readable-view-of-chunked-ads.md)
  (roadmap 2026-0016), shipped in v0.5.0
- [Try grammar's Overlap in sections.All, then take up grammar v0.3.2](2026-0056-try-overlap-and-take-up-grammar-v0.3.2.md),
  shipped in v0.3.5
- [Take up grammar v0.3.1, and read headings.txt with MapOnly](2026-0055-take-up-grammar-v0.3.1-and-read-headings-with-maponly.md),
  shipped in v0.3.5
- [Take up grammar v0.3.0 and scalar v0.6.0](2026-0054-take-up-grammar-v0.3.0-and-scalar-v0.6.0.md),
  shipped in v0.3.5
- [Check the rules on a second sample](2026-0001-check-the-rules-on-a-second-sample.md)
  (roadmap 2026-0001), shipped in v0.3.5
- [Stop shift times passing as headings](2026-0002-stop-shift-times-passing-as-headings.md)
  (roadmap 2026-0002), shipped in v0.4.0
