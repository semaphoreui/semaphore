#!/usr/bin/env bash
set -e

HERE="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mkdir -p /etc/forge
cp ${HERE}/../forge.service /etc/systemd/system
cp ${HERE}/../env /etc/forge/env
systemctl daemon-reload
systemctl start forge.service