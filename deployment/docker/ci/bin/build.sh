#!/bin/bash -l

# Set env
SUSER=$USER
SCRIPT_HOME=$(realpath $(dirname $0))
SEMAPHORE_HOME=$(realpath $SCRIPT_HOME/../../../../)
echo "Building in $SEMAPHORE_HOME"

if [[ ! -f $SCRIPT_HOME/../semaphore.simg ]]; then
    # Build Image
    sudo bash -lc "
    # Set env
    module purge
    module load singularity || echo 'Not using modules'
    export SINGULARITY_BIND=${SEMAPHORE_HOME}:/mnt
    
    # Build image
    cd $SCRIPT_HOME/../ && singularity build semaphore.simg Singularity

    # Update perms
    chmod -R ${SUSER}: $SEMAPHORE_HOME"
else
    # Build RPM
    module load singularity || echo 'Not using modules'
    cd ${SEMAPHORE_HOME} && singularity exec $SCRIPT_HOME/../semaphore.simg task release
    chmod a+rx ${SEMAPHORE_HOME}/bin
    chmod a+r ${SEMAPHORE_HOME}/bin/*
fi

