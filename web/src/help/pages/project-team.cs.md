---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Tým

Členové projektu a role, které mají. Role určuje, co člen v projektu smí: spouštět úlohy, měnit
šablony a další prostředky, spravovat tým. Člen může mít více rolí a získá každé oprávnění, které
mu některá z nich dává.

- **Nový člen týmu** přidá do projektu uživatele s rolí. Tlačítko vidí členové, kteří smějí
  spravovat tým.
- Sloupec **Role** mění role člena okamžitě, bez ukládání. Člen, který tým spravovat nesmí, role
  vidí, ale nemůže je měnit.
- Role se zámkem pochází ze skupiny uživatele v LDAP nebo OIDC, kterou správce k této roli
  přiřadil. Zde ji odebrat nelze: zmizí, když uživatel skupinu opustí nebo se přiřazení odstraní.
- Projekt potřebuje alespoň jednoho Owner. Semaphore odmítne odebrat posledního Owner nebo mu roli
  Owner odejmout, pokud to nedělá správce Semaphore.
- **Opustit projekt** vás z projektu odebere. Tlačítko se zobrazí jen tehdy, když správce členům
  dovolí projekty opouštět; ani poslední Owner projekt opustit nemůže.
