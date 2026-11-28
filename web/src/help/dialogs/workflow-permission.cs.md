---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Udělí roli další oprávnění k tomuto jedinému workflow. Člen, který má tuto roli, je dostane navíc
k oprávněním, která má role v celém projektu: udělení zde oprávnění pouze přidává, nikdy žádné
neodebírá. Člen s několika rolemi dostane každé oprávnění, které je uděleno kterékoli z nich.

## Role {#templatePerm-roleSlug}

Role, která oprávnění dostane. Seznam ukazuje role tohoto projektu, sdílené role a vestavěné
role. Pole je povinné a role může mít k workflow pouze jedno udělení: druhé server odmítne.

Roli uloženého udělení nelze změnit. Server roli ponechá a uloží pouze oprávnění, ať zde vyberete
cokoli. Chcete-li oprávnění dát jiné roli, odeberte toto udělení ze seznamu a přidejte nové.

## Oprávnění {#templatePerm-permissionsTitle}

Co smí člen, který má tuto roli, s tímto workflow dělat:

- **Can see this workflow**: workflow je v seznamu workflow i se svými běhy.
- **Can run this workflow**: spouštět běhy workflow.
- **Can edit this workflow**: měnit a odstranit workflow a spravovat tato oprávnění.
- **Can approve this workflow**: schvalovat nebo zamítat kroky schválení jeho běhů.

Spouštění, úpravy i schvalování zahrnují zobrazení workflow: když zaškrtnete kterékoli z nich,
zaškrtne se také **Can see this workflow** a nelze ho zrušit. Když necháte všechna políčka
nezaškrtnutá, uloží se udělení, které nedává nic.
