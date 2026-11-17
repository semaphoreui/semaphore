---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

Uma função global concede permissões sobre toda a instância do Semaphore aos utilizadores a quem
é atribuída; atribui-a na caixa de diálogo de um utilizador. Não dá qualquer acesso dentro dos
projetos: esse vem apenas das funções que um utilizador tem como membro de um projeto. Um
administrador tem todas as permissões globais, sejam quais forem as funções atribuídas. As funções
globais são uma funcionalidade Pro.

## Nome {#globalRole-name}

O nome com que a função aparece no separador Funções globais da página Usuários e na caixa de
diálogo de um utilizador. O campo é obrigatório. Pode mudar o nome de uma função a qualquer
momento.

## Slug {#globalRole-slug}

O identificador da função. O Semaphore preenche-o a partir do nome enquanto escreve o nome, até
que edite o slug manualmente. Aceita de 1 a 100 letras, algarismos, sublinhados e hífenes, por
exemplo `user-admins`.

O slug tem de ser único entre as funções globais; uma função de projeto pode ter o mesmo slug. Não
pode ser alterado depois de a função ser guardada: o campo fica bloqueado quando edita uma função.

## Global permissions {#globalRole-permissionsTitle}

O que um utilizador que tem a função pode fazer em toda a instância. Uma função nova não tem
nenhuma assinalada. As permissões de todas as funções globais de um utilizador somam-se.

- **Can manage users**: criar utilizadores, alterá-los e eliminá-los, e ver os seus dados na
  lista de utilizadores. Esse utilizador não pode tornar um utilizador administrador ou utilizador
  Pro, nem alterar ou eliminar um administrador ou um utilizador que tenha uma permissão global
  que ele próprio não tem.
- **Can manage global roles**: criar, alterar e eliminar as funções globais e as funções de
  projeto partilhadas, e decidir que utilizadores têm que função global. Atribua-a apenas a
  pessoas em quem confia toda a instância.
- **Can manage system settings**: alterar as configurações da instância, as aplicações, os
  executores globais e os mapeamentos de grupos; ver o cluster e as informações sobre o servidor;
  ver e remover as tarefas de todos os projetos; limpar a cache do servidor.
- **Can read the audit log**: ler o registo de auditoria da instância e exportá-lo.
