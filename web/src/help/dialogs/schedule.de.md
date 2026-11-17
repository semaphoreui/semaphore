---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Zeitplan bearbeiten

Ein Zeitplan startet einen Task einer Task-Vorlage, ohne dass jemand auf **Ausführen** klickt:
immer wieder nach einem Cron-Ausdruck oder einmal zu einem festgelegten Zeitpunkt. Der Task
erhält die in diesem Dialog gespeicherten Parameter. Alle Zeiten gelten in der Zeitzone des
Semaphore-Servers, die neben den Zeitfeldern angezeigt wird: `UTC`, sofern der Administrator nicht
`schedule.timezone` oder `SEMAPHORE_SCHEDULE_TIMEZONE` setzt.

## Name {#scheduleForm-name}

Der Name, unter dem der Zeitplan auf der Seite Zeitplan aufgeführt wird. Das Feld ist ein
Pflichtfeld. Er ändert nicht, was der Zeitplan tut.

## Template {#scheduleForm-templateId}

Die Task-Vorlage, die der Zeitplan startet. Das Feld ist ein Pflichtfeld. Der Task läuft, als
wäre er aus der Vorlage gestartet worden, mit dem Inventory, dem Repository, der Variablengruppe
und den übrigen Einstellungen der Vorlage und mit den Parametern unten, wo die Vorlage einen Task
sie ändern lässt.

## Task-Parameter {#scheduleForm-taskParams}

Wird angezeigt, wenn eine Vorlage ausgewählt ist.

Die Parameter jedes Tasks, den der Zeitplan startet. Welche Felder erscheinen, hängt von der
Vorlage ab: ihre Survey-Variablen und die Optionen, die sie einen Task ändern lässt, etwa den
Branch, das Inventory, **Limit**, **Tags**, **Tags überspringen** oder **CLI args**, sowie die
Optionen von Ansible oder von Terraform, OpenTofu und Terragrunt.

Der Server prüft die Vorlage bei jeder Auslösung des Zeitplans: Ein hier gespeicherter Wert wird
ignoriert, sobald die Vorlage einem Task nicht mehr erlaubt, ihn zu ändern.

## Nachricht {#taskParams-message}

Optional. Eine Notiz, die mit jedem vom Zeitplan gestarteten Task gespeichert und im Verlauf der
Tasks angezeigt wird, zum Beispiel `Nightly deploy`. Sie ändert nicht, wie der Task läuft.

## Survey-Variablen {#taskParams-survey}

Wird angezeigt, wenn die Vorlage Survey-Variablen hat.

Die Werte der Survey-Variablen der Vorlage für die geplanten Tasks. Sie erreichen den Task auf
dieselbe Weise, wie wenn jemand den Survey von Hand ausfüllt. Eine Pflichtvariable muss einen Wert
haben, bevor der Zeitplan gespeichert werden kann.

Die Standardwerte der Vorlage werden beim Öffnen des Dialogs übernommen und mit dem Zeitplan
gespeichert. Ein später in der Vorlage geänderter Standardwert erreicht einen zuvor gespeicherten
Zeitplan nicht.

Der Wert einer Variablen vom Typ Secret wird nicht mit dem Zeitplan gespeichert: Der geplante Task
läuft ohne ihn.

## Branch {#taskParams-gitBranch}

Wird angezeigt, wenn die Vorlage einem Task erlaubt, den Branch zu überschreiben.

Der Branch oder der Tag des Repositorys, den der geplante Task auscheckt. Das Feld beginnt mit dem
Branch der Vorlage. Der Server verwendet den Wert nur, solange die Vorlage das Überschreiben
erlaubt.

## Inventory {#taskParams-inventoryId}

Wird angezeigt, wenn die Vorlage einem Task erlaubt, das Inventory zu wählen.

Das Inventory, gegen das der geplante Task anstelle des Inventorys der Vorlage läuft. Bei einer
Terraform- oder OpenTofu-Vorlage heißt das Feld **Workspace**, der Server wendet die Auswahl aber
nur auf Ansible-Vorlagen an und behält bei den anderen den Workspace der Vorlage bei.

