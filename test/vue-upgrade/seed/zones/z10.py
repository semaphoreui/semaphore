"""Z10 Task log (TaskLogDialog: header, Log / Details / Summary tabs): tasks the base set has not.

Seed Demo, as admin, names with the prefix «Z10 »:

- repository «Z10 repo» — a local git repository runtime/visual/fixture/repos/z10 written from the
  files below (fixed author and date, as ctx.repo: one commit hash on every build); kept in this file
  so the shared seed/repos/* (and the commit hash of «Seed infra» other zones show) stay untouched;
- inventory «Z10 hosts» — z10-a, z10-b, z10-c (local), z10-down-1, z10-down-2 (ssh to 127.0.0.1:1,
  refused at once: «unreachable» without waiting for a timeout);
- «Z10 Ansible params» (ansible/site.yml of Seed infra) — one run with a message, Ansible parameters
  and an extra environment: the message bar and the Task parameters card;
- «Z10 Ansible failing» (z10-fail.yml of Z10 repo, Z10 hosts) — one run, status error: 2 failed hosts,
  2 unreachable, 1 ok with a change — the Summary tables;
- «Z10 OK script» (ok.sh of Seed scripts) + integration «Z10 hook» on it with an alias; the seed calls
  the alias once — a task started by an integration (no author);
- «Z10 long log» (z10-long.sh of Z10 repo) — one run, 300 coloured lines, some long enough to wrap.

The live log of a running task is tested on a patched task with a mocked websocket
(tests/zones/test_z10.py), so no slow template is needed.

Idempotent: objects are found by name; a template is run only while it has no task.
"""

import os
import subprocess
import time

from seed.ctx import GIT_DATE

DEMO = "Seed Demo"
REPO = "Z10 repo"
INVENTORY = "Z10 hosts"

T_PARAMS = "Z10 Ansible params"
T_FAILING = "Z10 Ansible failing"
T_OK = "Z10 OK script"
T_LONG = "Z10 long log"
HOOK = "Z10 hook"

MESSAGE = "Z10 run with parameters"
PARAMS = {"debug": True, "debug_level": 2, "diff": True, "dry_run": True, "limit": ["localhost"]}
ENVIRONMENT = '{"Z10_VAR": "1"}'

FAIL_PLAYBOOK = """\
- name: Z10 summary playbook
  hosts: all
  gather_facts: false
  tasks:
    - name: Z10 check the host
      ansible.builtin.fail:
        msg: "z10: {{ inventory_hostname }} is broken on purpose"
      when: inventory_hostname in ['z10-a', 'z10-b']

    - name: Z10 change the host
      ansible.builtin.command: "true"
      changed_when: true
"""

LONG_SCRIPT = r"""#!/usr/bin/env bash
# 300 numbered lines: ANSI colours every few lines, a long line every 25th (wraps in the log).
long=$(printf 'z10-long-text %.0s' $(seq 1 30))
for i in $(seq 1 300); do
  n=$(printf '%03d' "$i")
  if (( i % 25 == 0 )); then
    echo "z10: line $n $long"
  elif (( i % 5 == 1 )); then
    printf '\033[32mz10: line %s ok\033[0m\n' "$n"
  elif (( i % 5 == 2 )); then
    printf '\033[33mz10: line %s changed\033[0m\n' "$n"
  elif (( i % 5 == 3 )); then
    printf '\033[31mz10: line %s failed\033[0m\n' "$n"
  else
    echo "z10: line $n"
  fi
done
"""

FILES = {
    "z10-fail.yml": FAIL_PLAYBOOK,
    "z10-long.sh": LONG_SCRIPT,
}

HOSTS = """\
[ok]
z10-a ansible_connection=local
z10-b ansible_connection=local
z10-c ansible_connection=local

[ok:vars]
ansible_python_interpreter="{{ ansible_playbook_python }}"

[down]
z10-down-1 ansible_host=127.0.0.1 ansible_port=1
z10-down-2 ansible_host=127.0.0.1 ansible_port=1

[down:vars]
ansible_ssh_common_args='-o ConnectTimeout=2 -o StrictHostKeyChecking=no -o BatchMode=yes'
"""


def z10_repo(ctx):
    """runtime/visual/fixture/repos/z10 with FILES, one commit with a fixed author and date."""
    dst = ctx.repos_dir / "z10"
    env = dict(os.environ, GIT_AUTHOR_NAME="seed", GIT_AUTHOR_EMAIL="seed@seed.local",
               GIT_COMMITTER_NAME="seed", GIT_COMMITTER_EMAIL="seed@seed.local",
               GIT_AUTHOR_DATE=GIT_DATE, GIT_COMMITTER_DATE=GIT_DATE)
    git = ["git", "-C", str(dst), "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null"]
    if not (dst / ".git").is_dir():
        dst.mkdir(parents=True, exist_ok=True)
        subprocess.run(["git", "init", "-q", "-b", "main", str(dst)], check=True)
    for name, text in FILES.items():
        path = dst / name
        path.write_text(text)
        if name.endswith(".sh"):
            path.chmod(0o755)
    subprocess.run(git + ["add", "-A"], check=True, env=env)
    if subprocess.run(git + ["diff", "--cached", "--quiet"], env=env).returncode != 0:
        subprocess.run(git + ["commit", "-q", "-m", "seed: z10"], check=True, env=env)
    return str(dst)


