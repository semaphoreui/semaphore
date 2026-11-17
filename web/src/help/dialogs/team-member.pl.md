---
form: src/components/TeamMemberForm.vue
---

# Nowy członek zespołu

Dodaje użytkownika do projektu z rolą. Rola decyduje, co członek może widzieć i robić w projekcie;
nic innego na jego koncie się nie zmienia. Gdy administrator włączył zaproszenia, użytkownik
otrzymuje zamiast tego zaproszenie i staje się członkiem po jego przyjęciu.

## Email or Username {#teamMember-inviteType}

Wyświetlane, gdy użytkowników można zapraszać zarówno przez e-mail, jak i przez nazwę użytkownika.

Sposób wskazania osoby: **Email**, aby zaprosić adres e-mail, **Username**, aby wybrać
użytkownika, który ma już konto na tym serwerze Semaphore.

## Email {#teamMember-email}

Wyświetlane, gdy osoba jest zapraszana przez e-mail.

Adres e-mail osoby do zaproszenia. Staje się ona członkiem projektu po przyjęciu zaproszenia.

## Użytkownik {#teamMember-userId}

Wyświetlane, gdy osoba jest wybierana według nazwy użytkownika.

Użytkownik do dodania, wyświetlany z nazwą użytkownika i pełnym imieniem i nazwiskiem. Pole jest
wymagane. Lista zawiera każdego użytkownika serwera Semaphore, który nie jest jeszcze członkiem
projektu: osoba bez konta musi je najpierw otrzymać od administratora.

## Rola {#teamMember-role}

Co członek może robić w projekcie. Pole jest wymagane. Role wbudowane:

- **Owner**: wszystko, łącznie z ustawieniami projektu, jego zespołem i jego usunięciem.
- **Manager**: uruchamia zadania i przepływy pracy, zatwierdza przepływy pracy oraz tworzy,
  zmienia i usuwa szablony zadań, inwentarze, klucze, repozytoria, grupy zmiennych, harmonogramy,
  integracje i przepływy pracy. Menedżer nie może zmieniać ustawień projektu ani jego zespołu.
- **Task Runner**: widzi projekt, uruchamia zadania i przepływy pracy oraz zatwierdza przepływy
  pracy. Nie może niczego zmieniać w projekcie.
- **Guest**: widzi projekt, jego zadania i ich dzienniki i niczego nie zmienia.

Lista zawiera także role utworzone dla tego projektu (**Role projektu**) i role współdzielone przez
wszystkie projekty (**Shared roles**). Taka rola przyznaje uprawnienia ustawione dla niej na
stronie Roles. Zaproszenie może zawierać tylko rolę wbudowaną.

Członek, który może zarządzać zespołem, może nadać dowolną rolę, łącznie z **Owner**. Okno
dialogowe nadaje jedną rolę; kolejne dodaj później w kolumnie **Rola** listy zespołu.
