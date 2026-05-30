# TC-029 — GitHub webhook + HMAC + matcher triggers a template

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Integrations                     |
| Priority     | High                             |
| Type         | Integration                      |
| Automatable  | Yes                              |

## Objective

An integration configured with GitHub auth (HMAC-SHA256) and a header matcher
launches the bound template when a properly-signed webhook is delivered, and
rejects requests with invalid signatures.

## Preconditions

* Project `Infra QA`.
* Template `Deploy site (dev)` (TC-020).
* `curl` available; OpenSSL for HMAC signing.

## Test data

| Field        | Value                                |
|--------------|--------------------------------------|
| Integration  | `gh-push`                            |
| Auth method  | `github`                             |
| Auth secret  | stored in key `gh-hmac` with value `s3cret-hmac-key` |
| Matcher      | header `X-GitHub-Event` equals `push` |
| Extractor    | body JSON path `.ref` → env var `GIT_REF` |
| Alias        | `gh-push` (single)                   |

## Steps

1. Create the integration with the test data, binding to `Deploy site (dev)`.
2. Note the integration URL `/api/integrations/<alias>`.
3. Construct a JSON payload:
   ```
   {"ref":"refs/heads/main","commits":[{"id":"abc"}]}
   ```
4. Compute `X-Hub-Signature-256` = `sha256=` + `openssl dgst -sha256 -hmac
   "s3cret-hmac-key" <payload>`.
5. POST the payload with headers
   `X-GitHub-Event: push` and the signature header.
6. Repeat with a tampered signature.
7. Repeat with the correct signature but `X-GitHub-Event: pull_request`.
8. Inspect the triggered task's environment for `GIT_REF`.

## Expected results

* Step 5: HTTP 200/204; a task is created with `integration_id` set, and the
  task log shows `GIT_REF=refs/heads/main`.
* Step 6: HTTP 401/403; no task created.
* Step 7: HTTP 200/204 but no task — the matcher did not match the event.
* Aliases scope: hitting a different alias path returns 404.
* The webhook URL is reachable without UI authentication (it is the public
  integration endpoint).

## Postconditions

Integration left in place for use in audits.
