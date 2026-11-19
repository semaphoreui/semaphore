---
form: src/components/RepositoryForm.vue
---

# Bewerken Repository

Een repository is de plek waar de playbooks, Terraform-code of scripts van een taaksjabloon
vandaan komen: een Git-repository, of een map op de machine die de taken uitvoert. Vóór elke taak
kloont Semaphore de Git-repository, of werkt de eerder gekloonde kopie bij, checkt de branch uit
en voert de taak in die kopie uit.

## Naam {#repository-name}

De naam waaronder de repository wordt vermeld op de pagina Repositories en in de repositorykeuze
van een taaksjabloon en van een inventaris. Het veld is verplicht.

## URL of pad {#repository-gitUrl}

Waar de code staat. Het veld is verplicht. De chips eronder tonen het soort adres dat u hebt
getypt, en een klik op een chip schrijft het adres om naar dat soort.

- **ssh**: `git@github.com:org/repo.git` of `ssh://git@github.com/org/repo.git`. Git logt in met
  de gebruiker in het adres en de **SSH Sleutel** die in **Toegangssleutel** is gekozen.
- **https** en **http**: `https://github.com/org/repo.git`. Een privérepository heeft een sleutel
  van het type **Inloggen met wachtwoord** nodig; een openbare heeft **Geen** nodig. Vermijd
  `http`: de login en het wachtwoord gaan dan onversleuteld over het netwerk.
- **git**: `git://host/org/repo.git`, het alleen-lezen Git-protocol, dat geen login kent.
- **file**: `file:///srv/git/repo.git`, een Git-repository op de schijf van de machine die de
  taak uitvoert, gekloond zoals een externe repository.
- **absoluut pad**: een absoluut pad zoals `/srv/playbooks` of `D:\playbooks`. Semaphore kloont
  het niet: de taak draait direct in die map, die moet bestaan op de machine die de taak uitvoert
  (de Semaphore-server of een runner). De branch is er niet op van toepassing.

Het wijzigen van het adres verwijdert de kopieën die de Semaphore-server van het oude adres heeft
gekloond.

## Tak {#repository-gitBranch}

De branch die Semaphore kloont en bijhoudt, bijvoorbeeld `main`. Het veld is verplicht, behalve
voor een absoluut pad, waarvoor het is vergrendeld.

Een taaksjabloon kan een eigen branch noemen, die voor deze gaat, en een sjabloon kan zijn taken
een branch laten kiezen wanneer ze worden gestart. Een tag werkt ook, maar een tag kan niet worden
gepulld, dus kloont Semaphore de repository vóór elke taak opnieuw.

## Single checkout for all templates {#repository-monorepo}

Standaard uit: elk taaksjabloon krijgt een eigen kopie van de repository. Vink het vakje aan om de
repository één keer te klonen en die kopie te delen tussen alle sjablonen die haar gebruiken, voor
een repository die te groot is om per sjabloon een kopie te bewaren.

De gedeelde kopie blijft op de branch van deze repository: sjablonen die haar gebruiken, kunnen
geen eigen branch hebben, en hun taken kunnen geen andere branch of commit kiezen. Vergrendeld voor
een absoluut pad.

## Toegangssleutel {#repository-sshKeyId}

De sleutel uit de Sleutel Opslag waarmee Git op de repository inlogt. Het veld is verplicht. Het
biedt sleutels van drie typen aan; welke u kiest, hangt af van de URL:

- **SSH Sleutel** voor een `ssh`-URL. Semaphore laadt de privésleutel in een SSH-agent zolang het
  klonen of bijwerken duurt. De gebruiker is die in de URL, zoals `git` in
  `git@github.com:org/repo.git`.
- **Inloggen met wachtwoord** voor een `https`-URL: de login en het wachtwoord, of het
  toegangstoken van de Git-server. Met een lege login wordt alleen het wachtwoord verstuurd, als
  gebruikersnaam, wat sommige servers van een token verwachten. Dezelfde referenties worden aan
  `ansible-galaxy` gegeven voor rollen en collecties op dezelfde server.
- **Geen** voor een openbare repository, een `git`- of `file`-URL, of een absoluut pad.

Een sleutel die niet bij de URL past, geeft Git geen referenties, en het klonen van een
privérepository mislukt dan.
