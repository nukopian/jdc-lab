# Metadata as filters in the retrieval test

**Status:** in progress

**Origin:** roadmap 34

**Started:** 2026-11-03, in 10b3c74d171698c9c4168806c2bb7ad976bda2f4

## The plan

**Serves:** metadata

**What:** use the facts read as metadata (pay, hours, work pattern,
place) as filters beside the embeddings in the retrieval test, and
measure what they add.

**Why:** that facts can be filtered on rather than embedded is the
lab's central claim, and it's untested: structure work is judged so far
by figures that stand in for search, not by search.

**What's known:** the first retrieval test
([record](2026-10-27-r10-a-first-retrieval-test.md)) has job titles as
queries, an ad relevant if its title is similar
(`boilerplate.SimilarTitles`), over 5,000 ads, embedded with
`BAAI/bge-base-en-v1.5`; an ad without its boilerplate ("own") is found
as well as the whole. Of every ad of the latest run, the facts find pay
in 64%, a work pattern in 60% (permanent, full time, contract,
temporary, hybrid, part time, office based, temp to perm, remote), a
schedule in 32%, hours in 15%, and a postcode in only 3%. Every ad has
a place from the export, `location_canonical`, which nothing in the lab
reads or makes: London, Birmingham, Manchester, Leeds and 5,386 more.

A fact's answer key can't be the lab's own facts, or filtering on them
scores perfectly by construction. A place has one: the export's. A work
pattern has none, so its relevance is judged.

This slice is place and work pattern, the two the roadmap names. Pay and
hours want queries with a range ("over £30,000"), and come after, if
this says filters are worth it.

**Steps:**

1. Queries over every distinct ad of the latest run, as a title and a
   fact:
   - **place:** "Care Assistant in Leeds": the pairs of a group of
     similar titles and an export place with the most ads, up to 30, no
     title or place more than three times; relevant, an ad whose title
     is similar and whose `location_canonical` is the place.
   - **work pattern:** "part time Receptionist", for part time,
     temporary, contract, hybrid and remote: the groups of similar
     titles with most ads, up to 20 queries, four of each pattern; an ad
     is a candidate if its title is similar, and relevant if it's judged
     to be for that pattern (step 3).

   `cmd/retrieval-set` writes them beside the first test's, with each
   ad's export place and the work patterns its facts found.
2. Two ways of ranking, each with the own and whole representations:
   - **embedded:** the query whole ("part time Receptionist")
     against every ad;
   - **filtered:** the title alone, against the ads the fact lets
     through: the export place, or a work pattern among the ad's facts.

   `retrieval_eval.py` embeds every distinct ad once, ranks, and writes
   each query's top 10 of each way.
3. **Judging, by pooling,** as search evaluation does: each work-pattern
   query's top 10 of each way, those whose title is similar, are
   judged by Claude reading the ad's whole text, as for the pattern,
   against it, or not saying, with the reason, kept in the private data
   store (shown to Iain first). An ad not judged counts as not
   relevant, so recall is within the pool.
4. Scores for each way, representation and kind of query: precision
   and nDCG at 10, and recall at 100 (place) or within the pool
   (pattern); and how many relevant ads the filter can't let through
   (place: none; pattern: those whose facts miss the pattern).
5. The record, PIPELINE.md and the README's Where it stands say what
   it showed; the test can be run again on any run.

**Done when** the retrieval test has queries that name a fact as well
as a title ("in Leeds", "part-time"), scored with the fact as a filter
and with it left to the embeddings; the figures are in its record and
the README's Where it stands.

## What happened

Work in progress.
