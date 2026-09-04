#!/usr/bin/env bash
set -euo pipefail

sleep 120 &
echo "Started same-group process with PID $!"

setsid sleep 120 &
echo "Started escaped process with PID $!"
