---
form: src/components/RoleMappingForm.vue
---

# New mapping

Mapování přidělí roli v projektu každému uživateli, který se přihlašuje přes poskytovatele LDAP
nebo OpenID Connect a patří tam do určité skupiny. Semaphore při každém přihlášení přes
poskytovatele načte skupiny uživatele a podle nich mapované role přidělí nebo odebere, takže
uživatel, který skupinu opustí, roli při dalším přihlášení ztratí. Mapovaná role ze uživatele
udělá člena projektu. Mapování nelze upravit: odstraňte ho a přidejte nové.

## Provider {#roleMapping-provider}

Adresář LDAP nebo poskytovatel OpenID Connect, ze kterého skupina pochází. Seznam obsahuje
poskytovatele nastavené v konfiguraci serveru. Pole je povinné.

Skupiny se načítají pouze tehdy, když je k tomu poskytovatel nastaven: poskytovatel LDAP
potřebuje základ hledání skupin (`group_search_dn`), poskytovatel OpenID Connect uvádí claim se
skupinami (`groups_claim`). Mapování poskytovatele, který je později z konfigurace odebrán,
zůstane v seznamu s označením **Provider not configured**.

## Group {#roleMapping-group}

Název skupiny tak, jak ho uvádí poskytovatel. Pole je povinné.

- LDAP: hodnota atributu s názvem skupiny v záznamu skupiny, ve výchozím stavu `cn`, například
  `devops`. Úplné DN jako `cn=devops,ou=groups,dc=example,dc=com` se zkrátí na jeho `cn`.
- OpenID Connect: jedna z hodnot claimu se skupinami, tak jak ji poskytovatel posílá.

## Projekt {#roleMapping-projectId}

Projekt, ve kterém se role přiděluje. Pole je povinné. Administrátor vidí všechny projekty;
ostatní uživatelé vidí pouze projekty, jejichž jsou členy.

## Role {#roleMapping-role}

Role, kterou členové skupiny v projektu dostanou. Pole je povinné a je k dispozici, jakmile je
vybrán projekt. Seznam obsahuje role tohoto projektu, sdílené role a vestavěné role. Když vyberete
jiný projekt, role, která v něm neexistuje, se vymaže.

Role **Owner** přidělená mapováním se nepočítá, když vám Semaphore brání odebrat roli poslednímu
vlastníkovi projektu, protože může při dalším přihlášení zmizet. Ponechte alespoň jednoho
vlastníka přidaného ručně.
