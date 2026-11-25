---
form: src/components/RoleMappingForm.vue
---

# New mapping

Mapowanie nadaje rolę w projekcie każdemu użytkownikowi, który loguje się przez dostawcę LDAP lub
OpenID Connect i należy tam do grupy. Semaphore przy każdym logowaniu odczytuje grupy użytkownika
przez dostawcę i nadaje lub odbiera mapowane role tak, aby się zgadzały, więc użytkownik, który
opuści grupę, traci rolę przy następnym logowaniu. Mapowana rola czyni użytkownika członkiem
projektu. Mapowania nie można edytować: usuń je i dodaj nowe.

## Provider {#roleMapping-provider}

Katalog LDAP lub dostawca OpenID Connect, z którego pochodzi grupa. Lista zawiera dostawców
skonfigurowanych w konfiguracji serwera. Pole jest wymagane.

Grupy są odczytywane tylko wtedy, gdy dostawca jest do tego skonfigurowany: dostawca LDAP wymaga
bazy wyszukiwania grup (`group_search_dn`), a dostawca OpenID Connect wskazuje roszczenie z grupami
(`groups_claim`). Mapowanie dostawcy, który został później usunięty z konfiguracji, pozostaje na
liście z oznaczeniem **Provider not configured**.

## Group {#roleMapping-group}

Nazwa grupy w postaci podawanej przez dostawcę. Pole jest wymagane.

- LDAP: wartość atrybutu nazwy grupy we wpisie grupy, domyślnie `cn`, na przykład `devops`. Pełny
  DN, taki jak `cn=devops,ou=groups,dc=example,dc=com`, jest skracany do swojego `cn`.
- OpenID Connect: jedna z wartości roszczenia grup, w postaci wysyłanej przez dostawcę.

## Project {#roleMapping-projectId}

Projekt, w którym nadawana jest rola. Pole jest wymagane. Administrator widzi każdy projekt; każdy
inny użytkownik widzi tylko projekty, których jest członkiem.

## Rola {#roleMapping-role}

Rola, którą członkowie grupy otrzymują w projekcie. Pole jest wymagane i jest dostępne po wybraniu
projektu. Lista zawiera role tego projektu, role współdzielone i role wbudowane. Gdy wybierzesz
inny projekt, rola, która w nim nie istnieje, jest czyszczona.

Rola **Owner** nadana przez mapowanie nie jest brana pod uwagę, gdy Semaphore nie pozwala odebrać
roli ostatniemu właścicielowi projektu, ponieważ może zniknąć przy następnym logowaniu. Zachowaj co
najmniej jednego właściciela dodanego ręcznie.
