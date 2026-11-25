---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Koppelt deze workspace aan een sjabloon, zodat hij tussen de workspaces van het sjabloon verschijnt
met zijn state en aliassen. Een workspace hoort bij één sjabloon tegelijk: zodra hij is gekoppeld,
verdwijnt hij van de pagina Inventaris en wordt hij in het sjabloon beheerd. Koppelen maakt hem
niet de standaardworkspace van het sjabloon: open daarvoor het sjabloon en druk op
**Make default**.

## Template {#inventoryAttach-templateId}

Het sjabloon waaraan u de workspace koppelt. Het veld is verplicht. De lijst toont de sjablonen van
het project die voor dezelfde app als de workspace zijn gemaakt, Terraform of OpenTofu.

De workspace behoudt zijn naam, sleutel en state. Om hem later naar een ander sjabloon te
verplaatsen, drukt u op **Detach** in het sjabloon waaraan hij is gekoppeld: hij keert terug naar
de pagina Inventaris. De standaardworkspace van een sjabloon kan niet worden losgekoppeld.
