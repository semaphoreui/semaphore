---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Upravit Secret Storage

Úložiště tajných údajů je externí správce tajných údajů, do kterého Semaphore tajné údaje ukládá
a ze kterého je čte: HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets Manager nebo Azure
Key Vault. Skupina proměnných vytvořená s úložištěm v něm uchovává své tajné údaje a úložiště se
zapnutou synchronizací importuje své tajné údaje do Úložiště klíčů projektu. Druh úložiště se
vybírá v nabídce **New Storage** a v tomto dialogu ho nelze změnit; pole níže na něm závisí.

## Název {#secretStorage-name}

Název, pod kterým je úložiště uvedeno na kartě **Storages** Úložiště klíčů a ve výběru **Secret
storage** nové skupiny proměnných. Pole je povinné.

## Server URL {#secretStorage-vaultURL}

Zobrazí se, když je úložiště HashiCorp Vault, OpenBao nebo Devolutions Server.

Adresa serveru se schématem, například `https://vault.example.com:8200`. Pole je povinné.

## Mount {#secretStorage-vaultMount}

Zobrazí se, když je úložiště HashiCorp Vault nebo OpenBao. Volitelné.

Cesta připojení (mount) modulu tajných údajů, který tajné údaje obsahuje. Ponechte pole prázdné,
chcete-li použít výchozí `secret`.

## Namespace {#secretStorage-vaultNamespace}

Zobrazí se, když je úložiště HashiCorp Vault nebo OpenBao. Volitelné.

Jmenný prostor, ve kterém tajné údaje jsou. Jmenné prostory existují ve Vault Enterprise a HCP
Vault Dedicated a v OpenBao od verze 2.3. Pro server bez jmenných prostorů ponechte pole prázdné.

## Zdroj tajného údaje {#secretSource-options}

Zobrazí se u každého druhu úložiště, nad tajným údajem, kterým se Semaphore přihlašuje: **Token**
u Vaultu nebo OpenBao, tajný klíč aplikace u Devolutions Serveru, tajný klíč u AWS, tajný klíč
klienta u Azure.

Odkud Semaphore tento tajný údaj bere:

- **Store in DB**: tajný údaj zadáte do tohoto dialogu; Semaphore ho zašifruje a uchová ve své
  databázi.
- **From ENV**: zadáte název proměnné prostředí serveru Semaphore; tajný údaj se z ní čte
  pokaždé, když se úložiště použije.
- **From File**: zadáte absolutní cestu k souboru na serveru Semaphore; tajný údaj se z něj čte
  pokaždé, bez koncového zalomení řádku. Soubor musí být uvnitř adresáře tajných údajů serveru
  (`dirs.secrets` v konfiguraci, ve výchozím stavu `/tmp/semaphore`) a cesta obsahující `..` je
  odmítnuta.

Přepnutí zdroje vymaže pole pod ním.

## Token {#secretStorage-vaultToken}

Zobrazí se, když je úložiště HashiCorp Vault nebo OpenBao a zdroj tajného údaje je
**Store in DB**.

Token, kterým se Semaphore k serveru ověřuje. Musí mít oprávnění tajné údaje číst a, pokud
úložiště není **Read only**, také je zapisovat a odstraňovat.

U nového úložiště je pole povinné. Když otevřete existující úložiště, je prázdné: ponechte ho
prázdné, chcete-li zachovat uložený token, nebo zadejte nový, chcete-li ho nahradit.

## Název proměnné prostředí / cesta k souboru {#secretStorage-vaultTokenSource}

Zobrazí se, když je úložiště HashiCorp Vault nebo OpenBao a zdroj tajného údaje je
**From ENV** nebo **From File**.

Název proměnné prostředí, nebo cesta k souboru, který obsahuje token. U nového úložiště je pole
povinné.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Zobrazí se, když je úložiště Devolutions Server.

Semaphore se k serveru připojí bez kontroly jeho certifikátu TLS. Kdokoli mezi Semaphore
a serverem pak může tajné údaje číst. Ponechte volbu vypnutou a místo toho dejte serveru
certifikát, kterému Semaphore důvěřuje.

## Vault ID {#secretStorage-dvlsKey}

Zobrazí se, když je úložiště Devolutions Server.

ID trezoru Devolutions Serveru, který tajné údaje obsahuje. U nového úložiště je pole povinné.

## App Key {#secretStorage-dvlsAppKey}

Zobrazí se, když je úložiště Devolutions Server.

Klíč aplikace identity aplikace, kterou se Semaphore přihlašuje, té, jejíž tajný údaj patří do
**Secret**. U nového úložiště je pole povinné.

## Secret {#secretStorage-dvlsSecret}

Zobrazí se, když je úložiště Devolutions Server a zdroj tajného údaje je **Store in DB**.

Tajný klíč aplikace identity aplikace z **App Key**. U nového úložiště je pole povinné. Když
otevřete existující úložiště, je prázdné: ponechte ho prázdné, chcete-li zachovat uložený tajný
údaj, nebo zadejte nový, chcete-li ho nahradit.

## Název proměnné prostředí / cesta k souboru {#secretStorage-dvlsEnv}

Zobrazí se, když je úložiště Devolutions Server a zdroj tajného údaje je **From ENV** nebo
**From File**.

Název proměnné prostředí, nebo cesta k souboru, který obsahuje tajný klíč aplikace. U nového
úložiště je pole povinné.

## Region {#secretStorage-awsRegion}

