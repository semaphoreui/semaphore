---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Upravit skupinu proměnných

Skupina proměnných obsahuje doplňkové proměnné, proměnné prostředí a tajné údaje, které Semaphore
předává každé úloze šablony úlohy, jež skupinu používá. Ansible dostane doplňkové proměnné jako
`--extra-vars`, Terraform a OpenTofu jako `-var`; proměnné prostředí se nastaví pro proces úlohy.
Šablona může používat několik skupin: když dvě z nich definují stejný název, vyhraje skupina,
která je v šabloně uvedena později.

## Název skupiny {#varGroup-name}

Název, pod kterým je skupina uvedena na stránce Skupiny proměnných a ve výběru skupin proměnných
v šabloně úlohy. Pole je povinné.

## Secret storage {#varGroup-secretStorageId}

Zobrazí se, když je skupina nová a server podporuje úložiště tajných údajů. Volitelné.

Úložiště tajných údajů, ve kterém jsou uloženy tajné údaje této skupiny, například HashiCorp
Vault. Ponechte pole prázdné, chcete-li tajné údaje uchovávat v databázi Semaphore, zašifrované.

Úložiště se volí jednou, při vytvoření skupiny: existující skupina si své úložiště ponechá a pole
se už nezobrazuje. Úložiště označené **Read only** nepřijímá nové tajné údaje: uložení tajného
údaje do něj selže.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Zobrazí se, když je skupina nová a server podporuje úložiště tajných údajů. K dispozici, jakmile
je vybráno **Secret storage**.

Začátek cesty, kterou každý tajný údaj této skupiny v úložišti dostane. Semaphore k ní přidá 10
náhodných znaků: s předponou `semaphore/prod/` se tajný údaj uloží jako `semaphore/prod/`
následované těmito znaky. Předpona se bere tak, jak je, takže ji ukončete znakem `/`, když mají
tajné údaje ležet pod ní jako ve složce.

Předpona se nastavuje jednou, při vytvoření skupiny. Karta **Secrets** ji ukazuje jako vzor
zdrojové cesty.

## Table, JSON nebo YAML {#varGroup-extraVarsMode}

Jak upravujete doplňkové proměnné:

- **Table**: jeden řádek na proměnnou, každá s typem. Řádek přidáte tlačítkem **+** vedle
  přepínače.
- **JSON**: všechny proměnné jako jeden objekt JSON.
- **YAML**: všechny proměnné jako jedno mapování YAML.

Dialog se otevře v režimu **Table**. Přepnutí převede, co jste napsali; když to nelze přečíst,
dialog zůstane v režimu, ve kterém jste byli, a ukáže chybu. Ať je režim jakýkoli, proměnné se
uloží jako jeden objekt názvů a hodnot.

## Doplňkové proměnné v JSON {#varGroup-extraVarsJson}

Zobrazí se, když se doplňkové proměnné upravují jako **JSON**.

