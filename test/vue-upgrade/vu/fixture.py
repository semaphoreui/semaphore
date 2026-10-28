"""The frozen fixture: build it on a fresh stand, restore it into the stand of any work copy.

runtime/visual/fixture/
  database.sqlite  — the seeded database, one self-contained file (no -wal)
  config.json      — the config of the stand that seeded it: the encryption keys of the secrets
  meta.json        — seed time, browser clock, ids of projects and users, zones seeded
  repos/           — the local git repositories the seed templates use (fixed commit hashes)
"""

import datetime as dt
import json
import os
import shutil
import subprocess
import sqlite3
import time
from pathlib import Path

from . import stand as st
from .paths import FIXTURE, ROOT

CLOCK_OFFSET = dt.timedelta(hours=2)  # browser clock = end of the seed + 2 h: tasks are "2 hours ago"
PLAN = "enterprise"


def _copy_dir(ws):
    return ROOT / "work" / ws


def _check_binary(ws):
    copy = _copy_dir(ws)
    binary = copy / "bin" / "semaphore-pro"
    if not binary.is_file():
        raise SystemExit(f"vu: no {binary} — ./project.sh -w {ws} build pro")
    built = binary.stat().st_mtime
    newer = [p for p in (copy / "web" / "src").rglob("*") if p.is_file() and p.stat().st_mtime > built]
    if newer:
        print(f"vu: ⚠ {len(newer)} file(s) in web/src are newer than {binary.name} "
              f"(e.g. {newer[0].relative_to(copy)}) — the snapshots show the old build; "
              f"./project.sh -w {ws} build pro", flush=True)
    return binary


def _wait_up(stand, timeout=60):
    deadline = time.time() + timeout
    while time.time() < deadline:
        if st.is_up(stand):
            return
        time.sleep(0.5)
    raise SystemExit(f"vu: stand {stand.workspace} does not answer at {stand.url}")


def build(ws, with_zones=True):
    """Fresh database on the stand of <ws>, Enterprise, admin Pro seat, base set and zone seeds;
    then stop the stand and freeze the database, the config and meta.json into runtime/visual/fixture/."""
    from seed import base, zone_module, zone_names
    from seed.ctx import Ctx

    _check_binary(ws)
    slot = st.slot(ws)
    st.project_sh(ws, "stand", "stop", check=False)
    st.project_sh(ws, "stand", "reset")
    st.project_sh(ws, "stand", "init", env={"STAND_SLOT": slot} if slot else None)
    _claim_port(st.env(ws))
    st.project_sh(ws, "stand", "start")
    stand = st.env(ws)
    _wait_up(stand)
    st.project_sh(ws, "stand", "license", PLAN)

    if FIXTURE.exists():
        shutil.rmtree(FIXTURE)
    FIXTURE.mkdir(parents=True)
    os.chmod(FIXTURE, 0o700)

    started = dt.datetime.now(dt.timezone.utc)
    ctx = Ctx(stand.url, FIXTURE / "repos")
    print("== base set", flush=True)
    base.seed(ctx)
    zones = zone_names() if with_zones else []
    for z in zones:
        print(f"== zone {z}", flush=True)
        zone_module(z).seed(ctx)
    seeded = dt.datetime.now(dt.timezone.utc).replace(microsecond=0)

    meta = {
        "format": 1,
        "workspace": ws,
        "plan": PLAN,
        "seed_started": started.replace(microsecond=0).isoformat(),
        "seeded_at": seeded.isoformat(),
        "clock": (seeded + CLOCK_OFFSET).replace(second=0).isoformat(),
        "web_host": st.read_config(stand)["web_host"],
        "zones": zones,
        "users": {u["username"]: {"id": u["id"], "admin": u["admin"],
                                  "password": ctx.admin_password if u["username"] == "admin" else ctx.password}
                  for u in ctx.admin.get("/users")},
        "projects": {p["name"]: p["id"] for p in ctx.admin.get("/projects")},
        "version": ctx.admin.get("/info")["version"],
    }

    st.project_sh(ws, "stand", "stop")
    # one self-contained file: the backup API folds the WAL in
    src = sqlite3.connect(f"file:{stand.db}?mode=ro", uri=True)
    dst = sqlite3.connect(FIXTURE / "database.sqlite")
    src.backup(dst)
    dst.execute("PRAGMA journal_mode=DELETE")
    dst.close()
    src.close()
    shutil.copyfile(stand.config, FIXTURE / "config.json")
    os.chmod(FIXTURE / "config.json", 0o600)
    (FIXTURE / "meta.json").write_text(json.dumps(meta, indent=2) + "\n")
    print(f"vu: fixture frozen in {FIXTURE} (seeded {meta['seeded_at']}, browser clock {meta['clock']}, "
          f"zones: {', '.join(zones) or '—'})", flush=True)


