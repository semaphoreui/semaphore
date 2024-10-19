## Pull Requests

When creating a pull-request you should:

- __Open an issue first:__ Confirm that the change or feature will be accepted
- __gofmt and vet the code:__ Use  `gofmt`, `golint`, `govet` and `goimports` to clean up your code.
- __vendor dependencies with dep:__ Use `dep ensure --update` if you have added to or updated dependencies, so that they get added to the dependency manifest.
- __Update api documentation:__ If your pull-request adding/modifying an API request, make sure you update the swagger documentation (`api-docs.yml`)
- __Run Api Tests:__ If your pull request modifies the API make sure you run the integration tests using dredd.

## Installation in a development environment

- Check out the `develop` branch
- [Install Go](https://golang.org/doc/install). Go must be >= v1.10 for all the tools we use to work
- Install MySQL / MariaDB (OPTIONAL!!!)
- Install node.js

1) Set up GOPATH, GOBIN and Workspace.
```
mkdir -p $GOPATH/src/github.com/ansible-semaphore && cd $GOPATH/src/github.com/ansible-semaphore
```

2) Clone semaphore (with submodules)

```
git clone --recursive git@github.com:ansible-semaphore/semaphore.git && cd semaphore
```

3) Install dev dependencies

```
go install github.com/go-task/task/v3/cmd/task@latest
task deps
```
Windows users will additionally need to manually install goreleaser from https://github.com/goreleaser/goreleaser/releases

4) Create database if you want to use MySQL (Semaphore also supports [bbolt](https://github.com/etcd-io/bbolt), it doesn't require additional action) 

```
echo "create database semaphore;" | mysql -uroot -p
```

5) Compile, set up & run

```
task compile
go run cli/main.go setup
go run cli/main.go service --config ./config.json
```

Open [localhost:3000](http://localhost:3000)

Note: for Windows, you may need [Cygwin](https://www.cygwin.com/) to run certain commands because the [reflex](github.com/cespare/reflex) package probably doesn't work on Windows. 
You may encounter issues when running `task watch`, but running `task build` etc... will still be OK.

## Integration Tests

Dredd is used for API integration tests, if you alter the API in any way you must make sure that the information in the api docs
matches the responses.

As Dredd and the application database config may differ it expects it's own config.json in the .dredd folder.

### How to run Dredd tests locally

1) Build Dredd hooks:
    ````bash
    task compile:api:hooks
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
    ````bash
    ./bin/semaphore server
    ```
5) Start Dredd tests
    ```
    dredd --config ./.dredd/dredd.local.yml
    ```