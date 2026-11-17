---
form: src/components/TeamMemberForm.vue
---

# Novo Membro da Equipe

Adiciona um utilizador ao projeto com uma função. A função decide o que o membro pode ver e
fazer no projeto; nada mais na sua conta é alterado. Quando o administrador ativou os convites,
o utilizador recebe um convite e torna-se membro depois de o aceitar.

## Email ou Username {#teamMember-inviteType}

Apresentado quando é possível convidar utilizadores tanto por e-mail como por nome de
utilizador.

Como escolhe a pessoa: **Email** para convidar um endereço de e-mail, **Username** para escolher
um utilizador que já tem conta neste servidor Semaphore.

## Email {#teamMember-email}

Apresentado quando a pessoa é convidada por e-mail.

O endereço de e-mail da pessoa a convidar. Torna-se membro do projeto depois de aceitar o
convite.

## Usuário {#teamMember-userId}

Apresentado quando a pessoa é escolhida pelo nome de utilizador.

O utilizador a adicionar, apresentado com o nome de utilizador e o nome completo. O campo é
obrigatório. A lista contém todos os utilizadores do servidor Semaphore que ainda não são membros
do projeto: uma pessoa sem conta tem primeiro de obter uma junto de um administrador.

## Função {#teamMember-role}

O que o membro pode fazer no projeto. O campo é obrigatório. As funções incorporadas:

- **Owner**: tudo, incluindo as configurações do projeto, a sua equipa e a sua eliminação.
- **Manager**: executa tarefas e workflows, aprova workflows e cria, altera e elimina modelos de
  tarefa, inventários, chaves, repositórios, grupos de variáveis, agendas, integrações e
  workflows. Um manager não pode alterar as configurações do projeto nem a sua equipa.
- **Task Runner**: vê o projeto, executa tarefas e workflows e aprova workflows. Um task runner
  não pode alterar nada no projeto.
- **Guest**: vê o projeto, as suas tarefas e os respetivos logs, e não altera nada.

A lista contém também as funções criadas para este projeto (**Funções do projeto**) e as funções
partilhadas por todos os projetos (**Shared roles**). Uma função destas concede as permissões
definidas para ela na página Roles. Um convite só pode conter uma função incorporada.

Um membro que pode gerir a equipa pode atribuir qualquer função, **Owner** incluída. A caixa de
diálogo atribui uma função; acrescente outras depois na coluna **Função** da lista da equipa.
