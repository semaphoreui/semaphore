---
form: src/components/InventoryGroupForm.vue
---

# Editar grupo

Um grupo de hosts de um inventário gerido. Um playbook é executado numa parte dos hosts pelo nome do
grupo, e uma variável do grupo é definida uma vez para todos os seus hosts. Os hosts são colocados
num grupo no separador Hosts, na caixa de diálogo do host ou com as ações sobre a seleção.

Um grupo importado de uma fonte de inventário é gerido por essa fonte. Todos os seus campos são só de
leitura e são atualizados a cada sincronização.

## Nome {#inventoryGroupForm-name}

O nome pelo qual o Ansible conhece o grupo, por exemplo em `hosts: web` de um playbook. O campo é
obrigatório. Um nome é único no seu inventário e distingue maiúsculas de minúsculas: `web` e `Web`
são dois grupos. Não pode conter espaços, vírgulas, parênteses retos nem dois pontos, tem no máximo
250 caracteres e não pode ser `all` nem `ungrouped`. O Ansible avisa sobre nomes com caracteres que
não sejam letras, algarismos e sublinhados, como `web-eu`, mas o nome funciona na mesma. Os
caracteres `*`, `?`, `!`, `&`, `~` e `@` também não são permitidos, porque o Ansible os lê como um
padrão.

## Descrição {#inventoryGroupForm-description}

Opcional. Uma nota para as pessoas, mostrada na lista de grupos. O Ansible não a recebe.

## Child groups {#inventoryGroupForm-children}

Opcional. Grupos dentro deste grupo. Os hosts de um grupo filho também pertencem a este grupo. Um
grupo não pode ser o seu próprio antepassado.

## Variáveis {#inventoryGroupForm-variables}

Opcional. As variáveis deste grupo, escritas como um mapeamento YAML ou JSON:

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

Todos os hosts do grupo as recebem. Uma variável de um host tem precedência sobre uma variável do
grupo com o mesmo nome, e uma variável do grupo tem precedência sobre uma variável do inventário.
Âncoras, aliases e chaves de fusão de YAML não são aceites.
