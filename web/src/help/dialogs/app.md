---
form: src/components/AppForm.vue
---

# Edit App

An app is a tool Semaphore runs tasks with: Ansible, Terraform, OpenTofu, Terragrunt, Bash,
Python, PowerShell, or a command of your own. Every task template is made for one app, and this
dialog sets how the server starts it and how the app is shown in the interface.

The settings are stored in the database and take effect at once on the server that saves them;
the other nodes of a cluster read them when they restart. A field you empty keeps its old value
until the server restarts.

## ID {#appForm-id}

Required. The key the app is stored under; a task template refers to its app by this key. The
built-in apps are `ansible`, `terraform`, `tofu`, `terragrunt`, `pulumi`, `bash`, `python` and
`powershell`. Any other ID adds an app of your own: Semaphore runs it like a script, with the
command named in **Path** or, when **Path** is empty, the command named like the ID.

Use an ID without dots. Changing the ID of an existing app does not rename it: the settings are
saved as a new app under the new ID, and the old app stays.

## Icon {#appForm-icon}

Optional. The name of an icon of the Material Design Icons set without the `mdi-` prefix, for
example `language-python`. Leave it empty to keep the built-in icon of the app;
an app of your own then shows a question mark.

## Name {#appForm-title}

Optional. The name of the app in the list of applications and in the menu **New Template**. Leave
it empty to keep the built-in name, such as `Ansible Playbook`; an app of your own then has no
name, so give it one.

## Path {#appForm-path}

Optional. The command Semaphore runs instead of the default one, either a name looked up in the
`PATH` of the server or a full path, for example `/usr/local/bin/terraform-1.9`. The defaults are
`terraform`, `tofu`, `terragrunt`, `bash`, `python3` and `powershell`; an app of your own runs the
command named like its ID.

Ansible ignores this field: Semaphore always runs `ansible-playbook` and `ansible-galaxy`.

## Priority {#appForm-priority}

The place of the app in the list of applications and in the menu **New Template**: an app with a
higher number comes first. The built-in apps start at 1000 for Ansible down to 500 for Python. A
built-in app with a priority of `0` or less gets its built-in priority back when the server
starts.

## Args {#appForm-args}

Optional. Arguments put on the command line right after the command, one per chip. Click **+** to
add one, click a chip to change it, drag the chips to change their order.

- Terraform, OpenTofu and Terragrunt: the arguments come before the arguments Semaphore adds to
  every command, such as `init` and `apply`.
- Script apps: the arguments come before the script of the template and replace the default
  ones: PowerShell runs with `-File` by default, so put `-File` last in your list.
- Ansible ignores this field.

## Active {#appForm-active}

An active app is offered in the menu **New Template**. Clear the box to hide the app from that
menu: templates already made with it keep their app and still run.

The box is cleared for a new app: tick it to make the app available.
