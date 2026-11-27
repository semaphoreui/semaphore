---
form: src/components/InventorySourceForm.vue
---

# Upravit zdroj

Zdroj načítá hostitele a skupiny spravovaného inventáře ze souboru v repozitáři Git. Soubor se čte při každém stisku Sync a hostitelé i skupiny z něj se udržují v souladu se souborem. Server potřebuje Ansible a přístup k repozitáři.

## Název {#inventorySourceForm-name}

Název zdroje, zobrazuje se na jeho kartě a u hostitelů, které přináší. Pole je povinné.

## Repozitář {#inventorySourceForm-repository}

Repozitář projektu, ze kterého se soubor čte. Pole je povinné. Použije se jeho větev.

## Cesta {#inventorySourceForm-path}

Cesta k souboru inventáře v repozitáři, například `inventory/prod.yml`, nebo k adresáři se soubory inventáře. Pole je povinné.

Soubory se čtou jako statický YAML nebo INI, skripty a konfigurační soubory pluginů jsou odmítnuty. V adresáři musí být každý soubor, který Ansible nepřeskočí, platným inventářem, jinak synchronizace selže a nic se nezmění. Ansible v adresáři přeskakuje některé přípony, například `.cfg`, `.md` a `.txt`. Repozitář s odkazem, který vede mimo něj, je odmítnut.
