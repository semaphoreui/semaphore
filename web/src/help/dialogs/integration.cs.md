---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Upravit spuštění

Spuštění je to, co integrace spustí, když jiná služba zavolá webhook URL Semaphore, například při
události push na GitHubu: úlohu šablony úlohy nebo workflow. URL a kontrola požadavku patří k
endpointu integrace, ne ke spuštění. Stránka spuštění obsahuje jeho matchery a hodnoty, které
přebírá z požadavku.

## Název {#integrationForm-name}

Název, pod kterým je integrace uvedena na stránce Integrace. Úlohy, které integrace spustí, se
jí v auditním logu připisují. Pole je povinné.

## Integrace {#integrationForm-groupId}

Integrace, ke které spuštění patří. Spuštění se provede při požadavcích, které dorazí na endpoint
této integrace a projdou její autorizací. Pole je povinné.

## Cíl {#integrationForm-targetToggle}

Zobrazí se, když server podporuje workflow.

Co integrace spouští: **Template** spustí úlohu šablony úlohy, **Workflow** spustí běh workflow.
Přepnutím na **Workflow** se zahodí parametry úlohy níže.

## Workflow to run {#integrationForm-workflowId}

Zobrazí se, když je **Spouští** nastaveno na **Workflow**.

Workflow, které požadavek spustí. Každá hodnota, kterou integrace z požadavku vezme, se předá
workflow jako odpověď na jeho vstupní proměnnou se stejným názvem. Integrace s workflow vyžadují
edici Pro: bez ní je pole neaktivní a server změnu odmítne.

## Šablona úlohy ke spuštění {#integrationForm-templateId}

Zobrazí se, když je **Spouští** nastaveno na **Template**.

Šablona úlohy, jejíž úlohu požadavek spustí. Pole je povinné. Jakmile je šablona vybrána, pod ní
se objeví parametry jejích úloh.

## Parametry úlohy {#integrationForm-taskParams}

Zobrazí se, když je vybrána šablona úlohy.

Parametry, které dostane každá úloha spuštěná touto integrací, jako byste ji ze šablony spustili
sami. Jsou zde pouze volby, které šablona dovoluje v úloze přepsat. Hodnoty vzaté z požadavku je
přepisují: hodnota typu **Variables** odpověď se stejným názvem, hodnota typu **Task Params**
parametr se stejným názvem.

Když šablona dovoluje argumenty CLI v úloze, **CLI args** obsahuje doplňkové argumenty pro
příkazový řádek úlohy.

## Zpráva {#taskParams-message}

Volitelné. Zpráva každé úlohy, kterou integrace spustí, zobrazená v historii úloh.

## Vstupní proměnné {#taskParams-survey}

Zobrazí se, když má šablona vstupní proměnné.

Odpovědi na vstupní proměnné šablony, stejné pro každou úlohu. Hodnota typu **Variables** vzatá
z požadavku nahradí odpověď se stejným názvem.

Odpovědi na proměnné typu secret se s integrací neukládají: tajnou hodnotu místo toho předejte
přes skupinu proměnných šablony.

## Větev / tag {#taskParams-gitBranch}

Zobrazí se, když šablona dovoluje v úloze přepsat větev.

Větev nebo tag repozitáře, na které se úlohy přepnou. Ponechte hodnotu beze změny, chcete-li
použít větev šablony. Hodnota s názvem `git_branch` typu **Task Params** vzatá z požadavku ji
přepíše; úvodní `refs/heads/` se z této hodnoty odstraní.

## Inventář {#taskParams-inventoryId}

Zobrazí se, když šablona dovoluje v úloze přepsat inventář.

Inventář, nebo u Terraformu workspace, proti kterému úlohy běží místo inventáře šablony.

## Volby Ansible {#taskParams-ansible}

Zobrazí se, když šablona spouští playbook Ansible.

Volby `ansible-playbook` pro každou úlohu, každá zobrazená, když ji šablona dovoluje: **Limit**,
**Tagy**, **Přeskočit tagy**, **Ladění**, **Zkušební běh** (`--check`), **Rozdíl** a **Skip
Galaxy install**.

## Volby Terraformu {#taskParams-terraform}

Zobrazí se, když šablona spouští Terraform, OpenTofu nebo Terragrunt.

- **Plan**: úloha spustí plán a zastaví se, nic se neaplikuje.
- **Destroy**: plán a aplikace zničí prostředky.
- **Auto Approve**: plán se aplikuje bez potvrzení, když šablona dovoluje automatické schválení.
  Bez toho úloha čeká, dokud ji někdo v Semaphore nepotvrdí.
- **Upgrade** a **Reconfigure**: `init` se spustí s `-upgrade` a `-reconfigure`.
