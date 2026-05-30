# TC-022 — Survey variables: required, enum, default

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Task Templates                   |
| Priority     | High                             |
| Type         | Functional + Negative            |
| Automatable  | Yes                              |

## Objective

Survey variables defined on a template are presented in the launch dialog with
correct widgets, required validation, enum constraints, and defaults; their
values are passed to the underlying task.

## Preconditions

* Project `Infra QA`.
* Template `survey-demo` (Bash app) running:
  ```
  echo "env=$env"
  echo "replicas=$replicas"
  echo "feature_name=$feature_name"
  ```

## Test data — Survey variables

| Name          | Title         | Type | Required | Values             | Default |
|---------------|---------------|------|----------|--------------------|---------|
| `env`         | Environment   | enum | yes      | `dev`,`stg`,`prod` | `dev`   |
| `replicas`    | Replica count | int  | yes      | -                  | `1`     |
| `feature_name`| Feature       | str  | no       | -                  | empty   |

## Steps

1. Configure the three survey variables on `survey-demo`.
2. Click **Run** and observe the launch dialog.
3. Click **Submit** without filling `replicas` → expect validation error.
4. Enter `replicas = "abc"` → expect validation error.
5. Submit with `env=dev`, `replicas=3`, leave `feature_name` empty.
6. Submit a run with `env=stg`, `replicas=2`, `feature_name=login-redesign`.

## Expected results

* Step 2: `env` is a select with the three values, default `dev`; `replicas`
  is a numeric input prefilled with `1`; `feature_name` is a plain text input.
* Step 3-4: form blocks submission with inline error messages; no API call is
  made.
* Step 5: task succeeds, log shows `env=dev replicas=3 feature_name=`.
* Step 6: task succeeds, log shows `env=stg replicas=2
  feature_name=login-redesign`.
* Values are also visible on the **Task → Arguments** tab.

## Postconditions

Template retains survey config.
