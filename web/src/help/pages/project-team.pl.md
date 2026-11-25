---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Zespół

Członkowie projektu i ich role. Rola określa, co członek może robić w projekcie: uruchamiać
zadania, zmieniać szablony i inne zasoby, zarządzać zespołem. Członek może mieć kilka ról i
otrzymuje każde uprawnienie, które daje którakolwiek z nich.

- **Nowy członek zespołu** dodaje użytkownika do projektu z rolą. Przycisk widzą członkowie, którzy
  mogą zarządzać zespołem.
- Kolumna **Rola** zmienia role członka od razu, bez zapisywania. Członek bez prawa zarządzania
  zespołem widzi role, ale nie może ich zmienić.
- Rola z kłódką pochodzi z grupy użytkownika w LDAP lub OIDC, którą administrator przypisał do tej
  roli. Nie można jej tu usunąć: znika, gdy użytkownik opuści grupę lub przypisanie zostanie
  usunięte.
- Projekt musi mieć co najmniej jednego Owner. Semaphore nie pozwala usunąć ostatniego Owner ani
  odebrać mu roli Owner, chyba że robi to administrator Semaphore.
- **Opuszczam projekt** usuwa cię z projektu. Przycisk widać tylko wtedy, gdy administrator
  pozwala członkom opuszczać projekty; ostatni Owner również nie może opuścić projektu.
