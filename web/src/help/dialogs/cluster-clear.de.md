---
form: src/views/Cluster.vue
---

# Clear tasks from Redis

Die Knoten eines Clusters teilen den Zustand des Task-Pools in Redis: welche Tasks warten, welche
laufen und welcher Knoten jeden von ihnen besitzt. Dieser Dialog löscht die angehakten Gruppen von
Einträgen aus Redis, um einen festgefahrenen Cluster wiederherzustellen, zum Beispiel mit einem
beendeten Task, der noch als laufend angezeigt wird. Aus der Datenbank wird nichts gelöscht: Tasks
behalten ihren Status, ihr Log und ihren Verlauf.

Was immer Sie anhaken, jeder Knoten verwirft außerdem die Tasks, die er im Speicher hält, und liest
sie erneut aus Redis und der Datenbank. Beim Öffnen des Dialogs ist nur **Queue** angehakt; haken
Sie nur die Gruppen an, die festhängen. Das Ergebnis zeigt, wie viele Einträge entfernt wurden.

## Queue {#clusterClearDialog-queue}

Die Tasks, die darauf warten, dass ein Knoten sie startet. Geleerte Tasks werden nicht gestartet:
Sie bleiben mit dem Status Waiting in der Datenbank. Stoppen Sie sie oder führen Sie ihre Vorlagen
erneut aus.

## Running {#clusterClearDialog-running}

Die Tasks, die der Cluster als laufend zählt. Das Leeren stoppt sie nicht: Ein Task läuft auf
seinem Runner weiter. Der Cluster verliert ihn aber aus dem Blick: Der Task zählt nicht mehr gegen
die Limits paralleler Tasks des Servers und seines Runners, sodass mehr Tasks starten können, als
die Limits erlauben, und ein Task, den ein Runner noch nicht übernommen hat, wird ihm nicht mehr
übergeben.

## Active by project {#clusterClearDialog-active}

Die nicht beendeten Tasks jedes Projekts. Sie setzen das Limit paralleler Tasks eines Projekts
durch und verhindern, dass ein zweiter Task einer Vorlage startet, während einer läuft, wenn die
Vorlage keine parallelen Tasks erlaubt. Leeren Sie diese Gruppe, wenn die Tasks eines Projekts
endlos hinter Tasks warten, die bereits beendet sind. Tasks, die in diesem Moment laufen, zählen
nicht mehr, sodass neue Tasks des Projekts neben ihnen starten können.

## Aliases {#clusterClearDialog-aliases}

Die zufälligen Namen, die Terraform-, OpenTofu- und Terragrunt-Tasks für ihre Laufzeit erhalten,
dem Task zugeordnet. Ein laufender Task erreicht seinen Terraform-State auf dem Server über diesen
Namen: Nach dem Leeren kann ein laufender Task dieser Apps darüber nicht mehr gefunden werden.

## Claims {#clusterClearDialog-claims}

Welcher Knoten jeden Task besitzt. Ein Knoten erneuert die Claims seiner Tasks jede Minute; ein
Claim, der nicht erneuert wird, läuft nach fünf Minuten ab. Ein Task ohne Claim gilt als verwaist:
Ein Task, der noch keinen Runner erreicht hat, wird für einen anderen Knoten zurück in die
Warteschlange gestellt, und ein Task, den ein Runner bereits ausführt, darf dort zu Ende laufen.

## Runtime fields {#clusterClearDialog-runtimeFields}

Die Angaben zu jedem Task, die die Datenbank nicht enthält: der Benutzer, der ihn gestartet hat,
die Version, mit der er gestartet wurde, sein Alias, sein Runner und sein Status sowie das Projekt
jedes Tasks. Ohne sie zeigt ein Knoten, der einen Task nicht gestartet hat, weniger über ihn an,
und die Bereinigung verwaister Tasks überspringt einen laufenden Task, sodass ein festhängender
Eintrag nicht mehr von selbst entfernt wird.
