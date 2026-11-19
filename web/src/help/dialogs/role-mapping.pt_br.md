---
form: src/components/RoleMappingForm.vue
---

# New mapping

Um mapeamento dá uma função em um projeto a todo usuário que entra por um provedor LDAP ou OpenID
Connect e pertence a um grupo nele. O Semaphore lê os grupos do usuário a cada login pelo provedor
e concede ou retira as funções mapeadas de acordo, de modo que um usuário que sai do grupo perde a
função no próximo login. Uma função mapeada torna o usuário membro do projeto. Um mapeamento não
pode ser editado: exclua-o e adicione um novo.

## Provider {#roleMapping-provider}

O diretório LDAP ou o provedor OpenID Connect de onde vem o grupo. A lista contém os provedores
configurados na configuração do servidor. O campo é obrigatório.

Os grupos só são lidos quando o provedor está configurado para isso: um provedor LDAP precisa de
uma base de busca de grupos (`group_search_dn`), um provedor OpenID Connect indica a claim com os
grupos (`groups_claim`). Um mapeamento de um provedor que depois é removido da configuração
permanece na lista, marcado como **Provider not configured**.

## Group {#roleMapping-group}

O nome do grupo como o provedor o informa. O campo é obrigatório.

- LDAP: o valor do atributo de nome do grupo na entrada do grupo, `cn` por padrão, por exemplo
  `devops`. Um DN completo como `cn=devops,ou=groups,dc=example,dc=com` é reduzido ao seu `cn`.
- OpenID Connect: um dos valores da claim de grupos, como o provedor a envia.

## Project {#roleMapping-projectId}

O projeto em que a função é concedida. O campo é obrigatório. Um administrador vê todos os
projetos; qualquer outro usuário vê apenas os projetos dos quais é membro.

## Função {#roleMapping-role}

A função que os membros do grupo recebem no projeto. O campo é obrigatório e fica disponível
depois que um projeto é selecionado. A lista contém as funções desse projeto, as funções
compartilhadas e as funções incorporadas. Quando você escolhe outro projeto, uma função que não
existe nele é limpa.

Uma função **Owner** dada por um mapeamento não conta quando o Semaphore impede que você retire a
função do último proprietário de um projeto, porque ela pode desaparecer no próximo login. Mantenha
pelo menos um proprietário adicionado manualmente.
