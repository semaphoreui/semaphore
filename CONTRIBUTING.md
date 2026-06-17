## Developer documentation

In-repo guides for contributors and operators live in [`docs/`](docs/README.md): configuration, runners, executors, workflows, API pagination, and HA cluster operations.

## Pull Requests

When creating a pull-request you should:

- __Open an issue first:__ Confirm that the change or feature will be accepted
- __Update API documentation:__ If your pull-request adding/modifying an API request, make sure you update the Swagger documentation (`api-docs.yml`)
- __Run API Tests:__ If your pull request modifies the API make sure you run the integration tests using **dredd**.

## Installation in a development environment

- Check out the `develop` branch
- [Install Go](https://golang.org/doc/install). Go must be >= v1.21 for all the tools we use to work
- Install MySQL / MariaDB (Optional)
- Install node.js

1) Set up `GOPATH`
   * Set `GOPATH` in your shell (for example, in your `.bashrc` or `.zshrc`):
   
      ```bash
      export GOPATH=$HOME/go
      export PATH=$PATH:$GOPATH/bin
      ```
   * Create required directory and switch to it:
   
      ```bash
      mkdir -p $GOPATH/src/github.com/semaphoreui
      cd $GOPATH/src/github.com/semaphoreui
      ```

2) Clone semaphore (with submodules)

   ```
   git clone --recursive git@github.com:semaphoreui/semaphore.git && cd semaphore
   ```

3) Install dev dependencies

   ```
   go install github.com/go-task/task/v3/cmd/task@latest
   task deps
   ```
   Windows users will additionally need to manually install goreleaser from https://github.com/goreleaser/goreleaser/releases

4) Create database if you want to use MySQL or PostgreSQL (Semaphore also supports SQLite, which does not require
   additional setup)

   > **BoltDB removed in 2.19:** The embedded BoltDB backend was removed. Use SQLite for
   > local development or MySQL/PostgreSQL for production. Existing `database.boltdb`
   > files must be migrated before upgrading.

   ```
   echo "create database semaphore;" | mysql -uroot -p
   ```

5) Compile, set up & run

   ```
   task build
   go run cli/main.go setup
   go run cli/main.go service --config ./config.json
   ```

Open [localhost:3000](http://localhost:3000)

Note: for Windows, you may need [Cygwin](https://www.cygwin.com/) to run certain commands because the [reflex](github.com/cespare/reflex) package probably doesn't work on Windows.
You may encounter issues when running `task watch`, but running `task build` etc... will still be OK.

## Integration tests

Dredd is used for API integration tests, if you alter the API in any way you must make sure that the information in the api docs
matches the responses.

As Dredd and the application database config may differ it expects it's own config.json in the .dredd folder.

### How to run Dredd tests locally

1) Build Dredd hooks:

    ```bash
    task dredd:hooks
    ```
2) Install Dredd globally

    ```bash
    npm install -g dredd
    ```
3) Create `./dredd/config.json` for Dredd. It must contain database connection same as used in Semaphore server.
   You can use any supported database dialect for tests. For example SQLite.
    ```json
   {
        "sqlite": {
            "host": "/tmp/database.sqlite"
        },
        "dialect": "sqlite"
    }
    ```
4) Start Semaphore server (add `--config` option if required):

5) ```bash
    ./bin/semaphore server
    ```
5) Start Dredd tests

    ```
    dredd --config ./.dredd/dredd.local.yml
    ```

## Logging and diagnostics

Semaphore uses [logrus](https://github.com/sirupsen/logrus) for structured logging.
Log level and selective debug filtering are controlled via CLI flags or environment
variables (flags take precedence).

### Log level

| Flag / env var                        | Values                                             |
|---------------------------------------|----------------------------------------------------|
| `--log-level` / `SEMAPHORE_LOG_LEVEL` | `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, `PANIC` |

```bash
SEMAPHORE_LOG_LEVEL=DEBUG ./bin/semaphore server --config ./config.json
```

### Selective debug filter

When the log level is `DEBUG`, you can narrow output to specific subsystems using a
[Node.js `debug`](https://www.npmjs.com/package/debug)-style namespace filter. The filter
only acts on `DEBUG`-level entries; it does not raise the log level on its own.

| Flag / env var                              | Example values              |
|---------------------------------------------|-----------------------------|
| `--debug-filter` / `SEMAPHORE_DEBUG_FILTER` | `runner`, `task_*`, `*,-db` |

Syntax: comma- or space-separated patterns; `*` is a wildcard; a leading `-` excludes a
namespace (exclusions win over includes).

```bash
# Only runner subsystem debug output
SEMAPHORE_LOG_LEVEL=DEBUG SEMAPHORE_DEBUG_FILTER=runner ./bin/semaphore server

# All task-related namespaces except database tracing
SEMAPHORE_LOG_LEVEL=DEBUG SEMAPHORE_DEBUG_FILTER='task_*,*,-db' ./bin/semaphore server
```

Debug entries use a `context` field as the namespace. Common namespaces:

| Namespace          | Subsystem                                  |
|--------------------|--------------------------------------------|
| `runner`           | Runner registration, job hand-off, polling |
| `task_pool`        | Task queueing, scheduling, capacity        |
| `task_runner`      | Individual task lifecycle                  |
| `git`              | Repository clone/checkout                  |
| `terraform`        | Terraform/OpenTofu invocation              |
| `schedule`         | Cron schedules, dedup, fire/skip           |
| `db` / `migration` | Database queries and migrations            |
| `websocket`        | Real-time UI sockets                       |

When syslog is enabled (`syslog` in `config.json`), the same filter applies to syslog
output as well as stdout.

## Goland debug configuration

<img width="700" alt="image" src="https://github.com/user-attachments/assets/cc6132ee-b31e-424c-8ca9-4eba56bf7fb0" />

## Manual testing with using Semaphore MCP and Cursor agent

1. Install Semaphore MCP

   ```bash
   pipx install semaphore-mcp
   ```

   Upgrade:

   ```bash
   pipx upgrade semaphore-mcp
   ```

2. Install Cursor Agent CLI

   ```bash
   curl https://cursor.com/install -fsSL | bash
   ```

   You can check the agent using command:

   ```bash
   cursor-agent --version
   ```

3. Set up MCP server for Cursor

   Add following block to `~/.cursor/mcp.json`:

   ```json
	{
	  "mcpServers": {
	    "semaphore": {
	      "command": "semaphore-mcp",
	      "args": [],
	      "env": {
	        "SEMAPHORE_URL": "http://localhost:3000",
	        "SEMAPHORE_API_TOKEN": "<TOKEN>"
	      }
	    }
	  }
	}
   ```

4. Run tests

   ```bash
   cd tests/manual
   ./run.sh
   ```
