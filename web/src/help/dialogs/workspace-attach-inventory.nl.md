---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Koppelt een workspace die al in het project bestaat aan dit sjabloon, zodat hij tussen de
workspaces van het sjabloon verschijnt met zijn state en aliassen. Een workspace hoort bij één
sjabloon tegelijk. Koppelen maakt de workspace niet de standaardworkspace: taken blijven in de
standaardworkspace draaien totdat u op **Make default** drukt.

## Workspace {#workspaceSelect-inventoryId}

De workspace die u koppelt. Het veld is verplicht. De lijst toont de workspaces van het project die
zijn gemaakt voor de app van dit sjabloon (Terraform, OpenTofu of Terragrunt) en aan geen enkel
sjabloon zijn gekoppeld; een workspace die met **Detach** van zijn sjabloon is losgekoppeld,
verschijnt hier weer.

De workspace behoudt zijn naam, sleutel en state. De server weigert een workspace die een ander
sjabloon intussen heeft genomen.
