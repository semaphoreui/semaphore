# Contributing

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
- Install MySQL / MariaDB
- Install node.js

1. Set up GOPATH, GOBIN and Workspace.

```bash
cd {WORKING_DIRECTORY}
# Exports only needed pre Go 1.8 or for custom GOPATH location
export GOPATH=`pwd`
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN

mkdir -p $GOPATH/src/github.com/ansible-semaphore && cd $GOPATH/src/github.com/ansible-semaphore
```

2. Clone semaphore (with submodules)

```bash
git clone --recursive git@github.com:ansible-semaphore/semaphore.git && cd semaphore
```

3. Install dev dependencies

```bash
go get github.com/go-task/task/v2/cmd/task
task deps
```

Windows users will additionally need to manually install goreleaser from <https://github.com/goreleaser/goreleaser/releases>

4. Set up config, database & run migrations

```bash
cat config.json
{
    "mysql": {
        "host": "127.0.0.1:3306",
        "user": "root",
        "pass": "",
        "name": "semaphore"
    },
    "port": ":3000"
}

echo "create database semaphore;" | mysql -uroot -p
task compile
go run cli/main.go -config ./config.json -migrate
```

Now it's ready to start.. Run `task watch`

- Watches js files in `public/js/*` and compiles into a bundle
- Watches css files in `public/css/*` and compiles into css code
- Watches pug files in `public/html/*` and compiles them into html
- Watches go files and recompiles the binary
- Open [localhost:3000](http://localhost:3000)

Note: for Windows, you may need [Cygwin](https://www.cygwin.com/) to run certain commands because the [reflex](github.com/cespare/reflex) package probably doesn't work on Windows. 
You may encounter issues when running `task watch`, but running `task build` etc... will still be OK.

## Integration Tests

Dredd is used for API integration tests, if you alter the API in any way you must make sure that the information in the api docs
matches the responses.

As Dredd and the application database config may differ it expects it's own config.json in the .dredd folder.
The most basic configuration for this using a local docker container to run the database would be

```json
{
    "mysql": {
        "host": "0.0.0.0:3306",
        "user": "semaphore",
        "pass": "semaphore",
        "name": "semaphore"
    }
}
```

It is strongly advised to run these tests inside docker containers, as dredd will write a lot of test information and will __NOT__ clear it up.
This means that you should never run these tests against your productive database!
The best practice to run these tests is to use docker and the task commands.

```bash
context=dev task dc:build #build fresh semaphore images
context=dev task dc:up  #up semaphore and mysql
task dc:build:dredd #build fresh dredd image
task dc:up:dredd #run dredd over docker-compose stack
```
