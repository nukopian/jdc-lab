# Find near-duplicate boilerplate

**Status:** shipped, unreleased

**Origin:** roadmap 3

**Started:** 2026-10-27, in 47a0e01caeae782db372469a19bdcd3f5543f689

**Finished:** 2026-10-27

## The plan

**Serves:** structure, embeddings: what isn't found as boilerplate gets embedded

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** find boilerplate that's the same standard text with a few
words changed, which `internal/boilerplate` misses because it matches
word for word.

**Why:** most of it is likely in summaries, which are 28% of the text
but measured only 6% boilerplate. It's probably the largest amount of
boilerplate still hidden, and what isn't found gets embedded.

**What's known:** texts of eight words or more whose first 50
characters recur in five ads or more number 178, against 155 whole
texts: about 23 groups of variants. The commonest change is likely the
ad's own title or place written into the standard text ("we're
recruiting a … in …").

**Steps:**

1. Cheapest first: before keying a text, replace the words of the ad's
   own title and locations in it with one placeholder, so that
   templated text keys the same in every ad. Measure what that finds.
2. If that leaves much, compare texts by shingles (overlapping runs of
   five words) and their overlap, or MinHash for speed, grouping texts
   above a threshold; tune it on the variant groups above.
3. Keep the title-groups rule: variants count as one text across
   different jobs.

**Done when** boilerplate's share of summaries and of the whole has been
remeasured, a sample of the newly found texts has been checked by eye
to be standard text and not job text, tests cover the matching, and
PIPELINE.md describes it.

## What happened

Work started 2026-10-27, on the three samples together, as runs now
are; the plan's figures are from the first sample.

- **Outcome:** done. Boilerplate is 17.6% of the three samples' text,
  from 15.8%, and 10% of summaries, from 7%; ads over half boilerplate
  are 9%, from 6%. Template for the job rose from 2.9% to 3.1%. Headings
  and sections didn't move (`pipeline compare claude/before
  claude/after`). Tests cover the templated keys, the plain key still
  counting, long variants matched and short ones not, a text shared in
  part not matched, and groups worked out again after more ads.
  PIPELINE.md describes it.
- **Checked by eye:** fifteen new finds drawn at random, and the four
  that changed in the 20 ads of the view. Almost all are standard text:
  agencies' disclaimers and safeguarding notices with the agency's name
  changed, company introductions, apply and contact lines with the
  title written in. About one in seven is an advertiser's opener
  reused across related jobs, with the title and place swapped in ("…
  is currently recruiting an experienced Senior Health and Safety
  Advisor… based in Swindon"): about the job, but in summaries, so
  counted as boilerplate.
- **Changes from the plan:** both steps were needed. The placeholders
  alone found 0.4 points (to 16.2%); runs of five words (shingles)
  found the rest. A text keeps its own key beside the templated one,
  as templating alone broke exact repeats: standard text naming
  "support workers" keyed differently in an ad titled "Support
  Worker", and boilerplate fell to 14.9% when only the templated key
  counted. Placeholders replace the ad's whole title, as a phrase;
  replacing any of its words found no more and risked more. Shingles
  are compared in full, not by MinHash, with runs in over 300 texts
  left out of finding candidates, which is fast enough.
- **Measurements:** on the three samples together, at an overlap of
  0.5 boilerplate would be 18.1% and summaries 10.3%, at 0.6 17.6% and
  9.6%, at 0.7 17.3% and 9.3%; 0.6 was taken, the samples at 0.5 and
  0.6 being much alike. The boilerplate and view stages each take about
  8 seconds more on the three samples (about 20 in all), and peak at
  1.1 and 1.3 GB, from 0.7 and 0.9: the templated keys are a second
  entry for many texts. Holding runs of words as 64-bit hashes, not
  strings, took the peaks down from 1.3 and 1.7 GB, with the same
  figures.
- **Tried and reverted:** shingles on every text of 8 words or more.
  At 0.5 that found 19.2%, but the sample was mostly short lines about
  the job phrased alike ("strong knowledge of uk health and safety
  legislation", "assist with goods in and goods out activities"),
  joined across five groups and so counted as boilerplate wherever
  they were: job text kept out of the embeddings. Hence `LongWords`.
- **Found on the way:** the view's picks (roadmap 16) could move when a
  change moved boilerplate, as the ad chosen for being mostly
  boilerplate did here, shuffling the view and its diff. Sixteen ads
  are now always the first by the hash, and all are written in its
  order.
- **What's left:** openers reused across related jobs, which the
  verdict can't tell from standard text; standard text under 20 words
  with a word changed.
