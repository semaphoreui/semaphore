---
form: src/components/EditTemplatePermissionForm.vue
---

# Template permissions

Concede a uma função permissões extras sobre este modelo específico. Um membro que tem a função
as recebe além das permissões que a função tem no projeto inteiro: uma concessão aqui apenas
acrescenta, nunca retira uma permissão. Um membro com várias funções recebe todas as permissões
concedidas a qualquer uma delas.

## Função {#templatePerm-roleSlug}

A função que recebe as permissões. A lista mostra as funções deste projeto, as funções
compartilhadas e as funções incorporadas. O campo é obrigatório, e uma função pode ter apenas uma
concessão em um modelo: o servidor recusa uma segunda.

A função de uma concessão salva não pode ser alterada. O servidor mantém a função e salva apenas
as permissões, seja qual for a sua escolha aqui. Para dar as permissões a outra função, remova
esta concessão da lista e adicione uma nova.

## Permissions {#templatePerm-permissionsTitle}

O que um membro que tem a função pode fazer com este modelo:

- **Can see this template**: o modelo aparece na lista de modelos, com as tarefas dele e os logs
  delas.
- **Can run this template**: iniciar tarefas a partir do modelo, e parar, confirmar ou rejeitar
  as tarefas dele.
- **Can edit this template**: alterar e excluir o modelo, parar todas as tarefas dele e gerenciar
  estas permissões.

Executar ou editar inclui ver o modelo: quando você marca qualquer um deles, **Can see this
template** também é marcado e não pode ser desmarcado. Deixe todas as caixas desmarcadas para
salvar uma concessão que não dá nada.
