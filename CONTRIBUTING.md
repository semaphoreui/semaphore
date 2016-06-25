# Pull Requests

When creating a pull-request you should:

- __Open an issue first:__ Confirm that the change or feature will be accepted
- __gofmt and vet the code:__ Use  `gofmt`, `golint`, `govet` and `goimports` to clean up your code.
- __Update api documentation:__ If your pull-request adding/modifying an API request, make sure you update the swagger documentation (`swagger.yml`)

# Installing dependencies

Clone the project to `$GOPATH/src/github.com/ansible-semaphore/semaphore` (more on GOPATHS below)

> note: You should clone semaphore with all submodules
> - you should have latest go installed and node with ES6 (used to be a special `harmony` flag) capability

```
go get github.com/jteeuwen/go-bindata/...
go get github.com/mitchellh/gox
go get github.com/cespare/reflex
go get -u ./...

npm i -g nodemon less jade
npm i async
```

## Gopaths

To develop in Go, you need to setup a gopath where go code, libraries & executables live.

Follow either of these ([go wiki](https://github.com/golang/go/wiki/GOPATH), [tutorial](http://www.ryanday.net/2012/10/01/installing-go-and-gopath/), [SO question](https://stackoverflow.com/questions/21001387/how-do-i-set-the-gopath-environment-variable-on-ubuntu-what-file-must-i-edit)).

1. `mkdir -p $GOPATH/src/github.com/ansible-semaphore`
2. `cd $GOPATH/src/github.com/ansible-semaphore`
3. `git clone --recursive git@github.com:ansible-semaphore/semaphore.git`
4. Now install dependencies above

# Running in development

You will need to have a local `config.json` file because it is linked to. It should contain your local configuration.

```
$EDITOR config.json
./make.sh watch
```
