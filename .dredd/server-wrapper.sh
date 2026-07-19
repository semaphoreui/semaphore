#!/bin/sh

export SEMAPHORE_MAX_TASKS_PER_TEMPLATE=300
export SEMAPHORE_APPS='{"ansible": {}}'
./semaphore server --config .dredd/config.json