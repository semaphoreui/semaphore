---
form: src/components/TeamMemberForm.vue
---

# Novo Membro da Equipe

Adiciona um usuário ao projeto com uma função. A função decide o que o membro pode ver e fazer no
projeto; nada mais na conta dele muda. Quando o administrador ativou os convites, o usuário recebe
um convite e se torna membro depois de aceitá-lo.

## Email ou nome de usuário {#teamMember-inviteType}

Exibido quando os usuários podem ser convidados tanto por e-mail quanto por nome de usuário.

Como você escolhe a pessoa: **Email** para convidar um endereço de e-mail, **Username** para
escolher um usuário que já tem uma conta neste servidor Semaphore.

## Email {#teamMember-email}

Exibido quando a pessoa é convidada por e-mail.

O endereço de e-mail da pessoa a convidar. Ela se torna membro do projeto depois de aceitar o
convite.

## Usuário {#teamMember-userId}

Exibido quando a pessoa é escolhida pelo nome de usuário.

O usuário a adicionar, listado pelo nome de usuário e pelo nome completo. O campo é obrigatório. A
lista contém todos os usuários do servidor Semaphore que ainda não são membros do projeto: uma
pessoa sem conta precisa primeiro receber uma de um administrador.

## Função {#teamMember-role}

O que o membro pode fazer no projeto. O campo é obrigatório. As funções incorporadas:

- **Owner**: tudo, inclusive as configurações do projeto, a equipe dele e a exclusão dele.
- **Manager**: executa tarefas e workflows, aprova workflows, e cria, altera e exclui modelos de
  tarefa, inventários, chaves, repositórios, grupos de variáveis, agendas, integrações e workflows.
  Um manager não pode alterar as configurações do projeto nem a equipe dele.
- **Task Runner**: vê o projeto, executa tarefas e workflows e aprova workflows. Um task runner não
  pode alterar nada no projeto.
- **Guest**: vê o projeto, as tarefas dele e os logs delas, e não altera nada.

A lista também contém as funções criadas para este projeto (**Funções do projeto**) e as funções
compartilhadas por todos os projetos (**Shared roles**). Uma função assim concede as permissões
definidas para ela na página Roles. Um convite só pode levar uma função incorporada.

Um membro que pode gerenciar a equipe pode dar qualquer função, **Owner** inclusive. A caixa de
diálogo dá uma função; adicione outras depois na coluna **Função** da lista da equipe.
