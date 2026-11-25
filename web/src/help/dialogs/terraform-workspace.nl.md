---
form: src/components/TerraformInventoryForm.vue
---

# Nieuw Terraform Workspace

Een workspace is de inventaris van een Terraform-, OpenTofu- of Terragrunt-sjabloon: hij noemt de
workspace die een taak na `init` selecteert, en de sleutel waarmee `init` privémodules downloadt.
De workspace die u hier aanmaakt, wordt aan dit sjabloon gekoppeld en boven de state ervan
vermeld. Taken van het sjabloon draaien in de standaardworkspace, die met een vinkje is
gemarkeerd; **Make default** wijzigt die. Hetzelfde dialoogvenster bewerkt een workspace wanneer
u het opent met de potloodknop.

## Workspace name {#tfWorkspace-inventory}

De naam van de Terraform-workspace, bijvoorbeeld `production`. Het veld is verplicht.

Vóór een plan voert Semaphore `workspace select -or-create=true` uit met deze naam, zodat de
workspace wordt aangemaakt wanneer hij nog niet bestaat. Wanneer `workspace list` mislukt,
bijvoorbeeld omdat de backend geen workspaces ondersteunt, slaat Semaphore de selectie over.

Semaphore slaat de workspace op als `TEMPLATE - NAME`, de naam van het sjabloon en deze naam: u
ziet die volledige naam op de pagina Inventaris en wanneer u de workspace aan een sjabloon
koppelt. Het hernoemen van een workspace hernoemt hem niet in Terraform: de volgende taak
selecteert, of maakt, een workspace met de nieuwe naam.

## SSH key for private modules {#tfWorkspace-sshKeyId}

De sleutel die `init` gebruikt om modules via SSH uit private Git-repositories te downloaden. Het
veld is in dit dialoogvenster verplicht. De lijst biedt de sleutels van het type **SSH Sleutel**,
**Inloggen met wachtwoord** en **Geen** uit de Sleutel Opslag aan.

Alleen een **SSH Sleutel** heeft effect: Semaphore laadt hem in een SSH-agent zolang `init` duurt,
en Git gebruikt hem voor modulebronnen zoals `git::ssh://git@example.com/modules/vpc.git`. Een
sleutel van het type **Inloggen met wachtwoord** geeft geen referenties aan Git door. Kies
**Geen** wanneer de modules geen sleutel nodig hebben.
