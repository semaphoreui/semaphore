#!/usr/bin/env bash
# vue-upgrade browser suite: fixture, zone seeds, runs, comparison — see README.md.
#   ./vu.sh fixture build | fixture restore <copy> | seed zone <zone> | run <label> [pytest args] | compare <a> <b>
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
command -v uv > /dev/null || { echo "vu: no uv on PATH" >&2; exit 1; }
cd "$here"
exec uv run --quiet --project "$here" python -m vu "$@"
