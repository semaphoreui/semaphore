#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ "$1" == "build" ]; then
    docker-compose -f "${DIR}"/docker-compose.yml build semaphore_dev
fi

# If you use this you can call anything from the make.sh script in root inside a docker image
# ie. ./make.sh dev make run
if [ "$1" == "make" ]; then
    docker run -it ansible-semaphore/semaphore:test ./make.sh "$2"
fi

if [ "$1" == "cmd" ]; then
    docker run -it ansible-semaphore/semaphore:test /bin/bash
fi

if [ "$1" == "connect" ]; then
    docker exec -it dev_semaphore_dev_1 /bin/ash
fi

if [ "$1" == "up" ]; then
    docker-compose -f "${DIR}"/docker-compose.yml build semaphore_dev
    docker-compose -f "${DIR}"/docker-compose.yml up --force-recreate
fi

if [ "$1" == "start" ]; then
    docker-compose -f "${DIR}"/docker-compose.yml up
fi

if [ "$1" == "stop" ]; then
    docker-compose -f "${DIR}"/docker-compose.yml stop
fi

if [ "$1" == "down" ]; then
    docker-compose -f "${DIR}"/docker-compose.yml down
fi

if [ "$1" == "clean" ]; then
    docker-compose -f "${DIR}"/docker-compose.yml down
    rm -rf vendor
    rm -rf public/node_modules
    docker rmi -f ansible-semaphore/sempahore:test
fi

if [ "$1" == "generate-config" ]; then
    cat > "${APP_ROOT}config.json" <<EOF
{
	"mysql": {
		"host": "mysql:3306",
		"user": "semaphore",
		"pass": "semaphore",
		"name": "semaphore"
	},
	"port": ":3000"
}
EOF
fi