# Survey variable target — CLI argument vs process environment

A survey variable carries a `target` field: `""` (default) delivers it the app-specific way —
Ansible `--extra-vars`, Terraform/OpenTofu/Terragrunt `-var`, shell `name=value` argument —
and `"env"` delivers it as a process environment variable named exactly like the variable. The
field lives inside the existing `survey_vars` JSON column, so there is no migration and untouched
templates behave as before. Delivery is decided in one place, `LocalExecutor`, which the server
and the remote runner both run: the runner receives the whole `db.Template` (with `survey_vars`)
in its job payload.

Status (2026-09-27): implemented — commits bce105b46 + 487584918 (2026-07-14: model,
validation, executor, UI, `api-docs.yml`) and a6e899a2c (2026-08-03: multi-select values as JSON
for the env target and `-var`); present in tag trees from v2.19.5-beta7 and on
`origin/2-19-stable`; user docs in `docs/docs/user-guide/task-templates/survey-vars.md`
("Pass variable as (target)").

## Decisions and why

- Field name and values mirror `EnvironmentSecretType` (`var`/`env`, `db/Environment.go`); an
  empty default means zero migration and unchanged behaviour for every existing template.
- The env var name is the survey var name verbatim — no automatic `TF_VAR_` prefix. Users who need
  it name the variable `TF_VAR_foo`; keeps delivery app-agnostic and predictable.
- Env-target vars are removed from the extra-vars map in `getEnvironmentExtraVars` and appended
  to the process environment by `getSurveyEnvVars` in `Prepare`, so each value is delivered
  exactly once and never leaks into CLI arguments.
- Secret survey vars honour the target too — the main use case: env vars are invisible in
  process listings, CLI args are not. The secret JSON reaches `LocalExecutor.Secret` through the
  task-bound access key (see `survey-secrets-remote-runners.md`).
- Since 2026-10-10 (TASK@CEA87DDD08) a secret with the default target is no longer in clear
  anywhere either: for Ansible, the survey secrets and the variable-group secrets of type `var`
  go in a second `--extra-vars @task_<id>_secret_vars_*.yml`, encrypted in the Ansible Vault
  format with a one-off password that ansible-playbook receives on its `--vault-id <guid>@prompt`
  (`services/tasks/extra_vars_file.go`, `pkg/ansible_vault`; the full picture is in
  `AGENTS/work/workflow-editor/artifacts.md` § Trust). The env target remains the way to keep a
  secret out of extra vars altogether.
- `Template.Validate` rejects unknown targets (`ValidationError`).
- Deliberately not done (YAGNI): per-app target overrides, a prefix option, validation that an
  env-target name is a valid shell identifier, translations beyond `en.js`.
- Later change vs the plan: multi-select (`SurveyVarSelect`) values are serialised through
  `formatVarValue` (JSON array) for both the env target and terraform `-var`; the plan used
  plain `%v` formatting.

## Open or deferred

- Only `web/src/lang/en.js` has `survey_var_target`; other locales fall back to English.
- Whether the Pro container executors (`pro_impl/services/tasks/containerexec`) honour `target`
  when they build their own argument lists was not verified.

## Key files and references

- `db/Template.go` — `SurveyVarTarget`, `SurveyVar.Target`, check in `Validate`; tests
  `db/Template_test.go`.
- `services/tasks/local_executor.go` — `getEnvironmentExtraVars`, `getSurveyEnvVars`,
  `Prepare`; tests `services/tasks/local_executor_test.go`. `services/tasks/extra_vars_file.go`
  — `takeSecretExtraVars` (which keys are secrets, env-target ones excluded); tests
  `services/tasks/extra_vars_file_test.go`.
- `web/src/components/SurveyVars.vue`, `web/src/lang/en.js`; `api-docs.yml`
  (`TemplateSurveyVar.target`).
- Source plan: `AGENTS/plans/2_19/survey-var-target.md`.
