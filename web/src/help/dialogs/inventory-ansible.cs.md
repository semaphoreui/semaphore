---
form: src/components/InventoryForm.vue
---

# Upravit Ansible Inventory

Inventář určuje, na kterých hostech Ansible spustí playbook, a obsahuje přihlašovací údaje,
kterými se k nim Semaphore připojuje. Šablona úlohy odkazuje na inventář a každá úloha této
šablony se spouští proti němu. Formát inventáře popisuje
[průvodce inventářem Ansible](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html).

## Název {#inventoryForm-name}

Název, pod kterým je inventář uveden na stránce Inventář a ve výběru inventáře v šabloně úlohy.
Zvolte název, který hosty od sebe odliší, například `Production web servers`.

## Tag runneru {#inventoryForm-runnerTag}

Zobrazí se, když jsou k dispozici runnery projektu.

Úlohy, které používají tento inventář, běží na runneru s tímto tagem. Tag runneru nastavený
v šabloně úlohy má přednost před tagem inventáře. Ponechte pole prázdné, aby úlohy mohl převzít
kterýkoli runner.

## Přihlašovací údaje uživatele {#inventoryForm-sshKeyId}

Klíč z Úložiště klíčů, kterým se Ansible přihlašuje k hostům. Pole je povinné a přijímá tři typy
klíčů:

- **SSH klíč**: Ansible se připojí pomocí soukromého klíče a, pokud ho klíč má, i pomocí jeho
  přihlašovacího jména.
- **Přihlášení heslem**: Ansible se připojí pomocí přihlašovacího jména a hesla.
- **Žádný**: Semaphore nepředává žádné přihlašovací údaje. Použijte ho pro hosty, kteří žádné
  nepotřebují, například `localhost`, nebo když způsob připojení určuje samotný inventář.

## Přihlašovací údaje sudo {#inventoryForm-becomeKeyId}

Volitelné. Klíč typu **Přihlášení heslem** z Úložiště klíčů, který se použije, když play zvyšuje
oprávnění pomocí `become`: jeho přihlašovací jméno je uživatel, kterým se Ansible stane, a jeho
heslo je heslo, kterým Ansible odpoví na výzvu.

Ponechte pole prázdné, když playe oprávnění nezvyšují, nebo když uživatel uvedený výše k tomu
heslo nepotřebuje.

## Typ {#inventoryForm-type}

Odkud inventář pochází:

- **Static**: inventář se zadává v tomto dialogu ve formátu INI.
- **Static YAML**: inventář se zadává v tomto dialogu ve formátu YAML. Použijte ho také pro
  konfiguraci pluginu inventáře.
- **File**: inventář je soubor v repozitáři Git.
- **Managed**: hosté a jejich proměnné jsou uloženi jako záznamy. Chcete-li je upravit, otevřete
  inventář ze seznamu.

Pole pod typem se mění podle něj.

## Cesta k souboru inventáře {#inventoryForm-inventoryPath}

Zobrazí se, když je typ **File**.

Cesta k souboru inventáře od kořene jeho repozitáře, například `inventory/production.yml`.
Repozitář je ten, který je vybraný v poli Repozitář, nebo, když je toto pole prázdné, repozitář
šablony úlohy, která inventář používá.

## Repozitář {#inventoryForm-repositoryId}

Zobrazí se, když je typ **File**. Volitelné.

Repozitář, který obsahuje soubor inventáře. Semaphore před každou úlohou aktualizuje svou kopii
repozitáře. Ponechte pole prázdné, když je soubor ve stejném repozitáři jako playbook: cesta se
pak hledá v repozitáři šablony úlohy.

## Inventář {#inventoryForm-inventoryCode}

Zobrazí se, když je typ **Static** nebo **Static YAML**.

Text inventáře. Semaphore ho uloží a pro každou úlohu ho zapíše do dočasného souboru. Statický
inventář ve formátu INI:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

Tytéž hosty jako **Static YAML**:

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
