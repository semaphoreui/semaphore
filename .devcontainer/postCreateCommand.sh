go install github.com/go-task/task/v3/cmd/task@latest

(cd ./web && npm install)

task build

./bin/semaphore user add \
    --admin \
    --login user123 \
    --name User123 \
    --email user123@example.com \
    --password 123456 \
    --config ./.devcontainer/config.json