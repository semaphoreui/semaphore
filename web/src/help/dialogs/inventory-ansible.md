---
form: src/components/InventoryForm.vue
---

# Edit Ansible Inventory

An inventory tells Ansible which hosts to run a playbook against, and holds the credentials
Semaphore uses to connect to them. A task template refers to an inventory, and every task of that
template runs against it. The format of an inventory is described in the
[Ansible inventory guide](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html).

## Name {#inventoryForm-name}

The name the inventory is listed under on the Inventory page and in the inventory selector of a
task template. Pick a name that tells the hosts apart, for example `Production web servers`.

## Runner tag {#inventoryForm-runnerTag}

Shown when project runners are available.

Tasks that use this inventory run on a runner that carries this tag. A runner tag set on the task
template takes precedence over the tag of the inventory. Leave the field empty to let any runner
take the tasks.

## User Credentials {#inventoryForm-sshKeyId}

The key from the Key Store that Ansible logs in to the hosts with. The field is required and
accepts three types of keys:

- **SSH Key**: Ansible connects with the private key and, when the key has one, with its login.
- **Login with password**: Ansible connects with the login and the password.
- **None**: Semaphore passes no credentials. Use it for hosts that need none, such as
  `localhost`, or when the inventory itself defines how to connect.

## Sudo Credentials {#inventoryForm-becomeKeyId}

Optional. A key of the type **Login with password** from the Key Store, used when a play
escalates privileges with `become`: its login is the user Ansible becomes, and its password is
the password Ansible answers the prompt with.

Leave the field empty when the plays do not escalate privileges, or when the user above needs no
password to do so.

## Type {#inventoryForm-type}

Where the inventory comes from:

- **Static**: the inventory is typed into this dialog in the INI format.
- **Static YAML**: the inventory is typed into this dialog in the YAML format. Use it for the
  configuration of an inventory plugin as well.
- **File**: the inventory is a file in a Git repository.
- **Managed**: the hosts and their variables are kept as records. Open the inventory from the list
  to edit them.

The fields below the type change with it.

## Path to Inventory file {#inventoryForm-inventoryPath}

Shown when the type is **File**.

The path to the inventory file from the root of its repository, for example
`inventory/production.yml`. The repository is the one selected in the Repository field or, when
that field is empty, the repository of the task template that uses the inventory.

## Repository {#inventoryForm-repositoryId}

Shown when the type is **File**. Optional.

The repository that holds the inventory file. Semaphore updates its copy of the repository before
every task. Leave the field empty when the file lives in the same repository as the playbook: the
path is then looked up in the repository of the task template.

## Inventory {#inventoryForm-inventoryCode}

Shown when the type is **Static** or **Static YAML**.

The text of the inventory. Semaphore stores it and writes it to a temporary file for every task.
A static inventory in the INI format:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

The same hosts as **Static YAML**:

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
