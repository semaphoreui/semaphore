---
form: src/components/TemplateVaults.vue
---

# Kluis

Een kluis geeft Ansible het wachtwoord van de bestanden en variabelen die met Ansible Vault in de
playbookrepository zijn versleuteld. Elke kluis van het sjabloon wordt één `--vault-id`-optie van
`ansible-playbook`. De kluis wordt met het sjabloon opgeslagen, wanneer u het
sjabloondialoogvenster opslaat.

## Kluis ID {#templateVaultDialog-name}

Optioneel. De vault ID waarvoor het wachtwoord is, het label waarmee de inhoud is versleuteld
(`ansible-vault encrypt --vault-id prod@prompt`). Laat het veld leeg voor de vault ID `default`.

De ID wordt in kleine letters opgeslagen. Elke kluis van het sjabloon heeft een eigen ID nodig, en
er mag er maar één leeg zijn.

## Type {#templateVaultDialog-type}

Waar het wachtwoord vandaan komt:

- **Wachtwoord**: een sleutel uit de Sleutel Opslag. Semaphore typt het wachtwoord ervan wanneer
  Ansible erom vraagt, zodat het nooit naar een bestand wordt geschreven.
- **Client Script**: een vault password client script dat het wachtwoord afdrukt; Ansible voert
  het uit.

## Kluis Wachtwoord {#templateVaultDialog-vaultKeyId}

Wordt getoond wanneer het type **Wachtwoord** is.

De sleutel die het vault-wachtwoord bevat. Alleen sleutels van het type **Inloggen met
wachtwoord** worden vermeld; Semaphore gebruikt hun wachtwoord en negeert de login. Het veld is
verplicht.

## Script Pad {#templateVaultDialog-script}

Wordt getoond wanneer het type **Client Script** is.

Het pad naar het script, bijvoorbeeld `scripts/vault-pass-client.py`. Een relatief pad wordt
genomen vanaf de map waarin Ansible draait: de hoofdmap van de repository, of de werkmap van het
sjabloon wanneer het er een heeft. Het veld is verplicht.

De bestandsnaam moet eindigen op `-client`, met of zonder extensie: alleen dan voert Ansible het
bestand uit als client script en geeft het de vault ID door met `--vault-id`. Het script draait op
de machine die de taak uitvoert, als de gebruiker waaronder Semaphore draait.
