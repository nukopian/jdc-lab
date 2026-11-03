# Boilerplate's verdict checked

**Status:** in progress

**Origin:** roadmap 33, reshaped before work started

**Started:** 2026-11-03, in 050edf7eccfccf6da8e51d7bc89cae20bab21661

## The plan

**Serves:** structure

**What:** check that the text boilerplate's verdict takes is boilerplate:
not text about the job, nor a fact of it, filed there by mistake.

**Why:** text found as boilerplate isn't embedded. It isn't thrown away
either: the chunker keeps it in a chunk of its own, at the end, each
line with the section it came from. But what isn't embedded can't be
found by search, so text about the job filed there is lost to it, and
there's reason to doubt the verdict: about one in seven of the texts
roadmap 3 newly found are openers about the job, and of what all three
samples find that a slice of 2,000 ads doesn't, about one in five texts
is. Iain wants to be sure that all of it is boilerplate before relying
on it. Whether leaving it out costs retrieval needs queries a title
can't answer, which wait on roadmap idea 32's judged queries: roadmap
idea 49.

**What's known:** the latest run finds 13,480 distinct texts
boilerplate, about 2 million characters, 11 million in all the ads
they're in. A text is boilerplate by one of four rules (see
PIPELINE.md's Boilerplate): found more often outside the sections that
describe the job; found mostly at the end of its ads and talking as a
footer does; under `Widespread` groups of titles or more; or a line of a
wall judged boilerplate whole. The chunker reads a line that is mostly
facts into the details before asking whether it's boilerplate, but a
sentence found as boilerplate in a section that's embedded goes to the
boilerplate chunk without its facts being read: "We offer a salary of
£30,000 and …" would lose its pay from the details.

**Steps:**

1. Say why each text is boilerplate: a method on `boilerplate.Index`
   giving the rule that decided, tested to agree with `Boilerplate`
   wherever it's asked.
2. A program, `cmd/boilerplate-check`, writing every distinct text
   found as boilerplate, with its characters, ads, groups of titles and
   kinds of section, the rule that decided, and three flags for where
   something important is likeliest: it holds a fact (pay, hours, work
   pattern, schedule, place); it opens as a duty or requirement does
   (`JobLine`); it's found mostly in the sections that describe the
   job. A table any run can rebuild.
3. Judging, by rules written before any text is seen: every flagged
   text, or the most characters of them if there are too many to judge,
   and a random sample of the rest, weighted by characters, so the
   blocks repeated most count most. Each is boilerplate, about the job,
   or mixed, and what in it matters if it isn't boilerplate. Judged by
   Claude, as roadmap 34's work patterns were, and kept in the private
   data store, shown to Iain first.
4. How much of boilerplate, by characters, is about the job, flagged and
   not, by rule and by kind of section, with the judged examples: what
   gets misfiled, and which rule files it.
5. How many facts the chunker misses in sentences found as boilerplate,
   over every ad; fixed, with a test, if the fix is small, or an item
   on the roadmap if not.
6. Each kind of misfiling found becomes a roadmap item with its fix,
   and PIPELINE.md and the README say what was found.

**Done when** every text found as boilerplate has its rule and flags in a
table any run can rebuild; a judged sample says how much of
boilerplate, by characters, is about the job, flagged and not; each
kind of misfiling has a fix on the roadmap, or made; facts in sentences
found as boilerplate reach the details, or there's an item to make them;
and PIPELINE.md and the README's Where it stands say what was found.

## What happened

Work in progress.
