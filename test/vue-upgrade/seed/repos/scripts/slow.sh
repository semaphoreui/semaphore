#!/usr/bin/env bash
# Runs for SEED_SLEEP seconds (default 30): a task to stop or watch while it runs. Not run by the seed.
for i in $(seq 1 "${SEED_SLEEP:-30}"); do
  echo "seed: tick $i"
  sleep 1
done
