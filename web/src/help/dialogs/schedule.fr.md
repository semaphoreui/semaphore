---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Modifier le calendrier

Un calendrier lance une tâche d'un modèle de tâche sans que personne n'appuie sur **Exécuter** :
de façon répétée selon une expression cron, ou une seule fois à un moment donné. La tâche reçoit
les paramètres enregistrés dans cette boîte de dialogue. Toutes les heures sont exprimées dans le
fuseau horaire du serveur Semaphore, affiché à côté des champs d'heure : `UTC`, sauf si
l'administrateur définit `schedule.timezone` ou `SEMAPHORE_SCHEDULE_TIMEZONE`.

## Name {#scheduleForm-name}

Le nom sous lequel le calendrier apparaît sur la page Calendrier. Le champ est obligatoire. Il ne
modifie pas ce que fait le calendrier.

## Template {#scheduleForm-templateId}

Le modèle de tâche que le calendrier lance. Le champ est obligatoire. La tâche s'exécute comme si
elle avait été lancée depuis le modèle, avec l'inventaire, le dépôt, le groupe de variables et les
autres paramètres propres au modèle, et avec les paramètres ci-dessous lorsque le modèle permet à
une tâche de les modifier.

## Paramètres de la tâche {#scheduleForm-taskParams}

Affiché lorsqu'un modèle est sélectionné.

Les paramètres de chaque tâche lancée par le calendrier. Les champs affichés dépendent du modèle :
ses variables d'enquête et les options qu'il permet à une tâche de modifier, comme la branche,
l'inventaire, **Limite**, **Étiquettes**, **Ignorer les étiquettes** ou **Arguments CLI**, ainsi
que les options d'Ansible ou de Terraform, OpenTofu et Terragrunt.

Le serveur vérifie le modèle à chaque déclenchement du calendrier : une valeur enregistrée ici est
ignorée dès que le modèle ne permet plus à une tâche de la modifier.

## Message {#taskParams-message}

Optionnel. Une note enregistrée avec chaque tâche lancée par le calendrier et affichée dans
l'historique des tâches, par exemple `Nightly deploy`. Elle ne modifie pas la façon dont la tâche
s'exécute.

## Variables d'enquête {#taskParams-survey}

Affiché lorsque le modèle possède des variables d'enquête.

Les valeurs des variables d'enquête du modèle pour les tâches planifiées. Elles parviennent à la
tâche de la même manière que lorsque quelqu'un remplit l'enquête à la main. Une variable
obligatoire doit avoir une valeur avant que le calendrier puisse être enregistré.

Les valeurs par défaut du modèle sont copiées à l'ouverture de la boîte de dialogue et sont
enregistrées avec le calendrier. Une valeur par défaut modifiée plus tard dans le modèle n'atteint
pas un calendrier enregistré auparavant.

La valeur d'une variable de type secret n'est pas enregistrée avec le calendrier : la tâche
planifiée s'exécute sans elle.

## Branche {#taskParams-gitBranch}

Affiché lorsque le modèle permet à une tâche de remplacer la branche.

La branche ou l'étiquette du dépôt que la tâche planifiée extrait. Le champ contient d'abord la
branche du modèle. Le serveur n'utilise la valeur que tant que le modèle autorise le remplacement.

## Inventaire {#taskParams-inventoryId}

Affiché lorsque le modèle permet à une tâche de choisir l'inventaire.

L'inventaire sur lequel la tâche planifiée s'exécute, à la place de l'inventaire du modèle. Pour un
modèle Terraform ou OpenTofu, le champ s'appelle **Workspace**, mais le serveur n'applique le choix
qu'aux modèles Ansible et conserve le workspace du modèle pour les autres.

## Déboguer {#taskParams-debug}

Affiché pour un modèle Ansible qui autorise le débogage.

Exécute `ansible-playbook` avec une sortie plus détaillée : `-v` répété autant de fois que l'indique
**Niveau de débogage**.

## Niveau de débogage {#taskParams-debugLevel}

Affiché pour un modèle Ansible qui autorise le débogage ; actif lorsque **Déboguer** est coché.

Le niveau de détail de la sortie, de `1` (`-v`) à `6` (`-vvvvvv`). La valeur par défaut est `4`.

## Exécution à blanc {#taskParams-dryRun}

Affiché pour un modèle Ansible, sauf si le modèle le masque.

Exécute le playbook avec `--check` : Ansible indique ce qu'il modifierait et ne modifie rien.

## Diff {#taskParams-diff}

Affiché pour un modèle Ansible, sauf si le modèle le masque.

Exécute le playbook avec `--diff` : Ansible affiche les modifications qu'il apporte aux fichiers et
aux templates.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Affiché pour un modèle Ansible qui permet à une tâche de le modifier.

Ignore l'installation des rôles et des collections des fichiers `requirements.yml` avant
l'exécution du playbook. Cochez-le lorsque les dépendances sont déjà installées et que la tâche ne
doit pas contacter Ansible Galaxy.

## Plan {#taskParams-plan}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Exécute uniquement `plan`. La tâche se termine après le plan et n'applique rien.

## Destroy {#taskParams-destroy}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Ajoute `-destroy` au plan et à l'application : la tâche détruit les ressources du workspace au lieu
de les créer ou de les mettre à jour.

## Auto Approve {#taskParams-autoApprove}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Applique le plan sans attendre de confirmation. Cela ne fonctionne que si le modèle autorise
l'approbation automatique ; un modèle configuré pour approuver chaque tâche applique le plan dans
tous les cas.

