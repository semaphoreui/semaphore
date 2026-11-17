---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Editar Usuário

Um usuário é uma conta que entra no Semaphore. A aba **Settings** contém o perfil da conta e as
suas opções; **Salvar** as grava. A aba **Security** contém a senha, a autenticação de dois fatores
e as contas de login vinculadas, e a aba **Roles** as funções globais do usuário: ambas agem
imediatamente, sem **Salvar**. O que um usuário pode fazer dentro de um projeto é definido na
página **Equipe** desse projeto, não aqui.

## Nome {#user-name}

O nome completo do usuário, exibido na barra lateral e nas listas de usuários e de membros de
projeto. O campo é obrigatório.

Para um usuário **Externo**, o Semaphore obtém o nome do provedor de identidade no login, quando o
provedor envia um, então uma alteração feita aqui pode não durar além do próximo login.

## Nome de usuário {#user-username}

O login do usuário. O campo é obrigatório. O usuário entra com o nome de usuário ou com o email, e
a senha.

O nome de usuário de um usuário **Externo** não pode ser alterado: o servidor se recusa a salvá-lo.

## Email {#user-email}

O endereço de email do usuário. O campo é obrigatório. O Semaphore envia para ele os alertas de
tarefa do usuário, e um aplicativo autenticador o exibe como o nome da conta. Ele é privado: apenas
o usuário e quem pode gerenciar usuários o veem.

Para um usuário **Externo**, o email vem do provedor de identidade: o campo fica somente leitura, e
o Semaphore atualiza o endereço a cada login.

## Senha {#user-password}

Exibido quando um novo usuário é criado.

A senha com que o usuário entra. O campo é obrigatório, exceto para um usuário **Externo**, que não
tem senha no Semaphore. Para alterar a senha de um usuário existente, use **Change password** na
aba **Security**.

## Enviar alertas {#user-alert}

Envia ao usuário um email quando uma tarefa falha em um projeto em que o usuário pode ver a tarefa.
O projeto precisa permitir alertas, e o administrador do servidor precisa ter ativado os alertas
por email. Um modelo de tarefa pode impedir que as suas tarefas com falha enviem alertas.

A opção não afeta os alertas de um projeto enviados para o Telegram, o Slack e os outros
mensageiros.

## Usuário administrador {#user-admin}

Exibido quando você é administrador.

Um administrador pode fazer tudo no servidor: gerenciar os usuários, as funções globais e as
configurações do servidor, e abrir todos os projetos com todos os direitos.

Você não pode retirar a opção da sua própria conta, e o servidor mantém pelo menos um
administrador: ele se recusa a retirar a opção do último.

## Pro user {#user-pro}

Exibido na edição Pro. Apenas um administrador pode alterar a opção.

Dá ao usuário uma vaga da assinatura Pro. A assinatura tem um número fixo de usuários Pro: quando
ele é atingido, o servidor se recusa a salvar a opção. Sem uma assinatura ativa, a opção não tem
efeito.

## Externo {#user-external}

Exibido quando você é administrador. Só pode ser definido quando o usuário é criado.

Um usuário externo entra por LDAP ou OpenID Connect, não com uma senha do Semaphore: o campo de
senha fica desativado, e um login com senha é recusado. O nome e o email são obtidos do provedor de
identidade a cada login. Deixe a caixa desmarcada para um usuário que entra com senha.

## Change password {#user-changePassword}

Exibido na aba **Security** de um usuário existente que não é **Externo**, quando o servidor permite
entrar com senha.

Abre uma caixa de diálogo que define uma nova senha. Para alterar a sua própria senha, você precisa
informar a atual; para redefinir a senha de outro usuário, não.

## Time-based one-time password {#user-totp}

Exibido na aba **Security** quando o administrador ativou a autenticação de dois fatores no
servidor.

Ativa a autenticação de dois fatores para este usuário. Quando você liga o interruptor, o Semaphore
cria um segredo imediatamente e o mostra como um QR code: escaneie-o com um aplicativo
autenticador. A partir de então, o login pede o código do aplicativo depois da senha.

Quando o servidor permite a recuperação, um código de recuperação é exibido abaixo do QR code. Ele
é exibido uma única vez: copie-o e guarde-o em segurança. Desligar o interruptor remove o segredo
imediatamente.

## Provedor LDAP {#user-ldapProvider}

Exibido na aba **Security** da sua própria conta quando o servidor tem um provedor LDAP ao qual a
sua conta ainda não está vinculada.

O servidor LDAP ao qual vincular a sua conta, quando há vários.

## LDAP username {#user-ldapUsername}

Exibido com **Provedor LDAP**.

O seu nome de usuário no servidor LDAP. A vinculação permite que você entre nesta conta também por
LDAP.

## LDAP password {#user-ldapPassword}

Exibido com **Provedor LDAP**.

A sua senha no servidor LDAP. O Semaphore a usa uma vez, para verificar que a conta LDAP é sua, e
não a armazena.

## Link {#user-ldapLink}

Exibido com **Provedor LDAP**.

Vincula a sua conta à conta LDAP acima. A vinculação é recusada quando as credenciais estão
erradas, quando a conta LDAP está vinculada a outro usuário ou quando a sua conta já está vinculada
a esse provedor. Uma conta vinculada aparece em **Linked accounts**; feche o chip dela para
desvinculá-la.

## Global role {#global-role-select}

Exibido na aba **Roles** de um usuário existente, na edição Pro, quando você pode gerenciar funções
globais.

Uma função global a dar ao usuário, entre as funções que o usuário ainda não tem. Uma função
global concede permissões sobre o servidor inteiro: gerenciar usuários, funções globais ou as
configurações do servidor, ou ler o log de auditoria. As próprias funções globais são criadas na
aba **Funções globais** da página Usuários.

## Assign {#globalRoleAssign-assign}

Exibido com **Global role**.

Dá ao usuário a função selecionada imediatamente. As funções que o usuário tem são listadas
abaixo, cada uma com um botão para retirá-la. **Effective global permissions** mostra o que o
usuário pode fazer no fim das contas, e de onde vem cada permissão.
