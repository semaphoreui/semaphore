---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Team

De leden van het project en de rollen die ze hebben. Een rol bepaalt wat een lid in het project
mag: taken uitvoeren, sjablonen en andere resources wijzigen, het team beheren. Een lid kan meerdere
rollen hebben en krijgt elke bevoegdheid die een van die rollen geeft.

- **Nieuw Teamlid** voegt een gebruiker met een rol aan het project toe. Leden die het team mogen
  beheren zien de knop.
- De kolom **Rol** wijzigt de rollen van een lid direct, zonder opslaan. Een lid dat het team niet
  mag beheren ziet de rollen, maar kan ze niet wijzigen.
- Een rol met een slot komt uit een groep van de gebruiker in LDAP of OIDC die de beheerder aan die
  rol heeft gekoppeld. Hij is hier niet te verwijderen: hij verdwijnt wanneer de gebruiker de groep
  verlaat of de koppeling wordt verwijderd.
- Het project heeft minstens één Owner nodig. Semaphore weigert de laatste Owner te verwijderen of
  de rol Owner af te nemen, tenzij een beheerder van Semaphore dat doet.
- **Verlaat Project** haalt u uit het project. De knop verschijnt alleen als de beheerder leden
  toestaat projecten te verlaten; ook de laatste Owner kan het project niet verlaten.
