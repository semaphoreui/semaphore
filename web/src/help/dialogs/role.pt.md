---
form: src/components/EditRoleForm.vue
---

# Edit Role

Uma função é um conjunto de permissões com nome dentro de um projeto. Atribui funções aos membros
de um projeto, e um membro pode fazer o que qualquer uma das suas funções permite. Aberta a partir
da página Roles de um projeto, a caixa de diálogo edita apenas uma função desse projeto; aberta a
partir da página Roles fora de um projeto, edita uma função partilhada que todos os projetos podem
utilizar. As funções incorporadas Owner, Manager, Task Runner e Guest não podem ser alteradas.

## Nome {#role-name}

O nome com que a função aparece na página Roles e em todos os sítios onde se escolhe uma função,
por exemplo quando adiciona um membro a um projeto. O campo é obrigatório. Pode mudar o nome de
uma função a qualquer momento.

## Slug {#role-slug}

O identificador da função. O Semaphore preenche-o a partir do nome enquanto escreve o nome, até
que edite o slug manualmente. Aceita de 1 a 100 letras, algarismos, sublinhados e hífenes, por
exemplo `release-managers`.

O slug não pode ser alterado depois de a função ser guardada: o campo fica bloqueado quando edita
uma função.

- Um slug tem de ser único, e a verificação ignora maiúsculas e minúsculas. O slug de uma função
  partilhada tem de ser diferente de todas as outras funções; o slug de uma função de projeto tem
  de ser diferente das funções partilhadas e das outras funções do seu projeto.
- `owner`, `manager`, `task_runner`, `guest` e `all` estão reservados, com quaisquer maiúsculas ou
  minúsculas.

## Permissions {#role-permissionsTitle}

O que os membros que têm a função podem fazer no projeto. Uma função nova não tem nenhuma
assinalada. As permissões de todas as funções de um membro somam-se; um administrador do Semaphore
pode fazer tudo em todos os projetos, sejam quais forem as funções.

- **Can read project**: ver o projeto: os seus modelos de tarefa, as tarefas e os respetivos logs,
  inventários, chaves, repositórios, grupos de variáveis, agendas e membros. Todas as leituras no
  projeto exigem-na, e nenhuma outra permissão a inclui, por isso assinale-a em quase todas as
  funções. Sem ela, um membro vê apenas os modelos de tarefa e os workflows que concedem acesso
  próprio à função, com as respetivas tarefas.
- **Can run project tasks**: iniciar tarefas a partir de modelos de tarefa, pará-las, e confirmar
  ou rejeitar uma tarefa que aguarda confirmação.
- **Can update project**: alterar as configurações do projeto e eliminá-lo.
- **Can manage project resources**: criar, alterar e eliminar modelos de tarefa, inventários,
  chaves, repositórios, grupos de variáveis, agendas, visualizações, integrações, armazenamentos
  de segredos, executores de projeto e as funções do projeto; eliminar tarefas; definir que
  funções podem utilizar um modelo de tarefa.
- **Can manage project users**: adicionar membros ao projeto, alterar as suas funções e
  removê-los.
- **Can run workflows**: iniciar um workflow e parar as suas execuções.
- **Can edit workflows**: criar, alterar e eliminar workflows, e definir que funções podem
  utilizar um workflow.
- **Can approve workflows**: aprovar ou rejeitar um passo de uma execução de workflow que aguarda
  aprovação.
