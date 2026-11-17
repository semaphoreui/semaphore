---
form: src/components/RoleMappingForm.vue
---

# New mapping

Een koppeling geeft een rol in een project aan elke gebruiker die inlogt via een LDAP- of OpenID
Connect-provider en daar tot een groep behoort. Semaphore leest bij elke aanmelding via de
provider de groepen van de gebruiker en geeft of ontneemt de gekoppelde rollen zodat ze
overeenkomen, dus een gebruiker die de groep verlaat, verliest de rol bij zijn volgende
aanmelding. Een gekoppelde rol maakt de gebruiker lid van het project. Een koppeling kan niet
worden bewerkt: verwijder haar en voeg een nieuwe toe.

## Provider {#roleMapping-provider}

De LDAP-directory of de OpenID Connect-provider waar de groep vandaan komt. De lijst bevat de
providers die in de configuratie van de server zijn ingesteld. Het veld is verplicht.

De groepen worden alleen gelezen wanneer de provider daarvoor is ingesteld: een LDAP-provider heeft
een zoekbasis voor groepen nodig (`group_search_dn`), een OpenID Connect-provider noemt de claim
met de groepen (`groups_claim`). Een koppeling van een provider die later uit de configuratie wordt
verwijderd, blijft in de lijst staan, gemarkeerd met **Provider not configured**.

## Group {#roleMapping-group}

De naam van de groep zoals de provider die geeft. Het veld is verplicht.

- LDAP: de waarde van het attribuut met de groepsnaam van de groepsvermelding, standaard `cn`,
  bijvoorbeeld `devops`. Een volledige DN zoals `cn=devops,ou=groups,dc=example,dc=com` wordt
  teruggebracht tot zijn `cn`.
- OpenID Connect: een van de waarden van de groepsclaim, zoals de provider die verstuurt.

## Project {#roleMapping-projectId}

Het project waarin de rol wordt gegeven. Het veld is verplicht. Een beheerder ziet elk project;
elke andere gebruiker ziet alleen de projecten waarvan hij lid is.

## Rol {#roleMapping-role}

De rol die de leden van de groep in het project krijgen. Het veld is verplicht en is beschikbaar
zodra een project is geselecteerd. De lijst bevat de rollen van dat project, de gedeelde rollen en
de ingebouwde rollen. Wanneer u een ander project kiest, wordt een rol die daar niet bestaat
gewist.

Een rol **Owner** die door een koppeling is gegeven, telt niet mee wanneer Semaphore voorkomt dat u
de rol afneemt van de laatste eigenaar van een project, omdat ze bij de volgende aanmelding kan
verdwijnen. Houd minstens één eigenaar die met de hand is toegevoegd.
