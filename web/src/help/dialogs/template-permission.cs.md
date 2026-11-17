---
form: src/components/EditTemplatePermissionForm.vue
---

# Oprávnění k šablonám

Udělí roli další oprávnění k této jediné šabloně. Člen, který má tuto roli, je dostane navíc
k oprávněním, která má role v celém projektu: udělení zde oprávnění pouze přidává, nikdy žádné
neodebírá. Člen s několika rolemi dostane každé oprávnění, které je uděleno kterékoli z nich.

## Role {#templatePerm-roleSlug}

Role, která oprávnění dostane. Seznam ukazuje role tohoto projektu, sdílené role a vestavěné
role. Pole je povinné a role může mít k šabloně pouze jedno udělení: druhé server odmítne.

Roli uloženého udělení nelze změnit. Server roli ponechá a uloží pouze oprávnění, ať zde vyberete
cokoli. Chcete-li oprávnění dát jiné roli, odeberte toto udělení ze seznamu a přidejte nové.

## Oprávnění {#templatePerm-permissionsTitle}

Co smí člen, který má tuto roli, s touto šablonou dělat:

- **Can see this template**: šablona je v seznamu šablon i se svými úlohami a jejich
  logy.
- **Can run this template**: spouštět úlohy ze šablony a zastavovat, potvrzovat nebo zamítat její
  úlohy.
- **Can edit this template**: měnit a odstranit šablonu, zastavit všechny její úlohy a spravovat
  tato oprávnění.

Spouštění i úpravy zahrnují zobrazení šablony: když zaškrtnete kterékoli z nich, zaškrtne se také
**Can see this template** a nelze ho zrušit. Když necháte všechna políčka nezaškrtnutá, uloží se
udělení, které nedává nic.
