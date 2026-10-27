# TODO

## readme-pdf.sh refuses a change record by its full path

**From:** scalar's session, 2026-10-27 (scalar 4ecbc6a70c815b55c4e97e74593d8cd9ac6f57d8)
**Waiting on:** nothing

**What:** `build` in `scripts/readme-pdf.sh` matches the file name as
typed, so `cd docs && ../scripts/readme-pdf.sh changes/<record>.md`
builds a PDF of a change record that should be refused.

**Why:** the refusal exists so that the CHANGELOG and change records
never get PDFs; a relative path slips past it. scalar's wash-up before
v0.6.1 found it in scalar's copy, which came from jdc-lab's.

**Steps:**

1. Match the full path, as scalar's fix (the commit above) does:

   ```sh
   	# By its full path, so that changes/x.md, given from docs/, is refused too.
   	case $(cd "$(dirname "$1")" && pwd)/$(basename "$1") in
   	*/CHANGELOG.md | */docs/changes/*)
   ```

**Done when** `readme-pdf.sh` refuses, with exit status 2, each of
`CHANGELOG.md`, `../CHANGELOG.md` from `docs/`, and `changes/` files
from `docs/`, and still builds the other documents.
