---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Variablengruppe bearbeiten

Eine Variablengruppe enthält die zusätzlichen Variablen, die Umgebungsvariablen und die Secrets,
die Semaphore an jeden Task einer Task-Vorlage übergibt, die die Gruppe verwendet. Ansible erhält
die zusätzlichen Variablen als `--extra-vars`, Terraform und OpenTofu als `-var`; die
Umgebungsvariablen werden für den Prozess des Tasks gesetzt. Eine Vorlage kann mehrere Gruppen
verwenden: Wenn zwei davon denselben Namen definieren, gewinnt die Gruppe, die in der Vorlage
später kommt.

## Gruppenname {#varGroup-name}

Der Name, unter dem die Gruppe auf der Seite Variablengruppen und in der Variablengruppen-Auswahl
einer Task-Vorlage aufgeführt wird. Das Feld ist ein Pflichtfeld.

## Secret storage {#varGroup-secretStorageId}

Wird angezeigt, wenn die Gruppe neu ist und der Server Secret-Speicher unterstützt. Optional.

Der Secret-Speicher, in dem die Secrets dieser Gruppe aufbewahrt werden, zum Beispiel ein
HashiCorp Vault. Lassen Sie das Feld leer, um die Secrets verschlüsselt in der Datenbank von
Semaphore aufzubewahren.

Der Speicher wird einmal gewählt, beim Erstellen der Gruppe: Eine bestehende Gruppe behält ihren
Speicher, und das Feld wird nicht mehr angezeigt. Ein als **Read only** markierter Speicher nimmt
keine neuen Secrets an: Das Speichern eines Secrets darin schlägt fehl.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Wird angezeigt, wenn die Gruppe neu ist und der Server Secret-Speicher unterstützt. Verfügbar,
sobald ein **Secret storage** gewählt ist.

Der Anfang des Pfads, den jedes Secret dieser Gruppe im Speicher erhält. Semaphore hängt 10
zufällige Zeichen daran an: Mit dem Präfix `semaphore/prod/` wird ein Secret als
`semaphore/prod/` gefolgt von diesen Zeichen gespeichert. Das Präfix wird unverändert übernommen,
beenden Sie es also mit `/`, wenn die Secrets darunter wie in einem Ordner liegen sollen.

Das Präfix wird einmal festgelegt, beim Erstellen der Gruppe. Der Tab **Secrets** zeigt es als
Muster des Quellpfads.

## Table, JSON oder YAML {#varGroup-extraVarsMode}

Wie Sie die zusätzlichen Variablen bearbeiten:

- **Table**: eine Zeile pro Variable, jede mit einem Typ. Fügen Sie mit der Schaltfläche **+**
  neben dem Umschalter eine Zeile hinzu.
- **JSON**: alle Variablen als ein JSON-Objekt.
- **YAML**: alle Variablen als ein YAML-Mapping.

Der Dialog öffnet sich in **Table**. Ein Umschalten konvertiert, was Sie eingegeben haben; kann es
nicht gelesen werden, bleibt der Dialog im bisherigen Modus und zeigt den Fehler. In jedem Modus
werden die Variablen als ein Objekt aus Namen und Werten gespeichert.

## Zusätzliche Variablen in JSON {#varGroup-extraVarsJson}

Wird angezeigt, wenn die zusätzlichen Variablen als **JSON** bearbeitet werden.

