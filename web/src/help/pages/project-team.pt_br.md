---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Equipe

Os membros do projeto e as funções que eles têm. Uma função define o que um membro pode fazer no
projeto: executar tarefas, alterar modelos e outros recursos, gerenciar a equipe. Um membro pode ter
várias funções e recebe todas as permissões que qualquer uma delas concede.

- **Novo Membro da Equipe** adiciona um usuário ao projeto com uma função. O botão aparece para os
  membros que podem gerenciar a equipe.
- A coluna **Função** altera as funções de um membro na hora, sem salvar. Um membro que não pode
  gerenciar a equipe vê as funções, mas não consegue alterá-las.
- Uma função com cadeado vem de um grupo do usuário no LDAP ou OIDC que o administrador associou a
  essa função. Ela não pode ser removida aqui: some quando o usuário sai do grupo ou a associação é
  removida.
- O projeto precisa de pelo menos um Owner. O Semaphore se recusa a remover o último Owner ou a
  tirar dele a função Owner, a menos que um administrador do Semaphore faça isso.
- **Sair do Projeto** tira você do projeto. O botão só aparece se o administrador permitir que os
  membros saiam dos projetos; o último Owner também não pode sair.
