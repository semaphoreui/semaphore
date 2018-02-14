#!/bin/sh

set -e

echoerr() { printf "%s\n" "$*" >&2; }

SEMAPHORE_CONFIG_PATH="${SEMAPHORE_CONFIG_PATH:-/etc/semaphore}"

SEMAPHORE_TMP_PATH="${SEMAPHORE_TMP_PATH:-/tmp/semaphore_data}"
# Semaphore database env config
SEMAPHORE_DB_HOST="${SEMAPHORE_DB_HOST:-127.0.0.1}"
SEMAPHORE_DB_PORT="${SEMAPHORE_DB_PORT:-3306}"
SEMAPHORE_DB="${SEMAPHORE_DB:-semaphore}"
SEMAPHORE_DB_USER="${SEMAPHORE_DB_USER:-semaphore}"
SEMAPHORE_DB_PASS="${SEMAPHORE_DB_PASS:-semaphore}"
# Email alert env config
SEMAPHORE_WEB_ROOT="${SEMAPHORE_WEB_ROOT:-http://127.0.0.1:8081}"
# Semaphore Admin env config
SEMAPHORE_ADMIN="${SEMAPHORE_ADMIN:-admin}"
SEMAPHORE_ADMIN_EMAIL="${SEMAPHORE_ADMIN_EMAIL:-admin@localhost}"
SEMAPHORE_ADMIN_NAME="${SEMAPHORE_ADMIN_NAME:-Semaphore Admin}"
SEMAPHORE_ADMIN_PASSWORD="${SEMAPHORE_ADMIN_PASSWORD:-semaphorepassword}"
#Semaphore LDAP env config
SEMAPHORE_LDAP_ACTIVATED="${SEMAPHORE_LDAP_ACTIVATED:-no}"
SEMAPHORE_LDAP_HOST="${SEMAPHORE_LDAP_HOST:-}"
SEMAPHORE_LDAP_PORT="${SEMAPHORE_LDAP_PORT:-}"
SEMAPHORE_LDAP_NEEDTLS="${SEMAPHORE_LDAP_NEEDTLS:-no}"
SEMAPHORE_LDAP_DN_BIND="${SEMAPHORE_LDAP_DN_BIND:-}"
SEMAPHORE_LDAP_PASSWORD="${SEMAPHORE_LDAP_PASSWORD:-}"
SEMAPHORE_LDAP_DN_SEARCH="${SEMAPHORE_LDAP_DN_SEARCH:-}"
SEMAPHORE_LDAP_SEARCH_FILTER="${SEMAPHORE_LDAP_SEARCH_FILTER:-(uid=%s)}"
SEMAPHORE_LDAP_MAPPING_DN="${SEMAPHORE_LDAP_MAPPING_DN:-dn}"
SEMAPHORE_LDAP_MAPPING_USERNAME="${SEMAPHORE_LDAP_MAPPING_USERNAME:-uid}"
SEMAPHORE_LDAP_MAPPING_FULLNAME="${SEMAPHORE_LDAP_MAPPING_FULLNAME:-cn}"
SEMAPHORE_LDAP_MAPPING_EMAIL="${SEMAPHORE_LDAP_MAPPING_EMAIL:-mail}"

# create semaphore temporary directory if non existent
[ -d "${SEMAPHORE_TMP_PATH}" ] || mkdir -p "${SEMAPHORE_TMP_PATH}" || {
    echo "Can't create Semaphore tmp path ${SEMAPHORE_TMP_PATH}."
    exit 1
}
# create semaphore config directory if non existent
[ -d "${SEMAPHORE_CONFIG_PATH}" ] || mkdir -p "${SEMAPHORE_CONFIG_PATH}" || {
    echo "Can't create Semaphore Config path ${SEMAPHORE_CONFIG_PATH}."
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

# Create a config if it does not exist in the current config path
if [ ! -f "${SEMAPHORE_CONFIG_PATH}/semaphore_config.json" ]; then
    echoerr "Generating ${SEMAPHORE_TMP_PATH}/config.stdin ..."
    cat << EOF > "${SEMAPHORE_TMP_PATH}/config.stdin"
${SEMAPHORE_DB_HOST}:${SEMAPHORE_DB_PORT}
${SEMAPHORE_DB_USER}
${SEMAPHORE_DB_PASS}
${SEMAPHORE_DB}
${SEMAPHORE_TMP_PATH}
${SEMAPHORE_WEB_ROOT}
no
no
${SEMAPHORE_LDAP_ACTIVATED}
EOF

    if [ "${SEMAPHORE_LDAP_ACTIVATED}" = "yes" ]; then
        cat << EOF >> "${SEMAPHORE_TMP_PATH}/config.stdin"
${SEMAPHORE_LDAP_HOST}:${SEMAPHORE_LDAP_PORT}
${SEMAPHORE_LDAP_NEEDTLS}
${SEMAPHORE_LDAP_DN_BIND}
${SEMAPHORE_LDAP_PASSWORD}
${SEMAPHORE_LDAP_DN_SEARCH}
${SEMAPHORE_LDAP_SEARCH_FILTER}
${SEMAPHORE_LDAP_MAPPING_DN}
${SEMAPHORE_LDAP_MAPPING_USERNAME}
${SEMAPHORE_LDAP_MAPPING_FULLNAME}
${SEMAPHORE_LDAP_MAPPING_EMAIL}
EOF
    fi;

    cat << EOF >> "${SEMAPHORE_TMP_PATH}/config.stdin"
yes
${SEMAPHORE_ADMIN}
${SEMAPHORE_ADMIN_EMAIL}
${SEMAPHORE_ADMIN_NAME}
${SEMAPHORE_ADMIN_PASSWORD}
EOF

    cat "${SEMAPHORE_TMP_PATH}/config.stdin"
    $1 -setup - < "${SEMAPHORE_TMP_PATH}/config.stdin"


    echoerr "Moving config file to non temporary path ${SEMAPHORE_CONFIG_PATH}/semaphore_config.json"
    mv  "${SEMAPHORE_TMP_PATH}/semaphore_config.json" ${SEMAPHORE_CONFIG_PATH}/semaphore_config.json 2>/dev/null || true
    echoerr "Run Semaphore with semaphore -config ${SEMAPHORE_CONFIG_PATH}/semaphore_config.json"
fi

# run our command
exec "$@"
