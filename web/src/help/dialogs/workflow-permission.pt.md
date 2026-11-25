---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Concede a uma função permissões adicionais sobre este workflow em particular. Um membro que tenha
a função recebe-as além das permissões que a função tem em todo o projeto: uma concessão aqui
apenas acrescenta, nunca retira uma permissão. Um membro com várias funções recebe todas as
permissões concedidas a qualquer uma delas.

## Função {#templatePerm-roleSlug}

A função que recebe as permissões. A lista apresenta as funções deste projeto, as funções
partilhadas e as funções incorporadas. O campo é obrigatório, e uma função só pode ter uma
concessão sobre um workflow: o servidor recusa uma segunda.

A função de uma concessão guardada não pode ser alterada. O servidor mantém a função e guarda
apenas as permissões, seja qual for a escolha feita aqui. Para dar as permissões a outra função,
remova esta concessão da lista e adicione uma nova.

## Permissions {#templatePerm-permissionsTitle}

O que um membro que tenha a função pode fazer com este workflow:

- **Can see this workflow**: o workflow aparece na lista de workflows, com as suas execuções.
- **Can run this workflow**: iniciar execuções do workflow.
- **Can edit this workflow**: alterar e eliminar o workflow, e gerir estas permissões.
- **Can approve this workflow**: aprovar ou rejeitar os passos de aprovação das suas execuções.

Executar, editar ou aprovar inclui ver o workflow: quando assinala qualquer uma delas, **Can see
this workflow** fica também assinalada e não pode ser desmarcada. Deixe todas as caixas por
assinalar para guardar uma concessão que não dá nada.
