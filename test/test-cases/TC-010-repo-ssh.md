# TC-010 — Add Git repository with SSH deploy key

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Repositories                     |
| Priority     | Critical                         |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

A Git repository can be added using a stored SSH key, and Semaphore can clone
the configured branch on first task launch.

## Preconditions

* Project `Infra QA` exists.
* A Git repository (e.g. private GitLab repo) reachable from the Semaphore host
  over SSH on port 22.
* SSH private key (and matching deploy key registered on the Git provider)
  available as a file.

## Test data

| Field      | Value                                |
|------------|--------------------------------------|
| Key name   | `git-deploy-key`                     |
| Key type   | `SSH`                                |
| Repo name  | `playbooks`                          |
| Git URL    | `git@gitlab.example.com:qa/playbooks.git` |
| Branch     | `main`                               |

## Steps

1. Go to **Key Store → New Key**.
2. Choose **SSH**, name `git-deploy-key`, paste the private key body, leave
   passphrase empty (or fill it), save.
3. Open **Repositories → New Repository**.
4. Fill the repo form using the test data, selecting `git-deploy-key` as the
   access key.
5. Save the repository.
6. Create a minimal Ansible template that uses this repository and run it.

## Expected results

* Step 2: key is created with `type=ssh`; the private body is not displayed
  after save.
* Step 5: repository row appears with the URL and branch.
* Step 6: the task log shows the clone step succeeding (`Cloning into …`,
  `Switched to branch 'main'`), and the playbook executes.
* If the wrong key is configured (re-test by editing to a bogus key) the task
  fails with a clear `Permission denied (publickey)` style error.

## Postconditions

A working repository `playbooks` is bound to project `Infra QA`.