## Debug {#taskParams-debug}

Wird bei einer Ansible-Vorlage angezeigt, die Debugging erlaubt.

Führt `ansible-playbook` mit mehr Ausgabe aus: `-v` so oft wiederholt, wie **Debug-Level**
festlegt.

## Debug-Level {#taskParams-debugLevel}

Wird bei einer Ansible-Vorlage angezeigt, die Debugging erlaubt; aktiv, wenn **Debug** angehakt
ist.

Wie ausführlich die Ausgabe ist, von `1` (`-v`) bis `6` (`-vvvvvv`). Der Standardwert ist `4`.

## Testlauf {#taskParams-dryRun}

Wird bei einer Ansible-Vorlage angezeigt, sofern die Vorlage es nicht ausblendet.

Führt das Playbook mit `--check` aus: Ansible meldet, was es ändern würde, und ändert nichts.

## Diff {#taskParams-diff}

Wird bei einer Ansible-Vorlage angezeigt, sofern die Vorlage es nicht ausblendet.

Führt das Playbook mit `--diff` aus: Ansible zeigt die Änderungen, die es an Dateien und Templates
vornimmt.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Wird bei einer Ansible-Vorlage angezeigt, die einem Task erlaubt, es zu ändern.

Überspringt die Installation der Rollen und Collections aus den Dateien `requirements.yml`, bevor
das Playbook läuft. Haken Sie es an, wenn die Anforderungen bereits installiert sind und der Task
Ansible Galaxy nicht erreichen soll.

## Plan {#taskParams-plan}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Führt nur `plan` aus. Der Task endet nach dem Plan und wendet nichts an.

## Destroy {#taskParams-destroy}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Fügt dem Plan und dem Apply `-destroy` hinzu: Der Task zerstört die Ressourcen des Workspace,
statt sie zu erstellen oder zu aktualisieren.

## Auto Approve {#taskParams-autoApprove}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Wendet den Plan an, ohne auf eine Bestätigung zu warten. Das funktioniert nur, wenn die Vorlage
automatisches Akzeptieren erlaubt; eine Vorlage, die so eingestellt ist, dass sie jeden Task
akzeptiert, wendet den Plan in jedem Fall an.

Ohne automatisches Akzeptieren hält ein geplanter Task nach dem Plan an und wartet, bis jemand ihn
auf der Seite des Tasks bestätigt oder ablehnt.

## Upgrade {#taskParams-upgrade}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Führt `init` mit `-upgrade` aus: Die Provider und Module werden auf die neuesten Versionen
aktualisiert, die die Konfiguration erlaubt.

## Reconfigure {#taskParams-reconfigure}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Führt `init` mit `-reconfigure` statt `-migrate-state` aus: Das Backend wird neu eingerichtet, und
der bestehende State wird nicht migriert.

## Run at {#scheduleForm-runAt}

Wird bei einem Zeitplan angezeigt, der einmal läuft.

Datum und Uhrzeit, zu denen der Task startet, in der Zeitzone des Servers, die am Ende des Feldes
angezeigt wird. Ein neuer Zeitplan beginnt mit der nächsten vollen Stunde. Der Zeitpunkt muss in
der Zukunft liegen: Der Server lehnt eine vergangene Zeit ab.

Nachdem der Task gestartet ist, wird der Zeitplan ausgeschaltet oder, wenn **Delete after run**
angehakt ist, gelöscht. Läuft Semaphore zu diesem Zeitpunkt nicht, wird der Task nicht später
gestartet: Beim nächsten Start wird der Zeitplan ohne Lauf ausgeschaltet oder gelöscht.

## Delete after run {#scheduleForm-deleteAfterRun}

Wird bei einem Zeitplan angezeigt, der einmal läuft.

Löscht den Zeitplan, nachdem sein Task gestartet ist. Lassen Sie das Kästchen leer, um den
Zeitplan ausgeschaltet zu behalten, sodass Sie eine neue Zeit festlegen und ihn wieder
einschalten können.

## Show cron format {#scheduleForm-rawCron}

Wird bei einem Cron-Zeitplan angezeigt.

