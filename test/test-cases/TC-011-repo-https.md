# TC-011 — Add Git repository with HTTPS personal access token

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Repositories                     |
| Priority     | High                             |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A Git repository can be cloned over HTTPS using a `Login/Password` key whose
password field stores a PAT.

## Preconditions

* Project `Infra QA` exists.
* A private repository accessible via HTTPS using a PAT (e.g.
  `https://github.com/qa/playbooks.git`).

## Test data

| Field        | Value                              |
|--------------|------------------------------------|
| Key name     | `gh-pat`                           |
| Key type     | `Login With Password`              |
| Login        | `qa-bot` (or empty for token-only) |
| Password     | `<GitHub PAT>`                     |
| Repo URL     | `https://github.com/qa/playbooks.git` |
| Branch       | `main`                             |

## Steps

1. Create the `gh-pat` key in **Key Store**.
2. Create a repository `playbooks-https` with the test data above.
3. Create or reuse a template pointing at this repo and run it.
4. Rotate the PAT on GitHub (revoke) and run the task again.

## Expected results

* Step 3: task succeeds and pulls the latest commit on `main`.
* Step 4: task fails with a recognizable authentication error (`fatal:
  Authentication failed`); status is `error`, not silently `success`.
* No PAT value is logged in the task output or stored in plaintext in the DB
  (verified via direct DB query: `access_key.secret` is encrypted).

## Postconditions

`playbooks-https` repo present but with a revoked token; rotate before further
use.
