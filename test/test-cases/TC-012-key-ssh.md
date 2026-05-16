# TC-012 — Create SSH key and use it for inventory host access

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Key Store / Inventory            |
| Priority     | Critical                         |
| Type         | Functional                       |
| Automatable  | Yes                              |

## Objective

An SSH key registered in the Key Store is used by Ansible to connect to hosts
defined in the inventory.

## Preconditions

* Project `Infra QA`.
* A reachable test Linux host (`10.0.0.21`) with the matching public key
  authorized for user `qa`.
* Connectivity from Semaphore (or its runner) to the host on port 22.

## Test data

| Field      | Value         |
|------------|---------------|
| Key name   | `qa-host-ssh` |
| Inventory  | `qa-hosts`    |

## Steps

1. Create SSH key `qa-host-ssh` from the matching private key.
2. Create inventory `qa-hosts` (type `static`) with one host:
   ```
   [linux]
   db01 ansible_host=10.0.0.21 ansible_user=qa
   ```
   and SSH key `qa-host-ssh`.
3. Create or reuse an Ansible template that runs `ansible -m ping linux`.
4. Run the template.
5. Edit the inventory and detach the SSH key (set to *none*); run again.

## Expected results

* Step 4: task succeeds; `db01 | SUCCESS => { "ping": "pong" }` appears in the
  log.
* Step 5: task fails with `Permission denied` / `Authentication failed`.
* The private key file is written under the per-task temp directory and removed
  on task completion (verify the temp dir is cleaned).

## Postconditions

Re-attach `qa-host-ssh` to the inventory.
