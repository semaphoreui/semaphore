---
form: src/components/InventoryHostForm.vue
---

# Edit Host

A host of a managed inventory. A task that uses the inventory runs on its hosts that are switched on
in the list of hosts.

## Name {#inventoryHostForm-name}

The name Ansible knows the host by, and the address it connects to unless the variables set
`ansible_host`. The field is required. A name is unique in its inventory, and letter case matters:
`web1` and `Web1` are two hosts. It can not contain spaces, commas, square brackets or colons, and
is at most 250 characters long.

## Description {#inventoryHostForm-description}

Optional. A note for people, shown in the list of hosts. Ansible does not get it.

## Variables {#inventoryHostForm-variables}

Optional. Variables of this host, written as a YAML or JSON mapping:

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

A variable of the host takes precedence over a variable of the inventory with the same name. YAML
anchors, aliases and merge keys are not accepted.
