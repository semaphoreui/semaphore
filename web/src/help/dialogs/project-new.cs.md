---
form: src/components/ProjectForm.vue
---

# Nový projekt

Projekt obsahuje šablony úloh, inventáře, repozitáře, skupiny proměnných a klíče jedné oblasti
práce a lidi, kteří na ní pracují. Projektu, který vytvoříte, se stanete vlastníkem. Projekt může
vytvořit pouze administrátor, pokud to správce serveru nepovolí všem uživatelům. Každé nastavení
níže lze později změnit na stránce Nastavení projektu.

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

## Demo {#newProject-demo}

Naplní nový projekt příklady, na kterých si můžete Semaphore vyzkoušet: veřejným repozitářem
`https://github.com/semaphoreui/semaphore-demo.git`, inventáři, klíčem s heslem vaultu
a šablonami úloh pro Ansible, OpenTofu, Terragrunt, Bash a PowerShell, roztříděnými do zobrazení
**Build**, **Deploy** a **Tools**.

Pro projekt bez příkladů nechte přepínač vypnutý: dostane pouze klíč **None** a zobrazení
**All**. Příklady jsou běžné položky projektu: upravte
nebo odstraňte je podle potřeby.
