#!/usr/bin/env bash
set -euo pipefail

sleep 60 &
child_pid=$!

echo "Started 60-second background process with PID $child_pid"
echo "To stop it manually, run: kill $child_pid"
