# A readable view of chunked ads

**Status:** shipped, unreleased

**Origin:** roadmap 16

**Started:** 2026-10-27, in 1a734be8cf0260f9cf8d8313972a0b5ca925cb63

**Finished:** 2026-10-27

## The plan

**Serves:** structure

**From:** jdc-lab's session, 2026-10-27 (jdc-lab 1a734be8cf0260f9cf8d8313972a0b5ca925cb63)
**Waiting on:** nothing

**What:** each run writes, for each sample, about 20 ads in their
chunked form, readable by eye: each section with its kind and heading,
fields, repeats, list items, and the text found as boilerplate or
template marked where it sits. It also shows the text that no heading
placed: the opening that's summary only because nothing headed it, and
unmapped sections. And it checks that every ad's text is in some
section, naming any that's missing.

**Why:** the summaries count what moved, but not what a change did to
real ads; reading ads is how faults such as "Part-Time Teachers Wanted
for Day-to-Day Supply!" filed as responsibilities are found. With the
same ads in every run, a diff of two runs' views shows a change's
effect ad by ad.

**What's known:** the stages already hold everything: blocks, headings
and sections give each ad's sections, and boilerplate's Index gives
each text's verdict. No text is lost on the way today: comparing each
ad's HTML text with its sections' text, the only differences are words
an advertiser ran together ("Operative</strong>to"), kept as written.

**Steps:**

1. `internal/view`: choose ads by a fixed hash of their IDs, so the
   same ones come in every run, with one each that has no heading, an
   unmapped section, and over half its text boilerplate; render an ad
   as Markdown; and find text of the ad's that's in no section.
   Tests and examples use made-up ads.
2. `cmd/view`: a stage, as `cmd/boilerplate` is, writing
   `view/ads.md` and a summary: the ads shown, and how many of the
   sample's ads have text in no section.
3. `cmd/pipeline` runs it with the other stages.
4. PIPELINE.md describes the stage and how to diff two runs' views;
   the README's Layout lists the new program and package.

**Done when** each run has `data/runs/RUN/SAMPLE/view/ads.md`, the
same ads in every run, a diff of two runs' views shows what a change
did, and the summary says how many ads have text in no section (none,
on today's samples).

## What happened

- **Outcome:** done. Each run writes `view/ads.md` for each sample, 20
  ads each; two runs gave identical views, so a diff of two runs'
  views shows only what a change did. The first sample's first ad
  shows a fault no figure had: "Why join our client", "How you'll make
  an impact" and "Why you're a great fit", headings written as plain
  lines, go unseen, and the whole ad is one summary (roadmap 6).
- **Changes from the plan:** the check finds text out of the ad's
  order as well as text in no section, and tells them apart, and one
  ad with either is chosen on purpose, so the view shows one. Lists
  render as tight lists, and a location the same as the canonical one
  once.
- **Measurements:** across the three samples' 25,954 distinct ads, no
  text is in no section, and 20 ads have text out of order: a list
  item that holds a list or paragraphs of its own comes out after the
  blocks that follow it, so its text lands under the wrong heading.
  The view stage takes about 12.5 seconds and 870 MB on `all`, as it
  builds the boilerplate index as `cmd/boilerplate` does.
- **Tried and reverted:** a first check compared words, and took an
  advertiser's "Operative</strong>to", kept as written, for a loss; it
  compares letters and digits instead. Its first version called the
  out-of-order text missing, until both counts were compared.
- **What's left:** the blocks stage's order fault, idea 17 on the
  roadmap.
- **Afterwards,** in roadmap 3: the ads chosen on purpose came first,
  and could change with a change to boilerplate, shuffling the view.
  Sixteen ads are now always the first by the hash, and all are
  written in its order (see that
  [record](2026-10-27-find-near-duplicate-boilerplate.md)).
