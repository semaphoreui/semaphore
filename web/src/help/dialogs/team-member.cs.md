---
form: src/components/TeamMemberForm.vue
---

# Nový člen týmu

Přidá uživatele do projektu s určitou rolí. Role rozhoduje o tom, co člen smí v projektu vidět
a dělat; nic jiného se na jeho účtu nemění. Když správce zapnul pozvánky, uživatel místo toho
dostane pozvánku a stane se členem, jakmile ji přijme.

## E-mail nebo uživatelské jméno {#teamMember-inviteType}

Zobrazí se, když lze uživatele zvát e-mailem i uživatelským jménem.

Jak osobu vyberete: **Email** pro pozvání e-mailové adresy, **Username** pro výběr uživatele,
který už má na tomto serveru Semaphore účet.

## E-mail {#teamMember-email}

Zobrazí se, když se osoba zve e-mailem.

E-mailová adresa osoby, kterou chcete pozvat. Členem projektu se stane, jakmile pozvánku
přijme.

## Uživatel {#teamMember-userId}

Zobrazí se, když se osoba vybírá podle uživatelského jména.

Uživatel, kterého chcete přidat, uvedený uživatelským jménem a celým jménem. Pole je povinné.
Seznam obsahuje každého uživatele serveru Semaphore, který ještě není členem projektu: osoba bez
účtu ho musí nejprve dostat od administrátora.

## Role {#teamMember-role}

Co smí člen v projektu dělat. Pole je povinné. Vestavěné role:

- **Owner**: všechno, včetně nastavení projektu, jeho týmu a jeho odstranění.
- **Manager**: spouští úlohy a workflow, schvaluje workflow a vytváří, mění a odstraňuje šablony
  úloh, inventáře, klíče, repozitáře, skupiny proměnných, harmonogramy, integrace a workflow.
  Manager nemůže měnit nastavení projektu ani jeho tým.
- **Task Runner**: vidí projekt, spouští úlohy a workflow a schvaluje workflow. Task Runner
  nemůže v projektu nic měnit.
- **Guest**: vidí projekt, jeho úlohy a jejich logy a nic nemění.

Seznam obsahuje také role vytvořené pro tento projekt (**Projektové role**) a role sdílené všemi
projekty (**Shared roles**). Taková role uděluje oprávnění, která jsou pro ni nastavena na
stránce Role. Pozvánka může nést pouze vestavěnou roli.

Člen, který smí spravovat tým, může přidělit kteroukoli roli, včetně **Owner**. Dialog přidělí
jednu roli; další přidejte později ve sloupci **Role** v seznamu týmu.
