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

cd public
node ./bundler.js
cd -

echo "Adding bindata"

go-bindata $BINDATA_ARGS config.json database/sql_migrations/ $(find ./public -type d -print)

if [ "$1" == "ci_test" ]; then
	exit 0
fi

if [ "$1" == "watch" ]; then
	cd public

	nodemon -w js -i bundle.js -e js bundler.js &
	nodemon -w css -e less --exec "lessc css/semaphore.less > css/semaphore.css" &
	jade -w -P html/*.jade html/*/*.jade html/*/*/*.jade &

	cd ../
	reflex -r '\.go$' -s -d none -- sh -c 'go run main.go'
	exit 0
fi

gox -os="linux darwin windows openbsd" ./...

if [ "$CIRCLE_ARTIFACTS" != "" ]; then
	rsync -a semaphore_* $CIRCLE_ARTIFACTS/
	exit 0
fi