#!/usr/bin/env bash

export SEMAPHORE_MAX_TASKS_PER_TEMPLATE=300
export SEMAPHORE_APPS='{"ansible": {}}'
export SEMAPHORE_PORT=58427

semaphore=./semaphore
[[ -x "$semaphore" ]] || semaphore=./bin/semaphore

exec "$semaphore" server --config .dredd/config.json
