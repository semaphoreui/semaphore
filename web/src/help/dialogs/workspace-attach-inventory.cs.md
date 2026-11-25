---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Připojí k této šabloně workspace, který už v projektu existuje, takže se objeví mezi workspaces
šablony se svým stavem a aliasy. Workspace patří vždy jen jedné šabloně. Připojení z workspace
neudělá výchozí: úlohy dál běží ve výchozím workspace, dokud nestisknete **Make default**.

## Workspace {#workspaceSelect-inventoryId}

Workspace, který se má připojit. Pole je povinné. Seznam ukazuje workspaces projektu, které jsou
určené pro aplikaci této šablony (Terraform, OpenTofu nebo Terragrunt) a nejsou připojené
k žádné šabloně; workspace odpojený od své šablony tlačítkem **Detach** se zde znovu objeví.

Workspace si ponechá svůj název, klíč a stav. Server odmítne workspace, který mezitím převzala
jiná šablona.
