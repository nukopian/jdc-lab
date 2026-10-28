# Keep a list item's own blocks in the ad's order

**Status:** shipped, unreleased

**Origin:** roadmap 17

**Started:** 2026-10-28, in 7dbc459e5efbfa5891397f2b2510ee61bd316c67

**Finished:** 2026-10-28

## The plan

**Serves:** structure

**From:** jdc-lab's session, 2026-10-28 (jdc-lab 7dbc459e5efbfa5891397f2b2510ee61bd316c67)
**Waiting on:** nothing

**What:** `internal/blocks` writes out a list item's own blocks,
meaning the paragraphs, headings and lists inside an `<li>`, where they
stand in the ad. Today some of that text comes out after the blocks
that follow it, and so lands under the wrong heading.

**Why:** text out of order is text under the wrong heading. It gets
the wrong kind, and may be embedded or left out wrongly. One of the
pages of ads before and after shows the worst case. The whole ad sits
inside a single list item, and its pay, in a list inside that item,
comes out last, under the closing "Why …?" heading, and is left out as
benefits. That ad's empty "Skills and experience" chunk has the same
cause, as its list comes out at the end too.

**What's known:** `cmd/view` counts the ads whose sections hold all
their text, but out of the ad's order: 20 on the three samples (2 on
`ads-20261024`, 1 on `ads-20261025`, 17 on `ads-20261027`). Every one
has a list inside a list item, and 17 have paragraphs inside items too;
the same advertisers' markup accounts for most of the third sample's.
The order is decided in the builder's `start` and `end`
(`internal/blocks/blocks.go`): a `<ul>` inside an item ends the item's
text so far, and a `<p>` inside one is only a space, so that
`<li><p>Drive</p></li>` stays one item.

**Steps:**

1. A failing test in `internal/blocks` for each shape seen: an item
   holding a paragraph and then a list, followed by more of the ad; and
   the whole ad inside one item, headings and all. Each wants its
   blocks in the ad's order.
2. Fix the builder so that an item's own blocks are written out where
   they stand. Keep a plain item, and `<li><p>…</p></li>`, as one item.
3. Save a run before the fix and one after (`pipeline run RUN`), and
   compare them. Read the ads that move, as the pages show them.
4. PIPELINE.md: drop the "Text out of order" row from Known gaps,
   update what it says of blocks, and update the figures that move.
   Update the note for that page's ad in the data store's `ads.tsv`.

**Done when** `cmd/view` reports no ads with text out of order on any
of the three samples; the tests cover items holding lists and
paragraphs; and the compare's changes are read and explained here.

## What happened

- **Outcome:** done. `cmd/view` reports no ads with text out of order
  on any of the three samples (from 2, 1 and 17), and none with text in
  no section. Tests in `internal/blocks` cover an item's paragraph
  after its sublist, the whole ad in one item, and text between an
  item's sublists.
- **What was wrong:** two things in the builder. Inside a list, `add`
  held back the items gathered so far until the outermost list closed,
  so an item's own paragraphs and headings came out ahead of them, and
  every sublist of the item piled into one list at the end. Text in a
  list but in no item, such as a bold line between two sublists, was
  taken into the next item, or waited, as a line ended by `<br>`, until
  after the list. Now a block always ends the list before it, and text
  in a list but in no item is a block of its own, as it is outside a
  list.
- **Changes from the plan:** the plan saw one cause; the first fix (the
  held-back items) took the count from 20 to 8, and the remaining 8 were
  the second (text between sublists), which the third test covers.
- **Measurements:** `pipeline compare claude/before claude/after`, on
  all three samples. Of 28,000 ads, 22 have different blocks: the 20,
  and one ad posted twice whose markup puts "- " lines straight in
  `<ul>`, in no `<li>`, which are now blocks and, by their bullets, a
  list again, after the "Skills:" item. Paragraphs 259,908 to 259,921,
  lines 146,765 to 146,756, lists 74,613 to 74,637; one more heading
  use (of a known kind); ads with only a summary 1,038 to 1,035;
  apply's boilerplate share 56% to 57%. Where an ad's tokens go is
  unchanged at one decimal: 57.5% embedded, 3.0% metadata, 20.3% left
  out, 19.2% in the boilerplate chunk. The pages' "Why ...?" ad now has
  its pay at the top, where the ad puts it, and its requirements under
  "Skills and experience"; its embedded share is 44%.
- **Tried and reverted:** nothing.
- **What's left:** nothing.
