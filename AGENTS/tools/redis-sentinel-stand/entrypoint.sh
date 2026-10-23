#!/bin/bash
# Node entrypoint: a local git repo with the test scripts, schema migration (retried until
# PostgreSQL answers), the admin user, then the server. Everything comes from SEMAPHORE_* env.
set -euo pipefail
if [ ! -d /tmp/repo/.git ]; then
  cp -r /repo-src /tmp/repo
  git -C /tmp/repo -c init.defaultBranch=main init -q
  git -C /tmp/repo -c user.email=stand@localhost -c user.name=stand add -A
  git -C /tmp/repo -c user.email=stand@localhost -c user.name=stand commit -q -m init
fi
for _ in $(seq 1 60); do
  if semaphore migrate --no-config >/tmp/migrate.log 2>&1; then break; fi
  sleep 1
done
semaphore user add --admin --login admin --name Admin --email admin@localhost --password admin123 --no-config >/dev/null 2>&1 || true
exec semaphore server --no-config "$@"
