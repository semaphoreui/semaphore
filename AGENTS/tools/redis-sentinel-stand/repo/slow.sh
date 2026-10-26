#!/bin/bash
# Prints one line per second for N seconds (default 30), so a task is observably running while
# the Redis master is killed underneath the cluster.
N=${1:-30}
for i in $(seq 1 "$N"); do echo "tick $i/$N $(date -u +%T)"; sleep 1; done
echo "slow done"
