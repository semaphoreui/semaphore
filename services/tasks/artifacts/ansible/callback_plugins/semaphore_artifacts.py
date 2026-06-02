# Copyright (c) Semaphore UI contributors.
# SPDX-License-Identifier: MIT
#
# Ansible aggregate callback plugin that captures `set_stats` data from a
# play and writes it as a JSON object to the file referenced by the
# SEMAPHORE_ARTIFACTS_FILE environment variable. Semaphore reads that file
# after the task exits and forwards the contents as workflow-level extra
# vars to downstream task templates in the same WorkflowRun, mirroring
# AWX's "workflow artifacts" behaviour.
#
# Usage: Semaphore wires this plugin in automatically for Ansible templates;
# users only need `set_stats` in their playbook to pass values downstream.
from __future__ import absolute_import, division, print_function

__metaclass__ = type

DOCUMENTATION = """
    callback: semaphore_artifacts
    type: aggregate
    short_description: Persist set_stats output to SEMAPHORE_ARTIFACTS_FILE
    description:
      - Aggregates custom stats produced by Ansible's set_stats module and
        writes them as JSON to the file pointed to by the
        SEMAPHORE_ARTIFACTS_FILE environment variable.
      - Only aggregate (non per-host) stats are persisted, mirroring AWX's
        default workflow artifact behaviour.
    requirements: []
"""

import json
import os

from ansible.plugins.callback import CallbackBase


def _is_jsonable(value):
    try:
        json.dumps(value)
        return True
    except (TypeError, ValueError):
        return False


class CallbackModule(CallbackBase):
    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = "aggregate"
    CALLBACK_NAME = "semaphore_artifacts"
    CALLBACK_NEEDS_WHITELIST = True

    def __init__(self, *args, **kwargs):
        super(CallbackModule, self).__init__(*args, **kwargs)
        self._path = os.environ.get("SEMAPHORE_ARTIFACTS_FILE")

    def _safe_stats(self, custom_stats):
        # custom_stats is a dict of {scope: {key: value}}; we only forward the
        # play-level "_run" scope (aggregate stats), matching AWX semantics.
        run_stats = {}
        if isinstance(custom_stats, dict):
            run_stats = custom_stats.get("_run", {}) or {}
        cleaned = {}
        if isinstance(run_stats, dict):
            for key, value in run_stats.items():
                if not isinstance(key, str):
                    continue
                if _is_jsonable(value):
                    cleaned[key] = value
        return cleaned

    def _write(self, data):
        if not self._path or not data:
            return
        try:
            os.makedirs(os.path.dirname(self._path), exist_ok=True)
        except OSError:
            pass
        try:
            with open(self._path, "w") as fh:
                json.dump(data, fh)
        except OSError:
            # The Semaphore server logs a warning if it cannot read the file
            # afterwards; nothing useful to do from inside Ansible.
            pass

    def v2_playbook_on_stats(self, stats):
        # Newer Ansible exposes stats.custom; fall back gracefully on older
        # versions so installations that do not use set_stats are unaffected.
        custom = getattr(stats, "custom", None)
        if not custom:
            return
        data = self._safe_stats(custom)
        if data:
            self._write(data)
