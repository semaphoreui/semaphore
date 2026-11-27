---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

De abonnementssleutel ontgrendelt de functies van Semaphore PRO en Enterprise op deze server. Het
dialoogvenster activeert een sleutel en toont het plan van de actieve sleutel: de houder, de
status, wanneer hij verloopt en hoeveel Pro-gebruikers, runners en andere items hij toestaat en er
in gebruik zijn. Het wordt door een beheerder geopend vanuit het gebruikersmenu van de Pro-editie.
Het dialoogvenster heet **Upgrade to Semaphore PRO** zolang er geen abonnement actief is.

## Enter your PRO or EE key {#subscription-key}

De abonnementssleutel, zoals u die met uw PRO- of Enterprise-abonnement hebt gekregen. Plak hem
hier, of laad hem uit een bestand met de uploadknop, en druk daarna op **Activate New Key**. Het
veld is verplicht voor de activering.

Het veld is alleen-lezen wanneer de sleutel in de configuratie van de server is ingesteld
(`subscription.key` of `subscription.key_file`, of de variabelen `SEMAPHORE_SUBSCRIPTION_KEY` en
`SEMAPHORE_SUBSCRIPTION_KEY_FILE`): de sleutel kan dan alleen daar worden gewijzigd.

## Sleutelbestand uploaden {#subscription-upload}

Wordt getoond wanneer er nog geen abonnement is geactiveerd.

Leest de sleutel uit een bestand op uw computer (`.txt`, `.key`, `.pem` of `.lic`) in het veld
hierboven. De sleutel wordt pas geactiveerd wanneer u op **Activate New Key** drukt.

## Sleutelacties {#subscription-actionsMenu}

Wordt getoond wanneer een abonnement actief is of is verlopen.

Een menu met acties op de sleutel van de server:

- **Reload**: haalt de toestand van het abonnement opnieuw op, bijvoorbeeld nadat het is
  verlengd.
- **Upload**: leest een nieuwe sleutel uit een bestand in het veld hierboven, net als de
  uploadknop.
- **Reset**: verwijdert de sleutel direct van de server, zonder bevestiging.

## Activate New Key {#subscription-activate}

Activeert de sleutel uit het veld op deze server. Wanneer de sleutel geldig is, wordt het plan van
het abonnement eronder getoond. Wanneer uw account nog geen Pro-gebruiker is en het plan een vrije
plaats voor een Pro-gebruiker heeft, vraagt het dialoogvenster of uw account een Pro-gebruiker moet
worden.

Een sleutel die al is geactiveerd, wordt als waarschuwing gemeld. De knop is uitgeschakeld wanneer
de sleutel in de configuratie van de server is ingesteld.

## Buy Pro {#subscription-buy}

Opent de pagina van het Semaphore-portaal waar een PRO-abonnement wordt gekocht, in een nieuw
tabblad van de browser.

## Get 30-day free trial {#subscription-trial}

Wordt getoond zolang er geen abonnement actief is.

Opent de pagina van het Semaphore-portaal die een sleutel uitgeeft voor een gratis proefperiode van
30 dagen, in een nieuw tabblad van de browser. Plak de sleutel die u krijgt in het veld hierboven.
