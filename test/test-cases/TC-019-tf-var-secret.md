# TC-019 — `TF_VAR_` secrets reach Terraform plan

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Variable Groups / Terraform      |
| Priority     | High                             |
| Type         | Integration                      |
| Automatable  | Partial                          |

## Objective

Secrets prefixed with `TF_VAR_` in a variable group are converted into the
corresponding Terraform input variables (`var.<name>`) at runtime.

## Preconditions

* Project `Infra QA`.
* Terraform repo `tf-infra` with a module declaring:
  ```hcl
  variable "hcloud_token" { type = string }
  output "len" { value = length(var.hcloud_token) }
  ```
* Template `tf-plan` using `tf-infra` and inventory `tf-dev`.

## Test data

| Field          | Value                          |
|----------------|--------------------------------|
| Variable group | `tf-secrets`                   |
| Secret name    | `TF_VAR_hcloud_token`          |
| Secret value   | `xxxxxxxx-xxxxxxxxxxxxxxxxxxxx`|

## Steps

1. Create variable group `tf-secrets` with the secret above (Secrets tab, type
   `env`).
2. Attach `tf-secrets` to template `tf-plan`.
3. Run the plan.
4. In the task log inspect the `Outputs:` block and the diff.
5. Remove the secret and re-run.

## Expected results

* Step 4: `len = <length-of-token>` is shown in outputs, confirming the value
  was injected as `var.hcloud_token`.
* Plan/apply log lines do **not** print the token value.
* Step 5: terraform fails (`No value for required variable`) — proves the
  injection mechanism was the source of the value.

## Postconditions

Restore the secret for downstream tests.
