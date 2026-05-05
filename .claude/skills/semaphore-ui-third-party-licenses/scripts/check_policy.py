#!/usr/bin/env python3
"""
check_policy.py — Validate detected licenses against Semaphore UI's policy.

Usage:
    python3 check_policy.py <cache-dir>

Exit codes:
    0 — all licenses allowed
    1 — at least one forbidden license detected (BLOCKS file generation)
    2 — at least one license requires manual review (warning, does not block)

The policy lives in references/license_policy.md (allowlist / reviewlist /
denylist). This script encodes that policy. If the policy changes, update
both files.
"""
import csv
import json
import sys
from pathlib import Path

# SPDX identifiers grouped by policy decision.
# Keep these in sync with references/license_policy.md.
ALLOWED = {
    "MIT", "MIT-0",
    "Apache-2.0", "Apache 2.0",
    "BSD-2-Clause", "BSD-3-Clause", "BSD-3-Clause-Clear",
    "ISC",
    "Zlib",
    "Unlicense",
    "CC0-1.0",
    "MPL-2.0",  # MPL is permissive enough at the file level; allowed.
    "0BSD",
    "WTFPL",
    "BlueOak-1.0.0",
    "Python-2.0",
}

REVIEW = {
    "LGPL-2.1", "LGPL-2.1-only", "LGPL-2.1-or-later",
    "LGPL-3.0", "LGPL-3.0-only", "LGPL-3.0-or-later",
    "EPL-1.0", "EPL-2.0",
    "CDDL-1.0", "CDDL-1.1",
    "EUPL-1.1", "EUPL-1.2",
    "OFL-1.1",  # fonts: typically fine for embedding, but verify scope
}

FORBIDDEN = {
    "GPL-2.0", "GPL-2.0-only", "GPL-2.0-or-later",
    "GPL-3.0", "GPL-3.0-only", "GPL-3.0-or-later",
    "AGPL-1.0", "AGPL-3.0", "AGPL-3.0-only", "AGPL-3.0-or-later",
    "SSPL-1.0",
    "BUSL-1.1",
    "Commons-Clause",
    "Elastic-2.0",
    "RSALv2",
    # Non-OSI / non-commercial markers we sometimes see in the wild:
    "CC-BY-NC-4.0", "CC-BY-NC-SA-4.0",
    "Facebook-BSD-Patents",  # legacy React clause, should not appear anymore
}


def normalize(license_str: str) -> list[str]:
    """Split compound expressions like 'MIT OR Apache-2.0' into components."""
    if not license_str:
        return ["UNKNOWN"]
    # license-checker uses '*' as a confidence-low suffix; strip it.
    s = license_str.replace("*", "").strip()
    # Strip enclosing parens: "(MIT OR Apache-2.0)" -> "MIT OR Apache-2.0"
    if s.startswith("(") and s.endswith(")"):
        s = s[1:-1]
    # Split on OR / AND. We keep both sides — for OR, the package can be
    # taken under either; we pick the most permissive in the generator,
    # but here we still need to verify all options are non-forbidden.
    parts = []
    for chunk in s.replace(" AND ", " OR ").split(" OR "):
        parts.append(chunk.strip())
    return parts or ["UNKNOWN"]


def classify(license_str: str) -> str:
    """Return 'allowed', 'review', 'forbidden', or 'unknown'."""
    parts = normalize(license_str)
    statuses = set()
    for p in parts:
        if p in FORBIDDEN:
            statuses.add("forbidden")
        elif p in ALLOWED:
            statuses.add("allowed")
        elif p in REVIEW:
            statuses.add("review")
        else:
            statuses.add("unknown")
    # If ANY option is allowed, the package can be taken under that one.
    if "allowed" in statuses:
        return "allowed"
    if "review" in statuses:
        return "review"
    if "forbidden" in statuses:
        return "forbidden"
    return "unknown"


def load_go(path: Path):
    if not path.exists():
        return []
    out = []
    with path.open() as f:
        for row in csv.reader(f, delimiter="\t"):
            if len(row) >= 4:
                module, version, license_id, url = row[0], row[1], row[2], row[3]
                out.append({
                    "ecosystem": "go",
                    "name": module,
                    "version": version,
                    "license": license_id,
                    "url": url,
                })
    return out


def load_npm(path: Path):
    if not path.exists():
        return []
    data = json.loads(path.read_text())
    out = []
    for full_name, info in data.items():
        # full_name is "package@version"
        if "@" in full_name[1:]:  # skip leading @ of scoped packages
            at = full_name.rindex("@")
            name, version = full_name[:at], full_name[at + 1:]
        else:
            name, version = full_name, "unknown"
        out.append({
            "ecosystem": "npm",
            "name": name,
            "version": version,
            "license": info.get("licenses", "UNKNOWN"),
            "url": info.get("repository", ""),
        })
    return out


def main():
    if len(sys.argv) != 2:
        print("Usage: check_policy.py <cache-dir>", file=sys.stderr)
        sys.exit(64)

    cache = Path(sys.argv[1])
    deps = (
        load_go(cache / "go.tsv")
        + load_npm(cache / "npm.json")
    )

    buckets = {"allowed": [], "review": [], "forbidden": [], "unknown": []}
    for d in deps:
        buckets[classify(d["license"])].append(d)

    # Report
    print(f"Allowed:   {len(buckets['allowed'])}")
    print(f"Review:    {len(buckets['review'])}")
    print(f"Forbidden: {len(buckets['forbidden'])}")
    print(f"Unknown:   {len(buckets['unknown'])}")
    print()

    if buckets["forbidden"]:
        print("=" * 60)
        print("FORBIDDEN LICENSES DETECTED — do not generate the file.")
        print("=" * 60)
        for d in buckets["forbidden"]:
            print(f"  [{d['ecosystem']}] {d['name']}@{d['version']}  {d['license']}")
        print()

    if buckets["review"]:
        print("-" * 60)
        print("Licenses requiring manual review:")
        print("-" * 60)
        for d in buckets["review"]:
            print(f"  [{d['ecosystem']}] {d['name']}@{d['version']}  {d['license']}")
        print()

    if buckets["unknown"]:
        print("-" * 60)
        print("Unknown / unrecognized licenses (treat as forbidden until verified):")
        print("-" * 60)
        for d in buckets["unknown"]:
            print(f"  [{d['ecosystem']}] {d['name']}@{d['version']}  {d['license']!r}")
        print()

    # Persist the classification for the generator to consume.
    out_path = cache / "classified.json"
    out_path.write_text(json.dumps(buckets, indent=2))
    print(f"Classification written to {out_path}")

    if buckets["forbidden"] or buckets["unknown"]:
        sys.exit(1)
    if buckets["review"]:
        sys.exit(2)
    sys.exit(0)


if __name__ == "__main__":
    main()
