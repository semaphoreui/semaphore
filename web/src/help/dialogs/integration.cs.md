---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Upravit integraci

Integrace spustí šablonu úlohy nebo workflow, když jiná služba zavolá webhookovou URL Semaphore,
například při události push z GitHubu. URL se skládá z aliasu, `/api/integrations/` a náhodného
kódu: aliasy projektu jsou na stránce Integrace, aliasy jedné integrace na její vlastní stránce.
Tato stránka obsahuje také matchery a hodnoty, které integrace bere z požadavku.

## Name {#integrationForm-name}

Název, pod kterým je integrace uvedena na stránce Integrace. Úlohy, které integrace spustí, se
jí v auditním logu připisují. Pole je povinné.

## Spouští {#integrationForm-targetToggle}

Zobrazí se, když server podporuje workflow.

Co integrace spouští: **Template** spustí úlohu šablony úlohy, **Workflow** spustí běh workflow.
Přepnutím na **Workflow** se zahodí parametry úlohy níže.

## Workflow to run {#integrationForm-workflowId}

Zobrazí se, když je **Spouští** nastaveno na **Workflow**.

Workflow, které požadavek spustí. Každá hodnota, kterou integrace z požadavku vezme, se předá
workflow jako odpověď na jeho vstupní proměnnou se stejným názvem. Integrace s workflow vyžadují
edici Pro: bez ní je pole neaktivní a server změnu odmítne.

## Task Template to run {#integrationForm-templateId}

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

## Auth method {#integrationForm-authMethod}

Jak Semaphore ověřuje, že požadavek pochází od služby, kterou očekáváte. Požadavek, který
kontrolou neprojde, nic nespustí a webhook na něj odpoví stejně jako na kterýkoli jiný
požadavek.

- **None**: žádná kontrola. Úlohu může spustit kdokoli, kdo zná URL.
- **GitHub Webhooks**: hlavička `X-Hub-Signature-256` musí obsahovat podpis HMAC-SHA256 těla
  vytvořený pomocí tajného údaje, tak jak ho posílá GitHub.
- **Bitbucket Webhooks**: stejná kontrola hlavičky `X-Hub-Signature`, tak jak ji posílá
  Bitbucket.
- **Token**: hlavička uvedená v **Auth header** musí obsahovat samotný tajný údaj.
- **HMAC** a **HMAC (SHA-512)**: hlavička uvedená v **Auth header** musí obsahovat HMAC-SHA256
  nebo HMAC-SHA512 těla vytvořený pomocí tajného údaje, šestnáctkově malými písmeny, bez předpony
  jako `sha256=`.
- **BasicAuth**: požadavek musí v hlavičce `Authorization` nést přihlašovací jméno a heslo klíče
  z **Autentizační tajemství**.

## Auth header {#integrationForm-authHeader}

Zobrazí se, když je **Auth method** nastaveno na **Token**, **HMAC** nebo **HMAC (SHA-512)**.

Název hlavičky požadavku, která nese token nebo podpis, například `X-Webhook-Token`. Na velikosti
písmen v názvu nezáleží. Požadavek bez této hlavičky je odmítnut.

## Autentizační tajemství {#integrationForm-authSecretId}

Zobrazí se, když **Auth method** není **None**.

Klíč z Úložiště klíčů, proti kterému se požadavek ověřuje. Pro **BasicAuth** se hodí pouze klíč
typu **Přihlášení heslem**. Pro ostatní metody se hodí i klíč typu **Řetězec**: kontrola použije
jeho hodnotu, nebo heslo klíče **Přihlášení heslem**.

Pole potřebuje každá metoda kromě **None**: bez klíče, nebo s prázdným tajným údajem, je každý
požadavek odmítnut.
