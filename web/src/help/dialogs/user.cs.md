---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Upravit uživatele

Uživatel je účet, který se přihlašuje do Semaphore. Karta **Settings** obsahuje profil účtu a jeho
příznaky; **Uložit** je zapíše. Karta **Security** obsahuje heslo, dvoufaktorové ověřování
a propojené přihlašovací účty a karta **Roles** globální role uživatele: obě se projeví okamžitě,
bez **Uložit**. Co smí uživatel dělat uvnitř projektu, se nastavuje na stránce **Tým** daného
projektu, ne zde.

## Název {#user-name}

Celé jméno uživatele, zobrazené v postranním panelu a v seznamech uživatelů a členů projektů. Pole
je povinné.

U uživatele **Externí** převezme Semaphore jméno při přihlášení od poskytovatele identity, když
ho poskytovatel posílá, takže změna provedená zde nemusí vydržet déle než do dalšího přihlášení.

## Uživatelské jméno {#user-username}

Přihlašovací jméno uživatele. Pole je povinné. Uživatel se přihlašuje uživatelským jménem nebo
e-mailem a heslem.

Uživatelské jméno uživatele **Externí** nelze změnit: server odmítne ho uložit.

## E-mail {#user-email}

E-mailová adresa uživatele. Pole je povinné. Semaphore na ni posílá upozornění na úlohy
uživatele a aplikace pro ověřování ji ukazuje jako název účtu. Je soukromá: vidí ji pouze
uživatel a ti, kdo smí spravovat uživatele.

U uživatele **Externí** pochází e-mail od poskytovatele identity: pole je jen pro čtení
a Semaphore adresu aktualizuje při každém přihlášení.

## Heslo {#user-password}

Zobrazí se, když se vytváří nový uživatel.

Heslo, kterým se uživatel přihlašuje. Pole je povinné, kromě uživatele **Externí**, který
v Semaphore žádné heslo nemá. Chcete-li změnit heslo existujícího uživatele, použijte
**Change password** na kartě **Security**.

## Odesílat upozornění {#user-alert}

Pošle uživateli e-mail, když selže úloha v projektu, ve kterém smí uživatel úlohu vidět. Projekt
musí upozornění povolovat a správce serveru musí mít zapnutá upozornění e-mailem. Šablona úlohy
může zabránit tomu, aby její neúspěšné úlohy odesílaly upozornění.

Příznak nemá vliv na upozornění projektu odesílaná do Telegramu, Slacku a dalších messengerů.

## Administrátor {#user-admin}

Zobrazí se, když jste administrátor.

Administrátor smí na serveru všechno: spravovat uživatele, globální role a nastavení serveru
a otevřít každý projekt s plnými právy.

Příznak nemůžete odebrat svému vlastnímu účtu a server ponechává alespoň jednoho administrátora:
odmítne odebrat příznak poslednímu z nich.

## Pro user {#user-pro}

Zobrazí se v edici Pro. Příznak může změnit pouze administrátor.

Dá uživateli místo v předplatném Pro. Předplatné má pevný počet uživatelů Pro: když je dosažen,
server odmítne příznak uložit. Bez aktivního předplatného nemá příznak žádný účinek.

## Externí {#user-external}

Zobrazí se, když jste administrátor. Lze nastavit pouze při vytváření uživatele.

Externí uživatel se přihlašuje přes LDAP nebo OpenID Connect, ne heslem Semaphore: pole hesla je
neaktivní a přihlášení heslem je odmítnuto. Jméno a e-mail se přebírají od poskytovatele identity
při každém přihlášení. Pro uživatele, který se přihlašuje heslem, nechte políčko nezaškrtnuté.

## Change password {#user-changePassword}

Zobrazí se na kartě **Security** existujícího uživatele, který není **Externí**, když server
dovoluje přihlašování heslem.

Otevře dialog, který nastaví nové heslo. Chcete-li změnit vlastní heslo, musíte zadat současné;
při obnovení hesla jiného uživatele ne.

## Time-based one-time password {#user-totp}

Zobrazí se na kartě **Security**, když správce zapnul pro server dvoufaktorové ověřování.

Zapne pro tohoto uživatele dvoufaktorové ověřování. Když přepínač zapnete, Semaphore okamžitě
vytvoří tajný klíč a ukáže ho jako QR kód: naskenujte ho aplikací pro ověřování. Od té chvíle si
přihlášení po hesle vyžádá kód z aplikace.

Když server dovoluje obnovení, pod QR kódem se zobrazí kód pro obnovení. Zobrazí se pouze jednou:
zkopírujte si ho a uschovejte ho na bezpečném místě. Vypnutí přepínače tajný klíč okamžitě
odstraní.

## Poskytovatel LDAP {#user-ldapProvider}

Zobrazí se na kartě **Security** vašeho vlastního účtu, když má server poskytovatele LDAP, ke
kterému ještě váš účet není propojen.

Server LDAP, ke kterému se má váš účet propojit, když jich je několik.

## LDAP username {#user-ldapUsername}

Zobrazí se spolu s **Poskytovatel LDAP**.

Vaše uživatelské jméno na serveru LDAP. Propojení vám umožní přihlašovat se k tomuto účtu také
přes LDAP.

## LDAP password {#user-ldapPassword}

Zobrazí se spolu s **Poskytovatel LDAP**.

Vaše heslo na serveru LDAP. Semaphore ho použije jednou, aby ověřil, že účet LDAP patří vám, a
neukládá ho.

## Link {#user-ldapLink}

Zobrazí se spolu s **Poskytovatel LDAP**.

Propojí váš účet s účtem LDAP výše. Propojení je odmítnuto, když jsou přihlašovací údaje
nesprávné, když je účet LDAP propojen s jiným uživatelem, nebo když je váš účet s tímto
poskytovatelem už propojen. Propojený účet se objeví v části **Linked accounts**; zavřením jeho
štítku ho odpojíte.

## Global role {#global-role-select}

Zobrazí se na kartě **Roles** existujícího uživatele v edici Pro, když smíte spravovat globální
role.

Globální role, kterou chcete uživateli dát, z rolí, které uživatel ještě nemá. Globální role
uděluje oprávnění nad celým serverem: spravovat uživatele, globální role nebo nastavení serveru,
nebo číst auditní log. Samotné globální role se vytvářejí na kartě **Globální role** stránky
Uživatelé.

## Assign {#globalRoleAssign-assign}

Zobrazí se spolu s **Global role**.

Okamžitě dá uživateli vybranou roli. Role, které uživatel má, jsou uvedené níže, každá
s tlačítkem pro její odebrání. **Effective global permissions** ukazuje, co uživatel nakonec smí
dělat a odkud každé oprávnění pochází.