def meta():
    path = FIXTURE / "meta.json"
    if not path.is_file():
        raise SystemExit(f"vu: no fixture in {FIXTURE} — ./vu.sh fixture build")
    return json.loads(path.read_text())


def install(inst, port, tmp_name, m):
    """The frozen database and config into the folder <inst>: port <port>, its own DB and secrets
    paths, task temp /tmp/semaphore-<uid>/<tmp_name>; the fixture's keys and web_host."""
    cfg = json.loads((FIXTURE / "config.json").read_text())
    cfg["sqlite"] = {"host": str(inst / "database.sqlite")}
    cfg["port"] = f":{port}"
    # web_host stays the fixture's: the UI prints it (runner setup, raw log link, API docs link),
    # so every copy shows the same text; same-origin checks also accept the request host
    cfg["web_host"] = m["web_host"]
    # task temp and ssh-agent sockets: a unix socket path is limited to 107 bytes, so they live in
    # the short private dir of utils/common.sh (short_tmp_dir), /tmp/semaphore-<uid>/<name>
    task_tmp = Path(f"/tmp/semaphore-{os.getuid()}") / tmp_name
    task_tmp.mkdir(mode=0o700, parents=True, exist_ok=True)
    os.chmod(task_tmp.parent, 0o700)
    os.chmod(task_tmp, 0o700)
    cfg["tmp_path"] = str(task_tmp)
    cfg.setdefault("dirs", {})
    cfg["dirs"]["secrets"] = str(inst / "tmp")
    cfg["dirs"]["ssh_agent_sockets"] = str(task_tmp)
    (inst / "tmp").mkdir(parents=True, exist_ok=True)
    for suffix in ("", "-wal", "-shm", "-journal"):
        p = inst / f"database.sqlite{suffix}"
        if p.exists():
            p.unlink()
    shutil.copyfile(FIXTURE / "database.sqlite", inst / "database.sqlite")
    tmp = inst / "config.json.tmp"
    fd = os.open(tmp, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, "w") as f:
        json.dump(cfg, f, indent=2)
        f.write("\n")
    os.replace(tmp, inst / "config.json")


def _claim_port(stand):
    """Before `stand start`: the stand's port must be free — `stand start` takes any server answering
    there for the stand ("already up"). The at-baseline server of the suite gives it up; anything
    else — refuse."""
    from . import baseline

    port = st.port_of(stand.url)
    if baseline.owns(port):
        baseline.stop()
    busy, _ = st.listeners(port)
    if busy:
        raise SystemExit(f"vu: {stand.url} is taken by {st.describe(port)}, not by the stand "
                         f"{stand.workspace} — stop it first")


def restore(ws, start=True):
    """Put the frozen database and config into the stand of <ws> (its own slot and paths, the
    fixture's keys and web_host) and start it with the built binary."""
    m = meta()
    _check_binary(ws)
    if st.env(ws) is None:
        st.project_sh(ws, "stand", "init")
    # `stand stop` also ends `stand web` (the hot-reload UI on port +1): note it to bring it back
    web_was_up = _web_up(ws)
    st.project_sh(ws, "stand", "stop", check=False)
    stand = st.env(ws)
    inst = stand.instance
    install(inst, st.port_of(stand.url), ws, m)
    token = inst / "api.token"  # an admin token of the old database: stand license makes a new one
    if token.exists():
        token.unlink()
    print(f"vu: fixture ({m['seeded_at']}) restored into {inst}", flush=True)
    if start:
        _claim_port(stand)
        st.project_sh(ws, "stand", "start")
        _wait_up(stand)
        if web_was_up:
            _start_web(ws)
    return stand


def _web_up(ws):
    pid_file = ROOT / "runtime" / "instances" / ws / "web.pid"
    try:
        os.kill(int(pid_file.read_text().strip()), 0)
        return True
    except (OSError, ValueError):
        return False


def _start_web(ws):
    """Relaunch `stand web` detached (it runs in the foreground), log in the instance's logs/."""
    log = ROOT / "runtime" / "instances" / ws / "logs" / "stand-web.out"
    log.parent.mkdir(parents=True, exist_ok=True)
    with open(log, "ab") as out:
        subprocess.Popen([str(ROOT / "project.sh"), "-w", ws, "stand", "web"], cwd=ROOT, stdout=out,
                         stderr=subprocess.STDOUT, stdin=subprocess.DEVNULL, start_new_session=True)
    print(f"vu: stand web of {ws} relaunched (was running before the restore)", flush=True)
