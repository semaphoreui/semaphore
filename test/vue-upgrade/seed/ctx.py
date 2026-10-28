"""Seed context: API sessions, idempotent create-by-name, local git repositories, task runs.

Every helper looks an object up by its name first and creates it only when it is missing, so the
base set and every zone seed may run again on a stand that already has the data (`vu seed zone`).
No randomness: fixed names, fixed values, fixed git commit dates.
"""

import hashlib
import os
import shutil
import subprocess
import time
from pathlib import Path

from vu.api import Api, ApiError

TERMINAL = {"success", "error", "stopped", "rejected"}
REPOS_SRC = Path(__file__).resolve().parent / "repos"
GIT_DATE = "2026-01-15T10:00:00+00:00"  # every seed commit: same hash on every build


class Ctx:
    def __init__(self, url, repos_dir, admin_password="admin123", password="seed12345", verbose=True):
        self.url = url
        self.repos_dir = Path(repos_dir)
        self.password = password  # every seed_* user
        self.admin_password = admin_password
        self.verbose = verbose
        self.admin = Api(url, "admin", admin_password)
        self.sessions = {"admin": self.admin}

    # ---- output ---------------------------------------------------------------

    def log(self, mark, kind, name, note=""):
        if self.verbose:
            print(f"  {mark} {kind:<12} {name}{'  — ' + note if note else ''}", flush=True)

    # ---- sessions and lookups ------------------------------------------------

    def session(self, username):
        if username not in self.sessions:
            pw = self.admin_password if username == "admin" else self.password
            self.sessions[username] = Api(self.url, username, pw)
        return self.sessions[username]

    def user_id(self, username):
        user = self.admin.find("/users", username, key="username")
        if user is None:
            raise KeyError(f"no user {username}")
        return user["id"]

    def project_id(self, name):
        p = self.admin.find("/projects", name)
        if p is None:
            raise KeyError(f"no project {name}")
        return p["id"]

    def id_of(self, pid, kind, name, key="name"):
        """Id of a project object by name; kind is the API list: keys, repositories, inventory,
        environment, templates, views (key=title), schedules, integrations."""
        obj = self.admin.find(f"/project/{pid}/{kind}", name, key=key)
        if obj is None:
            raise KeyError(f"project {pid}: no {kind} '{name}'")
        return obj["id"]

    # ---- create-by-name -------------------------------------------------------

    def ensure(self, path, body, key="name", api=None, list_path=None, kind=None):
        """POST <body> to <path> unless the list at <list_path or path> has an item with the same
        <key>; returns the existing or the created object."""
        api = api or self.admin
        kind = kind or path.rstrip("/").rsplit("/", 1)[-1]
        found = api.find(list_path or path, body[key], key=key)
        if found is not None:
            self.log("=", kind, str(body[key]))
            return found
        created = api.post(path, body)
        self.log("+", kind, str(body[key]))
        if created is None:  # some endpoints answer 204
            created = api.find(list_path or path, body[key], key=key)
        return created

    def ensure_user(self, username, name, admin=False, email=None):
        found = self.admin.find("/users", username, key="username")
        if found:
            self.log("=", "user", username)
            return found
        user = self.admin.post("/users", {
            "username": username, "name": name, "email": email or f"{username}@seed.local",
            "password": self.password, "admin": admin,
        })
        self.log("+", "user", username)
        return user

    def ensure_project(self, name, **extra):
        found = self.admin.find("/projects", name)
        if found:
            self.log("=", "project", name, f"id {found['id']}")
            return found["id"]
        p = self.admin.post("/projects", dict({"name": name}, **extra))
        self.log("+", "project", name, f"id {p['id']}")
        return p["id"]

    def ensure_member(self, pid, username, roles):
        uid = self.user_id(username)
        members = {m["username"]: m for m in self.admin.get(f"/project/{pid}/users") or []}
        want = sorted(roles)
        if username in members:
            have = sorted(r["slug"] if isinstance(r, dict) else r for r in members[username].get("roles") or [])
            if have == want:
                self.log("=", "member", username, ", ".join(want))
                return
            self.admin.put(f"/project/{pid}/users/{uid}", {"roles": want})
            self.log("~", "member", username, ", ".join(want))
            return
        self.admin.post(f"/project/{pid}/users", {"user_id": uid, "roles": want})
        self.log("+", "member", username, ", ".join(want))

    def optional(self, kind, name, fn):
        """Pro-only objects: an OSS build answers 404, a plan without the feature 403."""
        try:
            return fn()
        except ApiError as e:
            if e.status in (403, 404):
                self.log("-", kind, name, f"skipped, server answered {e.status}")
                return None
            raise

    # ---- repositories ---------------------------------------------------------

    def repo(self, name):
        """A local git repository with the files of seed/repos/<name>, in runtime/visual/fixture/repos/.
        Commits have a fixed author and date, so the commit hash is the same on every build."""
        src = REPOS_SRC / name
        dst = self.repos_dir / name
        env = dict(os.environ, GIT_AUTHOR_NAME="seed", GIT_AUTHOR_EMAIL="seed@seed.local",
                   GIT_COMMITTER_NAME="seed", GIT_COMMITTER_EMAIL="seed@seed.local",
                   GIT_AUTHOR_DATE=GIT_DATE, GIT_COMMITTER_DATE=GIT_DATE)
        git = ["git", "-C", str(dst), "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null"]
        if not (dst / ".git").is_dir():
            dst.mkdir(parents=True, exist_ok=True)
            subprocess.run(["git", "init", "-q", "-b", "main", str(dst)], check=True)
        shutil.copytree(src, dst, dirs_exist_ok=True)
        subprocess.run(git + ["add", "-A"], check=True, env=env)
        if subprocess.run(git + ["diff", "--cached", "--quiet"], env=env).returncode != 0:
            subprocess.run(git + ["commit", "-q", "-m", f"seed: {name}"], check=True, env=env)
        return str(dst)

    # ---- tasks ----------------------------------------------------------------

    def run_task(self, pid, template, as_user, min_count=1, **params):
        """Run template <template> as <as_user> unless it already has <min_count> tasks; wait for
        the end. One task at a time, so task ids and times follow the order of the calls."""
        tid = self.id_of(pid, "templates", template)
        have = len(self.admin.get(f"/project/{pid}/templates/{tid}/tasks") or [])
        if have >= min_count:
            self.log("=", "task", f"{template} as {as_user}")
            return None
        body = dict({"template_id": tid, "project_id": pid}, **params)
        task = self.session(as_user).post(f"/project/{pid}/tasks", body)
        status = self.wait_task(pid, task["id"])
        self.log("+", "task", f"{template} as {as_user}", f"#{task['id']} {status}")
        return task["id"]

    def wait_task(self, pid, task_id, timeout=240):
        deadline = time.time() + timeout
        status = "?"
        while time.time() < deadline:
            status = self.admin.get(f"/project/{pid}/tasks/{task_id}")["status"]
            if status in TERMINAL:
                return status
            time.sleep(0.5)
        raise RuntimeError(f"task #{task_id} still {status} after {timeout}s")


def fixed_ssh_key(name):
    """An ed25519 key derived from its name: the same public key on every build, no key in the repo."""
    from cryptography.hazmat.primitives import serialization
    from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey

    key = Ed25519PrivateKey.from_private_bytes(hashlib.sha256(f"vu-seed:{name}".encode()).digest())
    return key.private_bytes(serialization.Encoding.PEM, serialization.PrivateFormat.OpenSSH,
                             serialization.NoEncryption()).decode()
