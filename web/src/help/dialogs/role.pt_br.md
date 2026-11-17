---
form: src/components/EditRoleForm.vue
---

# Edit Role

Uma função é um conjunto nomeado de permissões dentro de um projeto. Você atribui funções aos
membros de um projeto, e um membro pode fazer o que qualquer uma das suas funções permitir. Aberta
a partir da página Roles de um projeto, a caixa de diálogo edita apenas uma função desse projeto;
aberta a partir da página Roles fora de um projeto, ela edita uma função compartilhada que todos os
projetos podem usar. As funções incorporadas Owner, Manager, Task Runner e Guest não podem ser
alteradas.

## Nome {#role-name}

O nome com que a função aparece na página Roles e onde quer que uma função seja escolhida, por
exemplo quando você adiciona um membro a um projeto. O campo é obrigatório. Você pode renomear uma
função a qualquer momento.

## Slug {#role-slug}

O identificador da função. O Semaphore o preenche a partir do nome enquanto você digita o nome,
até que você edite o slug por conta própria. Ele aceita de 1 a 100 letras, dígitos, sublinhados e
hifens, por exemplo `release-managers`.

O slug não pode ser alterado depois que a função é salva: o campo fica bloqueado quando você edita
uma função.

- Um slug precisa ser único, e a verificação ignora maiúsculas e minúsculas. O slug de uma função
  compartilhada precisa ser diferente de todas as outras funções; o slug de uma função de projeto
  precisa ser diferente das funções compartilhadas e das outras funções do seu projeto.
- `owner`, `manager`, `task_runner`, `guest` e `all` são reservados, em maiúsculas ou minúsculas.

## Permissions {#role-permissionsTitle}

O que os membros que têm a função podem fazer no projeto. Uma nova função não tem nenhuma marcada.
As permissões de todas as funções de um membro se somam; um administrador do Semaphore pode fazer
tudo em todos os projetos, sejam quais forem as funções.

- **Can read project**: ver o projeto: os modelos de tarefa, as tarefas e os logs delas, os
  inventários, as chaves, os repositórios, os grupos de variáveis, as agendas e os membros. Toda
  leitura no projeto a exige, e nenhuma outra permissão a inclui, então marque-a para quase todas
  as funções. Sem ela, um membro vê apenas os modelos de tarefa e os workflows que concedem acesso
  próprio à função, com as tarefas deles.
- **Can run project tasks**: iniciar tarefas a partir dos modelos de tarefa, pará-las, e confirmar
  ou rejeitar uma tarefa que aguarda uma confirmação.
- **Can update project**: alterar as configurações do projeto e excluí-lo.
- **Can manage project resources**: criar, alterar e excluir modelos de tarefa, inventários,
  chaves, repositórios, grupos de variáveis, agendas, visualizações, integrações, armazenamentos de
  segredos, executores de projeto e as funções do projeto; excluir tarefas; definir quais funções
  podem usar um modelo de tarefa.
- **Can manage project users**: adicionar membros ao projeto, alterar as funções deles e removê-los.
- **Can run workflows**: iniciar um workflow e parar as execuções dele.
- **Can edit workflows**: criar, alterar e excluir workflows, e definir quais funções podem usar um
  workflow.
- **Can approve workflows**: aprovar ou rejeitar uma etapa de uma execução de workflow que aguarda
  uma aprovação.
