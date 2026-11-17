---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

Uma função global concede permissões sobre a instância inteira do Semaphore aos usuários aos quais
é atribuída; você a atribui na caixa de diálogo de um usuário. Ela não dá acesso dentro dos
projetos: esse acesso vem apenas das funções que um usuário tem como membro de um projeto. Um
administrador tem todas as permissões globais, sejam quais forem as funções atribuídas. As funções
globais são um recurso Pro.

## Nome {#globalRole-name}

O nome com que a função aparece na aba Funções globais da página Usuários e na caixa de diálogo de
um usuário. O campo é obrigatório. Você pode renomear uma função a qualquer momento.

## Slug {#globalRole-slug}

O identificador da função. O Semaphore o preenche a partir do nome enquanto você digita o nome,
até que você edite o slug por conta própria. Ele aceita de 1 a 100 letras, dígitos, sublinhados e
hifens, por exemplo `user-admins`.

O slug precisa ser único entre as funções globais; uma função de projeto pode ter o mesmo slug. Ele
não pode ser alterado depois que a função é salva: o campo fica bloqueado quando você edita uma
função.

## Global permissions {#globalRole-permissionsTitle}

O que um usuário que tem a função pode fazer na instância inteira. Uma nova função não tem nenhuma
marcada. As permissões de todas as funções globais de um usuário se somam.

- **Can manage users**: criar usuários, alterá-los e excluí-los, e ver os detalhes deles na lista
  de usuários. Esse usuário não pode tornar um usuário administrador ou usuário Pro, e não pode
  alterar nem excluir um administrador ou um usuário que tenha uma permissão global que ele mesmo
  não tem.
- **Can manage global roles**: criar, alterar e excluir as funções globais e as funções de projeto
  compartilhadas, e decidir quais usuários têm qual função global. Conceda-a apenas a pessoas a
  quem você confiaria a instância inteira.
- **Can manage system settings**: alterar as configurações da instância, os aplicativos, os
  executores globais e os mapeamentos de grupos; ver o cluster e as informações sobre o servidor;
  ver e remover as tarefas de todos os projetos; limpar o cache do servidor.
- **Can read the audit log**: ler o log de auditoria da instância e exportá-lo.
