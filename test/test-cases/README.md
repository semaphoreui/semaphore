# Semaphore UI — Manual QA Test Cases

This directory contains 30 real-world manual test cases for Semaphore UI, covering
end-to-end behavior of the web UI and API across the major feature areas:
authentication, projects, repositories, inventory, key store, variable groups,
task templates, tasks, schedules, runners, integrations / webhooks, RBAC, and
notifications.

## How to use

* Each test case is self-contained: preconditions, data, steps, expected results.
* Priority is one of `Critical`, `High`, `Medium`, `Low`.
* Run a clean Semaphore instance (Docker recommended) before executing the full
  suite. A single admin user and one project are assumed unless noted otherwise.

## Reference environment

* Semaphore UI deployed via Docker (`semaphoreui/semaphore:latest`).
* Database: SQLite (default) — re-run on Postgres/MySQL when DB compatibility is
  in scope.
* At least one remote host reachable over SSH for inventory-related cases.
* A Git repository (public or with deploy key) for repository-related cases.

## Index

| #   | ID      | Area              | Title                                                                          | Priority |
|-----|---------|-------------------|--------------------------------------------------------------------------------|----------|
| 01  | TC-001  | Auth              | [Admin login with valid credentials](TC-001-admin-login-valid.md)              | Critical |
| 02  | TC-002  | Auth              | [Login fails with wrong password and locks brute-force](TC-002-login-invalid.md)| Critical |
| 03  | TC-003  | Auth / 2FA        | [Enable TOTP and log in with one-time code](TC-003-totp-enrollment.md)         | High     |
| 04  | TC-004  | Users             | [Admin creates and deactivates a user](TC-004-user-lifecycle.md)               | High     |
| 05  | TC-005  | API Tokens        | [Create and revoke a personal API token](TC-005-api-token.md)                  | High     |
| 06  | TC-006  | Projects          | [Create a new project](TC-006-create-project.md)                               | Critical |
| 07  | TC-007  | Projects          | [Update `max_parallel_tasks` and observe queueing](TC-007-max-parallel-tasks.md)| Medium   |
| 08  | TC-008  | Projects          | [Backup and restore a project](TC-008-backup-restore.md)                       | High     |
| 09  | TC-009  | Projects          | [Delete project only after dependents are removed](TC-009-delete-project.md)   | High     |
| 10  | TC-010  | Repositories      | [Add Git repository with SSH deploy key](TC-010-repo-ssh.md)                   | Critical |
| 11  | TC-011  | Repositories      | [Add Git repository with HTTPS token](TC-011-repo-https.md)                    | High     |
| 12  | TC-012  | Key Store         | [Create SSH key and use it for inventory access](TC-012-key-ssh.md)            | Critical |
| 13  | TC-013  | Key Store         | [Create login/password key and bind become user](TC-013-key-login-password.md) | High     |
| 14  | TC-014  | Key Store         | [Bind HashiCorp Vault secret storage](TC-014-vault-storage.md)                 | Medium   |
| 15  | TC-015  | Inventory         | [Static inventory with two host groups](TC-015-inventory-static.md)            | Critical |
| 16  | TC-016  | Inventory         | [File-based Ansible inventory from repo](TC-016-inventory-file.md)             | High     |
| 17  | TC-017  | Inventory         | [Terraform workspace inventory binding](TC-017-inventory-terraform.md)         | Medium   |
| 18  | TC-018  | Variable Groups   | [JSON extra-vars + ENV vars + secret](TC-018-variable-group-mixed.md)          | High     |
| 19  | TC-019  | Variable Groups   | [`TF_VAR_` secrets reach Terraform plan](TC-019-tf-var-secret.md)              | High     |
| 20  | TC-020  | Templates         | [Create an Ansible task template and run it](TC-020-template-ansible.md)       | Critical |
| 21  | TC-021  | Templates         | [Build → Deploy template chain with artifact version](TC-021-build-deploy.md)  | High     |
| 22  | TC-022  | Templates         | [Survey variables: required, enum, default](TC-022-survey-vars.md)             | High     |
| 23  | TC-023  | Templates         | [Override branch, limit, tags at task launch](TC-023-template-overrides.md)    | Medium   |
| 24  | TC-024  | Tasks             | [Stop a running task and verify status `stopped`](TC-024-stop-task.md)         | High     |
| 25  | TC-025  | Schedules         | [Create a cron schedule and validate next run](TC-025-schedule-cron.md)        | High     |
| 26  | TC-026  | Schedules         | [One-shot `run_at` schedule with delete-after-run](TC-026-schedule-runat.md)   | Medium   |
| 27  | TC-027  | Runners           | [Register a remote runner via registration token](TC-027-runner-register.md)  | High     |
| 28  | TC-028  | Runners           | [Tag-scoped task routing to matching runner](TC-028-runner-tags.md)            | Medium   |
| 29  | TC-029  | Integrations      | [GitHub webhook + HMAC + matcher triggers template](TC-029-integration-github.md)| High |
| 30  | TC-030  | RBAC              | [Invited user with `Task Runner` role cannot edit](TC-030-rbac-task-runner.md) | Critical |
