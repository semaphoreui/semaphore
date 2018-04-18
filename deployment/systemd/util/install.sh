#!/usr/bin/env bash
set -e

HERE="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mkdir -p /etc/semaphore
cp ${HERE}/../semaphore.service /etc/systemd/system
cp ${HERE}/../env /etc/semaphore/env
systemctl daemon-reload
systemctl start semaphore.service