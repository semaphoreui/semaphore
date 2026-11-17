---
form: src/views/Cluster.vue
---

# Vymazat úlohy z Redisu

Uzly clusteru sdílejí stav fondu úloh v Redisu: které úlohy čekají, které běží a který uzel
vlastní kterou z nich. Tento dialog odstraní z Redisu zaškrtnuté skupiny záznamů, aby obnovil
zablokovaný cluster, například s dokončenou úlohou, která se stále zobrazuje jako běžící.
Z databáze se nic neodstraní: úlohy si ponechají svůj stav, log i historii.

Ať zaškrtnete cokoli, každý uzel také zahodí úlohy, které drží v paměti, a znovu je načte
z Redisu a z databáze. Při otevření dialogu je zaškrtnutá pouze **Fronta**; zaškrtněte jen
skupiny, které jsou zablokované. Výsledek ukazuje, kolik záznamů bylo odstraněno.

## Fronta {#clusterClearDialog-queue}

Úlohy, které čekají, až je některý uzel spustí. Vymazané úlohy se nespustí: zůstanou v databázi
se stavem Waiting. Zastavte je, nebo jejich šablony spusťte znovu.

## Probíhající {#clusterClearDialog-running}

Úlohy, které cluster počítá jako běžící. Vymazání je nezastaví: úloha dál běží na svém runneru.
Cluster ji ale ztratí z evidence: úloha se už nezapočítává do limitů paralelních úloh serveru
a svého runneru, takže se může spustit víc úloh, než limity dovolují, a úloha, kterou runner
ještě nepřevzal, mu už nebude předána.

## Aktivní podle projektu {#clusterClearDialog-active}

Nedokončené úlohy každého projektu. Hlídají limit paralelních úloh projektu a brání spuštění
druhé úlohy šablony, zatímco jedna běží, když šablona paralelní úlohy nepovoluje. Tuto skupinu
vymažte, když úlohy projektu donekonečna čekají za úlohami, které už skončily. Úlohy, které
v tu chvíli běží, se už nepočítají, takže vedle nich mohou začít nové úlohy projektu.

## Aliasy {#clusterClearDialog-aliases}

Náhodné názvy, které úlohy Terraform, OpenTofu a Terragrunt dostávají na dobu svého běhu,
přiřazené k úloze. Běžící úloha se pod tímto názvem dostane ke svému stavu Terraformu na serveru:
po vymazání ji pod ním už nelze najít.

## Claims {#clusterClearDialog-claims}

Který uzel vlastní kterou úlohu. Uzel obnovuje claims svých úloh každou minutu; claim, který není
obnoven, vyprší po pěti minutách. Úloha bez claimu se považuje za osiřelou: úloha, která se ještě
nedostala k runneru, se vrátí do fronty pro jiný uzel, a úloha, kterou runner už provádí, se
nechá na něm doběhnout.

## Provozní pole {#clusterClearDialog-runtimeFields}

Údaje o každé úloze, které databáze neuchovává: uživatel, který ji spustil, verze, se kterou byla
spuštěna, její alias, její runner a její stav a projekt každé úlohy. Bez nich uzel, který úlohu
nespustil, o ní ukazuje méně a čistič osiřelých úloh přeskočí běžící úlohu, takže zablokovaný
záznam už nebude sám odstraněn.
