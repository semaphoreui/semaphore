---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Edit Schedule

A schedule starts a task of a task template without anyone pressing **Run**: again and again by
a cron expression, or once at a set moment. The task gets the parameters saved in this dialog.
All times are in the timezone of the Semaphore server, shown next to the time fields: `UTC`
unless the administrator sets `schedule.timezone` or `SEMAPHORE_SCHEDULE_TIMEZONE`.

## Name {#scheduleForm-name}

The name the schedule is listed under on the Schedule page. The field is required. It does not
change what the schedule does.

## Template {#scheduleForm-templateId}

The task template the schedule starts. The field is required. The task runs as if it was
started from the template, with the template's own inventory, repository, variable group and
other settings, and with the parameters below where the template lets a task change them.

## Task parameters {#scheduleForm-taskParams}

Shown when a template is selected.

The parameters of every task the schedule starts. Which fields appear depends on the template:
its survey variables, and the options it allows a task to change, such as the branch, the
inventory, **Limit**, **Tags**, **Skip tags** or **CLI args**, and the options of
Ansible or of Terraform, OpenTofu and Terragrunt.

The server checks the template every time the schedule fires: a value saved here is ignored
once the template no longer allows a task to change it.

## Message {#taskParams-message}

Optional. A note stored with every task the schedule starts and shown in the task history, for
example `Nightly deploy`. It does not change how the task runs.

## Survey variables {#taskParams-survey}

Shown when the template has survey variables.

The values of the template's survey variables for the scheduled tasks. They reach the task the
same way as when someone fills in the survey by hand. A required variable must have a value
before the schedule can be saved.

The defaults of the template are copied in when the dialog opens and are saved with the
schedule. A default changed in the template later does not reach a schedule saved before.

The value of a variable of the type secret is not saved with the schedule: the scheduled task
runs without it.

## Branch / Tag {#taskParams-gitBranch}

Shown when the template allows a task to override the branch.

The branch or the tag of the repository the scheduled task checks out. The field starts with the
branch of the template. The server uses the value only while the template allows the override.

## Inventory {#taskParams-inventoryId}

Shown when the template allows a task to choose the inventory.

The inventory the scheduled task runs against, in place of the inventory of the template. For a
Terraform or OpenTofu template the field is called **Workspace**, but the server applies the
choice to Ansible templates only and keeps the workspace of the template for the others.

## Debug {#taskParams-debug}

Shown for an Ansible template that allows debugging.

Runs `ansible-playbook` with more output: `-v` repeated as many times as **Debug level** sets.

## Debug level {#taskParams-debugLevel}

Shown for an Ansible template that allows debugging; active when **Debug** is ticked.

How verbose the output is, from `1` (`-v`) to `6` (`-vvvvvv`). The default is `4`.

## Dry Run {#taskParams-dryRun}

Shown for an Ansible template unless the template hides it.

Runs the playbook with `--check`: Ansible reports what it would change and changes nothing.

## Diff {#taskParams-diff}

Shown for an Ansible template unless the template hides it.

Runs the playbook with `--diff`: Ansible shows the changes it makes to files and templates.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Shown for an Ansible template that allows a task to change it.

Skips the installation of the roles and collections of the `requirements.yml` files before the
playbook runs. Tick it when the requirements are installed already and the task should not
reach Ansible Galaxy.

## Plan {#taskParams-plan}

Shown for a Terraform, OpenTofu or Terragrunt template.

Runs `plan` only. The task ends after the plan and applies nothing.

## Destroy {#taskParams-destroy}

Shown for a Terraform, OpenTofu or Terragrunt template.

Adds `-destroy` to the plan and the apply: the task destroys the resources of the workspace
instead of creating or updating them.

## Auto Approve {#taskParams-autoApprove}

Shown for a Terraform, OpenTofu or Terragrunt template.

Applies the plan without waiting for a confirmation. It works only when the template allows
auto approve; a template set to approve every task applies the plan in either case.

