# Keep a list item's own blocks in the ad's order

**Status:** in progress

**Origin:** roadmap 17

**Started:** 2026-10-28, in 7dbc459e5efbfa5891397f2b2510ee61bd316c67

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
