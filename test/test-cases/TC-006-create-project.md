# TC-006 — Create a new project

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Projects                         |
| Priority     | Critical                         |
| Type         | Functional / Positive            |
| Automatable  | Yes                              |

## Objective

An admin can create a project and is granted the `Owner` role on it.

## Preconditions

* Logged in as admin.
* No prior project with the name being used.

## Test data

| Field               | Value                |
|---------------------|----------------------|
| Name                | `Infra QA`           |
| Max parallel tasks  | `2`                  |
| Alert               | unchecked            |

## Steps

1. From the top bar open the project switcher and click **+ New Project**.
2. Fill the form with the test data and submit.
3. Verify the redirect lands on the new project dashboard.
4. Open **Team** and inspect role assignment for the current user.
5. Open **Settings** and verify `Max parallel tasks = 2`.
6. Call `GET /api/projects` with the admin session and find the project.

## Expected results

* Step 2: HTTP 201 returned; UI lands on `/project/{id}`.
* Step 4: the creator is listed with role `Owner`.
* Step 5: the configured value matches what was entered.
* `GET /api/projects` includes the new project with the same id used in the UI.

## Postconditions

Project `Infra QA` exists and is the working project for many subsequent tests.
