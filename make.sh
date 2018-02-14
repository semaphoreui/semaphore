#!/bin/bash
set -e

# Overridable variables
BINDATA_ARGS="${BINDATA_ARGS:--o util/bindata.go -pkg util }"
DELVE_HOST="${DELVE_HOST:-localhost}"
DELVE_PORT="${DELVE_PORT:-57780}"
WATCH_STARTUP_CMD="${WATCH_STARTUP_CMD:-/tmp/semaphore}"

# run node bundle
function bundleJs {
    cd public
        node ./bundler.js
    cd -
}

# runs go-bindata to compile the assets
# excludes the node_modules folder and package.json
function createBinData {
    echo "Adding bindata"
    if [ "$1" == "watch" ]; then
	    BINDATA_ARGS="-debug ${BINDATA_ARGS}"
	    echo "Creating util/bindata.go with file proxy"
    else
	    echo "Creating util/bindata.go"
    fi
    go-bindata -ignore "/package.*/" $BINDATA_ARGS config.json db/migrations/ $(find public/* -type d -print | grep -v "node_modules")
}


# Handoff dev tasks to the dev make.sh file.
# Pass in keywords that are listed in that script
if [ "$1" == "dev" ]; then
    ./images/dev/make.sh "${@:2}"
    exit $?
fi


# Install dependencies of the project via node and dep
# Also installs required go tools
if [ "$1" == "deps" ]; then
    # FE Deps
    pushd public
        npm install
    popd
    # BE Deps
    dep ensure -vendor-only
    # BE Tools
    # This kind of sucks as dep has no way to install tools currently
    # As this uses go install be aware that binaries will end up in your $GOBIN even though the deps are vendored
    sed -n -e '/^required/p' Gopkg.toml | cut -d "[" -f2 | cut -d "]" -f1  | tr ',' '\n' | while read -r package; do
        pushd ./vendor/"$(sed -e 's/^"//' -e 's/"$//' <<<"$package")"
            go install
        popd
    done
    exit 0
fi

if [ "$1" == "deps-update" ]; then
    dep ensure
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
	"port": ":8010",
	"email_alert": false
}
EOF
    ./make.sh deps
	pushd public
	    PATH=$(npm bin):$PATH lessc css/semaphore.less > css/semaphore.css
	    PATH=$(npm bin):$PATH pug $(find ./html/ -name "*.pug")
	popd
fi


if [ "$1" == "compile-assets" ]; then
    bundleJs
    createBinData "$@"
	exit 0
fi


if [ "$1" == "ci_test" ]; then
    ./make.sh compile-assets
	exit 0
fi

# Compile assets and start watchers
if [ "$1" == "watch" ]; then
    ./make.sh deps
    ./make.sh compile-assets
    pushd public
	    PATH=$(npm bin):$PATH nodemon -w js -i bundle.js -e js bundler.js &
	    PATH=$(npm bin):$PATH nodemon -w css -e less --exec "lessc css/semaphore.less > css/semaphore.css" &
	    PATH=$(npm bin):$PATH pug -w -P --doctype html $(find ./html/ -name "*.pug") &
	popd

	reflex -r '\.go$' -R '^public/vendor/' -R '^node_modules/' -s -d none -- sh -c "go build -i -o /tmp/semaphore cli/main.go && ${WATCH_STARTUP_CMD}"
	exit 0
fi

# Run is a full package command, fetching deps, compiling assets, creating the application and running it
if [ "$1" == "run" ]; then
    ./make.sh deps
    ./make.sh compile-assets
    go build -a -o /tmp/semaphore_bin cli/main.go
    images/common/semaphore-startup.sh /tmp/semaphore_bin -config /etc/semaphore/semaphore_config.json
    exit $?
fi


# Runs a local delve server on a debug build of semaphore
if [ "$1" == "debug" ]; then
    pushd cli
        dlv debug --headless --listen="${DELVE_HOST}:${DELVE_PORT}" --api-version=2
    popd
    exit $?
fi


if [ "$1" == "release" ]; then
    bundleJs
	VERSION=$2
	cat <<HEREDOC > util/version.go
package util

var Version string = "$VERSION"

HEREDOC

	echo "Updating changelog:"
	set +e
	git changelog -t "v$VERSION"
	set -e

	echo "\nCommitting version.go and changelog update"
	git add util/version.go CHANGELOG.md && git commit -m "update changelog, bump version to $VERSION"
	echo "\nTagging release"
	git tag -m "v$VERSION release" "v$VERSION"
	echo "\nPushing to repository"
	git push origin develop "v$VERSION"
	echo "\nCreating draft release v$VERSION"
	github-release release --draft -u ansible-semaphore -r semaphore -t "v$VERSION" -d "## Special thanks to\n\n## Installation\n\nFollow [wiki/Installation](https://github.com/ansible-semaphore/semaphore/wiki/Installation)\n\n## Changelog"
fi


./make.sh compile-assets
cd cli
gox -output="semaphore_{{.OS}}_{{.Arch}}" ./...

if [ "$CIRCLE_ARTIFACTS" != "" ]; then
	rsync -a semaphore_* "$CIRCLE_ARTIFACTS"/
	exit 0
fi

if [ "$1" == "release" ]; then
	echo "Uploading files.."
	find . -name "semaphore_*" -exec sh -c 'gpg --armor --detach-sig "$1"' _ {} \;
	VERSION=$2 find . -name "semaphore_*" -exec sh -c 'github-release upload -u ansible-semaphore -r semaphore -t "v$VERSION" -n "${1/.\/}" -f "$1"' _ {} \;
	echo "Done"
	rm -f semaphore_*
fi
