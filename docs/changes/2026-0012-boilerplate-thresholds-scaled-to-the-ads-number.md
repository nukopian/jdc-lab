# 2026-0012. Boilerplate thresholds scaled to the ads' number

**Status:** dropped

**Origin:** roadmap 2026-0012

**Started:** 2026-11-03, in 383ddd5bd06cc12bf9063e8f04e8e7ed68d720ae

**Finished:** 2026-11-03

## The plan

The roadmap item, word for word, as it stood when work started; no
further plan was written, as the first measurement settled it.

**Serves:** structure

**Size:** M (3)

**What:** scale boilerplate's thresholds, now counts of title groups,
to the number of ads read, so its verdict doesn't depend on how many
there are.

**Why:** more ads pass a count more easily, so boilerplate's share
rises as each batch is added (12.5% of text on 5,000 ads, 15.1% on
21,000, by word for word matching), and figures from before and after a
batch don't compare. It's to be done before the next batch is added.

**Done when** boilerplate's share is steady, within a point, across the
samples and their union; tests cover the scaling; and PIPELINE.md's
figures and thresholds are updated.

## What happened

- **Outcome:** dropped before any code changed. Boilerplate's share
  does rise with the number of ads, but what a larger sample adds is
  mostly boilerplate, found because more of each advertiser's ads are
  seen, so scaling the thresholds down to the smallest sample's verdict
  would throw away correct finds. Iain chose to drop it.
- **Measurements,** `cmd/boilerplate` on each sample alone, and on
  random slices of the three together (the latest run's ads, shuffled):

  | ads read | distinct | boilerplate |
  |---|---:|---:|
  | first sample alone | 1,890 | 15.1% |
  | second sample alone | 4,296 | 14.9% |
  | third sample alone | 19,913 | 17.1% |
  | a slice of 2,000 | 1,934 | 11.2% |
  | a slice of 5,000 | 4,857 | 13.5% |
  | a slice of 10,000 | 9,374 | 15.0% |
  | a slice of 20,000 | 18,641 | 16.7% |
  | all three | 25,954 | 17.8% |

  The share climbs steadily with the ads read, by about a point and a
  half each time they double. A sample alone has a higher share than a
  slice of its size, as its ads come from a few days, when the same
  advertisers post many.
- **What the extra is:** 2,524 texts of the slice of 2,000 are
  boilerplate when all the ads are read and not in the slice alone,
  nearly all of them (2,458) not repeated there at all. Of 25 picked at
  random, about 20 are plainly boilerplate: equal opportunities and
  inclusion statements, an agency's notices, standard benefits lines,
  "To see our other available vacancies, please visit our website". In
  the slice, too few of each advertiser's ads were read for them to
  reach two groups of titles. About 5 are lines about the job, such as
  "SEN: Must be willing to support children with SEN": the share of the
  verdict that's about the job is what roadmap 33 judges by hand.
- **What instead:** the thresholds stay. Boilerplate figures compare
  only on the same ads, which is how `pipeline compare` compares runs;
  a batch added moves them, and PIPELINE.md says so. The doubtful finds
  are left to roadmap 33, whose Why now names them.
