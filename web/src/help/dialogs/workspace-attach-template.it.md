---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Collega questo workspace a un modello, in modo che compaia tra i workspace del modello con il suo
stato e i suoi alias. Un workspace appartiene a un solo modello alla volta: una volta collegato,
lascia la pagina Inventario e viene gestito nel modello. Il collegamento non lo rende il workspace
predefinito del modello: per farlo, apri il modello e premi **Make default**.

## Template {#inventoryAttach-templateId}

Il modello a cui collegare il workspace. Il campo è obbligatorio. L'elenco mostra i modelli del
progetto creati per la stessa app del workspace, Terraform o OpenTofu.

Il workspace mantiene il suo nome, la sua chiave e il suo stato. Per spostarlo in seguito in un
altro modello, premi **Detach** nel modello a cui è collegato: torna nella pagina Inventario. Il
workspace predefinito di un modello non può essere scollegato.
