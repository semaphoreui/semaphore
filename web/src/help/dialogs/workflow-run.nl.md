---
form: src/components/WorkflowRunDialog.vue
---

# Run Workflow

Start een run van de workflow zodra u zijn enquête hebt beantwoord. Het dialoogvenster opent alleen
voor een workflow met een enquête: een workflow zonder enquête start direct. Na **Uitvoeren** opent
de pagina van de nieuwe run; wanneer de server de waarden weigert, blijft het dialoogvenster open
met uw antwoorden.

## Variabelen {#workflowRunDialog-survey}

Eén veld per variabele van de enquête van de workflow, in de volgorde van de enquête, met de
standaardwaarde ingevuld. Een veld dat met `*` is gemarkeerd, is verplicht, een veld van het type
Integer accepteert alleen cijfers, en een lijstveld accepteert één waarde of, wanneer de variabele
er meerdere toestaat, een willekeurig aantal. **Uitvoeren** start niets totdat elk veld geldig is.

De waarden worden met de run opgeslagen en op de pagina ervan getoond onder **Variables**. Een taak
van de workflow krijgt ze niet vanzelf: een variabele van haar sjabloon ontvangt alleen een waarde
wanneer haar node die aan een variabele van de workflow koppelt. De naam alleen levert niets op.

Een workflowenquête bevat geen geheimen, omdat de run zijn waarden onversleuteld bewaart. Zet
wachtwoorden en tokens in plaats daarvan in een sleutel of een variabele groep.
