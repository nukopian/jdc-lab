# Boilerplate's verdict checked

**Status:** shipped, unreleased

**Origin:** roadmap 33, reshaped before work started

**Started:** 2026-11-03, in 050edf7eccfccf6da8e51d7bc89cae20bab21661

**Finished:** 2026-11-03

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

- **Steps 1 and 2:** `Index.Why` names the rule that made a text
  boilerplate and the key that found it; `Boilerplate` asks it, so the
  two can't disagree, and `TestWhy` covers each rule, each kind of key
  and a wall. `boilerplate.Check` and `cmd/boilerplate-check` list every
  text found as boilerplate, with the rule, the flags, its characters
  in all, in the sections that describe the job, and in those whose
  text is embedded, and the title of the first ad it's in. The run's
  output is unchanged (`pipeline compare`).
- **What the table shows,** on the latest run: 13,480 texts, 11.2
  million characters. By rule: found mostly elsewhere 84% of the
  characters, a footer 13.5%, widespread 1.8%, a wall's verdict 0.7%.
  Flagged: 3,597 texts, 27% of the characters; mostly in the sections
  that describe the job 14.7%, holding a fact 13.0%, a job line 0.7%.
  What the verdict takes out of search is its text in the embedded
  sections (summary, company, role, responsibilities, requirements,
  desirable): 4.5 million characters, 41% of it. Benefits, apply and
  legal are left out whatever their text, so a misfiled benefit loses
  nothing to search.
- **Change from the plan:** 3,597 flagged texts were too many to judge
  each, so both strata were sampled, by probability proportional to
  characters, without replacement, seed 33: 300 flagged and 150 not,
  450 texts, 89,000 characters. Judged by Claude in four batches, blind
  to the rule and the flags, by rules written before any was seen:
  boilerplate, about the job, or mixed, and what in it matters (fact,
  duty, requirement, conditions, setting). Kept in the private data
  store.
- **Outcome so far:** 292 judged boilerplate, 118 about the job, 40
  mixed. Weighted to the characters in embedded sections, 35% (±8) of
  what the verdict takes out of search isn't boilerplate alone, and
  26% (±7) is about the job; 31% if checks every job of its kind has
  (an enhanced DBS check, the right to work) aren't counted as about
  the job. That's about 1.5 million characters, 2.4% of all the ads'
  text. The flags don't single it out: flagged and unflagged text is
  misfiled about as often (35% each, embedded). By rule, roughly, of
  each rule's embedded text: found mostly elsewhere about 45%, a footer
  about 25%, widespread about 60%, from samples too small for more.
- **What's misfiled,** from the largest judged texts:
  - a large employer's own account of the role, repeated across its
    ads under titles in different groups ("As a Bank Care Assistant,
    you could be doing anything from …", "As a Roadside Technician …
    you'll deliver expert mechanical roadside repair"), filed as a
    footer or as found elsewhere;
  - the setting: a school group's or care provider's account of what
    it provides, in the company's section;
  - requirements and checks written into an agency's notices, and
    paragraphs mixing an agency's notice with the job's terms ("This
    advert is for a temporary position");
  - an employer's standard terms in the benefits, which are left out
    anyway.
- **Facts in sentences found as boilerplate (step 5):** over every ad,
  2,375 such sentences in the embedded sections hold a fact the chunker
  doesn't read, and in 1,884 ads one isn't in the details otherwise:
  2,802 work patterns, 370 pay, 153 schedules, 100 hours, 22 places.
  Most are an agency describing itself ("an employment agency for
  permanent recruitment and … temporary workers") or a company's
  turnover read as pay, which the details are right to leave out; of
  the 120 judged texts holding a fact, 45 were judged to state one of
  the job. So reading every fact of those sentences would add more
  wrong facts than right: not a small fix.
- **Judging:** the calls the judges found hardest were checks written
  as notices (counted as requirements, by the rules), an agency's
  generic perks named by brand, personal qualities ("a genuine passion
  for supporting children"), and a run of keywords repeated across an
  agency's ads. One model judged once; a second judge, or Iain, on a
  sample would say how far to trust it.

- **Fixes (step 6),** each on the roadmap, in Next, as Iain chose:
  an employer's own account of the role kept (roadmap 50); repeated
  paragraphs judged sentence by sentence where they mix standard text
  and the job's terms (51); facts of this job read from sentences found
  as boilerplate (52), which is step 5's fix; and the setting kept from
  the company's text (53). Each is measured on this sample. Meanwhile
  boilerplate stays out of the embeddings, as Iain chose, with the
  known cost written down here.
- **Done when,** against it: the table any run can rebuild, yes; the
  judged sample's share, yes; a fix on the roadmap for each kind of
  misfiling, yes; facts in sentences found as boilerplate, an item
  (52); PIPELINE.md and the README, yes.
