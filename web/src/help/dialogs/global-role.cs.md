---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

Globální role uděluje uživatelům, kterým je přidělena, oprávnění nad celou instancí Semaphore;
přiděluje se v dialogu uživatele. Nedává žádný přístup uvnitř projektů: ten pochází pouze z rolí,
které uživatel má jako člen projektu. Administrátor má všechna globální oprávnění bez ohledu na
přidělené role. Globální role jsou funkcí verze Pro.

## Název {#globalRole-name}

Název, pod kterým je role uvedena na kartě Globální role stránky Uživatelé a v dialogu uživatele.
Pole je povinné. Roli můžete kdykoli přejmenovat.

## Slug {#globalRole-slug}

Identifikátor role. Semaphore ho během psaní názvu vyplňuje podle názvu, dokud slug neupravíte
sami. Přijímá 1 až 100 písmen, číslic, podtržítek a spojovníků, například `user-admins`.

Slug musí být jedinečný mezi globálními rolemi; projektová role může mít stejný slug. Po uložení
role ho nelze změnit: při úpravě role je pole uzamčené.

## Global permissions {#globalRole-permissionsTitle}

Co smí uživatel, který má tuto roli, dělat v celé instanci. Nová role nemá zaškrtnuté žádné.
Oprávnění všech globálních rolí uživatele se sčítají.

- **Can manage users**: vytvářet uživatele, měnit je a odstraňovat a vidět jejich údaje
  v seznamu uživatelů. Takový uživatel nemůže z uživatele udělat administrátora ani uživatele
  Pro a nemůže změnit ani odstranit administrátora nebo uživatele, který má globální oprávnění,
  jež sám nemá.
- **Can manage global roles**: vytvářet, měnit a odstraňovat globální role a sdílené projektové
  role a rozhodovat, kteří uživatelé mají kterou globální roli. Dejte ho jen lidem, kterým
  svěříte celou instanci.
- **Can manage system settings**: měnit nastavení instance, aplikace, globální runnery
  a mapování skupin; vidět cluster a informace o serveru; vidět a odstraňovat úlohy všech
  projektů; vymazat mezipaměť serveru.
- **Can read the audit log**: číst auditní log instance a exportovat ho.
