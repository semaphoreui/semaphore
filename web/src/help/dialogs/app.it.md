---
form: src/components/AppForm.vue
---

# Edit App

Un'app è uno strumento con cui Semaphore esegue i compiti: Ansible, Terraform, OpenTofu, Terragrunt,
Bash, Python, PowerShell, o un tuo comando. Ogni modello di compito è creato per un'app, e questa
finestra di dialogo imposta come il server la avvia e come l'app viene mostrata nell'interfaccia.

Le impostazioni sono memorizzate nel database e hanno effetto subito sul server che le salva; gli
altri nodi di un cluster le leggono al loro riavvio. Un campo che svuoti mantiene il suo vecchio
valore finché il server non si riavvia.

## ID {#appForm-id}

Obbligatorio. La chiave con cui l'app è memorizzata; un modello di compito fa riferimento alla sua
app tramite questa chiave. Le app integrate sono `ansible`, `terraform`, `tofu`, `terragrunt`,
`pulumi`, `bash`, `python` e `powershell`. Qualsiasi altro ID aggiunge un'app tua: Semaphore la
esegue come uno script, con il comando indicato in **Path** o, quando **Path** è vuoto, con il
comando che ha lo stesso nome dell'ID.

Usa un ID senza punti. Cambiare l'ID di un'app esistente non la rinomina: le impostazioni vengono
salvate come una nuova app con il nuovo ID, e la vecchia app resta.

## Icon {#appForm-icon}

Opzionale. Il nome di un'icona del set Material Design Icons senza il prefisso `mdi-`, ad esempio
`language-python`. Lascialo vuoto per mantenere l'icona integrata dell'app;
un'app tua mostra allora un punto interrogativo.

## Nome {#appForm-title}

Opzionale. Il nome dell'app nell'elenco delle applicazioni e nel menu **Nuovo modello**. Lascialo
vuoto per mantenere il nome integrato, come `Ansible Playbook`; un'app tua non ha allora alcun nome,
quindi dagliene uno.

## Path {#appForm-path}

Opzionale. Il comando che Semaphore esegue al posto di quello predefinito, un nome cercato nel
`PATH` del server o un percorso completo, ad esempio `/usr/local/bin/terraform-1.9`. I valori
predefiniti sono `terraform`, `tofu`, `terragrunt`, `bash`, `python3` e `powershell`; un'app tua
esegue il comando che ha lo stesso nome del suo ID.

Ansible ignora questo campo: Semaphore esegue sempre `ansible-playbook` e `ansible-galaxy`.

## Priority {#appForm-priority}

La posizione dell'app nell'elenco delle applicazioni e nel menu **Nuovo modello**: un'app con un
numero più alto viene prima. Le app integrate vanno da 1000 per Ansible fino a 500 per Python.
Un'app integrata con una priorità di `0` o inferiore riprende la sua priorità integrata all'avvio
del server.

## Args {#appForm-args}

Opzionale. Argomenti inseriti nella riga di comando subito dopo il comando, uno per chip. Fai clic
su **+** per aggiungerne uno, fai clic su un chip per modificarlo, trascina i chip per cambiarne
l'ordine.

- Terraform, OpenTofu e Terragrunt: gli argomenti vengono prima degli argomenti che Semaphore
  aggiunge a ogni comando, come `init` e `apply`.
- App di script: gli argomenti vengono prima dello script del modello e sostituiscono quelli
  predefiniti: PowerShell viene eseguito con `-File` per impostazione predefinita, quindi metti
  `-File` per ultimo nel tuo elenco.
- Ansible ignora questo campo.

## Active {#appForm-active}

Un'app attiva viene offerta nel menu **Nuovo modello**. Deseleziona la casella per nascondere l'app
da quel menu: i modelli già creati con essa mantengono la loro app e continuano a essere eseguiti.

La casella è deselezionata per una nuova app: selezionala per rendere l'app disponibile.
