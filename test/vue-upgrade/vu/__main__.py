"""vu — data, stand, runs and comparison of the vue-upgrade browser suite. Call through ./vu.sh.

  fixture build [-w <copy>] [--no-zones]   fresh DB on the copy's stand, Enterprise, seed, freeze
  fixture restore [<copy>] [--no-start] [--at-baseline]
                                            frozen DB into the copy's stand, start the built binary;
                                            --at-baseline: the copy's binary on the fixture's web_host
  fixture stop                              stop the at-baseline server, free the baseline address
  fixture info                              what the frozen fixture holds, who serves the baseline
  seed zone <zone>… [-w <copy>] [--at-baseline]
                                            add a zone's data to the running stand (idempotent)
  run <label> [-w <copy>] [--at-baseline] [--no-restore] [--keep] [pytest args]
                                            restore, run the suite, shots in runtime/visual/<label>/
  compare <label-a> <label-b> [--threshold T]
                                            pixels + aria + inventory, report in runtime/visual/compare/
"""

import argparse
import json
import os
import shutil
import subprocess
import sys

from . import baseline, fixture
from . import stand as st
from .paths import FIXTURE, SUITE, label_dir, workspace


def cmd_fixture(args):
    if args.action == "build":
        fixture.build(args.workspace, with_zones=not args.no_zones)
    elif args.action == "restore":
        restore = baseline.restore if args.at_baseline else fixture.restore
        restore(args.target or args.workspace, start=not args.no_start)
    elif args.action == "stop":
        baseline.stop()
    else:
        m = fixture.meta()
        print(json.dumps({k: v for k, v in m.items() if k != "users"}, indent=2))
        print(f"users: {', '.join(m['users'])}\nfiles: {FIXTURE}")
        s = baseline.state()
        print(f"at-baseline: {s['workspace']} ({s['version']}, pid {s['pid']}) at {s['url']}" if s else
              f"at-baseline: — ({m['web_host']}: {st.describe(st.port_of(m['web_host']))})")


def _stand(ws, at_baseline):
    """The running stand the suite talks to: the copy's own, or its at-baseline server."""
    if at_baseline:
        return baseline.stand(ws)
    stand = st.env(ws)
    if stand is None or not st.is_up(stand):
        raise SystemExit(f"vu: stand {ws} is not running — ./vu.sh fixture restore {ws}")
    return stand


def cmd_seed(args):
    from seed import zone_module
    from seed.ctx import Ctx

    stand = _stand(args.workspace, args.at_baseline)
    ctx = Ctx(stand.url, FIXTURE / "repos")
    for z in args.zones:
        mod = zone_module(z)
        print(f"== zone {z} -> {stand.url}", flush=True)
        mod.seed(ctx)


def cmd_run(args, pytest_args):
    ws = args.workspace
    if not args.no_restore:
        (baseline.restore if args.at_baseline else fixture.restore)(ws)
    stand = _stand(ws, args.at_baseline)
    print(f"vu: testing {ws} at {stand.url}", flush=True)
    out = label_dir(args.label)
    if out.exists() and not args.keep:
        shutil.rmtree(out)
    out.mkdir(parents=True, exist_ok=True)
    env = dict(os.environ, VU_WORKSPACE=ws)
    if args.at_baseline:
        env["VU_AT_BASELINE"] = "1"  # conftest: the stand is the at-baseline server of <ws>
    base = [sys.executable, "-m", "pytest", "--label", args.label, "--workspace", ws]
    workers = args.parallel
    if workers is None:
        # a whole-suite run is parallel by default; a narrowed one (-k, -m, a path) stays sequential
        narrowed = any(a in ("-k", "-m") or a.startswith(("-k", "-m", "tests")) for a in pytest_args)
        workers = 0 if narrowed else int(os.environ.get("VU_PARALLEL", "4"))
    if workers <= 1:
        cmd = [*base, *pytest_args]
        print(f"$ {' '.join(cmd[1:])}", flush=True)
        return subprocess.run(cmd, cwd=SUITE, env=env).returncode
    # snapshot tests only read data (writes are blocked), so they run in workers, one zone file per
    # worker (--dist loadfile keeps a file's order); functional tests change data — after, in one process
    junit = [a for a in pytest_args if a.startswith("--junitxml")]
    rest = [a for a in pytest_args if not a.startswith("--junitxml")]
    def with_junit(suffix):
        return [f"{a}.{suffix}.xml" if a.endswith(".xml") else a for a in junit]
    steps = [
        [*base, "-m", "snapshot", "-n", str(workers), "--dist", "loadfile", *with_junit("snapshot"), *rest],
        [*base, "-m", "functional", *with_junit("functional"), *rest],
    ]
    code = 0
    for cmd in steps:
        print(f"$ {' '.join(cmd[1:])}", flush=True)
        rc = subprocess.run(cmd, cwd=SUITE, env=env).returncode
        code = code or (0 if rc == 5 else rc)  # 5: no tests collected for this marker
    return code


def main(argv=None):
    ap = argparse.ArgumentParser(prog="vu.sh", description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)

    def ws_opt(p):
        p.add_argument("-w", "--workspace", default=workspace(), help="work copy (default: %(default)s)")

    def baseline_opt(p):
        p.add_argument("--at-baseline", action="store_true",
                       help="serve the copy's built binary on the fixture's web_host (the address of "
                            "`before`), outside the copy's stand; refused while a stand listens there")

    f = sub.add_parser("fixture", help="build / restore / stop / info of the frozen fixture")
    f.add_argument("action", choices=["build", "restore", "stop", "info"])
    f.add_argument("target", nargs="?", help="restore: the work copy (default: -w)")
    f.add_argument("--no-zones", action="store_true", help="build: base set only")
    f.add_argument("--no-start", action="store_true", help="restore: do not start the stand")
    ws_opt(f)
    baseline_opt(f)

    s = sub.add_parser("seed", help="seed zone <zone>…: add zone data to the running stand")
    s.add_argument("what", choices=["zone"])
    s.add_argument("zones", nargs="+")
    ws_opt(s)
    baseline_opt(s)

    r = sub.add_parser("run", help="run the suite against the stand; extra args go to pytest")
    r.add_argument("label")
    r.add_argument("--no-restore", action="store_true", help="keep the stand's current data")
    r.add_argument("--keep", action="store_true", help="keep earlier shots of the label")
    r.add_argument("--parallel", type=int, default=None, metavar="N",
                   help="workers for snapshot tests (default: VU_PARALLEL or 4 for a whole-suite run, "
                        "sequential when -k/-m/a path narrows it; 0 or 1 — sequential)")
    ws_opt(r)
    baseline_opt(r)

    c = sub.add_parser("compare", help="compare two runs")
    c.add_argument("a")
    c.add_argument("b")
    c.add_argument("--threshold", type=float, default=0.1, help="colour tolerance 0..1 (pixelmatch)")

    args, rest = ap.parse_known_args(argv)
    if rest and args.cmd != "run":
        ap.error(f"unrecognized arguments: {' '.join(rest)}")
    if args.cmd == "fixture" and args.at_baseline and args.action != "restore":
        ap.error("--at-baseline goes with fixture restore only")
    if args.cmd == "fixture":
        cmd_fixture(args)
    elif args.cmd == "seed":
        cmd_seed(args)
    elif args.cmd == "run":
        return cmd_run(args, rest)
    elif args.cmd == "compare":
        from . import compare
        return compare.run(args.a, args.b, args.threshold)
    return 0


if __name__ == "__main__":
    sys.exit(main())
