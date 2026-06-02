# Quick-test the Workflow Templates release in Docker on WSL

This guide shows the fastest way to spin up the `copilot/implement-workflow-templates`
branch (DAG-based Workflow Templates, PR #1) inside Docker on WSL2 and exercise
the new endpoints end-to-end.

It uses the existing `deployment/docker/server/Dockerfile` and BoltDB so you do
not need to provision an external database.

---

## 1. Prerequisites (one-time, on WSL)

In your WSL distro (Ubuntu/Debian shown):

```bash
# WSL2 must be enabled. Verify from PowerShell:
#   wsl --status            -> "Default Version: 2"
#   wsl --update            -> keep WSL up to date

# Inside WSL, you need Docker. Either:
#   a) Install Docker Desktop on Windows and enable
#      "Settings -> Resources -> WSL integration" for your distro, or
#   b) Install the Docker engine directly inside WSL:
#         curl -fsSL https://get.docker.com | sh
#         sudo usermod -aG docker "$USER"
#         newgrp docker

docker version            # sanity check
docker compose version    # sanity check
git --version
```

> Tip: clone into the WSL filesystem (e.g. `~/src`), **not** under `/mnt/c/...`.
> Building from `/mnt/c` is dramatically slower because of the 9P file-system
> bridge.

---

## 2. Get the branch

```bash
mkdir -p ~/src && cd ~/src
git clone https://github.com/makaiver/semaphore.git
cd semaphore
git fetch origin copilot/implement-workflow-templates
git checkout copilot/implement-workflow-templates
```

---

## 3. Build the server image

The repo already has a multi-stage Dockerfile that compiles both the Go backend
and the Vue frontend, so a single `docker build` is enough:

```bash
docker build \
  -f deployment/docker/server/Dockerfile \
  -t semaphore-workflows:dev \
  .
```

This takes a few minutes the first time (Go modules + npm install). Subsequent
builds are cached.

If you prefer Taskfile and have [Task](https://taskfile.dev) installed:

```bash
task docker:build:server tag=workflows-dev
# produces semaphoreui/semaphore:workflows-dev
```

---

## 4. Run it (BoltDB, single container)

```bash
docker volume create semaphore-workflows-data

docker run -d \
  --name semaphore-workflows \
  -p 3000:3000 \
  -e SEMAPHORE_DB_DIALECT=bolt \
  -e SEMAPHORE_DB_PATH=/var/lib/semaphore \
  -e SEMAPHORE_TMP_PATH=/tmp/semaphore \
  -e SEMAPHORE_ADMIN_NAME=Admin \
  -e SEMAPHORE_ADMIN=admin \
  -e SEMAPHORE_ADMIN_EMAIL=admin@localhost \
  -e SEMAPHORE_ADMIN_PASSWORD=changeme123 \
  -e SEMAPHORE_WEB_ROOT=http://localhost:3000 \
  -e SEMAPHORE_ACCESS_KEY_ENCRYPTION=IlRqgrrO5Gp27MlWakDX1xVrPv4jhoUx+ARY+qGyDxQ= \
  -v semaphore-workflows-data:/var/lib/semaphore \
  semaphore-workflows:dev
```

Open http://localhost:3000 from Windows (WSL forwards `localhost` automatically)
and log in as `admin` / `changeme123`.

Tail the logs while you test:

```bash
docker logs -f semaphore-workflows
```

You should see the new SQL migration tag in the log (BoltDB just creates the
buckets on demand):

```
Executing migration v2.18.6 ...   # only on SQL backends
```

### Optional: SQL backend via compose

If you want to test the actual SQL migration (`db/sql/migrations/v2.18.6.sql`)
use any of the snippets under `deployment/compose/` (postgres / mysql) and
override the image build to use this branch:

```bash
cd deployment/compose
# pick one of: store/postgres, store/mysql
SEMAPHORE_VERSION=workflows-dev \
  docker compose \
    -f server/base.yml \
    -f server/build.yml \
    -f store/postgres/base.yml \
    up --build
```

`server/build.yml` already points the build context at the repo root, so it
will build from your checked-out branch.

---

## 5. Smoke-test the new workflow API

Create a project + two task templates from the UI first (any two trivial
templates will do), then grab their IDs from `/project/<id>/templates`.

Create an API token from **User settings -> API tokens** in the UI, export it,
and call the new endpoints:

```bash
export TOKEN="<paste token>"
export PID=1                  # your project id
export T_BUILD=1              # template id for "build"
export T_DEPLOY=2             # template id for "deploy"
export T_RECOVER=3            # template id for failure-path

# 1. Create a workflow: build -> deploy on success, build -> recover on failure
curl -sS -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:3000/api/project/$PID/workflows \
  -d @- <<EOF
{
  "name": "Build/Deploy with failure recovery",
  "description": "Run recover when build fails",
  "nodes": [
    { "id": 1, "template_id": $T_BUILD },
    { "id": 2, "template_id": $T_DEPLOY },
    { "id": 3, "template_id": $T_RECOVER }
  ],
  "edges": [
    { "source_node_id": 1, "destination_node_id": 2, "condition": "on_success" },
    { "source_node_id": 1, "destination_node_id": 3, "condition": "on_failure" }
  ]
}
EOF

# 2. List
curl -sS -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/api/project/$PID/workflows | jq

# 3. Run it (replace WID with the id returned above)
export WID=1
curl -sS -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/api/project/$PID/workflows/$WID/run

# 4. Inspect runs / per-node task status
curl -sS -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/api/project/$PID/workflows/$WID/runs | jq
curl -sS -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/api/project/$PID/workflows/$WID/runs/1 | jq
```

What to verify:

* Workflow creation **rejects** cycles and multi-root graphs (HTTP 400).
* On a successful build, only the `on_success` edge is taken (deploy runs,
  recover does not).
* On a failing build, only the `on_failure` edge is taken (recover runs,
  deploy does not).
* `WorkflowRun.status` aggregates to `success` / `failed` once all reachable
  node tasks have terminated.
* Each spawned task carries `workflow_run_id` and `workflow_node_id`
  (visible in `GET /api/project/$PID/tasks/<id>`).

---

## 6. Reset / cleanup

```bash
docker rm -f semaphore-workflows
docker volume rm semaphore-workflows-data
docker rmi semaphore-workflows:dev
```

---

## Troubleshooting

| Symptom | Fix |
|---|---|
| `Cannot connect to the Docker daemon` in WSL | Start Docker Desktop, or `sudo service docker start` if using engine-in-WSL. |
| Build is extremely slow | Make sure the repo lives on the WSL filesystem (`~/src/...`), not on `/mnt/c/...`. |
| `port is already allocated` on 3000 | Change the host port: `-p 3001:3000`. |
| 401 on API calls | Token expired or missing `Bearer` prefix. Re-issue from the UI. |
| Want to wipe DB only | `docker volume rm semaphore-workflows-data` and re-run step 4. |
| Need to test SQL migration `v2.18.6` specifically | Use the compose recipe in step 4 with the postgres or mysql store snippet. |
