---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Equipe

Os membros do projeto e as funções que têm. Uma função determina o que um membro pode fazer no
projeto: executar tarefas, alterar modelos e outros recursos, gerir a equipa. Um membro pode ter
várias funções e recebe todas as permissões que qualquer uma delas concede.

- **Novo Membro da Equipe** adiciona um utilizador ao projeto com uma função. O botão é visível
  para os membros que podem gerir a equipa.
- A coluna **Função** altera as funções de um membro de imediato, sem guardar. Um membro que não
  pode gerir a equipa vê as funções, mas não as pode alterar.
- Uma função com um cadeado vem de um grupo do utilizador no LDAP ou OIDC que o administrador
  associou a essa função. Não pode ser retirada aqui: desaparece quando o utilizador sai do grupo
  ou a associação é removida.
- O projeto precisa de pelo menos um Owner. O Semaphore recusa remover o último Owner ou retirar-lhe
  a função Owner, a menos que seja um administrador do Semaphore a fazê-lo.
- **Sair do Projeto** retira-o do projeto. O botão só aparece se o administrador permitir que os
  membros saiam dos projetos; o último Owner também não pode sair.
