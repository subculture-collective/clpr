#!/usr/bin/env python3
"""Fail closed if a protected engagement component is absent or unexercised."""
import pathlib
import sys

contracts = {
    "internal/repository/engagement_repository.go": 65,
    "internal/scheduler/engagement_scheduler.go": 65,
    "internal/handlers/engagement_cursor.go": 90,
}
measured = {}
for filename in sys.argv[1:]:
    for line in pathlib.Path(filename).read_text().splitlines()[1:]:
        location, statements, count = line.split()
        source = location.split(":")[0].split("/clpr/")[-1]
        if source in contracts:
            measured[(source, location)] = (int(statements), int(count))
failures = []
for source, floor in contracts.items():
    if not pathlib.Path("backend", source).is_file():
        failures.append(f"Missing protected engagement source: {source}")
    blocks = [value for (file, _), value in measured.items() if file == source]
    total = sum(size for size, _ in blocks)
    covered = sum(size for size, count in blocks if count > 0)
    percent = 100 * covered / total if total else 0
    print(f"{source}: {percent:.1f}% statements; required {floor}%")
    if percent < floor:
        failures.append(f"Engagement contract was not sufficiently exercised: {source}")
if failures:
    raise SystemExit("\n".join(failures))
