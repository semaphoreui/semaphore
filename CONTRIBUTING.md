# Pull Requests

When creating a pull-request you should:

- __Open an issue first:__ Confirm that the change or feature will be accepted
- __gofmt and vet the code:__ Use  `gofmt`, `golint`, `govet` and `goimports` to clean up your code.
- __Update api documentation:__ If your pull-request adding/modifying an API request, make sure you update the swagger documentation (`api-docs.yml`)

## Docker

The easiest way to get started developing semaphore is to use docker. To create a development image you should run
`./make.sh dev up`. 

This will create the image and start both the application and a mysql database, inside this image the project route will be mounted as a volume so local changes are reflected in container. The default image start command is `./make.sh watch`

- On start the container will install all dependencies and rebuild tools and assets. 
- To avoid a rebuild on every start use `./make.sh dev start` instead of `up`
- To work directly in the container you can run `make.sh dev cmd`
- For other possible dev commands see images/dev/make.sh

## Installation in a local development environment

- Check out the `develop` branch
- [Install Go](https://golang.org/doc/install)
- Install MySQL / MariaDB
- Install node.js

1) Set up GOPATH, GOBIN and Workspace

```
cd {WORKING_DIRECTORY}
export GOPATH=`pwd`
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN

mkdir -p $GOPATH/src/github.com/ansible-semaphore && cd $GOPATH/src/github.com/ansible-semaphore
```

2) Clone semaphore (with submodules)

```
git clone --recursive git@github.com:ansible-semaphore/semaphore.git && cd semaphore
```

3) Install dev dependencies

```
./make.sh deps
```

4) Set up config, database & run migrations

```
cat <<EOT >> config.json
{
    "mysql": {
        "host": "127.0.0.1:3306",
        "user": "root",
        "pass": "",
        "name": "semaphore"
    },
    "port": ":3000"
}
EOT

echo "create database semaphore;" | mysql -uroot -p
go-bindata -debug -o util/bindata.go -pkg util config.json db/migrations/ $(find public/* -type d -print)
go run cli/main.go -config ./config.json -migrate
```

Now it's ready to start.. Run `./make.sh watch`

- Watches js files in `public/js/*` and compiles into a bundle
- Watches css files in `public/css/*` and compiles into css code
- Watches pug files in `public/html/*` and compiles them into html
- Watches go files and recompiles the binary
- Open [localhost:3000](http://localhost:3000)

Note: for Windows, you may need [Cygwin](https://www.cygwin.com/) to run certain commands. And because the [reflex](github.com/cespare/reflex) package probably doesn't work on Windows, you may encounter issues when running `./make.sh watch`, but running `./make.sh` will still be OK.
