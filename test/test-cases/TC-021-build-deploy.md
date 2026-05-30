# TC-021 — Build → Deploy template chain with artifact version

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Task Templates                   |
| Priority     | High                             |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A `build` template can emit an artifact version, and a `deploy` template
linked to that build can be triggered to deploy the produced version. The
`build_task_id` and `version` propagate from build to deploy.

## Preconditions

* Project `Infra QA`.
* Build template `image-build` (e.g. Bash app) that runs a script which prints
  a version string and exits 0. The script should call the API or use the
  documented mechanism to set the build version (e.g. via a marker file or
  `--build-version` argument).
* Deploy template `image-deploy` of type `deploy` referencing `image-build` as
  its build template.

## Steps

1. Run `image-build`. Wait for `success`.
2. Inspect the **Version** column on the task — confirm a non-empty version
   value (e.g. `1.2.3`).
3. Open `image-deploy` and click **Run**.
4. In the launch dialog inspect the build picker — the latest successful
   `image-build` task should be preselected; pick the one from step 1.
5. Submit and watch the deploy task.
6. After it completes, open the deploy task detail.

## Expected results

* Step 2: build task has `Version` set; UI shows it on the row.
* Step 4: the deploy launch dialog lists eligible build tasks (only successful
  builds of the linked build template).
* Step 6: the deploy task detail shows the linked `build_task_id` and the
  same `version` propagated from the chosen build.
* `GET /api/project/{id}/tasks/{tid}` for the deploy task returns
  `build_task_id` and `version` matching the build.

## Postconditions

Both tasks present in history.
