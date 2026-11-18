---
form: src/components/EditViewsForm.vue
---

# Weergaven Bewerken

Weergaven zijn de tabbladen boven de lijst met taaksjablonen. Een weergave toont de sjablonen die
u erin plaatst, of elk sjabloon van het project. Het dialoogvenster slaat elke wijziging direct
op: er is geen knop Opslaan voor het hele dialoogvenster. Sleep een weergave aan zijn greep om de
volgorde van de tabbladen te wijzigen.

## Titel {#viewsForm-title}

De naam van het tabblad. Typ hem en bevestig hem met het vinkje naast het veld; het kruisje zet de
opgeslagen titel terug. Een nieuwe weergave wordt opgeslagen zodra ze een titel heeft en u die
bevestigt of een van haar instellingen wijzigt. Het sluiten van het dialoogvenster slaat een titel
die u niet hebt bevestigd niet op.

Het tandwiel opent de instellingen van de weergave onder haar titel.

## Verwijderen {#viewsForm-viewDelete}

Het prullenbakpictogram verwijdert de weergave direct, zonder te vragen. De sjablonen van de
weergave worden niet verwijderd: ze blijven in het project en staan in geen enkele weergave meer.

## Type {#viewsForm-type}

Welke sjablonen het tabblad toont:

- **Custom**: de sjablonen waarvan het veld **Weergave**, in het sjabloondialoogvenster, deze
  weergave noemt. Dit is de standaard.
- **All**: elk sjabloon van het project, wat de weergave ervan ook is.

Een lid ziet alleen de sjablonen die het lid mag zien, in elke weergave.

## Sort by {#viewsForm-sortColumn}

De kolom waarop de sjablonen van het tabblad worden gesorteerd; **Name** is de enige. Een tabblad
zonder deze instelling wordt ook op naam gesorteerd, dus het veld doet er pas toe samen met
**Reverse**.

## Reverse {#viewsForm-sortReverse}

Sorteert het tabblad in aflopende volgorde. Dit werkt alleen wanneer **Sort by** is ingesteld.

## Hidden {#viewsForm-hidden}

Verbergt het tabblad voor leden die de resources van het project niet mogen beheren. Wie dat wel
mag, ziet het nog, doorgestreept. Het verbergen van een tabblad verbergt de sjablonen ervan niet:
een weergave van het type **All** vermeldt ze nog steeds.

## Show task counts {#viewsForm-showTemplateStats}

Toont op het tabblad hoeveel van de sjablonen ervan in welke toestand zijn, als één chip van
gekleurde segmenten: groen voor een laatste taak die is geslaagd, rood voor een mislukte, grijs voor
een gestopte of een sjabloon dat nooit is gestart, oranje voor een taak die op bevestiging wacht,
blauw voor een taak in de wachtrij of in uitvoering. Het staat standaard uit, en de aantallen worden
alleen berekend voor de weergaven waarbij het aan staat.

## Weergave Toevoegen {#viewsForm-add}

Voegt een weergave toe aan het eind van de lijst, met een lege titel om in te typen.
