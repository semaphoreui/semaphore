---
form: src/components/TerraformInventoryForm.vue
---

# Nuovo Terraform Workspace

Un workspace è l'inventario di un modello Terraform, OpenTofu o Terragrunt: indica il workspace che
un compito seleziona dopo `init` e la chiave con cui `init` scarica i moduli privati. Il workspace
che crei qui viene collegato a questo modello ed elencato sopra il suo stato. I compiti del modello
vengono eseguiti nel suo workspace predefinito, quello contrassegnato da un segno di spunta; **Make
default** lo cambia. La stessa finestra di dialogo modifica un workspace quando la apri con il
pulsante a forma di matita.

## Workspace name {#tfWorkspace-inventory}

Il nome del workspace Terraform, ad esempio `production`. Il campo è obbligatorio.

Prima di un piano, Semaphore esegue `workspace select -or-create=true` con questo nome, quindi il
workspace viene creato quando non esiste ancora. Quando `workspace list` fallisce, ad esempio perché
il backend non supporta i workspace, Semaphore salta la selezione.

Semaphore memorizza il workspace come `TEMPLATE - NAME`, il nome del modello e questo nome: vedi
questo nome completo nella pagina Inventario e quando colleghi il workspace a un modello.
Rinominare un workspace non lo rinomina in Terraform: il compito successivo seleziona, o crea, un
workspace con il nuovo nome.

## SSH key for private modules {#tfWorkspace-sshKeyId}

La chiave che `init` usa per scaricare moduli da repository Git privati tramite SSH. In questa
finestra di dialogo il campo è obbligatorio. L'elenco offre le chiavi di tipo **Chiave SSH**,
**Accesso con password** e **Nessuno** dal Negozio chiavi.

Solo una **Chiave SSH** ha effetto: Semaphore la carica in un agente SSH per la durata di `init`, e
Git la usa per sorgenti di moduli come `git::ssh://git@example.com/modules/vpc.git`. Una chiave di
tipo **Accesso con password** non passa alcuna credenziale a Git. Scegli **Nessuno** quando i moduli
non hanno bisogno di una chiave.
