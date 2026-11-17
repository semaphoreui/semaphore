---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Concede a uma função permissões extras sobre este workflow específico. Um membro que tem a função
as recebe além das permissões que a função tem no projeto inteiro: uma concessão aqui apenas
acrescenta, nunca retira uma permissão. Um membro com várias funções recebe todas as permissões
concedidas a qualquer uma delas.

## Função {#templatePerm-roleSlug}

A função que recebe as permissões. A lista mostra as funções deste projeto, as funções
compartilhadas e as funções incorporadas. O campo é obrigatório, e uma função pode ter apenas uma
concessão em um workflow: o servidor recusa uma segunda.

A função de uma concessão salva não pode ser alterada. O servidor mantém a função e salva apenas
as permissões, seja qual for a sua escolha aqui. Para dar as permissões a outra função, remova
esta concessão da lista e adicione uma nova.

## Permissions {#templatePerm-permissionsTitle}

O que um membro que tem a função pode fazer com este workflow:

- **Can see this workflow**: o workflow aparece na lista de workflows, com as execuções dele.
- **Can run this workflow**: iniciar execuções do workflow.
- **Can edit this workflow**: alterar e excluir o workflow e gerenciar estas permissões.
- **Can approve this workflow**: aprovar ou rejeitar as etapas de aprovação das execuções dele.

Executar, editar ou aprovar inclui ver o workflow: quando você marca qualquer um deles, **Can
see this workflow** também é marcado e não pode ser desmarcado. Deixe todas as caixas desmarcadas
para salvar uma concessão que não dá nada.
