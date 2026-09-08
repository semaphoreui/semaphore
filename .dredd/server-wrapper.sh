#!/usr/bin/env bash

export SEMAPHORE_MAX_TASKS_PER_TEMPLATE=300
export SEMAPHORE_APPS='{"ansible": {}}'

semaphore=./semaphore
[[ -x "$semaphore" ]] || semaphore=./bin/semaphore

"$semaphore" server --config .dredd/config.json
