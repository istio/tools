# Usage will likely be `python results_to_csv.py ---`
import csv
from typing import Set, List, Dict
import sys

TEST_RUN_SEPARATOR=sys.argv[1]

fieldnames: Set[str] = set()
rows: List[Dict[str,str]] = []
row: Dict[str, str] = dict()

for line in sys.stdin:
    line = line.strip()
    if line.strip() == TEST_RUN_SEPARATOR:
        fieldnames.update(row.keys())
        rows.append(row)
        row = dict()
        continue

    line = line.split("=")
    if (len(line) != 2):
        continue

    row[line[0]] = line[1]

writer = csv.DictWriter(sys.stdout, fieldnames=fieldnames)

writer.writeheader()
for row in rows:
    writer.writerow(row)



