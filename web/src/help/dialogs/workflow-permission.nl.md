---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Geeft een rol extra rechten op deze ene workflow. Een lid met de rol krijgt ze bovenop de rechten
die de rol in het hele project heeft: een toekenning hier voegt alleen toe, ze neemt nooit een
recht weg. Een lid met meerdere rollen krijgt elk recht dat aan een van die rollen is toegekend.

## Rol {#templatePerm-roleSlug}

De rol die de rechten krijgt. De lijst toont de rollen van dit project, de gedeelde rollen en de
ingebouwde rollen. Het veld is verplicht, en een rol kan maar één toekenning op een workflow
hebben: de server weigert een tweede.

De rol van een opgeslagen toekenning kan niet worden gewijzigd. De server behoudt de rol en slaat
alleen de rechten op, wat u hier ook kiest. Om de rechten aan een andere rol te geven, verwijdert
u deze toekenning uit de lijst en voegt u een nieuwe toe.

## Permissions {#templatePerm-permissionsTitle}

Wat een lid met de rol met deze workflow mag doen:

- **Can see this workflow**: de workflow staat in de lijst met workflows, met zijn runs.
- **Can run this workflow**: runs van de workflow starten.
- **Can edit this workflow**: de workflow wijzigen en verwijderen, en deze rechten beheren.
- **Can approve this workflow**: de goedkeuringsstappen van zijn runs goedkeuren of afwijzen.

Uitvoeren, bewerken of goedkeuren omvat het zien van de workflow: wanneer u een van deze aanvinkt,
wordt **Can see this workflow** ook aangevinkt en kan het niet worden uitgevinkt. Laat alle vakjes
leeg om een toekenning op te slaan die niets geeft.
