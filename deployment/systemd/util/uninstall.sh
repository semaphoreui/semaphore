#!/usr/bin/env bash
set -e

systemctl stop semaphore.service
systemctl disable semaphore.service
rm /etc/systemd/system/semaphore.service
rm -rf /etc/semaphore