Wechselt zwischen der Eingabe des Cron-Ausdrucks (ein) und seinem Aufbau aus **Timing** und den
Zeit-Kästchen (aus). Beide bearbeiten denselben Ausdruck. Der Browser merkt sich die Wahl.

Der Schalter ist fest eingeschaltet, wenn der Ausdruck nicht als Kästchen dargestellt werden
kann: wenn er ungültig ist oder eine Form wie `@daily` oder `CRON_TZ=` verwendet.

## Cron {#scheduleForm-cronFormat}

Wird bei einem Cron-Zeitplan angezeigt, wenn **Show cron format** eingeschaltet ist.

Der Cron-Ausdruck in fünf Feldern: Minute, Stunde, Tag des Monats, Monat, Wochentag (`0` ist
Sonntag). Das Feld ist ein Pflichtfeld, und der Server prüft den Ausdruck während der Eingabe. Ein
neuer Zeitplan beginnt mit `* * * * *`, jede Minute.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

Der Server akzeptiert außerdem `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` und
`@every 1h30m` sowie ein Präfix wie `CRON_TZ=Europe/Berlin` für eine andere Zeitzone. Die Vorschau
darunter kann den nächsten Lauf dieser Ausdrücke nicht anzeigen.

## Timing {#scheduleForm-timing}

Wird bei einem Cron-Zeitplan angezeigt, wenn **Show cron format** ausgeschaltet ist.

Wie oft der Task läuft. Die Kästchen darunter ändern sich damit:

- **Hourly**: **Minutes**.
- **Daily**: **Hours** und **Minutes**.
- **Weekly**: **Weekdays**, **Hours** und **Minutes**.
- **Monthly**: **Days**, **Hours** und **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** und **Minutes**.

## Months {#scheduleForm-months}

Wird angezeigt, wenn **Timing** auf **Yearly** steht.

Die Monate, in denen der Task läuft. Haken Sie mehrere an, damit er in jedem davon läuft. Ist
keiner angehakt, läuft der Task in jedem Monat.

## Weekdays {#scheduleForm-weekdays}

Wird angezeigt, wenn **Timing** auf **Weekly** steht.

Die Wochentage, an denen der Task läuft. Ist keiner angehakt, läuft der Task an jedem Tag.

## Days {#scheduleForm-days}

Wird angezeigt, wenn **Timing** auf **Monthly** oder **Yearly** steht.

Die Tage des Monats, an denen der Task läuft. Ein Monat ohne diesen Tag wird übersprungen: Ein
Task, der auf den 31. eingestellt ist, läuft im April nicht. Ist keiner angehakt, läuft der Task
an jedem Tag.

## Hours {#scheduleForm-hours}

Wird angezeigt, wenn **Timing** auf **Daily**, **Weekly**, **Monthly** oder **Yearly** steht.

Die Stunden, zu denen der Task läuft, in der rechts angezeigten Zeitzone. Ist keine angehakt,
läuft der Task in jeder Stunde.

## Minutes {#scheduleForm-minutes}

Wird bei einem Cron-Zeitplan angezeigt, wenn **Show cron format** ausgeschaltet ist.

Die Minuten der Stunde, zu denen der Task läuft, in Schritten von fünf. Ist keine angehakt, läuft
der Task jede Minute: Haken Sie mindestens eine an, zum Beispiel `:00`, für einen Task, der einmal
pro Stunde oder einmal pro Tag laufen soll.

## Next run time {#scheduleForm-nextRun}

Wann der nächste Task starten würde, gemäß dem Zeitplan, wie er im Formular steht: in der Zeitzone
des Servers und in der Ortszeit Ihres Browsers. Ein Strich bedeutet, dass der Ausdruck ungültig
ist oder einer, den die Vorschau nicht lesen kann, etwa `@daily`.

## Aktiviert {#scheduleForm-active}

Ob der Zeitplan Tasks startet. Entfernen Sie den Haken, um einen Zeitplan zu behalten, ohne ihn
auszuführen. Der Schalter in der Liste der Zeitpläne bewirkt dasselbe. Ein neuer Zeitplan ist
aktiviert.
