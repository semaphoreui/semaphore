#!/usr/bin/env python3
"""
generate_md.py — Render THIRD-PARTY-LICENSES.md from classified license data.

Usage:
    python3 generate_md.py <cache-dir> > THIRD-PARTY-LICENSES.md

Reads <cache-dir>/classified.json (produced by check_policy.py) and emits
the human-readable attribution document.
"""
import json
import sys
from collections import Counter, defaultdict
from datetime import datetime, timezone
from pathlib import Path


def section_header(title: str, level: int = 2) -> str:
    return f"{'#' * level} {title}\n"


def render_summary(deps_by_ecosystem: dict) -> str:
    lines = ["## Summary\n"]
    total = sum(len(v) for v in deps_by_ecosystem.values())
    lines.append(f"This document lists **{total}** third-party components ")
    lines.append("distributed with Semaphore UI, grouped by ecosystem.\n\n")

    lines.append("| Ecosystem | Components |\n")
    lines.append("|-----------|------------|\n")
    for eco, items in sorted(deps_by_ecosystem.items()):
        eco_label = {
            "go": "Go (backend)",
            "npm": "npm (frontend)",
        }.get(eco, eco)
        lines.append(f"| {eco_label} | {len(items)} |\n")
    lines.append("\n")

    # License distribution
    licenses_count = Counter()
    for items in deps_by_ecosystem.values():
        for d in items:
            licenses_count[d["license"]] += 1
    lines.append("### License distribution\n\n")
    lines.append("| License | Count |\n")
    lines.append("|---------|-------|\n")
    for lic, n in licenses_count.most_common():
        lines.append(f"| {lic} | {n} |\n")
    lines.append("\n")
    return "".join(lines)


def render_ecosystem(title: str, intro: str, items: list) -> str:
    out = [section_header(title)]
    out.append(intro + "\n\n")

    if not items:
        out.append("_No components in this category._\n\n")
        return "".join(out)

    # Deduplicate by name, collapsing versions.
    by_name = defaultdict(list)
    for d in items:
        by_name[d["name"]].append(d)

    out.append("| Component | Version(s) | License | Source |\n")
    out.append("|-----------|------------|---------|--------|\n")
    for name in sorted(by_name.keys(), key=str.lower):
        entries = by_name[name]
        versions = ", ".join(sorted({e["version"] for e in entries}))
        # All entries for the same name should have the same license; if not,
        # join them — that's a signal worth showing in the file.
        licenses = sorted({e["license"] for e in entries})
        license_str = " / ".join(licenses)
        url = entries[0].get("url", "") or ""
        url_md = f"[link]({url})" if url else "—"
        out.append(f"| `{name}` | {versions} | {license_str} | {url_md} |\n")
    out.append("\n")
    return "".join(out)


def main():
    if len(sys.argv) != 2:
        print("Usage: generate_md.py <cache-dir>", file=sys.stderr)
        sys.exit(64)

    cache = Path(sys.argv[1])
    classified_path = cache / "classified.json"
    if not classified_path.exists():
        print(f"Missing {classified_path}. Run check_policy.py first.", file=sys.stderr)
        sys.exit(1)

    classified = json.loads(classified_path.read_text())

    # If anything is forbidden or unknown, refuse — the policy check already
    # exited non-zero, but guard here too in case the user piped past it.
    if classified.get("forbidden") or classified.get("unknown"):
        print(
            "Refusing to generate: forbidden or unknown licenses present. "
            "Re-run check_policy.py and resolve them first.",
            file=sys.stderr,
        )
        sys.exit(1)

    # Combine allowed + review (review is acceptable, just flagged).
    all_deps = classified.get("allowed", []) + classified.get("review", [])

    by_ecosystem = defaultdict(list)
    for d in all_deps:
        by_ecosystem[d["ecosystem"]].append(d)

    timestamp = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")

    # Header
    out = []
    out.append("# Third-Party Licenses\n\n")
    out.append(
        "Semaphore UI is built on the work of many open-source projects. "
        "This document identifies every third-party component distributed "
        "with Semaphore UI, in compliance with the attribution requirements "
        "of the respective licenses and with §3.6 of our Master Service "
        "Agreement (identification of open-source components by name, "
        "version, and license type).\n\n"
    )
    out.append(f"_Generated on **{timestamp}** by `scripts/collect_licenses.sh`._\n\n")
    out.append(
        "To regenerate this file, run:\n\n"
        "```bash\n"
        "scripts/collect_licenses.sh\n"
        "scripts/check_policy.py .licenses-cache/\n"
        "scripts/generate_md.py .licenses-cache/ > THIRD-PARTY-LICENSES.md\n"
        "```\n\n"
    )

    # Summary
    out.append(render_summary(by_ecosystem))

    # Backend
    out.append(render_ecosystem(
        "Go Backend Dependencies",
        "Modules statically linked into the Semaphore UI server binary. "
        "Sourced from `go.mod` (production dependencies only).",
        by_ecosystem.get("go", []),
    ))

    # Frontend
    out.append(render_ecosystem(
        "Frontend Dependencies (npm)",
        "Packages bundled into the web UI assets, which are embedded in "
        "the server binary at build time. Sourced from the frontend "
        "`package.json` (production dependencies only; dev dependencies "
        "are not distributed).",
        by_ecosystem.get("npm", []),
    ))

    # Footer
    out.append("---\n\n")
    out.append("## License texts\n\n")
    out.append(
        "Full license texts for each component are available at the source "
        "URLs listed above. For permissively-licensed packages (MIT, BSD, "
        "ISC, Apache-2.0), the original LICENSE and NOTICE files are "
        "preserved in their respective package directories within the "
        "Semaphore UI distribution.\n\n"
    )
    out.append(
        "If you believe a component is missing from this list or "
        "incorrectly attributed, please open an issue at "
        "https://github.com/semaphoreui/semaphore/issues.\n\n"
    )
    out.append("<!-- end of generated file -->\n")

    sys.stdout.write("".join(out))


if __name__ == "__main__":
    main()
