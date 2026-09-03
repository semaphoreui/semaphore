#!/usr/bin/env bash
set -euo pipefail

sleep 120 &
child_pid=$!

echo "Started 120-second background process with PID $child_pid"
