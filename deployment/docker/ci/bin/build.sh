#!/bin/bash -l

# Set env
SCRIPT_HOME=$(realpath $(dirname $0))
SEMAPHORE_HOME=$(realpath $SCRIPT_HOME/../../../../)
echo "Building in $SEMAPHORE_HOME"

# Build
sudo bash -lc "
    # Set env
    module purge
    module load singularity || echo 'Not using modules'
    export SINGULARITY_BIND=${SEMAPHORE_HOME}:/mnt
    
    # Build image
    cd $SCRIPT_HOME/../ && singularity build semaphore.simg Singularity || exit

    # Build RPM
    cd ${SEMAPHORE_HOME} && singularity exec semaphore.simg task release
"

