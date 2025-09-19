#!/bin/sh

export SEMAPHORE_MAX_TASKS_PER_TEMPLATE=300
./forge server --config .dredd/config.json