Sans approbation automatique, une tâche planifiée s'arrête après le plan et attend que quelqu'un
le confirme ou le rejette sur la page de la tâche.

## Upgrade {#taskParams-upgrade}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Exécute `init` avec `-upgrade` : les providers et les modules sont mis à jour vers les versions les
plus récentes que la configuration autorise.

## Reconfigure {#taskParams-reconfigure}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Exécute `init` avec `-reconfigure` au lieu de `-migrate-state` : le backend est configuré à nouveau
et l'état existant n'est pas migré.

## Run at {#scheduleForm-runAt}

Affiché pour un calendrier qui s'exécute une seule fois.

La date et l'heure de démarrage de la tâche, dans le fuseau horaire du serveur affiché à la fin du
champ. Un nouveau calendrier commence par la prochaine heure pleine. Le moment doit se situer dans
le futur : le serveur refuse une heure déjà passée.

Une fois la tâche démarrée, le calendrier est désactivé, ou supprimé lorsque **Delete after run**
est coché. Si Semaphore ne fonctionne pas à ce moment-là, la tâche n'est pas lancée plus tard : au
démarrage suivant, le calendrier est désactivé, ou supprimé, sans exécution.

## Delete after run {#scheduleForm-deleteAfterRun}

Affiché pour un calendrier qui s'exécute une seule fois.

Supprime le calendrier après le démarrage de sa tâche. Laissez la case décochée pour conserver le
calendrier, désactivé, afin de pouvoir définir une nouvelle heure et le réactiver.

## Show cron format {#scheduleForm-rawCron}

Affiché pour un calendrier cron.

Bascule entre la saisie de l'expression cron (activé) et sa construction à partir de **Timing** et
des cases à cocher d'heure (désactivé). Les deux modifient la même expression. Le navigateur
mémorise le choix.

Le commutateur reste bloqué sur activé lorsque l'expression ne peut pas être représentée par des
cases à cocher : lorsqu'elle est invalide, ou qu'elle utilise une forme comme `@daily` ou
`CRON_TZ=`.

## Cron {#scheduleForm-cronFormat}

Affiché pour un calendrier cron lorsque **Show cron format** est activé.

L'expression cron en cinq champs : minute, heure, jour du mois, mois, jour de la semaine (`0` est
le dimanche). Le champ est obligatoire, et le serveur vérifie l'expression pendant la saisie. Un
nouveau calendrier commence par `* * * * *`, chaque minute.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

Le serveur accepte aussi `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` et `@every 1h30m`,
ainsi qu'un préfixe comme `CRON_TZ=Europe/Berlin` pour un autre fuseau horaire. L'aperçu ci-dessous
ne peut pas afficher la prochaine exécution de ces expressions.

## Timing {#scheduleForm-timing}

Affiché pour un calendrier cron lorsque **Show cron format** est désactivé.

La fréquence d'exécution de la tâche. Les cases à cocher ci-dessous changent en conséquence :

- **Hourly** : **Minutes**.
- **Daily** : **Hours** et **Minutes**.
- **Weekly** : **Weekdays**, **Hours** et **Minutes**.
- **Monthly** : **Days**, **Hours** et **Minutes**.
- **Yearly** : **Months**, **Days**, **Hours** et **Minutes**.

## Months {#scheduleForm-months}

Affiché lorsque **Timing** vaut **Yearly**.

Les mois pendant lesquels la tâche s'exécute. Cochez-en plusieurs pour l'exécuter dans chacun
d'eux. Si aucun n'est coché, la tâche s'exécute tous les mois.

## Weekdays {#scheduleForm-weekdays}

Affiché lorsque **Timing** vaut **Weekly**.

Les jours de la semaine où la tâche s'exécute. Si aucun n'est coché, la tâche s'exécute tous les
jours.

## Days {#scheduleForm-days}

Affiché lorsque **Timing** vaut **Monthly** ou **Yearly**.

Les jours du mois où la tâche s'exécute. Un mois qui n'a pas ce jour est ignoré : une tâche réglée
sur le 31 ne s'exécute pas en avril. Si aucun n'est coché, la tâche s'exécute tous les jours.

## Hours {#scheduleForm-hours}

Affiché lorsque **Timing** vaut **Daily**, **Weekly**, **Monthly** ou **Yearly**.

Les heures auxquelles la tâche s'exécute, dans le fuseau horaire affiché à droite. Si aucune n'est
cochée, la tâche s'exécute toutes les heures.

## Minutes {#scheduleForm-minutes}

Affiché pour un calendrier cron lorsque **Show cron format** est désactivé.

Les minutes de l'heure auxquelles la tâche s'exécute, par pas de cinq. Si aucune n'est cochée, la
tâche s'exécute chaque minute : cochez-en au moins une, par exemple `:00`, pour une tâche qui doit
s'exécuter une fois par heure ou une fois par jour.

## Next run time {#scheduleForm-nextRun}

Le moment où la prochaine tâche démarrerait, d'après le calendrier tel qu'il figure dans le
formulaire : dans le fuseau horaire du serveur et dans l'heure locale de votre navigateur. Un
tiret signifie que l'expression est invalide, ou que l'aperçu ne sait pas la lire, comme `@daily`.

## Activé {#scheduleForm-active}

Indique si le calendrier lance des tâches. Décochez la case pour conserver un calendrier sans
l'exécuter. Le commutateur de la liste des calendriers fait la même chose. Un nouveau calendrier
est activé.
