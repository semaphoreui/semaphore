# Pull Requests

When creating a pull-request you should:

- __Open an issue first:__ Confirm that the change or feature will be accepted
- __gofmt and vet the code:__ Use  `gofmt`, `golint`, `govet` and `goimports` to clean up your code.
- __Update api documentation:__ If your pull-request adding/modifying an API request, make sure you update the swagger documentation (`api-docs.yml`)


# Installing dependencies

First of all you'll need the Go Language and Node.js installed. If you dont have them installed, you can follow these steps:

1) Create a directory called 'bin' in home that will receive the lib files from GoLang and Node.js:

```
mkdir ~/bin
cd ~/bin
wget https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz
wget https://nodejs.org/dist/v6.9.1/node-v6.9.1-linux-x64.tar.xz
tar xvf go1.7.3.linux-amd64.tar.gz
tar xvf node-v6.9.1-linux-x64.tar.xz
mv go golang
mv node-v6.9.1-linux-x64 node-js
ln -s ~/bin/golang/bin/go ./go
ln -s ~/bin/golang/bin/godoc ./godoc
ln -s ~/bin/golang/bin/gofmt ./gofmt
ln -s ~/bin/node-js/bin/node ./node
ln -s ~/bin/node-js/bin/npm ./npm
```

2) Create project path and clone the repository with --recursive:

```
mkdir ~/GoProjects/src/github.com/ansible-semaphore/
cd ~/GoProjects/src/github.com/ansible-semaphore/
git clone --recursive git@github.com:ansible-semaphore/semaphore.git
```

3) Add environment variables to .profile (or .bashrc) need by go and npm (Don't forget to use source on file or reopen the terminal after):

```
PATH=/~/bin:/~/GoProjects/bin:/~/.npm-global/bin:$PATH
GOPATH=~/GoProjects
GOROOT=~/bin/golang
NPM_CONFIG_PREFIX=~/.npm-global
export PATH GOPATH GOROOT NPM_CONFIG_PREFIX
```

4) Clone project and Install go dependencies:

```
cd ~/GoProjects/src/github.com/ansible-semaphore/
git clone --recursive https://github.com/ansible-semaphore/semaphore.git
go get github.com/jteeuwen/go-bindata/...
go get github.com/mitchellh/gox
go get github.com/cespare/reflex
go get -u ./...
```

5)  Install node.js dependencies:

```
cd ~/GoProjects/src/github.com/ansible-semaphore/
npm install -g less pug-cli
npm install async
```

6) OPTIONAL: Create a MySQL Container to develop (or install and setup One):

```
docker run -d --name semaphore-db -v semaphore-data:/var/lib/mysql -p 3306:3306 -e MYSQL_USER=semaphore -e MYSQL_DATABASE=semaphore -e MYSQL_PASSWORD=semaphore -e MYSQL_ROOT_PASSWORD=semaphore mysql
```

7) Download the 2.0.4 version into main directory and run the setup to trigger the db migration:

```
wget https://github.com/ansible-semaphore/semaphore/releases/download/v2.0.4/semaphore_linux_amd64
chmod +x semaphore_linux_amd64
./semaphore_linux_amd64 -setup
```

8) Created a config.json which have the Database information to start the development:

```
cat <<EOT >> config.json
{
    "mysql": {
        "host": "127.0.0.1:3306",
        "user": "semaphore",
        "pass": "semaphore",
        "name": "semaphore"
    },
    "session_db": "127.0.0.1:6379",
    "port": ":8010"
}
EOT
```

9) Start the Watch process:

```
./make.sh watch
```

10) Point your browser to localhost:8010. Enjoy.


# Other Informations

For more information about GoPaths take a look at ([go wiki](https://github.com/golang/go/wiki/GOPATH), [tutorial](http://www.ryanday.net/2012/10/01/installing-go-and-gopath/), [SO question](https://stackoverflow.com/questions/21001387/how-do-i-set-the-gopath-environment-variable-on-ubuntu-what-file-must-i-edit)).


You will need to have a local `config.json` file because it is linked to. It should contain your local configuration.
