#!/bin/bash

BINDATA_ARGS="-o util/bindata.go -pkg util"

if [ "$1" == "dev" ]; then
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

go-bindata $BINDATA_ARGS config.json database/sql_migrations/ $(find ./public -type d -print)

echo "Building into build/"

mkdir -p build
GOOS=linux GOARCH=amd64 go build -o build/amd64 main.go
GOOS=windows GOARCH=?? go build -o build/windows main.go
GOOS=macos GOARCH=darwin go build -o build/darwin main.go

chmod +x build/*

echo "Build finished"
