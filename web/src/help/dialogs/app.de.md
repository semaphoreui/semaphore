---
form: src/components/AppForm.vue
---

# Edit App

Eine App ist ein Werkzeug, mit dem Semaphore Tasks ausführt: Ansible, Terraform, OpenTofu,
Terragrunt, Bash, Python, PowerShell oder ein eigener Befehl. Jede Task-Vorlage ist für eine App
gemacht, und dieser Dialog legt fest, wie der Server sie startet und wie die App in der Oberfläche
angezeigt wird.

Die Einstellungen werden in der Datenbank gespeichert und wirken sofort auf dem Server, der sie
speichert; die anderen Knoten eines Clusters lesen sie bei ihrem Neustart. Ein Feld, das Sie
leeren, behält seinen alten Wert, bis der Server neu startet.

## ID {#appForm-id}

Pflichtfeld. Der Schlüssel, unter dem die App gespeichert wird; eine Task-Vorlage verweist über
diesen Schlüssel auf ihre App. Die eingebauten Apps sind `ansible`, `terraform`, `tofu`,
`terragrunt`, `pulumi`, `bash`, `python` und `powershell`. Jede andere ID fügt eine eigene App
hinzu: Semaphore führt sie wie ein Skript aus, mit dem in **Path** genannten Befehl oder, wenn
**Path** leer ist, mit dem Befehl, der wie die ID heißt.

Verwenden Sie eine ID ohne Punkte. Das Ändern der ID einer bestehenden App benennt sie nicht um:
Die Einstellungen werden als neue App unter der neuen ID gespeichert, und die alte App bleibt
bestehen.

## Icon {#appForm-icon}

Optional. Der Name eines Icons aus dem Satz Material Design Icons ohne das Präfix `mdi-`, zum
Beispiel `language-python`. Lassen Sie das Feld leer, um das eingebaute Icon der App zu behalten;
eine eigene App zeigt dann ein Fragezeichen.

## Name {#appForm-title}

Optional. Der Name der App in der Liste der Anwendungen und im Menü **Neue Vorlage**. Lassen Sie
das Feld leer, um den eingebauten Namen zu behalten, etwa `Ansible Playbook`; eine eigene App hat
dann keinen Namen, geben Sie ihr also einen.

## Path {#appForm-path}

Optional. Der Befehl, den Semaphore statt des Standardbefehls ausführt, entweder ein Name, der im
`PATH` des Servers gesucht wird, oder ein vollständiger Pfad, zum Beispiel
`/usr/local/bin/terraform-1.9`. Die Standardbefehle sind `terraform`, `tofu`, `terragrunt`,
`bash`, `python3` und `powershell`; eine eigene App führt den Befehl aus, der wie ihre ID heißt.

Ansible ignoriert dieses Feld: Semaphore führt immer `ansible-playbook` und `ansible-galaxy` aus.

## Priority {#appForm-priority}

Die Position der App in der Liste der Anwendungen und im Menü **Neue Vorlage**: Eine App mit einer
höheren Zahl steht weiter oben. Die eingebauten Apps beginnen bei 1000 für Ansible und gehen
hinunter bis 500 für Python. Eine eingebaute App mit einer Priorität von `0` oder weniger erhält
beim Start des Servers ihre eingebaute Priorität zurück.

## Args {#appForm-args}

Optional. Argumente, die in der Befehlszeile direkt nach dem Befehl stehen, eines pro Chip.
Klicken Sie auf **+**, um eines hinzuzufügen, klicken Sie auf einen Chip, um ihn zu ändern, und
ziehen Sie die Chips, um ihre Reihenfolge zu ändern.

- Terraform, OpenTofu und Terragrunt: Die Argumente stehen vor den Argumenten, die Semaphore jedem
  Befehl hinzufügt, etwa `init` und `apply`.
- Skript-Apps: Die Argumente stehen vor dem Skript der Vorlage und ersetzen die
  Standardargumente: PowerShell läuft standardmäßig mit `-File`, setzen Sie `-File` also an das
  Ende Ihrer Liste.
- Ansible ignoriert dieses Feld.

## Active {#appForm-active}

Eine aktive App wird im Menü **Neue Vorlage** angeboten. Entfernen Sie den Haken, um die App in
diesem Menü auszublenden: Bereits damit erstellte Vorlagen behalten ihre App und laufen weiterhin.

Bei einer neuen App ist das Kästchen nicht angehakt: Haken Sie es an, um die App verfügbar zu
machen.
