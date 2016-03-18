#!/bin/bash
set -e

BINDATA_ARGS="-o util/bindata.go -pkg util"

if [ "$1" == "watch" ]; then
	BINDATA_ARGS="-debug ${BINDATA_ARGS}"
	echo "Creating util/bindata.go with file proxy"
else
	echo "Creating util/bindata.go"
fi

if [ "$1" == "ci_test" ]; then
	echo "Creating CI Test config.json"

	cat > config.json <<EOF
{
	"mysql": {
		"host": "127.0.0.1:3306",
		"user": "ubuntu",
		"pass": "",
		"name": "circle_test"
	},
	"session_db": "127.0.0.1:6379",
	"port": ":8010"
}
EOF
fi

if [ "$1" == "production" ]; then
	cd public

	# html
	cd html
	jade *.jade */*.jade */*/*.jade
	cd -

	# css
	cd css
	lessc --clean-css="--s1 --advanced --compatibility=ie8" --autoprefix="ie 8-10,iOS 7," semaphore.less > semaphore.css
	cd -
fi

cd public
node ./bundler.js
cd -

echo "Adding bindata"

go-bindata $BINDATA_ARGS config.json database/sql_migrations/ $(find ./public -type d -print)

if [ "$1" == "watch" ]; then
	cd public

	nodemon -w js -i bundle.js -e js bundler.js &
	nodemon -w css -e less --exec "lessc css/semaphore.less > css/semaphore.css" &
	jade -w -P html/*.jade html/*/*.jade html/*/*/*.jade
fi

echo ""

mkdir -p build
echo "build/darwin_amd64"
GOOS=darwin GOARCH=amd64 go build -o build/darwin_amd64 main.go
echo "build/linux_386"
GOOS=linux GOARCH=386 go build -o build/linux_386 main.go
echo "build/linux_amd64"
GOOS=linux GOARCH=amd64 go build -o build/linux_amd64 main.go
echo "build/linux_arm"
GOOS=linux GOARCH=arm go build -o build/linux_arm main.go
# echo "build/windows_386"
# GOOS=windows GOARCH=386 go build -o build/windows_386 main.go
# echo "build/windows_amd64"
# GOOS=windows GOARCH=amd64 go build -o build/windows_amd64 main.go

chmod +x build/*

echo "Build finished"