"""Converts job ads exported by MySQL's --xml option to JSON Lines.

Each line of the output is one ad, with the lab's field names: id, title,
location, location_canonical and html. The work is in jdclab.mysql_xml,
which python/tests/test_mysql_xml.py tests.

From the lab's folder, with direnv on (it puts python/lib on PYTHONPATH):

    python python/bin/xml_to_jsonl.py data/extract.xml data/ads.jsonl
"""

import argparse
from pathlib import Path

from jdclab.mysql_xml import convert


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("source", type=Path, help="the export, as XML")
    parser.add_argument("target", type=Path, help="where to write the JSON Lines")
    args = parser.parse_args()
    count = convert(args.source, args.target)
    print(f"{count} ads written to {args.target}")


if __name__ == "__main__":
    main()
