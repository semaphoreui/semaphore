---
form: src/components/RoleMappingForm.vue
---

# New mapping

Um mapeamento atribui uma função num projeto a todos os utilizadores que iniciam sessão através de
um fornecedor LDAP ou OpenID Connect e que pertencem aí a um grupo. O Semaphore lê os grupos do
utilizador em cada início de sessão através do fornecedor e atribui ou retira as funções
mapeadas em conformidade, pelo que um utilizador que sai do grupo perde a função no seu próximo
início de sessão. Uma função mapeada torna o utilizador membro do projeto. Um mapeamento não pode
ser editado: elimine-o e adicione um novo.

## Provider {#roleMapping-provider}

O diretório LDAP ou o fornecedor OpenID Connect de onde provém o grupo. A lista contém os
fornecedores configurados na configuração do servidor. O campo é obrigatório.

Os grupos só são lidos quando o fornecedor está configurado para isso: um fornecedor LDAP precisa
de uma base de pesquisa de grupos (`group_search_dn`), um fornecedor OpenID Connect indica a claim
com os grupos (`groups_claim`). Um mapeamento de um fornecedor que é mais tarde removido da
configuração permanece na lista, marcado como **Provider not configured**.

## Group {#roleMapping-group}

O nome do grupo tal como o fornecedor o indica. O campo é obrigatório.

- LDAP: o valor do atributo de nome do grupo na entrada do grupo, `cn` por omissão, por exemplo
  `devops`. Um DN completo como `cn=devops,ou=groups,dc=example,dc=com` é reduzido ao seu `cn`.
- OpenID Connect: um dos valores da claim de grupos, tal como o fornecedor o envia.

## Project {#roleMapping-projectId}

O projeto em que a função é atribuída. O campo é obrigatório. Um administrador vê todos os
projetos; qualquer outro utilizador vê apenas os projetos de que é membro.

## Função {#roleMapping-role}

A função que os membros do grupo recebem no projeto. O campo é obrigatório e fica disponível
depois de selecionado um projeto. A lista contém as funções desse projeto, as funções partilhadas
e as funções incorporadas. Quando escolhe outro projeto, uma função que aí não existe é limpa.

Uma função **Owner** atribuída por um mapeamento não conta quando o Semaphore o impede de retirar
a função ao último dono de um projeto, porque pode desaparecer no próximo início de sessão.
Mantenha pelo menos um dono adicionado manualmente.
