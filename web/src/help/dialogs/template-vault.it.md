---
form: src/components/TemplateVaults.vue
---

# Vault

Un vault fornisce ad Ansible la password dei file e delle variabili cifrati con Ansible Vault nel
repository del playbook. Ogni vault del modello diventa un'opzione `--vault-id` di
`ansible-playbook`. Il vault viene salvato con il modello, quando salvi la finestra di dialogo del
modello.

## ID Vault {#templateVaultDialog-name}

Opzionale. L'ID vault a cui si riferisce la password, l'etichetta con cui il contenuto è stato
cifrato (`ansible-vault encrypt --vault-id prod@prompt`). Lascialo vuoto per l'ID vault `default`.

L'ID viene salvato in minuscolo. Ogni vault del modello ha bisogno di un proprio ID, e solo uno può
essere vuoto.

## Tipo {#templateVaultDialog-type}

Da dove proviene la password:

- **Password**: una chiave dal Negozio chiavi. Semaphore digita la sua password quando Ansible la
  chiede, quindi non viene mai scritta in un file.
- **Script client**: uno script client per la password del vault che stampa la password; Ansible
  lo esegue.

## Password del Vault {#templateVaultDialog-vaultKeyId}

Visibile quando il tipo è **Password**.

La chiave che contiene la password del vault. Sono elencate solo le chiavi di tipo **Accesso con
password**; Semaphore usa la loro password e ignora il nome utente. Il campo è obbligatorio.

## Percorso script {#templateVaultDialog-script}

Visibile quando il tipo è **Script client**.

Il percorso dello script, ad esempio `scripts/vault-pass-client.py`. Un percorso relativo parte
dalla directory in cui viene eseguito Ansible: la radice del repository, o la directory di lavoro
del modello quando ne ha una. Il campo è obbligatorio.

Il nome del file deve terminare con `-client`, con o senza estensione: solo allora Ansible esegue
il file come script client e gli passa l'ID vault con `--vault-id`. Lo script viene eseguito sulla
macchina che esegue il compito, con l'utente con cui è in esecuzione Semaphore.
