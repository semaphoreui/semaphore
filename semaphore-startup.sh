#!/bin/sh

echoerr() { printf "%s\n" "$*" >&2; }

SEMAPHORE_PLAYBOOK_PATH="${SEMAPHORE_PLAYBOOK_PATH:-/semaphore}"
# Semaphore database env config
SEMAPHORE_DB_HOST="${SEMAPHORE_DB_HOST:-127.0.0.1}"
SEMAPHORE_DB_PORT="${SEMAPHORE_DB_PORT:-3306}"
SEMAPHORE_DB="${SEMAPHORE_DB:-semaphore}"
SEMAPHORE_DB_USER="${SEMAPHORE_DB_USER:-semaphore}"
SEMAPHORE_DB_PASS="${SEMAPHORE_DB_PASS:-semaphore}"
# Semaphore Admin env config
SEMAPHORE_ADMIN="${SEMAPHORE_ADMIN:-admin}"
SEMAPHORE_ADMIN_EMAIL="${SEMAPHORE_ADMIN_EMAIL:-admin@localhost}"
SEMAPHORE_ADMIN_NAME="${SEMAPHORE_ADMIN_NAME:-Semaphore Admin}"
SEMAPHORE_ADMIN_PASSWORD="${SEMAPHORE_ADMIN_PASSWORD:-semaphorepassword}"

# create semaphore playbook directory
mkdir -p "${SEMAPHORE_PLAYBOOK_PATH}" || {
    echo "Can't create Semaphore playbook path '$SEMAPHORE_PLAYBOOK_PATH'."
    exit 1
}

# wait on db to be up
echoerr "Attempting to connect to database ${SEMAPHORE_DB} on ${SEMAPHORE_DB_HOST}:${SEMAPHORE_DB_PORT} with user ${SEMAPHORE_DB_USER} ..."
TIMEOUT=30
while ! mysqladmin ping -h"$SEMAPHORE_DB_HOST" -P "$SEMAPHORE_DB_PORT" -u "$SEMAPHORE_DB_USER" --password="$SEMAPHORE_DB_PASS" --silent >/dev/null 2>&1; do
    TIMEOUT=$(expr $TIMEOUT - 1)
    if [ $TIMEOUT -eq 0 ]; then
        echoerr "Could not connect to database server. Exiting."
        exit 1
    fi
    echo -n "."
    sleep 1
done

if [ -f "${SEMAPHORE_PLAYBOOK_PATH}/semaphore_config.json" ]; then
    ln -s "${SEMAPHORE_PLAYBOOK_PATH}/semaphore_config.json" /etc/semaphore/semaphore_config.json
fi
if [ ! -f /etc/semaphore/semaphore_config.json ]; then
    echoerr "Generating ${SEMAPHORE_PLAYBOOK_PATH}/config.stdin ..."
    cat << EOF > "${SEMAPHORE_PLAYBOOK_PATH}/config.stdin"
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
EOF
    /usr/bin/semaphore -setup < "${SEMAPHORE_PLAYBOOK_PATH}/config.stdin"
fi

# run our command
exec "$@"
