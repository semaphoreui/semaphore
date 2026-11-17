---
form: src/components/EditRoleForm.vue
---

# Edit Role

Een rol is een benoemde set rechten binnen een project. U geeft rollen aan de leden van een
project, en een lid mag doen wat een van zijn rollen toestaat. Geopend vanaf de pagina Roles van
een project, bewerkt het dialoogvenster alleen een rol van dat project; geopend vanaf de pagina
Roles buiten een project, bewerkt het een gedeelde rol die elk project kan gebruiken. De
ingebouwde rollen Owner, Manager, Task Runner en Guest kunnen niet worden gewijzigd.

## Naam {#role-name}

De naam waaronder de rol wordt vermeld op de pagina Roles en overal waar een rol wordt gekozen,
bijvoorbeeld wanneer u een lid aan een project toevoegt. Het veld is verplicht. U kunt een rol op
elk moment hernoemen.

## Slug {#role-slug}

De identificatie van de rol. Semaphore vult deze in op basis van de naam terwijl u de naam typt,
totdat u de slug zelf bewerkt. De slug bestaat uit 1 tot 100 letters, cijfers, underscores en
koppeltekens, bijvoorbeeld `release-managers`.

De slug kan niet meer worden gewijzigd nadat de rol is opgeslagen: het veld is vergrendeld wanneer
u een rol bewerkt.

- Een slug moet uniek zijn, en de controle negeert hoofdletters en kleine letters. De slug van een
  gedeelde rol moet verschillen van elke andere rol; de slug van een projectrol moet verschillen
  van de gedeelde rollen en van de andere rollen van zijn project.
- `owner`, `manager`, `task_runner`, `guest` en `all` zijn gereserveerd, in elke schrijfwijze.

## Permissions {#role-permissionsTitle}

Wat de leden met de rol in het project mogen doen. Bij een nieuwe rol is niets aangevinkt. De
rechten van alle rollen van een lid tellen bij elkaar op; een beheerder van Semaphore mag alles in
elk project, welke rollen er ook zijn.

- **Can read project**: het project zien: zijn taaksjablonen, taken en hun logs, inventarissen,
  sleutels, repositories, variabele groepen, schema's en leden. Elke leesactie in het project
  vraagt om dit recht, en geen ander recht omvat het, dus vink het aan voor bijna elke rol. Zonder
  dit recht ziet een lid alleen de taaksjablonen en workflows die de rol zelf toegang geven, met
  hun taken.
- **Can run project tasks**: taken starten vanuit taaksjablonen, ze stoppen, en een taak die op
  een bevestiging wacht bevestigen of afwijzen.
- **Can update project**: de instellingen van het project wijzigen en het project verwijderen.
- **Can manage project resources**: taaksjablonen, inventarissen, sleutels, repositories,
  variabele groepen, schema's, weergaven, integraties, geheimenopslagen, projectrunners en de
  rollen van het project aanmaken, wijzigen en verwijderen; taken verwijderen; instellen welke
  rollen een taaksjabloon mogen gebruiken.
- **Can manage project users**: leden aan het project toevoegen, hun rollen wijzigen en ze
  verwijderen.
- **Can run workflows**: een workflow starten en zijn runs stoppen.
- **Can edit workflows**: workflows aanmaken, wijzigen en verwijderen, en instellen welke rollen
  een workflow mogen gebruiken.
- **Can approve workflows**: een stap van een workflowrun die op goedkeuring wacht goedkeuren of
  afwijzen.
