# Check the rules on a second sample

**Status:** in progress

**Origin:** roadmap 1

**Started:** 2026-10-25, in dedfbc7b228ede0069cd12c0393301288fe88f0f

## The plan

**Serves:** structure

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** work down the second sample's unmapped table, and change any
rule that fitted only the first.

**Why:** every rule, phrase and threshold was tuned on the first
sample, which leans towards education jobs. The second leans towards
vehicle jobs.

**What's known:** both samples are in `data/ingestion/`, and their
figures in PIPELINE.md's What it measured. The rules mostly hold. Where
they don't: heading uses of a known kind fall from 76% to 73%, unmapped
text rises from 14% to 16%, and a quarter of that is boilerplate. The
biggest unmapped heading is "and" (39 ads, 82,000 characters), with
"OR" beside it: one advertiser's template puts them alone in bold
between requirements (Known gaps: lone conjunctions). Then phrases
that look like kinds: "who thrives here", "what to expect", "the
business", "what is in it for you", "vacancy type", "working
arrangement", "important notice". "The business" also maps "The
business has a strong presence across:", so check each in context.

**Steps:**

1. `go run ./cmd/pipeline run before`.
2. Stop a lone conjunction passing as a heading, in
   `internal/headings`, with a case in `TestRead`.
3. Work down `data/runs/before/ads-20261025/sections/unmapped.tsv` as
   PIPELINE.md's Improving it says, comparing each change on every
   sample.

**Done when** the second sample's known headings are at least the
first's 76%, neither sample's figures got worse, the lone conjunctions
gap is out of PIPELINE.md, and its figures are updated.

## What happened

Nothing yet: the item has been in Now since the roadmap began, but no
work on it has been committed.
