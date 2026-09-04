#!/usr/bin/env bash
set -euo pipefail

sleep 180 &
echo "Started same-group process with PID $!"

setsid sleep 180 &
echo "Started escaped process with PID $!"

if [[ "${WAIT_FOR_SIGNAL:-}" == "1" ]]; then
    echo "Waiting for a termination signal"
    wait
fi
