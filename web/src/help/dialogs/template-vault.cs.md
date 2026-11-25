---
form: src/components/TemplateVaults.vue
---

# Vault

Vault předává Ansiblu heslo k souborům a proměnným, které jsou v repozitáři playbooku zašifrované
pomocí Ansible Vault. Každý vault šablony se stane jednou volbou `--vault-id` příkazu
`ansible-playbook`. Vault se uloží se šablonou, když uložíte dialog šablony.

## ID Vaultu {#templateVaultDialog-name}

Volitelné. Vault ID, ke kterému heslo patří, tedy označení, se kterým byl obsah zašifrován
(`ansible-vault encrypt --vault-id prod@prompt`). Pro vault ID `default` pole ponechte prázdné.

ID se ukládá malými písmeny. Každý vault šablony potřebuje vlastní ID a prázdné smí být pouze
jedno.

## Typ {#templateVaultDialog-type}

Odkud heslo pochází:

- **Heslo**: klíč z Úložiště klíčů. Semaphore zadá jeho heslo, když se na něj Ansible zeptá,
  takže se nikdy nezapíše do souboru.
- **Klientský skript**: klientský skript pro heslo vaultu, který heslo vypíše; Ansible ho
  spustí.

## Heslo Vaultu {#templateVaultDialog-vaultKeyId}

Zobrazí se, když je typ **Heslo**.

Klíč, který obsahuje heslo vaultu. V seznamu jsou pouze klíče typu **Přihlášení heslem**;
Semaphore použije jejich heslo a přihlašovací jméno ignoruje. Pole je povinné.

## Cesta ke skriptu {#templateVaultDialog-script}

Zobrazí se, když je typ **Klientský skript**.

Cesta ke skriptu, například `scripts/vault-pass-client.py`. Relativní cesta se bere od adresáře,
ve kterém Ansible běží: od kořene repozitáře, nebo od pracovního adresáře šablony, když ho má.
Pole je povinné.

Název souboru musí končit na `-client`, s příponou nebo bez ní: jen tehdy Ansible spustí soubor
jako klientský skript a předá mu vault ID pomocí `--vault-id`. Skript běží na stroji, který úlohu
spouští, pod uživatelem, pod kterým běží Semaphore.