Without auto approve a scheduled task stops after the plan and waits until someone confirms or
rejects it on the task page.

## Upgrade {#taskParams-upgrade}

Shown for a Terraform, OpenTofu or Terragrunt template.

Runs `init` with `-upgrade`: the providers and modules are updated to the newest versions the
configuration allows.

## Reconfigure {#taskParams-reconfigure}

Shown for a Terraform, OpenTofu or Terragrunt template.

Runs `init` with `-reconfigure` instead of `-migrate-state`: the backend is set up anew and the
existing state is not migrated.

## Run at {#scheduleForm-runAt}

Shown for a schedule that runs once.

The date and time the task starts, in the timezone of the server shown at the end of the field.
A new schedule starts with the next full hour. The moment must be in the future: the server
refuses a time that has passed.

After the task has started, the schedule is turned off, or deleted when **Delete after run** is
ticked. When Semaphore is not running at that moment, the task is not started later: on the next
start the schedule is turned off, or deleted, without a run.

## Delete after run {#scheduleForm-deleteAfterRun}

Shown for a schedule that runs once.

Deletes the schedule after its task has started. Leave it unticked to keep the schedule, turned
off, so that you can set a new time and turn it on again.

## Show cron format {#scheduleForm-rawCron}

Shown for a cron schedule.

Switches between typing the cron expression (on) and building it from **Timing** and the time
checkboxes (off). Both edit the same expression. The browser remembers the choice.

The switch is locked on when the expression can not be shown as checkboxes: when it is invalid,
or uses a form such as `@daily` or `CRON_TZ=`.

## Cron {#scheduleForm-cronFormat}

Shown for a cron schedule when **Show cron format** is on.

The cron expression in five fields: minute, hour, day of the month, month, day of the week (`0`
is Sunday). The field is required, and the server checks the expression as you type. A new
schedule starts with `* * * * *`, every minute.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

The server also accepts `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` and
`@every 1h30m`, and a prefix such as `CRON_TZ=Europe/Berlin` for another timezone. The preview
below can not show the next run of these.

## Timing {#scheduleForm-timing}

Shown for a cron schedule when **Show cron format** is off.

How often the task runs. The checkboxes below change with it:

- **Hourly**: **Minutes**.
- **Daily**: **Hours** and **Minutes**.
- **Weekly**: **Weekdays**, **Hours** and **Minutes**.
- **Monthly**: **Days**, **Hours** and **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** and **Minutes**.

## Months {#scheduleForm-months}

Shown when **Timing** is **Yearly**.

The months the task runs in. Tick several to run in each of them. With none ticked the task runs
in every month.

## Weekdays {#scheduleForm-weekdays}

Shown when **Timing** is **Weekly**.

The days of the week the task runs on. With none ticked the task runs on every day.

## Days {#scheduleForm-days}

Shown when **Timing** is **Monthly** or **Yearly**.

The days of the month the task runs on. A month without the day is skipped: a task set to the
31st does not run in April. With none ticked the task runs on every day.

## Hours {#scheduleForm-hours}

Shown when **Timing** is **Daily**, **Weekly**, **Monthly** or **Yearly**.

The hours the task runs at, in the timezone shown on the right. With none ticked the task runs
in every hour.

## Minutes {#scheduleForm-minutes}

Shown for a cron schedule when **Show cron format** is off.

The minutes of the hour the task runs at, in steps of five. With none ticked the task runs every
minute: tick at least one, for example `:00`, for a task that should run once an hour or once a
day.

## Next run time {#scheduleForm-nextRun}

When the next task would start, by the schedule as it is in the form: in the timezone of the
server and in the local time of your browser. A dash means the expression is invalid, or is one
the preview can not read, such as `@daily`.

## Enabled {#scheduleForm-active}

Whether the schedule starts tasks. Untick it to keep a schedule without running it. The switch
in the list of schedules does the same. A new schedule is enabled.
