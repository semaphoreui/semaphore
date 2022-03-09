#!/bin/bash -l

SCRIPT_HOME=$(realpath $(dirname $0))
cd $SCRIPT_HOME/../

SINGULARITY_BIND=/home/jhayes/projects/semaphore:/mnt

sudo apptainer build -B $SINGULARITY_BIND semaphore.simg Singularity

