---
name: semaphore-checkout
description: >-
    Create and switch to a branch with the same name in both repos — the root repo and pro_impl. 
    Usage: /semaphore-checkout <branch_name>
---

Create a branch named `<branch_name>` in both repositories at once — the root repo
(current directory) and the PRO module repo (`./pro_impl`) — and switch both to it.

CI depends on branch names matching: GitHub Actions clones the `pro_impl` branch
with the same name as the root repo branch and falls back to `main` if it does not
exist. This skill keeps the two repos in sync.

## Arguments

- `<branch_name>` (required) — the branch to create. If not provided, ask the user
  for it. Validate it with `git check-ref-format --branch <branch_name>` before use.

## Behavior

For each repo — root (`.`) first, then `./pro_impl`:

1. Check the working tree is clean (`git status --porcelain`). If there are
   uncommitted changes in either repo, stop and report which repo is dirty —
   do not stash or discard anything.
2. Fetch the latest state: `git fetch origin`.
3. Create the branch from the up-to-date default branch and switch to it:
   - root repo: `git checkout -b <branch_name> origin/develop`
   - pro_impl: `git checkout -b <branch_name> origin/main`
4. If the branch already exists locally, just `git checkout <branch_name>`
   instead of creating it. If it exists only on origin, check it out tracking
   the remote branch.

Do not push the new branches — they are pushed later with the first commit.

## Rules

- Never force-create (`-B`) or reset an existing branch.
- Both repos must end up on the same branch name; if step for `pro_impl` fails
  after the root repo already switched, report the inconsistent state clearly.
- Report the final state at the end: branch name and start point in each repo.

## Example

```
/semaphore-checkout fix/task-env-vars
```

Result:

```
Root repo:  fix/task-env-vars (from origin/develop)
pro_impl:   fix/task-env-vars (from origin/main)
```
