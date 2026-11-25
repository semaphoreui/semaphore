---
form: src/views/Cluster.vue
---

# Clear tasks from Redis

The nodes of a cluster share the state of the task pool in Redis: which tasks wait, which run, and
which node owns each of them. This dialog deletes the groups of records you tick from Redis, to
recover a cluster that is stuck, for example with a finished task that still shows as running.
Nothing is deleted from the database: tasks keep their status, their log and their history.

Whatever you tick, every node also drops the tasks it keeps in memory and reads them again from
Redis and the database. Only **Queue** is ticked when the dialog opens; tick only the groups that
are stuck. The result shows how many records were removed.

## Queue {#clusterClearDialog-queue}

The tasks that wait for a node to start them. Cleared tasks are not started: they stay in the
database with the status Waiting. Stop them, or run their templates again.

## Running {#clusterClearDialog-running}

The tasks the cluster counts as running. Clearing does not stop them: a task keeps running on its
runner. But the cluster loses track of it: the task no longer counts against the limits of
parallel tasks of the server and of its runner, so more tasks may start than the limits allow,
and a task that a runner has not picked up yet is no longer handed to it.

## Active by project {#clusterClearDialog-active}

The unfinished tasks of each project. They enforce the limit of parallel tasks of a project and
keep a second task of a template from starting while one runs, when the template does not allow
parallel tasks. Clear this group when the tasks of a project wait forever behind tasks that have
finished. Tasks that run at that moment no longer count, so new tasks of the project may start
beside them.

## Aliases {#clusterClearDialog-aliases}

The random names Terraform, OpenTofu and Terragrunt tasks get for their lifetime, mapped to the
task. A running task reaches its Terraform state on the server by this name: after clearing, a
running task of these apps can no longer be found by it.

## Claims {#clusterClearDialog-claims}

Which node owns each task. A node renews the claims of its tasks every minute; a claim that is not
renewed expires after five minutes. A task without a claim counts as orphaned: a task that has
not reached a runner yet is put back in the queue for another node, and a task a runner already
runs is left to finish there.

## Runtime fields {#clusterClearDialog-runtimeFields}

The details of each task that the database does not hold: the user who started it, the version it
was started with, its alias, its runner and its status, and the project of each task. Without
them, a node that did not start a task shows less about it, and the cleaner of orphaned tasks
skips a running task, so a stuck record is no longer removed by itself.