Zobrazí se, když je úložiště AWS Secrets Manager.

Region AWS, ve kterém tajné údaje jsou, například `us-east-1`. Pole je povinné.

## Endpoint URL {#secretStorage-awsEndpointURL}

Zobrazí se, když je úložiště AWS Secrets Manager. Volitelné.

Adresa API Secrets Manager, která se má volat místo výchozího endpointu regionu, například
endpoint VPC. Ponechte pole prázdné, chcete-li použít výchozí endpoint AWS.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Zobrazí se, když je úložiště AWS Secrets Manager.

Když je volba zapnutá, Semaphore neukládá žádný přístupový klíč a přihlašovací údaje AWS bere
z prostředí, ve kterém server běží, například z role IAM jeho instance. Pole přístupového klíče
níže jsou pak neaktivní.

Zapnutí volby při uložení odstraní uložený přístupový klíč. Opětovné vypnutí vyžaduje nový
**Secret Access Key**.

## Access Key ID {#secretStorage-awsAccessKeyId}

Zobrazí se, když je úložiště AWS Secrets Manager. Neaktivní, když je zapnuto
**Use IAM Role / Instance Profile**.

ID přístupového klíče, kterým se Semaphore přihlašuje, například `AKIAIOSFODNN7EXAMPLE`. Pole je
povinné, když je role IAM vypnutá. ID není tajné: ukládá se a zobrazuje tak, jak je.

## Secret Access Key {#secretStorage-awsSecretKey}

Zobrazí se, když je úložiště AWS Secrets Manager a zdroj tajného údaje je **Store in DB**.
Neaktivní, když je zapnuto **Use IAM Role / Instance Profile**.

Tajná část přístupového klíče. Pole je povinné u nového úložiště a po vypnutí role IAM. Jinak je
při otevření úložiště prázdné: ponechte ho prázdné, chcete-li zachovat uložený klíč, nebo zadejte
nový, chcete-li ho nahradit.

## Název proměnné prostředí / cesta k souboru {#secretStorage-awsSecretKeySource}

Zobrazí se, když je úložiště AWS Secrets Manager a zdroj tajného údaje je **From ENV** nebo
**From File**. Neaktivní, když je zapnuto **Use IAM Role / Instance Profile**.

Název proměnné prostředí, nebo cesta k souboru, který obsahuje tajný přístupový klíč. Pole je
povinné ve stejných případech jako **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Zobrazí se, když je úložiště Azure Key Vault.

Adresa trezoru klíčů, například `https://my-vault.vault.azure.net`. Pole je povinné.

## Tenant ID {#secretStorage-azureTenantId}

Zobrazí se, když je úložiště Azure Key Vault.

ID tenanta Microsoft Entra aplikace, jako kterou se Semaphore přihlašuje. Pole je povinné.

## Client ID {#secretStorage-azureClientId}

Zobrazí se, když je úložiště Azure Key Vault.

ID aplikace (klienta) této aplikace. Potřebuje přístup k tajným údajům trezoru klíčů. Pole je
povinné.

## Client Secret {#secretStorage-azureClientSecret}

Zobrazí se, když je úložiště Azure Key Vault a zdroj tajného údaje je **Store in DB**.

Tajný klíč klienta aplikace. U nového úložiště je pole povinné. Když otevřete existující
úložiště, je prázdné: ponechte ho prázdné, chcete-li zachovat uložený tajný údaj, nebo zadejte
nový, chcete-li ho nahradit.

## Název proměnné prostředí / cesta k souboru {#secretStorage-azureClientSecretSource}

Zobrazí se, když je úložiště Azure Key Vault a zdroj tajného údaje je **From ENV** nebo
**From File**.

Název proměnné prostředí, nebo cesta k souboru, který obsahuje tajný klíč klienta. U nového
úložiště je pole povinné.

## Read only {#secretStorage-readonly}

Když je volba zapnutá, Semaphore z úložiště pouze čte. Tajný údaj skupiny proměnných v něm nelze
vytvořit ani změnit: jeho uložení selže. Odstranění tajného údaje, klíče nebo skupiny proměnných
v Semaphore ponechá jeho tajný údaj v úložišti.

Když je volba vypnutá, Semaphore do úložiště zapisuje tajné údaje skupin proměnných, které ho
používají, a odstraňuje je z něj spolu se skupinami.

## Sync keys enabled {#secretStorage-syncEnabled}

Zapněte, chcete-li importovat tajné údaje z úložiště do Úložiště klíčů projektu, z cest
nastavených v **Sync paths**. Import se spustí, když stisknete tlačítko synchronizace úložiště na
kartě **Storages**, a sám v intervalu nastaveném v **Sync paths**.

Importovaný klíč je označen jako synchronizovaný: jeho název a typ nelze změnit a jeho
odstranění v Semaphore ponechá tajný údaj v úložišti. Odstranění úložiště, když je synchronizace
zapnutá, odstraní také klíče, které importovalo.

## Sync paths {#secretStorage-syncPaths}

Zobrazí se, když je zapnuto **Sync keys enabled**.

Otevře dialog s cestami v úložišti, ze kterých se importuje, a s intervalem automatického
importu. Číslo na tlačítku je počet cest. Tlačítko synchronizace úložiště zůstává neaktivní,
dokud úložiště nemá alespoň jednu cestu.

Dialog má pouze **Zavřít**: co v něm změníte, se uloží s úložištěm tlačítkem **Uložit** tohoto
dialogu (u nového úložiště **Vytvořit**).
