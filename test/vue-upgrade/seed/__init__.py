"""Fixture data: the base set (seed/base.py) and the zone seeds (seed/zones/<zone>.py)."""

import importlib
import pkgutil
from pathlib import Path

ZONES_DIR = Path(__file__).resolve().parent / "zones"


def zone_names():
    """Zone seed modules, sorted: seed/zones/<zone>.py with a seed(ctx) function."""
    return sorted(m.name for m in pkgutil.iter_modules([str(ZONES_DIR)]) if not m.name.startswith("_"))


def zone_module(name):
    if name not in zone_names():
        raise SystemExit(f"vu: no zone seed seed/zones/{name}.py; zones: {', '.join(zone_names()) or '—'}")
    mod = importlib.import_module(f"seed.zones.{name}")
    if not callable(getattr(mod, "seed", None)):
        raise SystemExit(f"vu: seed/zones/{name}.py has no seed(ctx) function")
    return mod
