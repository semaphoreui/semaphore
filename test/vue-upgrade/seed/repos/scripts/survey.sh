#!/usr/bin/env bash
# Prints its arguments and the survey variables passed as environment variables.
echo "seed: args: $*"
echo "TARGET=${TARGET:-<unset>}"
echo "REPLICAS=${REPLICAS:-<unset>}"
