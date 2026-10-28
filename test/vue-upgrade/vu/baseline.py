"""The baseline address: the built binary of any work copy serving the frozen fixture at the address
`before` was shot on — the fixture's web_host (meta.json, http://localhost:3100).

The config keeps the fixture's web_host (the UI prints it), and the server writes web_host into
<base href> of index.html: a copy's own stand on its slot port makes the browser load the frontend and
the API from the baseline address — another app. --at-baseline serves the copy's binary on that
address itself, as a server of the suite in runtime/visual/_at-baseline/ (config, DB, task secrets, log,
server.json), outside the copy's stand: its slot, `stand start` and `stand web` are not touched.

The server takes the identity of the baseline stand except the binary: port and web_host of the
fixture, task temp of the stand that seeded it (/tmp/semaphore-<uid>/<meta workspace>: System info and
task logs print it). One server owns the address: while a stand listens there the command refuses;
`fixture restore` / `fixture build` of the baseline copy stop this server first.
"""

import json
import os
import signal
import subprocess
import time

from . import fixture
from . import stand as st
from .paths import ROOT, VISUAL

DIR = VISUAL / "_at-baseline"  # not a valid label name: no run clears it
STATE = DIR / "server.json"  # {workspace, binary, version, pid, url}


def url():
    return fixture.meta()["web_host"].rstrip("/")


def _alive(pid):
    try:
        os.kill(pid, 0)
        return True
    except OSError:
        return False


def state():
    """server.json of the running at-baseline server, None when it is not running."""
    try:
        s = json.loads(STATE.read_text())
    except (OSError, ValueError):
        return None
    return s if _alive(s.get("pid", 0)) else None


def owns(port):
    s = state()
    return bool(s) and st.port_of(s["url"]) == port


def stop(quiet=False):
    s = state()
    if s:
        os.kill(s["pid"], signal.SIGTERM)
        for _ in range(40):
            if not _alive(s["pid"]):
                break
            time.sleep(0.25)
        else:
            os.kill(s["pid"], signal.SIGKILL)
        print(f"vu: at-baseline server of {s['workspace']} (pid {s['pid']}) stopped, {s['url']} is free",
              flush=True)
    elif not quiet:
        print("vu: no at-baseline server running", flush=True)
    STATE.unlink(missing_ok=True)


def own_stand(ws):
    """The copy's own stand when it already lives on the baseline address (the copy that seeded the
    fixture): --at-baseline is then the plain restore."""
    s = st.env(ws)
    return s if s is not None and s.url.rstrip("/") == url() else None


def stand(ws):
    """The Stand the tests talk to in --at-baseline mode; SystemExit when <ws> is not served there."""
    own = own_stand(ws)
    if own is not None:
        return own
    s = state()
    if s is None or s["workspace"] != ws:
        who = f"the binary of {s['workspace']}" if s else st.describe(st.port_of(url()))
        raise SystemExit(f"vu: {url()} does not serve {ws} (now: {who}) — "
                         f"./vu.sh fixture restore {ws} --at-baseline")
    s = st.Stand(ws, s["url"], "", DIR)
    if not st.is_up(s):
        raise SystemExit(f"vu: the at-baseline server of {ws} does not answer at {s.url}, "
                         f"see {DIR / 'logs' / 'server.log'}")
    return s


def restore(ws, start=True):
    """The frozen fixture into runtime/visual/_at-baseline/, served by the built binary of <ws> on the
    baseline address. Refuses while a stand listens there."""
    m = fixture.meta()
    if own_stand(ws) is not None:
        print(f"vu: {ws} is the baseline stand itself ({url()}) — plain restore", flush=True)
        return fixture.restore(ws, start)
    binary = fixture._check_binary(ws)
    base = url()
    port = st.port_of(base)
    stop(quiet=True)
    busy, _ = st.listeners(port)
    if busy:
        raise SystemExit(f"vu: {base} is taken by {st.describe(port)} — one server per address; free it "
                         f"first (the baseline stand: ./project.sh -w {m['workspace']} stand stop; "
                         f"back to it: ./vu.sh fixture restore {m['workspace']})")
    DIR.mkdir(parents=True, exist_ok=True)
    os.chmod(DIR, 0o700)
    fixture.install(DIR, port, m["workspace"], m)
    (DIR / "logs").mkdir(exist_ok=True)
    print(f"vu: fixture ({m['seeded_at']}) restored into {DIR} for {ws} at {base}", flush=True)
    if start:
        _start(ws, binary, base)
    return st.Stand(ws, base, "", DIR)


def _start(ws, binary, base):
    cfg = DIR / "config.json"
    log = DIR / "logs" / "server.log"
    env = dict(os.environ)
    apps = str(ROOT / "runtime" / "toolchain" / "apps" / "bin")  # apps_env of utils/common.sh
    if apps not in env.get("PATH", "").split(":"):
        env["PATH"] = f"{apps}:{env.get('PATH', '')}"
    version = subprocess.run([str(binary), "version"], text=True, capture_output=True).stdout.strip()
    with open(log, "ab") as out:
        res = subprocess.run([str(binary), "migrate", "--config", str(cfg)], stdout=out,
                             stderr=subprocess.STDOUT, stdin=subprocess.DEVNULL, env=env)
        if res.returncode != 0:
            raise SystemExit(f"vu: migrate failed ({res.returncode}), see {log}")
        # own session (setsid): outlives this process, like `stand start`
        proc = subprocess.Popen([str(binary), "server", "--config", str(cfg)], cwd=DIR, stdout=out,
                                stderr=subprocess.STDOUT, stdin=subprocess.DEVNULL, env=env,
                                start_new_session=True)
    STATE.write_text(json.dumps({"workspace": ws, "binary": str(binary), "version": version,
                                 "pid": proc.pid, "url": base}, indent=2) + "\n")
    port = st.port_of(base)
    s = st.Stand(ws, base, "", DIR)
    for _ in range(120):
        if proc.poll() is not None:
            STATE.unlink(missing_ok=True)
            raise SystemExit(f"vu: the server of {ws} exited ({proc.returncode}), see {log}")
        if st.is_up(s) and proc.pid in st.listeners(port)[1]:
            print(f"vu: {ws} binary ({version}) serves the fixture at {base} (pid {proc.pid}, data {DIR}); "
                  f"stop: ./vu.sh fixture stop", flush=True)
            return
        time.sleep(0.5)
    raise SystemExit(f"vu: the server of {ws} does not answer at {base} (now: {st.describe(port)}), see {log}")
