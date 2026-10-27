# Does it make search better? A first test

**Status:** shipped, unreleased

**Origin:** roadmap 10, a first slice

**Started:** 2026-10-27, in f4d5cdc22b59590e0b2e3ea22502b8b1692b3a86

**Finished:** 2026-10-27

## The plan

**Serves:** search, embeddings

**From:** jdc-lab's session, 2026-10-27 (jdc-lab f4d5cdc22b59590e0b2e3ea22502b8b1692b3a86)
**Waiting on:** nothing

**What:** a first retrieval test. Does embedding an ad's chosen
sections find ads for a job better than embedding the whole ad, and
for how much less text? This first slice compares local
representations only; the managed embedding service, metadata to
filter on, and judged queries come later.

**Why:** the lab has measured structure, not its value. A summary that
holds 17% of all text by default, or misfiled footers, matter only if
they hurt retrieval; this says which chunking faults are worth fixing.

**What's known:** every run has each ad's sections, and boilerplate's
verdict on each text. `BAAI/bge-base-en-v1.5` and sentence-transformers
are in the venv. No queries or judgements of relevance exist. The
model reads at most 512 tokens, about 350 words, so a whole ad is cut
short, and what's cut is its end.

**Steps:**

1. A Go program builds the test set from a sample's ads: about 5,000
   distinct ads chosen by a fixed hash of their IDs, and for each, the
   texts of each representation:
   - **whole:** the ad's text, as the description is embedded today
     (the title field left out, as it's embedded apart);
   - **own:** the same without the text found as boilerplate;
   - **core:** responsibilities, requirements and desirable only;
   - **summary:** the summary only;
   - **sections:** each section's own text, for one vector per section,
     an ad scoring by its best.
2. The queries are the 50 commonest groups of similar titles in the
   sample, each written as its commonest spelling, with every ad of the
   sample whose title is similar to it relevant (boilerplate's test of
   similar titles): so a query asks whether a description finds ads
   for the same job. Go writes the queries and the relevant ads.
3. A Python program embeds each representation with BGE (its query
   instruction on the queries), ranks the sample's ads for each query,
   and reports, for each representation, precision at 10, nDCG at 10
   and recall at 100, and the characters embedded per ad.
4. PIPELINE.md and the README say what it showed.

**Done when** the report gives each representation's scores and cost on
the sample, the record says what that means for the chunking work, and
the test can be run again on any run.

## What happened

- **Outcome:** done, for this slice. On the three samples together, 5,000
  ads and 50 job titles as queries (1,813 relevant ads; chance would
  put 0.007 of the top 10 right), with `BAAI/bge-base-en-v1.5`:

  | Representation | P@10 | nDCG@10 | R@100 | Tokens an ad | Vectors an ad |
  |---|---:|---:|---:|---:|---:|
  | whole | 0.822 | 0.848 | 0.892 | 473 | 1.00 |
  | own (no boilerplate) | 0.822 | 0.850 | 0.890 | 396 | 1.00 |
  | summary | 0.770 | 0.816 | 0.784 | 126 | 0.89 |
  | sections, best of | 0.784 | 0.832 | 0.885 | about 400 | 4.35 |
  | core | 0.608 | 0.676 | 0.625 | 157 | 0.85 |

  Tokens are BGE's. 37% of whole ads are over the model's 512, so 15%
  of their tokens are never read; 23% of the own texts are.
- **What it means, and doesn't:** for finding ads for a job by its
  title, leaving out the text found as boilerplate loses nothing, for
  16% fewer tokens; the summary carries most of what says which job an
  ad is for; and a vector for each section, or the sections that
  describe the job alone, don't help. But a job title is the one query
  the title's own embedding already answers, so this can't see what's
  lost if text called boilerplate is about the job, nor what sections
  add for queries beyond the title: skills, duties, conditions. That
  boilerplate's verdict is right isn't shown here, and is doubted: about
  one in seven of the texts roadmap 3 newly found are openers about the
  job. So none of this says to drop boilerplate yet.
- **Changes from the plan:** two faults in choosing queries, each
  reproduced by a test first: two groups could share their commonest
  title, as grouping goes in order, giving one query twice; and a
  query's fewest relevant ads was counted on its group, not on its own
  title, letting through queries with two. The Python program is
  `retrieval_eval.py`, as a name ending `_test.py` reads as a test.
  `boilerplate.SimilarTitles` is exported for the queries.
- **Found on the way:** `headings.Key` keeps a typographic dash as
  `-`, where it makes an ASCII hyphen a space, so "Title – Retail" and
  "Title - Retail" key differently; 1,587 headings' spellings have such
  a dash. Fixing it moves keys throughout, so it's left for a change of
  its own. `cmd/view`, `cmd/boilerplate` and `cmd/retrieval-set` now
  each read the ads and index boilerplate the same way, for a wash-up
  to share.
- **What's left:** queries judged by hand, beyond titles; a check of
  boilerplate's verdict on a sample; metadata to filter on; a managed
  embedding service beside the local model.
