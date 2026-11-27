---
form: src/components/TeamMemberForm.vue
---

# Nieuw Teamlid

Voegt een gebruiker met een rol aan het project toe. De rol bepaalt wat het lid in het project mag
zien en doen; verder verandert er niets aan zijn account. Wanneer de beheerder uitnodigingen heeft
ingeschakeld, krijgt de gebruiker in plaats daarvan een uitnodiging en wordt hij lid zodra hij die
accepteert.

## E-mail of gebruikersnaam {#teamMember-inviteType}

Wordt getoond wanneer gebruikers zowel per e-mail als per gebruikersnaam kunnen worden uitgenodigd.

Hoe u de persoon kiest: **Email** om een e-mailadres uit te nodigen, **Username** om een gebruiker
te kiezen die al een account op deze Semaphore-server heeft.

## E-mail {#teamMember-email}

Wordt getoond wanneer de persoon per e-mail wordt uitgenodigd.

Het e-mailadres van de persoon die u uitnodigt. Hij wordt lid van het project zodra hij de
uitnodiging accepteert.

## Gebruiker {#teamMember-userId}

Wordt getoond wanneer de persoon op gebruikersnaam wordt gekozen.

De gebruiker die u toevoegt, vermeld met gebruikersnaam en volledige naam. Het veld is verplicht.
De lijst bevat elke gebruiker van de Semaphore-server die nog geen lid van het project is: een
persoon zonder account moet er eerst een van een beheerder krijgen.

## Rol {#teamMember-role}

Wat het lid in het project mag doen. Het veld is verplicht. De ingebouwde rollen:

- **Owner**: alles, inclusief de instellingen van het project, zijn team en het verwijderen
  ervan.
- **Manager**: voert taken en workflows uit, keurt workflows goed, en maakt taaksjablonen,
  inventarissen, sleutels, repositories, variabele groepen, schema's, integraties en workflows
  aan, wijzigt en verwijdert ze. Een manager kan de instellingen van het project en zijn team niet
  wijzigen.
- **Task Runner**: ziet het project, voert taken en workflows uit en keurt workflows goed. Een task
  runner kan niets in het project wijzigen.
- **Guest**: ziet het project, zijn taken en hun logs, en wijzigt niets.

De lijst bevat ook de rollen die voor dit project zijn gemaakt (**Projectrollen**) en de rollen
die door elk project worden gedeeld (**Shared roles**). Zo'n rol verleent de rechten die er op de
pagina Roles voor zijn ingesteld. Een uitnodiging kan alleen een ingebouwde rol meekrijgen.

Een lid dat het team mag beheren, kan elke rol geven, **Owner** inbegrepen. Het dialoogvenster
geeft één rol; voeg er daarna meer toe in de kolom **Rol** van de teamlijst.
