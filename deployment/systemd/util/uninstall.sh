#!/usr/bin/env bash
set -e

systemctl stop forge.service
systemctl disable forge.service
rm /etc/systemd/system/forge.service
rm -rf /etc/forge