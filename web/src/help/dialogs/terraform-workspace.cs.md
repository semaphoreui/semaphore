---
form: src/components/TerraformInventoryForm.vue
---

# Nový Terraform Workspace

Workspace je inventář šablony Terraform, OpenTofu nebo Terragrunt: určuje workspace, který úloha
vybere po `init`, a klíč, kterým `init` stahuje soukromé moduly. Workspace, který zde vytvoříte,
se připojí k této šabloně a je uveden nad jejím stavem. Úlohy šablony běží v jejím výchozím
workspace, tom označeném zaškrtnutím; **Make default** ho změní. Stejný dialog upravuje
workspace, když ho otevřete tlačítkem s tužkou.

## Workspace name {#tfWorkspace-inventory}

Název workspace v Terraformu, například `production`. Pole je povinné.

Před plánem Semaphore spustí `workspace select -or-create=true` s tímto názvem, takže se
workspace vytvoří, když ještě neexistuje. Když `workspace list` selže, například proto, že backend
workspaces nepodporuje, Semaphore výběr přeskočí.

Semaphore ukládá workspace jako `TEMPLATE - NAME`, tedy název šablony a tento název: tento úplný
název uvidíte na stránce Inventář a při připojování workspace k šabloně. Přejmenování workspace
ho v Terraformu nepřejmenuje: další úloha vybere, nebo vytvoří, workspace s novým názvem.

## SSH key for private modules {#tfWorkspace-sshKeyId}

Klíč, kterým `init` stahuje moduly ze soukromých repozitářů Git přes SSH. V tomto dialogu je pole
povinné. Seznam nabízí klíče typu **SSH klíč**, **Přihlášení heslem** a **Žádný** z Úložiště
klíčů.

Účinek má pouze **SSH klíč**: Semaphore ho po dobu `init` načte do SSH agenta a Git ho použije
pro zdroje modulů jako `git::ssh://git@example.com/modules/vpc.git`. Klíč typu **Přihlášení
heslem** nepředává Gitu žádné přihlašovací údaje. Když moduly žádný klíč nepotřebují, zvolte
**Žádný**.
