#!/bin/sh

echoerr() { printf "%s\n" "$*" >&2; }

# wait on db to be up
echoerr "Attempting to connect to database ${SEMAPHORE_DB} on ${SEMAPHORE_DB_HOST} with user:pass ${SEMAPHORE_DB_USER}:${SEMAPHORE_DB_PASS}"
until mysql -h ${SEMAPHORE_DB_HOST} -u ${SEMAPHORE_DB_USER} --password=${SEMAPHORE_DB_PASS} ${SEMAPHORE_DB} -e "select version();" &>/dev/null;
do
    echoerr "waiting";
    sleep 3;
done

# generate stdin
if [ -f ${SEMAPHORE_PLAYBOOK_PATH}/config.stdin ]
then
    echoerr "already generated stdin"
else
    echoerr "generating ${SEMAPHORE_PLAYBOOK_PATH}/config.stdin"
    cat << EOF > ${SEMAPHORE_PLAYBOOK_PATH}/config.stdin
${SEMAPHORE_DB_HOST}:${SEMAPHORE_DB_PORT}
${SEMAPHORE_DB_USER}
${SEMAPHORE_DB_PASS}
${SEMAPHORE_DB}
${SEMAPHORE_PLAYBOOK_PATH}
yes
${SEMAPHORE_ADMIN}
${SEMAPHORE_ADMIN_EMAIL}
${SEMAPHORE_ADMIN_NAME}
${SEMAPHORE_ADMIN_PASSWORD}
n
EOF
fi

# test to see if initialzation is needed
if [ -f ${SEMAPHORE_PLAYBOOK_PATH}/semaphore_config.json ]
then
    echoerr "already initialized"
else
    echoerr "Initializing semaphore"
    /usr/bin/semaphore -setup < ${SEMAPHORE_PLAYBOOK_PATH}/config.stdin
fi

# run our command
exec "$@"
