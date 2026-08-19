#!/usr/bin/env bash
# collect_licenses.sh — Collect raw license data from Go and npm dependencies.
#
# Outputs:
#   .licenses-cache/go.tsv    — TSV: <module>\t<version>\t<license>\t<url>
#   .licenses-cache/npm.json  — license-checker JSON output
#
# Run from the repository root.

set -euo pipefail

CACHE_DIR=".licenses-cache"
mkdir -p "$CACHE_DIR"

# ----------------------------------------------------------------------------
# Go backend
# ----------------------------------------------------------------------------
if [[ -f "go.mod" ]]; then
  echo "==> Collecting Go module licenses"

  if ! command -v go-licenses >/dev/null 2>&1; then
    cat <<'EOF' >&2
go-licenses is not installed. Install it with:

  go install github.com/google/go-licenses@latest

Then ensure $(go env GOPATH)/bin is in your PATH.

Alternative tools:
  - github.com/CycloneDX/cyclonedx-gomod  (produces CycloneDX SBOM)
  - github.com/elastic/go-licence-detector (has SPDX detection)
EOF
    exit 1
  fi

  # Disable workspace mode so go-licenses only sees the root module's
  # dependency graph, not sibling workspace modules (e.g. pro_impl).
  export GOWORK=off

  # Discover the root module path so we can filter out first-party packages
  # (the root module itself plus any local `replace` targets that resolve to
  # paths inside this repo, like `pro/`).
  ROOT_MODULE="$(awk '/^module / { print $2; exit }' go.mod)"

  # Output format: <module>\t<url>\t<license>
  # We append our own version column from `go list -m`.
  go-licenses report ./... \
    --template "$(dirname "$0")/go_template.tpl" \
    > "$CACHE_DIR/go-raw.tsv" 2>"$CACHE_DIR/go-errors.log" || {
      echo "go-licenses reported issues. See $CACHE_DIR/go-errors.log" >&2
      echo "Continuing with partial data — review the errors before shipping." >&2
    }

  # Enrich with versions from `go list -m all`
  go list -m -f '{{.Path}}|{{.Version}}' all 2>/dev/null \
    | awk -F'|' '{ versions[$1] = $2 } END { for (m in versions) print m"\t"versions[m] }' \
    > "$CACHE_DIR/go-versions.tsv"

  # Join, roll up sub-packages to parent modules, drop first-party rows,
  # and apply manual overrides for packages with detection failures.
  ROOT_MODULE="$ROOT_MODULE" python3 - "$CACHE_DIR" <<'PY'
import csv
import os
import sys
from pathlib import Path

cache = Path(sys.argv[1])
root_module = os.environ["ROOT_MODULE"]

# Manual overrides for packages where go-licenses can't detect a license
# (typically because the LICENSE file uses an unusual layout). Keyed by
# module path. Extend this when new "Unknown" rows appear.
OVERRIDES = {
    "modernc.org/mathutil": {
        "license": "BSD-3-Clause",
        "url": "https://gitlab.com/cznic/mathutil/-/blob/master/LICENSE",
    },
}

# Module → version map from `go list -m all`.
versions = {}
with (cache / "go-versions.tsv").open() as f:
    for line in f:
        parts = line.rstrip("\n").split("\t")
        if len(parts) == 2:
            versions[parts[0]] = parts[1]

# Resolve a package path to its declared module (longest matching prefix).
sorted_modules = sorted(versions.keys(), key=len, reverse=True)
def find_module(pkg):
    for m in sorted_modules:
        if pkg == m or pkg.startswith(m + "/"):
            return m
    return None

def is_first_party(module):
    return module == root_module or module.startswith(root_module + "/")

# Read raw rows, roll up to parent module, dedupe, and apply overrides.
seen = {}  # module -> (version, license, url)
with (cache / "go-raw.tsv").open() as f:
    for row in csv.reader(f, delimiter="\t"):
        if len(row) < 3:
            continue
        pkg, url, license_id = row[0], row[1], row[2]
        module = find_module(pkg) or pkg
        if is_first_party(module):
            continue
        version = versions.get(module, "unknown")
        # Prefer the row whose package path == the module root (its LICENSE
        # is the canonical one); otherwise keep the first hit.
        if module not in seen or pkg == module:
            seen[module] = (version, license_id, url)

# Apply overrides last so they always win.
for module, fix in OVERRIDES.items():
    if module in seen:
        ver, lic, url = seen[module]
        seen[module] = (
            ver,
            fix.get("license", lic),
            fix.get("url", url),
        )

with (cache / "go.tsv").open("w") as f:
    w = csv.writer(f, delimiter="\t")
    for module in sorted(seen):
        ver, lic, url = seen[module]
        w.writerow([module, ver, lic, url])

print(f"  ({len(seen)} third-party modules)")
PY
  echo "    -> $CACHE_DIR/go.tsv"
else
  echo "==> No go.mod found at repo root, skipping Go collection"
fi

# ----------------------------------------------------------------------------
# Frontend (npm)
# ----------------------------------------------------------------------------
FRONTEND_DIR=""
for candidate in "web" "web2" "frontend" "ui" "."; do
  if [[ -f "$candidate/package.json" && "$candidate" != "." ]]; then
    FRONTEND_DIR="$candidate"
    break
  fi
done

if [[ -n "$FRONTEND_DIR" ]]; then
  echo "==> Collecting npm licenses from $FRONTEND_DIR"

  if ! command -v license-checker >/dev/null 2>&1; then
    cat <<'EOF' >&2
license-checker is not installed. Install it with:

  npm install -g license-checker

Or use it without global install:

  npx license-checker --production --json
EOF
    exit 1
  fi

  (cd "$FRONTEND_DIR" && license-checker --production --json --excludePrivatePackages) \
    > "$CACHE_DIR/npm.json"

  # Strip license-checker's low-confidence "*" suffix from license strings.
  # The marker is informational and not part of any SPDX identifier; leaving
  # it in produces ugly distinct buckets like "MIT" vs "MIT*".
  python3 - "$CACHE_DIR" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1]) / "npm.json"
data = json.loads(path.read_text())
for v in data.values():
    lic = v.get("licenses")
    if isinstance(lic, str) and lic.endswith("*"):
        v["licenses"] = lic.rstrip("*")
path.write_text(json.dumps(data, indent=2))
PY

  count=$(python3 -c "import json; print(len(json.load(open('$CACHE_DIR/npm.json'))))")
  echo "    -> $CACHE_DIR/npm.json ($count packages)"
else
  echo "==> No frontend package.json found, skipping npm collection"
fi

echo ""
echo "Done. Run scripts/check_policy.py $CACHE_DIR/ next."