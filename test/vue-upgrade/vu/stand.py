"""The stand of a work copy, driven through ./project.sh -w <copy> stand …"""

import json
import os
import subprocess
from dataclasses import dataclass
from pathlib import Path

from .paths import PROJECT_SH, ROOT


def project_sh(ws, *args, env=None, check=True, capture=False):
    cmd = [str(PROJECT_SH), "-w", ws, *args]
    full_env = dict(os.environ, **(env or {}))
    if capture:
        res = subprocess.run(cmd, env=full_env, text=True, capture_output=True)
        if check and res.returncode != 0:
            raise SystemExit(f"vu: {' '.join(cmd[1:])} failed:\n{res.stdout}{res.stderr}")
        return res
    print(f"$ ./project.sh -w {ws} {' '.join(args)}", flush=True)
    res = subprocess.run(cmd, env=full_env)
    if check and res.returncode != 0:
        raise SystemExit(f"vu: ./project.sh -w {ws} {' '.join(args)} failed ({res.returncode})")
    return res


@dataclass
class Stand:
    workspace: str
    url: str  # backend of the slot: the built binary, embedded UI
    web_url: str  # hot-reload UI (stand web), when it runs
    instance: Path

    @property
    def config(self):
        return self.instance / "config.json"

    @property
    def db(self):
        return self.instance / "database.sqlite"


def env(ws):
    """Values of `stand env`; None when the stand is not initialized."""
    res = project_sh(ws, "stand", "env", capture=True, check=False)
    if res.returncode != 0:
        return None
    vals = dict(line.split("=", 1) for line in res.stdout.splitlines() if "=" in line)
    return Stand(ws, vals["SEMAPHORE_STAND_URL"], vals["SEMAPHORE_STAND_WEB_URL"],
                 ROOT / "runtime" / "instances" / ws)


def is_up(stand):
    import requests
    try:
        return requests.get(stand.url + "/api/ping", timeout=2).ok
    except requests.RequestException:
        return False


def slot(ws):
    meta = ROOT / "runtime" / "instances" / ws / "instance.env"
    if meta.is_file():
        for line in meta.read_text().splitlines():
            if line.startswith("SLOT="):
                return line[5:].strip()
    return None


def read_config(stand):
    return json.loads(stand.config.read_text())
