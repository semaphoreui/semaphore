---
form: src/components/InventoryGroupForm.vue
---

# Edit Group

A group of hosts in a managed inventory. A playbook runs on a part of the hosts by the name of the
group, and a variable of the group is set once for all its hosts. Hosts are put into a group on the
Hosts tab, in the host dialog or with the selection actions.

## Name {#inventoryGroupForm-name}

The name Ansible knows the group by, for example in `hosts: web` of a playbook. The field is
required. A name is unique in its inventory, and letter case matters: `web` and `Web` are two
groups. It can not contain spaces, commas, square brackets or colons, is at most 250 characters
long, and can not be `all` or `ungrouped`. Ansible warns about names with characters other than
letters, digits and underscores, such as `web-eu`, but the name still works. The characters `*`,
`?`, `!`, `&`, `~` and `@` are not allowed either, because Ansible reads them as a pattern.

## Description {#inventoryGroupForm-description}

Optional. A note for people, shown in the list of groups. Ansible does not get it.

## Child groups {#inventoryGroupForm-children}

Optional. Groups inside this group. The hosts of a child group also belong to this group. A group
can not be its own ancestor.

## Variables {#inventoryGroupForm-variables}

Optional. Variables of this group, written as a YAML or JSON mapping:

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

Every host of the group gets them. A variable of a host takes precedence over a variable of the
group with the same name, and a variable of the group takes precedence over a variable of the
inventory. YAML anchors, aliases and merge keys are not accepted.
