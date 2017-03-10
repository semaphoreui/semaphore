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

	cd public
	lessc css/semaphore.less > css/semaphore.css
	pug html/*.pug html/*/*.pug html/*/*/*.pug
	cd -
fi

cd public
node ./bundler.js
cd -

echo "Adding bindata"

if [ "$1" == "release" ]; then
	VERSION=$2
	cat <<HEREDOC > util/version.go
package util

var Version string = "$VERSION"

HEREDOC

	echo "Updating changelog:"
	git changelog -t "v$VERSION"

	echo "\nCommitting version.go and changelog update"
	git add util/version.go CHANGELOG.md && git commit -m "update changelog, bump version to $VERSION"
	echo "\nTagging release"
	git tag -m "v$VERSION release" "v$VERSION"
	echo "\nPushing to repository"
	git push origin master "v$VERSION"
	echo "\nCreating draft release v$VERSION"
	github-release release --draft -u ansible-semaphore -r semaphore -t "v$VERSION" -d "## Special thanks to\n\n## Installation\n\nFollow [wiki/Installation](https://github.com/ansible-semaphore/semaphore/wiki/Installation)\n\n## Changelog"
fi

go-bindata $BINDATA_ARGS config.json db/migrations/ $(find public/* -type d -print)

if [ "$1" == "ci_test" ]; then
	exit 0
fi

if [ "$1" == "watch" ]; then
	cd public

	nodemon -w js -i bundle.js -e js bundler.js &
	nodemon -w css -e less --exec "lessc css/semaphore.less > css/semaphore.css" &
	pug -w -P html/*.pug html/*/*.pug html/*/*/*.pug &

	cd -
	reflex -r '\.go$' -R '^public/vendor/' -R '^node_modules/' -s -d none -- sh -c 'go build -i -o /tmp/semaphore_bin cli/main.go && /tmp/semaphore_bin'
	
	exit 0
fi

cd cli
gox -output="semaphore_{{.OS}}_{{.Arch}}" ./...

if [ "$CIRCLE_ARTIFACTS" != "" ]; then
	rsync -a semaphore_* $CIRCLE_ARTIFACTS/
	exit 0
fi

if [ "$1" == "release" ]; then
	echo "Uploading files.."
	VERSION=$2 find . -name "semaphore_*" -exec sh -c 'github-release upload -u ansible-semaphore -r semaphore -t "v$VERSION" -n "${1/.\/}" -f "$1"' _ {} \;
	echo "Done"
	rm -f semaphore_*
fi