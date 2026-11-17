---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Nastavení projektu

Nastavení projektu jako celku: jeho název, kolik jeho úloh smí běžet současně a zda odesílá
upozornění. Nebezpečná zóna pod formulářem obsahuje akce, které se týkají celého projektu: jeho
zálohu, jeho mezipaměť a jeho odstranění. Změny ve formuláři se projeví po kliknutí na **Uložit**.

## Název projektu {#newProject-name}

Název, pod kterým je projekt uveden v nabídce projektů v horní části postranního panelu. Pole je
povinné.

## Maximální počet paralelních úloh {#newProject-maxParallelTasks}

Kolik úloh tohoto projektu smí běžet současně. Úloha spuštěná nad limit čeká ve frontě, dokud
některá z běžících úloh neskončí.

Ponechte pole prázdné nebo ho nastavte na `0`, pokud projekt nemá mít vlastní limit. Limit celého
serveru, který nastavuje jeho správce, platí v obou případech.

## ID chatu Telegramu {#newProject-tg}

Volitelné. Chat Telegramu, který přijímá upozornění tohoto projektu. Když je pole prázdné,
upozornění jdou do chatu nastaveného pro celý server.

Pole má účinek pouze tehdy, když správce zapnul upozornění přes Telegram pro server a upozornění
jsou pro tento projekt povolena.

## Povolit upozornění pro tento projekt {#newProject-alert}

Zapne upozornění pro úlohy tohoto projektu. Kam upozornění chodí, nastavuje pro celý server jeho
správce: e-mail, Telegram, Slack, Rocket.Chat, Microsoft Teams, DingTalk nebo Gotify.

- E-mail se odešle, když úloha selže, a to členům projektu, kteří smí úlohu vidět a mají ve svém
  účtu zapnutá upozornění.
- Zpráva do messengerů se odešle, když úloha uspěje, selže nebo čeká na potvrzení.

Šablona úlohy může zabránit tomu, aby její úspěšné nebo neúspěšné úlohy odesílaly upozornění.

## Test Alerts {#settings-testAlerts}

Odešle testovací upozornění každým kanálem, který má server zapnutý, abyste si mohli ověřit, že
upozornění tohoto projektu dorazí.

Tlačítko funguje pouze tehdy, když je volba **Povolit upozornění pro tento projekt** zaškrtnutá
a uložená. Nejprve ji zaškrtněte a klikněte na **Uložit**.

## Zálohovat projekt {#settings-exportProject}

Stáhne soubor JSON s konfigurací projektu: šablonami úloh, inventáři, repozitáři, skupinami
proměnných, klíči, harmonogramy, integracemi a dalšími věcmi, které jsou v něm nastavené. Soubor
lze znovu převést na projekt pomocí **Obnovit projekt** v nabídce projektů v horní části
postranního panelu.

Soubor vynechává tajné údaje klíčů a tajné údaje na kartě **Secrets** skupin proměnných: po obnovení
je nutné je zadat znovu. Vše ostatní se zkopíruje tak, jak je, i heslo zapsané přímo do jiného pole:
adresa repozitáře, ve které je heslo, text statického inventáře, hodnoty běžných proměnných a
proměnných prostředí. Než soubor sdílíte, zkontrolujte, zda takové hodnoty neobsahuje. Historie úloh
není součástí zálohy.

## Vymazat mezipaměť {#settings-clearCache}

Odstraní soubory, které Semaphore pro tento projekt uchovává na serveru, například své kopie
repozitářů. Znovu se stáhnou při spuštění další úlohy.

Použijte to, když úloha selhává kvůli zastaralé nebo poškozené kopii repozitáře. Akci nelze
vrátit zpět a nastavení projektu se nedotkne.

## Odstranit projekt {#settings-deleteProject}

Odstraní projekt se vším, co obsahuje: šablonami úloh, historií úloh, inventáři, repozitáři,
skupinami proměnných, klíči a harmonogramy. Semaphore si nejprve vyžádá potvrzení.

Odstraněný projekt nelze obnovit. Než odstraníte projekt, který můžete znovu potřebovat, stáhněte
si jeho zálohu.