Doplňkové proměnné jako jeden objekt JSON. Název nesmí být prázdný; hodnota může být řetězec,
číslo, seznam nebo objekt:

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Doplňkové proměnné v YAML {#varGroup-extraVarsYaml}

Zobrazí se, když se doplňkové proměnné upravují jako **YAML**.

Doplňkové proměnné jako jedno mapování YAML. Text se čte s typy, které zná JSON: datum bez
uvozovek jako `2024-01-01` zůstane řetězcem a `.inf` nebo `.nan` jsou odmítnuty. Seznam nebo
skalár na nejvyšší úrovni jsou také odmítnuty.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Název proměnné {#varGroup-varName}

Zobrazí se, když se doplňkové proměnné upravují jako **Table**, na každém řádku.

Název proměnné tak, jak ji vidí playbook nebo kód Terraformu, například `app_version`.

## Typ proměnné {#varGroup-varType}

Zobrazí se, když se doplňkové proměnné upravují jako **Table**, na každém řádku.

Jak se čte hodnota řádku:

- **String**: text tak, jak je.
- **Number**: číslo, například `42`.
- **List**: pole JSON, například `["a", "b"]`.
- **Dict**: objekt JSON, například `{"key": "value"}`.

Hodnota, která neodpovídá svému typu, zastaví uložení s názvem proměnné.

## Hodnota proměnné {#varGroup-varValue}

Zobrazí se, když se doplňkové proměnné upravují jako **Table**, na každém řádku.

Hodnota proměnné ve formátu jejího typu. U **List** nebo **Dict** otevře tlačítko na konci pole
větší editor.

Semaphore předává hodnoty Ansiblu jako data: výraz `{{ … }}` uvnitř hodnoty se nevyhodnotí,
playbook dostane text tak, jak je.

## Název proměnné prostředí {#varGroup-envName}

Na kartě **Variables**, v části **Proměnné prostředí**, na každém řádku. Řádek přidáte tlačítkem
**+** této části.

Název proměnné nastavené v prostředí procesu, který úlohu spouští, například `AWS_REGION`.

## Hodnota proměnné prostředí {#varGroup-envValue}

Na kartě **Variables**, v části **Proměnné prostředí**, na každém řádku.

Hodnota proměnné jako text. Ukládá se jako prostý text a zobrazí se každému, kdo skupinu otevře:
hesla a tokeny místo toho uložte na kartu **Secrets**.

## Sync keys enabled {#varGroup-syncEnabled}

Zobrazí se na kartě **Secrets** skupiny, která má úložiště tajných údajů.

Zapněte, chcete-li importovat tajné údaje z úložiště do této skupiny, z cest nastavených v **Sync
paths**. Import se spustí, když stisknete tlačítko synchronizace skupiny na stránce Skupiny
proměnných, a sám v intervalu nastaveném v **Sync paths**. Když je volba vypnutá, nic se
neimportuje.

## Sync paths {#varGroup-syncPaths}

Zobrazí se, když je zapnuto **Sync keys enabled**.

Otevře dialog s cestami v úložišti, ze kterých se importuje, a s intervalem automatického
importu. Číslo na tlačítku je počet cest. Tlačítko synchronizace skupiny na stránce Skupiny
proměnných zůstává neaktivní, dokud skupina nemá alespoň jednu cestu.

Dialog má pouze **Zavřít**: co v něm změníte, se uloží se skupinou tlačítkem **Uložit** tohoto
dialogu.

## Název tajné proměnné {#varGroup-secretVarName}

Na kartě **Secrets**, v části **Doplňkové proměnné**, na každém řádku. Řádek přidáte tlačítkem
**+** této části.

Název doplňkové proměnné, jejíž hodnota je tajná. Úloha ji dostane stejně jako doplňkové proměnné
z karty **Variables**; tajný údaj vyhraje nad doplňkovou proměnnou se stejným názvem. Ansible
dostane tajné údaje v souboru zašifrovaném pomocí Ansible Vault, Terraform a OpenTofu jako
`-var`. Jak dialog varuje, tajný údaj předaný tímto způsobem se přesto může objevit jako prostý
text v logu Ansiblu.

## Hodnota tajné proměnné {#varGroup-secretVarValue}

Na kartě **Secrets**, v části **Doplňkové proměnné**, na každém řádku.

Hodnota tajného údaje. Semaphore ji zašifruje, nebo zapíše do úložiště tajných údajů skupiny,
a už ji nikdy neukáže: když skupinu později otevřete, pole je prázdné. Ponechte pole prázdné,
chcete-li zachovat uloženou hodnotu, nebo zadejte novou, chcete-li ji nahradit. Odstranění řádku
odstraní tajný údaj při uložení skupiny.

## Název tajné proměnné prostředí {#varGroup-secretEnvName}

Na kartě **Secrets**, v části **Proměnné prostředí**, na každém řádku. Řádek přidáte tlačítkem
**+** této části.

Název proměnné prostředí, jejíž hodnota je tajná, například `DB_PASSWORD`. Nastaví se
v prostředí procesu, který úlohu spouští, stejně jako proměnné prostředí z karty **Variables**.

## Hodnota tajné proměnné prostředí {#varGroup-secretEnvValue}

Na kartě **Secrets**, v části **Proměnné prostředí**, na každém řádku.

Hodnota tajného údaje. Stejně jako u tajné doplňkové proměnné se už nikdy neukáže: ponechte pole
prázdné, chcete-li zachovat uloženou hodnotu, nebo zadejte novou, chcete-li ji nahradit.
