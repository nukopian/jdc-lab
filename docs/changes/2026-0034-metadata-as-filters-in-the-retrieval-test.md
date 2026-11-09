# 2026-0034. Metadata as filters in the retrieval test

**Status:** shipped in v0.7.1

**Origin:** roadmap 2026-0034

**Started:** 2026-11-03, in 10b3c74d171698c9c4168806c2bb7ad976bda2f4

**Finished:** 2026-11-03

## The plan

**Serves:** metadata

**What:** use the facts read as metadata (pay, hours, work pattern,
place) as filters beside the embeddings in the retrieval test, and
measure what they add.

**Why:** that facts can be filtered on rather than embedded is the
lab's central claim, and it's untested: structure work is judged so far
by figures that stand in for search, not by search.

**What's known:** the first retrieval test
([record](2026-0010-a-first-retrieval-test.md)) has job titles as
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

- **Outcome:** done, for place and work pattern. A fact as a filter
  finds the ads asked for far better than the fact left to the
  embeddings, wherever the fact is read right; a fact read wrongly
  helps little, and remote, as the facts read it, not at all. Over all
  25,954 distinct ads of the latest run, with `BAAI/bge-base-en-v1.5`:

  | Queries | Way | Representation | P@10 | nDCG@10 | Recall | Filtered out |
  |---|---|---|---:|---:|---:|---:|
  | 30 place | embedded | own | 0.313 | 0.447 | 0.766 at 100 | |
  | | embedded | whole | 0.310 | 0.450 | 0.746 at 100 | |
  | | filtered | own | 0.657 | 0.936 | 1.000 at 100 | 0.000 |
  | | filtered | whole | 0.647 | 0.923 | 0.999 at 100 | 0.000 |
  | 20 work pattern | embedded | own | 0.394 | 0.472 | 0.426 at 10, in pool | |
  | | embedded | whole | 0.400 | 0.471 | 0.432 at 10, in pool | |
  | | filtered | own | 0.589 | 0.709 | 0.680 at 10, in pool | 0.114 |
  | | filtered | whole | 0.617 | 0.747 | 0.723 at 10, in pool | 0.114 |

  The best precision at 10 the queries allow, as some have fewer than
  ten relevant ads, is 0.707 for places and 0.706 for work patterns: a
  place as a filter comes within 0.05 of it. "Filtered out" is the share
  of relevant ads the filter can't let through: none for a place, whose
  filter is the answer key's own field, and 11% for a work pattern, ads
  judged for it whose facts didn't find it. 18 of the 20 pattern
  queries have a relevant ad; "hybrid Electrician" and "remote
  Warehouse Operative" have none, every ad judged against or silent.
- **By work pattern,** precision at 10, own text, embedded against
  filtered, and how the ads whose facts found the pattern were judged
  (for, against, silent):

  | Pattern | Embedded | Filtered | Facts that found it, judged |
  |---|---:|---:|---|
  | part time | 0.40 | 0.90 | 44, 2, 0 |
  | temporary | 0.70 | 0.80 | 52, 0, 5 |
  | hybrid | 0.38 | 0.53 | 28, 3, 0 |
  | contract | 0.28 | 0.38 | 25, 7, 14 |
  | remote | 0.03 | 0.05 | 2, 19, 0 |

  Part time, temporary and hybrid are read right nearly always, and
  filtering on them pays most where the embedding does worst: part time
  doubles. Contract is read where the word stands for something else,
  and its judging was hardest (below). Remote is read in negations
  ("This is NOT a hybrid or remote role", "Remote working is not
  available"), in other senses ("radio remote and pendant controls",
  "24/7 remote GP service", "the remote location of the site") and for
  part of the week ("3 days in the office, 2 days remote"): roadmap idea
  47.
- **Judging:** 307 ads, each work-pattern query's top ten of each way
  and representation whose title is similar, judged by Claude, in four
  batches, by rules written before any ad was seen (kept with the
  judgements in the private data store): for, if the job is offered on
  the pattern, alone or as an option; against, if the ad says another;
  silent, if it doesn't say. 163 for, 83 against, 61 silent. The calls
  the judges found hardest, and how they went: a full school week
  ("8.50am - 3.20pm, 5 days per week, term time only") counted part
  time, as the rules named school hours, though other ads call such
  weeks full time; a rolling agency contract counted temporary, and
  daily-rate supply work not a contract; "ongoing temporary contract"
  counted a contract; a "social housing contract" or "MOD Contract",
  the employer's with its client, silent; work at customers' homes or
  on site against hybrid and remote; a base town alone ("London-based
  position") silent. An ad no way ranked in its top ten isn't judged,
  so recall is within the pool, and the pool is judged once, by one
  model: a second judge, or Iain, on a sample would say how far to
  trust it.
- **Changes from the plan:** a pattern query needs at least three ads
  whose facts found the pattern, so there's something to find, which
  chooses queries by the facts, though it doesn't judge them. The
  ranking embeds every ad's own and whole texts, about 52,000 texts, in
  about 40 minutes on a Mac's GPU. `cmd/retrieval-set` groups 26,000
  titles twice, which takes a minute: fine for a test run rarely. The
  score functions of the first test are generic now, for IDs as well as
  places in a list.
- **What it means:** the lab's central claim holds for these two facts:
  filtering on a fact finds what's asked for far better than embedding
  it, so a fact read reliably is better metadata than text. How
  reliably the fact is read decides how much it pays, so a fact's
  rules are worth checking by search, not only by count. Pay and hours,
  which want queries with a range, are roadmap idea 48.
