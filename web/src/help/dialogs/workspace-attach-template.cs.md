---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Připojí tento workspace k šabloně, takže se objeví mezi workspaces šablony se svým stavem
a aliasy. Workspace patří vždy jen jedné šabloně: po připojení zmizí ze stránky Inventář a spravuje
se v šabloně. Připojení z něj neudělá výchozí workspace šablony: k tomu otevřete šablonu
a stiskněte **Make default**.

## Template {#inventoryAttach-templateId}

Šablona, ke které se workspace připojí. Pole je povinné. Seznam ukazuje šablony projektu určené
pro stejnou aplikaci jako workspace, Terraform nebo OpenTofu.

Workspace si ponechá svůj název, klíč a stav. Chcete-li ho později přesunout k jiné šabloně,
stiskněte **Detach** v šabloně, ke které je připojený: vrátí se na stránku Inventář. Výchozí
workspace šablony odpojit nelze.
