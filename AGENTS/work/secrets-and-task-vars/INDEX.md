# secrets-and-task-vars — zone index

Decisions and open state around what a task may receive as variables, how survey variables are
delivered and defaulted, how survey secrets reach remote runners in HA, and external secret
storages.

| When you need it | Read |
| --- | --- |
| Why task-API variables are default-deny with a per-template opt-in, and where the gate lives | `decisions.md` |
| Why survey secrets for remote runners are stored as task-bound access keys (not a task column, not Redis) | `decisions.md` |
| Touching the CyberArk PVWA secret storage branch | `decisions.md` |
| Changing how a task-bound survey secret is created, read at dispatch, expired or swept; wiring survey secrets into the Pro Docker/K8s executors | `survey-secrets-remote-runners.md` |
| Changing how a survey variable reaches the process (extra-vars / `-var` / argument vs environment variable), or adding a survey var delivery option | `survey-var-target.md` |
| Picking up issue #2212 — survey variables as a reusable project entity assignable to many templates | `reusable-surveys.md` |
| Picking up issue #2244 — survey defaults missing for scheduled, API-created or integration tasks | `schedule-survey-defaults.md` |
| Touching `TaskRunner.populateTaskEnvironment` or `SurveyVar.DefaultValue` | `schedule-survey-defaults.md` |