Die zusätzlichen Variablen als ein JSON-Objekt. Ein Name darf nicht leer sein; ein Wert kann eine
Zeichenkette, eine Zahl, eine Liste oder ein Objekt sein:

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Zusätzliche Variablen in YAML {#varGroup-extraVarsYaml}

Wird angezeigt, wenn die zusätzlichen Variablen als **YAML** bearbeitet werden.

Die zusätzlichen Variablen als ein YAML-Mapping. Der Text wird mit den Typen gelesen, die JSON
kennt: Ein Datum ohne Anführungszeichen wie `2024-01-01` bleibt eine Zeichenkette, und `.inf` oder
`.nan` werden abgelehnt. Eine Liste oder ein Skalar auf oberster Ebene wird ebenfalls abgelehnt.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Variablenname {#varGroup-varName}

Wird angezeigt, wenn die zusätzlichen Variablen als **Table** bearbeitet werden, in jeder Zeile.

Der Name der Variablen, wie das Playbook oder der Terraform-Code sie sieht, zum Beispiel
`app_version`.

## Variablentyp {#varGroup-varType}

Wird angezeigt, wenn die zusätzlichen Variablen als **Table** bearbeitet werden, in jeder Zeile.

Wie der Wert der Zeile gelesen wird:

- **String**: der Text, wie er ist.
- **Number**: eine Zahl, zum Beispiel `42`.
- **List**: ein JSON-Array, zum Beispiel `["a", "b"]`.
- **Dict**: ein JSON-Objekt, zum Beispiel `{"key": "value"}`.

Ein Wert, der nicht zu seinem Typ passt, stoppt das Speichern mit dem Namen der Variablen.

## Variablenwert {#varGroup-varValue}

Wird angezeigt, wenn die zusätzlichen Variablen als **Table** bearbeitet werden, in jeder Zeile.

Der Wert der Variablen im Format ihres Typs. Bei einer **List** oder einem **Dict** öffnet die
Schaltfläche am Ende des Feldes einen größeren Editor.

Semaphore übergibt die Werte als Daten an Ansible: Ein Ausdruck `{{ … }}` innerhalb eines Wertes
wird nicht ausgewertet, das Playbook erhält den Text, wie er ist.

## Name der Umgebungsvariablen {#varGroup-envName}

Auf dem Tab **Variables**, unter **Umgebungsvariablen**, in jeder Zeile. Fügen Sie mit der
Schaltfläche **+** des Abschnitts eine Zeile hinzu.

Der Name einer Variablen, die in der Umgebung des Prozesses gesetzt wird, der den Task ausführt,
zum Beispiel `AWS_REGION`.

## Wert der Umgebungsvariablen {#varGroup-envValue}

Auf dem Tab **Variables**, unter **Umgebungsvariablen**, in jeder Zeile.

Der Wert der Variablen, als Text. Er wird als Klartext gespeichert und jedem angezeigt, der die
Gruppe öffnet: Tragen Sie Passwörter und Tokens stattdessen auf dem Tab **Secrets** ein.

## Sync keys enabled {#varGroup-syncEnabled}

Wird auf dem Tab **Secrets** einer Gruppe angezeigt, die einen Secret-Speicher hat.

Schalten Sie die Option ein, um Secrets aus dem Speicher in diese Gruppe zu importieren, aus den
unter **Sync paths** festgelegten Pfaden. Der Import läuft, wenn Sie auf der Seite
Variablengruppen auf die Synchronisierungsschaltfläche der Gruppe klicken, und von selbst in dem
in **Sync paths** festgelegten Intervall. Ist die Option ausgeschaltet, wird nichts importiert.

## Sync paths {#varGroup-syncPaths}

Wird angezeigt, wenn **Sync keys enabled** eingeschaltet ist.

Öffnet den Dialog mit den Pfaden des Speichers, aus denen importiert wird, und dem Intervall des
automatischen Imports. Die Zahl auf der Schaltfläche ist die Anzahl der Pfade. Die
Synchronisierungsschaltfläche der Gruppe auf der Seite Variablengruppen bleibt deaktiviert, bis
die Gruppe mindestens einen Pfad hat.

Der Dialog hat nur **Schließen**: Was Sie darin ändern, wird mit der Gruppe gespeichert, durch
**Speichern** dieses Dialogs.

## Name der geheimen Variablen {#varGroup-secretVarName}

Auf dem Tab **Secrets**, unter **Zusätzliche Variablen**, in jeder Zeile. Fügen Sie mit der
Schaltfläche **+** des Abschnitts eine Zeile hinzu.

Der Name einer zusätzlichen Variablen, deren Wert ein Secret ist. Der Task erhält sie so, wie er
die zusätzlichen Variablen des Tabs **Variables** erhält; ein Secret hat Vorrang vor einer
zusätzlichen Variablen mit demselben Namen. Ansible erhält die Secrets in einer mit Ansible Vault
verschlüsselten Datei, Terraform und OpenTofu als `-var`. Wie der Dialog warnt, kann ein so
übergebenes Secret dennoch im Klartext im Log von Ansible erscheinen.

## Wert der geheimen Variablen {#varGroup-secretVarValue}

Auf dem Tab **Secrets**, unter **Zusätzliche Variablen**, in jeder Zeile.

Der Wert des Secrets. Semaphore verschlüsselt ihn oder schreibt ihn in den Secret-Speicher der
Gruppe und zeigt ihn nie wieder an: Wenn Sie die Gruppe später öffnen, ist das Feld leer. Lassen
Sie es leer, um den gespeicherten Wert zu behalten, oder geben Sie einen neuen ein, um ihn zu
ersetzen. Das Löschen der Zeile löscht das Secret, wenn Sie die Gruppe speichern.

## Name der geheimen Umgebungsvariablen {#varGroup-secretEnvName}

Auf dem Tab **Secrets**, unter **Umgebungsvariablen**, in jeder Zeile. Fügen Sie mit der
Schaltfläche **+** des Abschnitts eine Zeile hinzu.

Der Name einer Umgebungsvariablen, deren Wert ein Secret ist, zum Beispiel `DB_PASSWORD`. Sie
wird in der Umgebung des Prozesses gesetzt, der den Task ausführt, so wie die Umgebungsvariablen
des Tabs **Variables**.

## Wert der geheimen Umgebungsvariablen {#varGroup-secretEnvValue}

Auf dem Tab **Secrets**, unter **Umgebungsvariablen**, in jeder Zeile.

Der Wert des Secrets. Wie bei einer geheimen zusätzlichen Variablen wird er nie wieder angezeigt:
Lassen Sie das Feld leer, um den gespeicherten Wert zu behalten, oder geben Sie einen neuen ein,
um ihn zu ersetzen.