def ran(ctx, pid, task_id, want, ansible_hosts=None, commit=True):
    """A task the seed has just run ended as the shots need it — otherwise stop the seed: a run broken
    by the stand (a task directory or ssh-agent socket it cannot use) ends in «error» with no commit,
    no log and no Ansible summary, and the Z10 shots of it would be wrong. None — the task was there."""
    if task_id is None:
        return
    task = ctx.admin.get(f"/project/{pid}/tasks/{task_id}")
    problems = []
    if task["status"] != want:
        problems.append(f"status {task['status']}, expected {want}")
    if commit and not task.get("commit_hash"):
        problems.append("no commit (the repository was not cloned)")
    if ansible_hosts is not None:
        # the PLAY RECAP hosts are stored after the task turns terminal — wait for them
        deadline = time.time() + 30
        while True:
            hosts = ctx.admin.get(f"/project/{pid}/tasks/{task_id}/ansible/hosts") or []
            if len(hosts) >= ansible_hosts or time.time() > deadline:
                break
            time.sleep(0.5)
        if len(hosts) != ansible_hosts:
            problems.append(f"{len(hosts)} Ansible hosts, expected {ansible_hosts}")
    if problems:
        raise RuntimeError(f"Z10 task #{task_id} (template {task['template_id']}): "
                           + "; ".join(problems) + " — check the stand's tmp_path and its server log")


def seed(ctx):
    pid = ctx.project_id(DEMO)
    base = f"/project/{pid}"
    none = ctx.id_of(pid, "keys", "None")

    repo = ctx.ensure(f"{base}/repositories", {
        "name": REPO, "git_url": "file://" + z10_repo(ctx), "git_branch": "main", "ssh_key_id": none,
        "project_id": pid}, kind="repository")["id"]
    hosts = ctx.ensure(f"{base}/inventory", {
        "name": INVENTORY, "type": "static", "inventory": HOSTS, "ssh_key_id": none, "project_id": pid},
        kind="inventory")["id"]

    def template(name, app, playbook, repository, inventory, environment, view=None, **extra):
        body = {"name": name, "app": app, "playbook": playbook, "project_id": pid,
                "repository_id": repository, "inventory_id": inventory,
                "environment_ids": [ctx.id_of(pid, "environment", environment)]}
        if view:
            body["view_id"] = ctx.id_of(pid, "views", view, key="title")
        body.update(extra)
        return ctx.ensure(f"{base}/templates", body, kind="template")["id"]

    localhost = ctx.id_of(pid, "inventory", "Localhost")
    infra = ctx.id_of(pid, "repositories", "Seed infra")
    scripts = ctx.id_of(pid, "repositories", "Seed scripts")

    template(T_PARAMS, "ansible", "ansible/site.yml", infra, localhost, "Default", "Infrastructure")
    template(T_FAILING, "ansible", "z10-fail.yml", repo, hosts, "Empty", "Infrastructure")
    ok = template(T_OK, "bash", "ok.sh", scripts, localhost, "Default", "Scripts")
    template(T_LONG, "bash", "z10-long.sh", repo, localhost, "Empty", "Scripts")

    print("tasks")
    ran(ctx, pid, ctx.run_task(pid, T_PARAMS, "admin", 1, message=MESSAGE, params=PARAMS,
                               environment=ENVIRONMENT), "success")
    ran(ctx, pid, ctx.run_task(pid, T_FAILING, "admin", 1), "error", ansible_hosts=5)
    ran(ctx, pid, ctx.run_task(pid, T_LONG, "admin", 1), "success")

    print("integration")
    hook = ctx.ensure(f"{base}/integrations", {
        "name": HOOK, "project_id": pid, "template_id": ok, "auth_method": "", "searchable": False},
        kind="integration")
    ipath = f"{base}/integrations/{hook['id']}/aliases"
    aliases = ctx.admin.get(ipath) or []
    if aliases:
        ctx.log("=", "alias", HOOK)
    else:
        ctx.admin.post(ipath, {})
        aliases = ctx.admin.get(ipath) or []
        ctx.log("+", "alias", HOOK)
    if ctx.admin.get(f"{base}/templates/{ok}/tasks"):
        ctx.log("=", "task", f"{T_OK} by {HOOK}")
        return
    alias = aliases[0]["url"].rstrip("/").rsplit("/", 1)[-1]  # the list gives the full URL only
    ctx.admin.call("POST", f"/integrations/{alias}", {"ref": "refs/heads/main"})
    deadline = time.time() + 60
    tasks = []
    while not tasks and time.time() < deadline:
        tasks = ctx.admin.get(f"{base}/templates/{ok}/tasks") or []
        if not tasks:
            time.sleep(0.5)
    if not tasks:
        raise RuntimeError(f"{HOOK}: the integration made no task")
    status = ctx.wait_task(pid, tasks[0]["id"])
    ctx.log("+", "task", f"{T_OK} by {HOOK}", f"#{tasks[0]['id']} {status}")
    ran(ctx, pid, tasks[0]["id"], "success", commit=False)  # Seed scripts: a local path, no git
