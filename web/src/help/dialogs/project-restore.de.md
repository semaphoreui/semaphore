---
form: src/components/RestoreProjectForm.vue
---

# Projekt wiederherstellen...

Erstellt ein neues Projekt aus einer Backup-Datei, die mit **Projekt-Backup** auf der Seite
Einstellungen eines Projekts erstellt wurde. Das neue Projekt erhält die Task-Vorlagen,
Inventorys, Repositories, Variablengruppen, Schlüssel, Zeitpläne, Integrationen und alles andere,
was im Backup eingerichtet ist, und Sie werden sein Eigentümer. Nur ein Administrator kann ein
Projekt wiederherstellen, es sei denn, der Administrator des Servers erlaubt jedem Benutzer,
Projekte zu erstellen.

## Projektname {#restoreProject-name}

Der Name des neuen Projekts. Das Feld ist ein Pflichtfeld, und es ersetzt den Namen, der in der
Backup-Datei steht.

Der Name muss sich von den Namen aller Projekte auf dem Server unterscheiden: Der Server lehnt es
ab, ein Projekt unter einem bereits vergebenen Namen wiederherzustellen.

## Backup file {#restoreProject-file}

Die JSON-Datei des Backups. Wählen Sie die Datei aus, bevor Sie auf **Restore** klicken.

Ein Backup lässt die Secrets der Schlüssel und die Secrets auf dem Tab **Secrets** der
Variablengruppen weg: Nach der Wiederherstellung teilt Semaphore Ihnen mit, wie viele Schlüssel leer
sind und ihre Secrets erneut eingegeben werden müssen, und die Secrets der Variablengruppen müssen
ebenfalls erneut eingegeben werden. Ein Passwort, das direkt in ein anderes Feld eingegeben wurde,
etwa in die Adresse eines Repositorys, kommt so zurück, wie es in der Datei steht. Die Mitglieder
des Projekts und der Verlauf seiner Tasks sind nicht Teil des Backups. Ein Backup, das von einer
neueren Version von Semaphore geschrieben wurde, wird abgelehnt.
