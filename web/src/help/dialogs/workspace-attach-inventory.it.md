---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Collega a questo modello un workspace che esiste già nel progetto, in modo che compaia tra i
workspace del modello con il suo stato e i suoi alias. Un workspace appartiene a un solo modello
alla volta. Il collegamento non rende il workspace quello predefinito: i compiti continuano a essere
eseguiti nel workspace predefinito finché non premi **Make default**.

## Workspace {#workspaceSelect-inventoryId}

Il workspace da collegare. Il campo è obbligatorio. L'elenco mostra i workspace del progetto creati
per l'app di questo modello (Terraform, OpenTofu o Terragrunt) e non collegati ad alcun modello; un
workspace scollegato dal suo modello con **Detach** ricompare qui.

Il workspace mantiene il suo nome, la sua chiave e il suo stato. Il server rifiuta un workspace che
nel frattempo è stato preso da un altro modello.
