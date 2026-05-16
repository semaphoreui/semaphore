# TC-017 — Terraform workspace inventory binding

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Inventory / Terraform            |
| Priority     | Medium                           |
| Type         | Functional                       |
| Automatable  | Partial                          |

## Objective

A Terraform task template wires an inventory of type
`terraform-workspace`/`tofu-workspace` so the selected workspace is used at
task execution.

## Preconditions

* Project `Infra QA`.
* Repository `tf-infra` containing a minimal Terraform module with workspaces
  `dev` and `prod`.
* A backend configured (e.g. `local` for the test, or a remote backend with
  credentials in the variable group).

## Test data

| Field          | Value                          |
|----------------|--------------------------------|
| Inventory name | `tf-dev`                       |
| Type           | `terraform-workspace`          |
| Workspace      | `dev`                          |
| Inventory      | `tf-prod`                      |
| Type           | `terraform-workspace`          |
| Workspace      | `prod`                         |

## Steps

1. Create both inventories `tf-dev` and `tf-prod`.
2. Create a Terraform template `tf-plan` using repo `tf-infra` and inventory
   `tf-dev`.
3. Run **Plan** action.
4. Edit the template to use inventory `tf-prod` and run **Plan** again.
5. Check the task log for the workspace selection step.

## Expected results

* Step 3: log contains `Switched to workspace "dev"` (or equivalent
  `terraform workspace select dev` exit 0).
* Step 4: log contains `Switched to workspace "prod"`.
* The `Allow override inventory` setting controls whether the user can pick a
  different workspace at launch time. Toggle off and ensure the UI hides the
  selector.

## Postconditions

Templates and inventories remain for later reuse.
