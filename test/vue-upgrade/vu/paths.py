"""Where things live: the suite, the workspace root with project.sh, runtime/visual."""

import os
import re
from pathlib import Path

SUITE = Path(__file__).resolve().parent.parent  # <copy>/test/vue-upgrade
COPY = SUITE.parent.parent  # work/<copy>


def _find_root():
    env = os.environ.get("VU_ROOT")
    if env:
        return Path(env).resolve()
    for d in SUITE.parents:
        if (d / "project.sh").is_file() and (d / "work").is_dir():
            return d
    raise SystemExit("vu: no workspace root with project.sh above the suite — set VU_ROOT")


ROOT = _find_root()
PROJECT_SH = ROOT / "project.sh"
VISUAL = ROOT / "runtime" / "visual"
FIXTURE = VISUAL / "fixture"
SEED_DIR = SUITE / "seed"
ZONE_SEEDS = SEED_DIR / "zones"

DEFAULT_WORKSPACE = "dry_vue_upgrade"


def workspace():
    """The work copy whose stand the suite talks to: VU_WORKSPACE, else dry_vue_upgrade."""
    return os.environ.get("VU_WORKSPACE") or DEFAULT_WORKSPACE


def label_dir(label):
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]*", label):
        raise SystemExit(f"vu: label '{label}' — letters, digits, '.', '_', '-' only")
    return VISUAL / label
