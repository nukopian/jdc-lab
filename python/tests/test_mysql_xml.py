"""Tests for jdclab.mysql_xml, on a small made-up export in MySQL's layout."""

import json

import pytest

from jdclab.mysql_xml import FIELDS, convert, record, rows

# Two ads as MySQL's --xml writes them: HTML escaped as entities, a NULL
# location, a CRLF inside a value, and characters beyond ASCII.
EXPORT = """<?xml version="1.0"?>

<resultset statement="select ..." xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <row>
\t<field name="ref">2</field>
\t<field name="location">Leeds, West Yorkshire</field>
\t<field name="location_canonical">Leeds</field>
\t<field name="jobtitle">Café Manager</field>
\t<field name="description_formatted">&lt;p&gt;&lt;strong&gt;The role&lt;/strong&gt;&lt;/p&gt;&lt;p&gt;£30,000 ✨&amp;nbsp;\r\nnext&lt;/p&gt;</field>
\t<field name="extra">ignored</field>
  </row>
  <row>
\t<field name="ref">1</field>
\t<field name="location" xsi:nil="true" />
\t<field name="location_canonical"></field>
\t<field name="jobtitle">Driver</field>
\t<field name="description_formatted">&lt;p&gt;Drive.&lt;/p&gt;</field>
\t<field name="extra">ignored</field>
  </row>
</resultset>
"""


@pytest.fixture
def export(tmp_path):
    path = tmp_path / "export.xml"
    # newline="" writes the \r\n as it is, as MySQL does.
    path.write_text(EXPORT, encoding="utf-8", newline="")
    return path


def test_rows_reads_every_row_in_order(export):
    assert [row["ref"] for row in rows(export)] == ["2", "1"]


def test_rows_unescapes_html_and_keeps_characters(export):
    html = next(rows(export))["description_formatted"]
    assert html is not None
    assert html.startswith("<p><strong>The role</strong></p>")
    assert "£30,000 ✨&nbsp;" in html


def test_rows_turns_crlf_into_lf(export):
    # XML parsers must (XML 1.0, section 2.11), unless the export had
    # escaped the \r, which MySQL doesn't.
    html = next(rows(export))["description_formatted"]
    assert html is not None
    assert "\nnext" in html and "\r" not in html


def test_rows_tells_null_from_empty(export):
    second = list(rows(export))[1]
    assert second["location"] is None
    assert second["location_canonical"] == ""


def test_record_renames_and_drops_other_columns(export):
    first = record(next(rows(export)))
    assert list(first) == list(FIELDS.values())
    assert first["id"] == "2"
    assert first["title"] == "Café Manager"


def test_rows_refuses_a_field_without_a_name(tmp_path):
    path = tmp_path / "bad.xml"
    path.write_text("<resultset><row><field>1</field></row></resultset>")
    with pytest.raises(ValueError, match="without a name"):
        list(rows(path))


def test_record_refuses_a_row_missing_a_column():
    with pytest.raises(KeyError, match="jobtitle"):
        record({"ref": "1", "location": "", "location_canonical": "",
                "description_formatted": ""})  # fmt: skip


def test_convert_writes_one_json_object_per_line(export, tmp_path):
    target = tmp_path / "out" / "ads.jsonl"
    assert convert(export, target) == 2
    lines = target.read_text(encoding="utf-8").splitlines()
    assert [json.loads(line) for line in lines] == [record(r) for r in rows(export)]


def test_convert_writes_characters_not_escapes(export, tmp_path):
    target = tmp_path / "ads.jsonl"
    convert(export, target)
    text = target.read_text(encoding="utf-8")
    assert "Café" in text and "£" in text and "✨" in text
