---
form: src/components/InventorySourceForm.vue
---

# Edit Source

A source imports the hosts and groups of a managed inventory from a file in a Git repository. The file is read each time Sync is pressed, and the hosts and groups it describes are kept in step with it. The server needs Ansible and access to the repository.

## Name {#inventorySourceForm-name}

The name of the source, shown on its card and next to the hosts it brings. The field is required.

## Repository {#inventorySourceForm-repository}

The repository of the project the file is read from. The field is required. Its branch is used.

## Path {#inventorySourceForm-path}

The path of an inventory file in the repository, for example `inventory/prod.yml`, or of a directory of inventory files. The field is required.

Files are read as static YAML or INI, scripts and plugin configuration files are refused. In a directory, every file Ansible does not skip must be a valid inventory, otherwise the sync fails and nothing changes. Ansible skips some extensions in a directory, such as `.cfg`, `.md` and `.txt`. A repository with a link that leads outside of it is refused.
