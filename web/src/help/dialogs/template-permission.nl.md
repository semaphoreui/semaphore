---
form: src/components/EditTemplatePermissionForm.vue
---

# Template permissions

Geeft een rol extra rechten op dit ene sjabloon. Een lid met de rol krijgt ze bovenop de rechten
die de rol in het hele project heeft: een toekenning hier voegt alleen toe, ze neemt nooit een
recht weg. Een lid met meerdere rollen krijgt elk recht dat aan een van die rollen is toegekend.

## Rol {#templatePerm-roleSlug}

De rol die de rechten krijgt. De lijst toont de rollen van dit project, de gedeelde rollen en de
ingebouwde rollen. Het veld is verplicht, en een rol kan maar één toekenning op een sjabloon
hebben: de server weigert een tweede.

De rol van een opgeslagen toekenning kan niet worden gewijzigd. De server behoudt de rol en slaat
alleen de rechten op, wat u hier ook kiest. Om de rechten aan een andere rol te geven, verwijdert
u deze toekenning uit de lijst en voegt u een nieuwe toe.

## Permissions {#templatePerm-permissionsTitle}

Wat een lid met de rol met dit sjabloon mag doen:

- **Can see this template**: het sjabloon staat in de lijst met sjablonen, met zijn taken en hun
  logs.
- **Can run this template**: taken starten vanuit het sjabloon, en zijn taken stoppen, bevestigen
  of afwijzen.
- **Can edit this template**: het sjabloon wijzigen en verwijderen, al zijn taken stoppen, en deze
  rechten beheren.

Uitvoeren of bewerken omvat het zien van het sjabloon: wanneer u een van beide aanvinkt, wordt
**Can see this template** ook aangevinkt en kan het niet worden uitgevinkt. Laat alle vakjes leeg
om een toekenning op te slaan die niets geeft.
