---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Rattache cet espace de travail à un modèle, afin qu'il apparaisse parmi les espaces de travail du
modèle avec son état et ses alias. Un espace de travail appartient à un seul modèle à la fois : une
fois rattaché, il quitte la page Inventaire et se gère dans le modèle. Le rattachement n'en fait
pas l'espace de travail par défaut du modèle : pour cela, ouvrez le modèle et appuyez sur **Make
default**.

## Template {#inventoryAttach-templateId}

Le modèle auquel rattacher l'espace de travail. Le champ est obligatoire. La liste affiche les
modèles du projet créés pour la même application que l'espace de travail, Terraform ou OpenTofu.

L'espace de travail conserve son nom, sa clé et son état. Pour le déplacer plus tard vers un autre
modèle, appuyez sur **Detach** dans le modèle auquel il est rattaché : il revient sur la page
Inventaire. L'espace de travail par défaut d'un modèle ne peut pas être détaché.
