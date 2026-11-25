---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

Klíč předplatného odemyká na tomto serveru funkce Semaphore PRO a Enterprise. Dialog aktivuje
klíč a ukazuje plán aktivního klíče: jeho držitele, jeho stav, kdy vyprší a kolik uživatelů Pro,
runnerů a dalších položek povoluje a kolik se jich používá. Otevírá ho administrátor z nabídky
uživatele v edici Pro. Dokud není aktivní žádné předplatné, dialog se jmenuje
**Upgrade to Semaphore PRO**.

## Enter your PRO or EE key {#subscription-key}

Klíč předplatného tak, jak jste ho dostali se svým předplatným PRO nebo Enterprise. Vložte ho
sem, nebo ho načtěte ze souboru tlačítkem pro nahrání, a pak stiskněte **Activate New Key**. Pro
aktivaci je pole povinné.

Pole je jen pro čtení, když je klíč nastaven v konfiguraci serveru (`subscription.key` nebo
`subscription.key_file`, případně proměnné `SEMAPHORE_SUBSCRIPTION_KEY`
a `SEMAPHORE_SUBSCRIPTION_KEY_FILE`): klíč pak lze změnit pouze tam.

## Nahrát soubor s klíčem {#subscription-upload}

Zobrazí se, když ještě nebylo aktivováno žádné předplatné.

Načte klíč ze souboru ve vašem počítači (`.txt`, `.key`, `.pem` nebo `.lic`) do pole výše. Klíč
se neaktivuje, dokud nestisknete **Activate New Key**.

## Akce s klíčem {#subscription-actionsMenu}

Zobrazí se, když je předplatné aktivní nebo vypršelo.

Nabídka akcí s klíčem serveru:

- **Reload**: znovu načte stav předplatného, například po jeho obnovení.
- **Upload**: načte nový klíč ze souboru do pole výše, stejně jako tlačítko pro nahrání.
- **Reset**: okamžitě odebere klíč ze serveru, bez potvrzení.

## Activate New Key {#subscription-activate}

Aktivuje klíč z pole na tomto serveru. Když je klíč platný, zobrazí se níže plán předplatného.
Když váš účet ještě není uživatelem Pro a plán má volné místo pro uživatele Pro, dialog se zeptá,
zda z vašeho účtu udělat uživatele Pro.

Klíč, který je už aktivovaný, se ohlásí jako varování. Tlačítko je neaktivní, když je klíč
nastaven v konfiguraci serveru.

## Buy Pro {#subscription-buy}

Otevře v nové kartě prohlížeče stránku portálu Semaphore, kde se kupuje předplatné PRO.

## Get 30-day free trial {#subscription-trial}

Zobrazí se, dokud není aktivní žádné předplatné.

Otevře v nové kartě prohlížeče stránku portálu Semaphore, která vydává klíč pro bezplatnou
zkušební verzi na 30 dní. Klíč, který dostanete, vložte do pole výše.
