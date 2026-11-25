---
form: src/components/WorkflowRunDialog.vue
---

# Exécuter le workflow

Lance une exécution du workflow une fois que vous avez répondu à son enquête. La boîte de dialogue
ne s'ouvre que pour un workflow doté d'une enquête : un workflow sans enquête démarre
immédiatement. Après **Exécuter**, la page de la nouvelle exécution s'ouvre ; lorsque le serveur
refuse les valeurs, la boîte de dialogue reste ouverte avec vos réponses.

## Variables {#workflowRunDialog-survey}

Un champ par variable de l'enquête du workflow, dans l'ordre de l'enquête, avec sa valeur par
défaut pré-remplie. Un champ marqué `*` est obligatoire, un champ Integer n'accepte que des
chiffres, et un champ de liste accepte une valeur ou, lorsque la variable en autorise plusieurs,
un nombre quelconque de valeurs. **Exécuter** ne lance rien tant que tous les champs ne sont pas
valides.

Les valeurs sont enregistrées avec l'exécution et affichées sur sa page sous **Variables**. Une
tâche du workflow ne les reçoit pas d'elle-même : une variable de son modèle ne reçoit une valeur
que si son nœud la lie à une variable du workflow. Le nom seul ne transmet rien.

Une enquête de workflow ne contient aucun secret, car l'exécution conserve ses valeurs en clair.
Placez les mots de passe et les jetons dans une clé ou un groupe de variables.
