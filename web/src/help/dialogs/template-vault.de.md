---
form: src/components/TemplateVaults.vue
---

# Vault

Ein Vault gibt Ansible das Passwort der Dateien und Variablen, die im Playbook-Repository mit
Ansible Vault verschlüsselt sind. Jeder Vault der Vorlage wird zu einer Option `--vault-id` von
`ansible-playbook`. Der Vault wird mit der Vorlage gespeichert, wenn Sie den Vorlagendialog
speichern.

## Vault ID {#templateVaultDialog-name}

Optional. Die Vault-ID, für die das Passwort gilt, die Bezeichnung, mit der der Inhalt
verschlüsselt wurde (`ansible-vault encrypt --vault-id prod@prompt`). Lassen Sie das Feld leer
für die Vault-ID `default`.

Die ID wird in Kleinbuchstaben gespeichert. Jeder Vault der Vorlage braucht eine eigene ID, und
nur eine darf leer sein.

## Typ {#templateVaultDialog-type}

Woher das Passwort kommt:

- **Passwort**: ein Schlüssel aus dem Schlüsselspeicher. Semaphore gibt sein Passwort ein, wenn
  Ansible danach fragt, sodass es nie in eine Datei geschrieben wird.
- **Client-Skript**: ein Vault-Passwort-Client-Skript, das das Passwort ausgibt; Ansible führt
  es aus.

## Vault-Passwort {#templateVaultDialog-vaultKeyId}

Wird angezeigt, wenn der Typ **Passwort** ist.

Der Schlüssel, der das Vault-Passwort enthält. Nur Schlüssel vom Typ **Login mit Passwort**
werden aufgeführt; Semaphore verwendet ihr Passwort und ignoriert den Login. Das Feld ist ein
Pflichtfeld.

## Skriptpfad {#templateVaultDialog-script}

Wird angezeigt, wenn der Typ **Client-Skript** ist.

Der Pfad zum Skript, zum Beispiel `scripts/vault-pass-client.py`. Ein relativer Pfad wird ab dem
Verzeichnis gerechnet, in dem Ansible läuft: dem Stammverzeichnis des Repositorys oder dem
Arbeitsverzeichnis der Vorlage, wenn sie eines hat. Das Feld ist ein Pflichtfeld.

Der Dateiname muss auf `-client` enden, mit oder ohne Dateiendung: Nur dann führt Ansible die
Datei als Client-Skript aus und übergibt ihr die Vault-ID mit `--vault-id`. Das Skript läuft auf
dem Rechner, der den Task ausführt, als der Benutzer, unter dem Semaphore läuft.
