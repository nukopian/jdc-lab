"""Job ads exported by MySQL's --xml option, turned into the lab's records.

The export is a <resultset> of <row>s, each holding a <field name="..."> per
column, with NULL written as xsi:nil="true". XML is awkward to work with, so
the lab converts it once to JSON Lines: one ad per line, as a JSON object
with the lab's own field names (FIELDS), which jq, Python and Go all read
without help.

The conversion keeps every row and every value as it was: duplicates, HTML
and odd characters are for later steps to deal with. The one change is
XML's own: parsing turns each CRLF in a value into LF.
"""

import json
import xml.etree.ElementTree as ET
from collections.abc import Iterator
from pathlib import Path

NIL = "{http://www.w3.org/2001/XMLSchema-instance}nil"

# The lab's name for each column of the export.
FIELDS = {
    "ref": "id",
    "jobtitle": "title",
    "location": "location",
    "location_canonical": "location_canonical",
    "description_formatted": "html",
}


def rows(path: Path) -> Iterator[dict[str, str | None]]:
    """Yield each row of the export as a dict of column name to value."""
    # iterparse reads a row at a time; clearing each one when done keeps
    # memory flat however large the export.
    for _, element in ET.iterparse(path):
        if element.tag != "row":
            continue
        row: dict[str, str | None] = {}
        for field in element.iter("field"):
            name = field.get("name")
            if name is None:
                raise ValueError("a <field> without a name: not a MySQL export")
            row[name] = None if field.get(NIL) == "true" else field.text or ""
        yield row
        element.clear()


def record(row: dict[str, str | None]) -> dict[str, str | None]:
    """Rename a row's columns to the lab's names, dropping any not in FIELDS.

    A missing column is an error, not a silent None: it means the export
    isn't the one this was written for.
    """
    missing = FIELDS.keys() - row.keys()
    if missing:
        raise KeyError(f"row lacks {', '.join(sorted(missing))}")
    return {name: row[column] for column, name in FIELDS.items()}


def convert(source: Path, target: Path) -> int:
    """Write source's rows to target as JSON Lines; return how many."""
    target.parent.mkdir(parents=True, exist_ok=True)
    count = 0
    with target.open("w", encoding="utf-8") as out:
        for row in rows(source):
            # ensure_ascii=False keeps £, ’ and emoji readable in the file.
            out.write(json.dumps(record(row), ensure_ascii=False) + "\n")
            count += 1
    return count
