---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

Een globale rol verleent rechten over de hele Semaphore-instantie aan de gebruikers aan wie hij is
toegewezen; u wijst hem toe in het dialoogvenster van een gebruiker. Hij geeft geen toegang binnen
projecten: die komt alleen van de rollen die een gebruiker als lid van een project heeft. Een
beheerder heeft elk globaal recht, welke rollen er ook zijn toegewezen. Globale rollen zijn een
Pro-functie.

## Naam {#globalRole-name}

De naam waaronder de rol wordt vermeld op het tabblad Globale rollen van de pagina Gebruikers en in
het dialoogvenster van een gebruiker. Het veld is verplicht. U kunt een rol op elk moment
hernoemen.

## Slug {#globalRole-slug}

De identificatie van de rol. Semaphore vult deze in op basis van de naam terwijl u de naam typt,
totdat u de slug zelf bewerkt. De slug bestaat uit 1 tot 100 letters, cijfers, underscores en
koppeltekens, bijvoorbeeld `user-admins`.

De slug moet uniek zijn onder de globale rollen; een projectrol mag dezelfde slug hebben. De slug
kan niet meer worden gewijzigd nadat de rol is opgeslagen: het veld is vergrendeld wanneer u een
rol bewerkt.

## Global permissions {#globalRole-permissionsTitle}

Wat een gebruiker met de rol in de hele instantie mag doen. Bij een nieuwe rol is niets
aangevinkt. De rechten van alle globale rollen van een gebruiker tellen bij elkaar op.

- **Can manage users**: gebruikers aanmaken, wijzigen en verwijderen, en hun gegevens zien in de
  lijst met gebruikers. Zo'n gebruiker kan van een gebruiker geen beheerder of Pro-gebruiker maken,
  en kan geen beheerder wijzigen of verwijderen, noch een gebruiker die een globaal recht heeft dat
  hij zelf niet heeft.
- **Can manage global roles**: de globale rollen en de gedeelde projectrollen aanmaken, wijzigen en
  verwijderen, en bepalen welke gebruikers welke globale rol hebben. Geef dit alleen aan mensen
  die u de hele instantie toevertrouwt.
- **Can manage system settings**: de instellingen van de instantie, de applicaties, de globale
  runners en de groepskoppelingen wijzigen; het cluster en de informatie over de server bekijken;
  de taken van alle projecten bekijken en verwijderen; de cache van de server wissen.
- **Can read the audit log**: het auditlogboek van de instantie lezen en exporteren.
