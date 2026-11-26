---
form: src/components/RoleMappingForm.vue
---

# New mapping

Eine Zuordnung gibt jedem Benutzer, der sich über einen LDAP- oder OpenID-Connect-Provider
anmeldet und dort einer Gruppe angehört, eine Rolle in einem Projekt. Semaphore liest die Gruppen
des Benutzers bei jeder Anmeldung über den Provider und vergibt oder entzieht die zugeordneten
Rollen entsprechend, sodass ein Benutzer, der die Gruppe verlässt, die Rolle bei seiner nächsten
Anmeldung verliert. Eine zugeordnete Rolle macht den Benutzer zum Mitglied des Projekts. Eine
Zuordnung kann nicht bearbeitet werden: Löschen Sie sie und fügen Sie eine neue hinzu.

## Provider {#roleMapping-provider}

Das LDAP-Verzeichnis oder der OpenID-Connect-Provider, aus dem die Gruppe stammt. Die Liste
enthält die Provider, die in der Konfiguration des Servers eingerichtet sind. Das Feld ist ein
Pflichtfeld.

Die Gruppen werden nur gelesen, wenn der Provider dafür eingerichtet ist: Ein LDAP-Provider
braucht eine Suchbasis für Gruppen (`group_search_dn`), ein OpenID-Connect-Provider nennt den
Claim mit den Gruppen (`groups_claim`). Eine Zuordnung eines Providers, der später aus der
Konfiguration entfernt wird, bleibt in der Liste und wird mit **Provider not configured**
markiert.

## Group {#roleMapping-group}

Der Name der Gruppe, wie der Provider ihn liefert. Das Feld ist ein Pflichtfeld.

- LDAP: der Wert des Gruppennamen-Attributs des Gruppeneintrags, standardmäßig `cn`, zum Beispiel
  `devops`. Ein vollständiger DN wie `cn=devops,ou=groups,dc=example,dc=com` wird auf seinen
  `cn` reduziert.
- OpenID Connect: einer der Werte des Gruppen-Claims, wie der Provider ihn sendet.

## Project {#roleMapping-projectId}

Das Projekt, in dem die Rolle vergeben wird. Das Feld ist ein Pflichtfeld. Ein Administrator
sieht jedes Projekt; jeder andere Benutzer sieht nur die Projekte, deren Mitglied er ist.

## Rolle {#roleMapping-role}

Die Rolle, die die Mitglieder der Gruppe im Projekt erhalten. Das Feld ist ein Pflichtfeld und
steht zur Verfügung, sobald ein Projekt ausgewählt ist. Die Liste enthält die Rollen dieses
Projekts, die gemeinsamen Rollen und die eingebauten Rollen. Wenn Sie ein anderes Projekt wählen,
wird eine Rolle, die dort nicht existiert, geleert.

Eine durch eine Zuordnung vergebene Rolle **Owner** zählt nicht, wenn Semaphore Sie daran
hindert, dem letzten Eigentümer eines Projekts die Rolle zu entziehen, da sie bei der nächsten
Anmeldung wegfallen kann. Behalten Sie mindestens einen von Hand hinzugefügten Eigentümer.
