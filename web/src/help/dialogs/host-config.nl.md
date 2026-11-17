---
form: src/components/HostConfigForm.vue
---

# Edit mapping

Een koppeling vertelt Semaphore welke referentie uit de Sleutel Opslag moet worden gebruikt voor een
SSH-host of een Git-URL. Een repository heeft één sleutel, maar een taak kan ook andere servers
bereiken: submodules, rollen en collecties uit `requirements.yml`, Terraform-modules,
inventarisrepositories. De koppelingen van het project gelden voor elke Git- en SSH-verbinding van
zijn taken, en voor de Git-opdrachten die Semaphore voor het project uitvoert, zoals de controle
van een repository op nieuwe commits. Geen enkel bestand van uw repository verandert.

## Type {#hostConfigForm-type}

Waarop de koppeling overeenkomt:

- **Host**: een SSH-hostnaam. Elke SSH-verbinding van een taak met die host gebruikt de sleutel
  van de koppeling: Git via SSH, en Ansible dat verbinding maakt met een inventarishost met die
  naam. De koppeling gaat voor de SSH-configuratie van de server voor dezelfde host.
- **URL**: een Git-URL of het begin ervan. Git maakt verbinding met elke repository onder die URL
  met de referentie van de koppeling.

Overschakelen naar **Host** wist een **Credential** die een hostkoppeling niet kan gebruiken.

## Host of URL {#hostConfigForm-host}

Het veld is verplicht. Per host of URL kan er één koppeling van een type bestaan.

Voor **Host**: de hostnaam precies zoals Git of Ansible er verbinding mee maken, bijvoorbeeld
`github.com` of `10.0.0.21`. Alleen letters, cijfers, punten, koppeltekens en underscores: geen
gebruiker, geen poort, geen jokertekens.

Voor **URL**: een adres dat begint met `https://` of `http://`, zonder login erin en zonder
spaties, aanhalingstekens of `=`. Geef een hele repository op, of een voorvoegsel dat eindigt op
`/` om een groep repositories te dekken:

```
https://github.com/acme/
https://github.com/acme/infra.git
```

Wanneer meerdere koppelingen overeenkomen, wint de langste: de koppeling van de repository
hierboven gaat voor die ene repository voor de koppeling van `acme`.

Met een SSH-sleutel laat een URL-koppeling Git via SSH verbinding maken met de host van de URL op
poort `22`, welke poort de URL ook noemt.

## Credential {#hostConfigForm-sshKeyId}

De sleutel uit de Sleutel Opslag die wordt gebruikt voor de verbindingen waarop de koppeling
overeenkomt. Het veld is verplicht.

- **Host** accepteert alleen een **SSH Sleutel**. Wanneer de sleutel een login heeft, maakt SSH
  verbinding als die gebruiker.
- **URL** accepteert een **SSH Sleutel** of een **Inloggen met wachtwoord**. Een SSH-sleutel maakt
  verbinding als zijn login, of als `git` wanneer hij er geen heeft. Een login met wachtwoord werkt
  alleen met een `https://`-URL; zet voor een token het token in het wachtwoord en laat de login
  leeg.

Een login met wachtwoord bereikt Git via de omgeving, niet via de opdrachtregel, en het taaklog
toont alleen de oorspronkelijke URL.
