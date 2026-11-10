# Copyright (c) Semaphore UI contributors.
# SPDX-License-Identifier: MIT
#
# Ansible aggregate callback plugin that writes the per-host counters of a
# playbook run as JSON to the file referenced by the SEMAPHORE_HOSTS_FILE
# environment variable. Semaphore reads that file after the run and stores
# the hosts of the task.
#
# Usage: Semaphore wires this plugin in automatically for Ansible tasks.
from __future__ import absolute_import, division, print_function

__metaclass__ = type

DOCUMENTATION = """
    callback: semaphore_hosts
    type: aggregate
    short_description: Write per-host stats to SEMAPHORE_HOSTS_FILE
    description:
      - Writes the counters of every host (ok, changed, unreachable, failed,
        skipped, rescued, ignored) as JSON to the file pointed to by the
        SEMAPHORE_HOSTS_FILE environment variable.
      - The plugin needs no enabling, so it does not touch the callbacks_enabled
        setting of the playbook's ansible.cfg.
    requirements: []
"""

import json
import os

from ansible.plugins.callback import CallbackBase

COUNTERS = ("ok", "changed", "unreachable", "failed", "skipped", "rescued", "ignored")


class CallbackModule(CallbackBase):
    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = "aggregate"
    CALLBACK_NAME = "semaphore_hosts"

    def __init__(self, *args, **kwargs):
        super(CallbackModule, self).__init__(*args, **kwargs)
        self._path = os.environ.get("SEMAPHORE_HOSTS_FILE")

    def v2_playbook_on_stats(self, stats):
        if not self._path:
            return
        hosts = {}
        for host in sorted(stats.processed.keys()):
            summary = stats.summarize(host)
            # Ansible calls the failed counter "failures"
            summary["failed"] = summary.get("failures", 0)
            hosts[host] = {key: summary.get(key, 0) for key in COUNTERS}
        try:
            with open(self._path, "w") as fh:
                json.dump({"version": 1, "hosts": hosts}, fh)
        except OSError as e:
            self._display.warning("Could not write the Semaphore hosts file: %s" % e)
