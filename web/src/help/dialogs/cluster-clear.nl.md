---
form: src/views/Cluster.vue
---

# Clear tasks from Redis

De nodes van een cluster delen de toestand van de takenpool in Redis: welke taken wachten, welke
draaien en welke node eigenaar is van elk ervan. Dit dialoogvenster verwijdert de groepen records
die u aanvinkt uit Redis, om een vastgelopen cluster te herstellen, bijvoorbeeld met een voltooide
taak die nog als lopend wordt getoond. Er wordt niets uit de database verwijderd: taken behouden
hun status, hun log en hun geschiedenis.

Wat u ook aanvinkt, elke node laat ook de taken los die hij in het geheugen bewaart en leest ze
opnieuw uit Redis en de database. Alleen **Queue** is aangevinkt wanneer het dialoogvenster opent;
vink alleen de groepen aan die vastzitten. Het resultaat toont hoeveel records zijn verwijderd.

## Queue {#clusterClearDialog-queue}

De taken die wachten tot een node ze start. Gewiste taken worden niet gestart: ze blijven in de
database staan met de status Waiting. Stop ze, of voer hun sjablonen opnieuw uit.

## Running {#clusterClearDialog-running}

De taken die het cluster als lopend telt. Wissen stopt ze niet: een taak blijft op zijn runner
draaien. Maar het cluster verliest het overzicht: de taak telt niet meer mee voor de limieten van
parallelle taken van de server en van zijn runner, zodat er meer taken kunnen starten dan de
limieten toestaan, en een taak die een runner nog niet heeft opgepakt, wordt hem niet meer
overgedragen.

## Active by project {#clusterClearDialog-active}

De onvoltooide taken van elk project. Ze handhaven de limiet van parallelle taken van een project
en voorkomen dat een tweede taak van een sjabloon start terwijl er een draait, wanneer het sjabloon
geen parallelle taken toestaat. Wis deze groep wanneer de taken van een project eindeloos wachten
achter taken die al voltooid zijn. Taken die op dat moment draaien, tellen niet meer mee, zodat
nieuwe taken van het project ernaast kunnen starten.

## Aliases {#clusterClearDialog-aliases}

De willekeurige namen die taken van Terraform, OpenTofu en Terragrunt voor hun levensduur krijgen,
gekoppeld aan de taak. Een lopende taak bereikt met deze naam zijn Terraform-state op de server:
na het wissen kan een lopende taak van deze apps er niet meer mee worden gevonden.

## Claims {#clusterClearDialog-claims}

Welke node eigenaar is van elke taak. Een node vernieuwt de claims van zijn taken elke minuut; een
claim die niet wordt vernieuwd, verloopt na vijf minuten. Een taak zonder claim geldt als verweesd:
een taak die nog geen runner heeft bereikt, wordt teruggezet in de wachtrij voor een andere node,
en een taak die een runner al uitvoert, mag daar afronden.

## Runtime fields {#clusterClearDialog-runtimeFields}

De gegevens van elke taak die de database niet bevat: de gebruiker die de taak startte, de versie
waarmee de taak werd gestart, de alias, de runner en de status ervan, en het project van elke
taak. Zonder deze gegevens toont een node die een taak niet heeft gestart er minder over, en slaat
de opruimer van verweesde taken een lopende taak over, zodat een vastgelopen record niet meer
vanzelf wordt verwijderd.
