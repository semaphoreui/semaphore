#!/usr/bin/env bash
# Always succeeds: gives the task history a success.
echo "seed: ok"
echo "SEED_ENV=${SEED_ENV:-<unset>}"
if [ -n "${SEED_TOKEN:-}" ]; then echo "SEED_TOKEN is set"; else echo "SEED_TOKEN is unset"; fi
