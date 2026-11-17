---
form: src/components/EditRoleForm.vue
---

# Upravit roli

Role je pojmenovaná sada oprávnění uvnitř projektu. Role přidělujete členům projektu a člen smí
dělat to, co dovoluje kterákoli z jeho rolí. Když dialog otevřete ze stránky Role projektu,
upravuje pouze roli tohoto projektu; když ho otevřete ze stránky Role mimo projekt, upravuje
sdílenou roli, kterou může používat každý projekt. Vestavěné role Owner, Manager, Task Runner
a Guest nelze změnit.

## Název {#role-name}

Název, pod kterým je role uvedena na stránce Role a všude, kde se role vybírá, například když
přidáváte člena do projektu. Pole je povinné. Roli můžete kdykoli přejmenovat.

## Slug {#role-slug}

Identifikátor role. Semaphore ho během psaní názvu vyplňuje podle názvu, dokud slug neupravíte
sami. Přijímá 1 až 100 písmen, číslic, podtržítek a spojovníků, například `release-managers`.

Po uložení role slug nelze změnit: při úpravě role je pole uzamčené.

- Slug musí být jedinečný a kontrola nerozlišuje velikost písmen. Slug sdílené role se musí lišit
  od všech ostatních rolí; slug projektové role se musí lišit od sdílených rolí a od ostatních
  rolí svého projektu.
- `owner`, `manager`, `task_runner`, `guest` a `all` jsou rezervované, v jakékoli velikosti
  písmen.

## Oprávnění {#role-permissionsTitle}

Co smí členové, kteří mají tuto roli, v projektu dělat. Nová role nemá zaškrtnuté žádné.
Oprávnění všech rolí, které člen má, se sčítají; administrátor Semaphore smí v každém projektu
všechno bez ohledu na role.

- **Can read project**: vidět projekt: jeho šablony úloh, úlohy a jejich logy, inventáře, klíče,
  repozitáře, skupiny proměnných, harmonogramy a členy. Každé čtení v projektu ho vyžaduje
  a žádné jiné oprávnění ho nezahrnuje, takže ho zaškrtněte téměř u každé role. Bez něj člen vidí
  pouze šablony úloh a workflow, které roli udělují vlastní přístup, s jejich úlohami.
- **Může spouštět úlohy projektu**: spouštět úlohy ze šablon úloh, zastavovat je a potvrzovat
  nebo zamítat úlohu, která čeká na potvrzení.
- **Může aktualizovat projekt**: měnit nastavení projektu a odstranit ho.
- **Může spravovat prostředky projektu**: vytvářet, měnit a odstraňovat šablony úloh, inventáře,
  klíče, repozitáře, skupiny proměnných, harmonogramy, zobrazení, integrace, úložiště tajných
  údajů, runnery projektu a role projektu; odstraňovat úlohy; nastavovat, které role smí šablonu
  úlohy používat.
- **Může spravovat uživatele projektu**: přidávat členy do projektu, měnit jejich role a odebírat
  je.
- **Can run workflows**: spustit workflow a zastavovat jeho běhy.
- **Can edit workflows**: vytvářet, měnit a odstraňovat workflow a nastavovat, které role smí
  workflow používat.
- **Can approve workflows**: schválit nebo zamítnout krok běhu workflow, který čeká na
  schválení.
