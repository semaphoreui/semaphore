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

4) Create database if you want to use MySQL (Semaphore also supports SQLite, it doesn't require additional action)

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
   You can use any supported database dialect for tests. For example BoltDB.
    ```json
   {
        "bolt": {
            "host": "/tmp/database.boltdb"
        },
        "dialect": "bolt"
    }
    ```
4) Start Semaphore server (add `--config` option if required):

   ```bash
   ./bin/semaphore server
   ```
5) Start Dredd tests

    ```
    dredd --config ./.dredd/dredd.local.yml
    ```

## Project backup and restore

Project export/import is implemented in `services/project` (`GetBackup`, `Restore`, JSON marshalling). The backup payload includes a `runners` array for **project-scoped runners** (loaded with tag filter `ignore_tags`, so all project runners are included regardless of tags).

**Restore behavior:** each runner is inserted via `CreateRunner`, which **generates a new authentication token** for the row. After restoring a project on a new server, you must **reconfigure runners** with the new tokens (or recreate runners in the UI). Names must be unique within the `runners` section of the backup, same as other named entities.

Older backup files without a `runners` field are still valid; treat missing `runners` as an empty list.

## Secret storage background sync

When sync is enabled on a secret storage or on a variable group’s sync settings, the server runs `SecretStorageSyncScheduler` (`services/server/secret_storage_sync_scheduler.go`). It wakes every **60 seconds**, loads all sync-enabled `SecretSync` rows (`GetSyncEnabledSecretSyncs`), and calls `SyncSecrets` when the configured **interval in minutes** has elapsed since the last successful sync or the last failed attempt (whichever is later). A **zero** interval disables automatic sync for that row.

Sync scope is defined per row: `EnvironmentID == nil` is **storage-level** sync (imports access keys from paths); a non-nil `EnvironmentID` is **environment-scoped** sync (imports variables for that variable group). See `db/SecretSync.go` and the secret storage / environment APIs in Swagger (`web/public/swagger/api-docs.yml`).

## Task home directory mode (`home_dir_mode`)

`SEMAPHORE_HOME_DIR_MODE` / `home_dir_mode` in `config.json` controls how task processes see `HOME` and (for Ansible) `ANSIBLE_HOME`. Implementation: `util/config.go` constants and `db_lib/getHomeDir` plus `db_lib/AnsiblePlaybook.makeCmd`.

| Value | `HOME` for tasks | Ansible notes |
|-------|------------------|---------------|
| `template_dir` (default) | Process user `HOME` (not overridden) | `ANSIBLE_HOME` set to `<repo home for template>/.ansible` |
| `project_home` | Project temp directory | No extra `ANSIBLE_HOME` in playbook runner |
| `user_home` | Process user `HOME` | Same `ANSIBLE_HOME` isolation as `template_dir` |

Use `template_dir` or `user_home` when you need isolation between parallel Ansible tasks; `project_home` is legacy and can cause contention on shared `.ansible` caches.

## BoltDB dialect (deprecated)

`SEMAPHORE_DB_DIALECT=bolt` is still accepted for compatibility but **BoltDB is deprecated** in favor of SQLite (`sqlite`). New installs should use SQLite or MySQL/PostgreSQL. The CLI setup still lists BoltDB as an option with a deprecation warning (`cli/setup/setup.go`).

The API exposes `boltdb_used` in system info when the active dialect is `bolt` (`api/system_info.go`); the web UI may show a notice in that case. Prefer migrating with `cli/cmd/migrate.go` and the documented migration path rather than relying on Bolt long term.

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
