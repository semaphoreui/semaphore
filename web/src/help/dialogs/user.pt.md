---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Editar Usuário

Um utilizador é uma conta que inicia sessão no Semaphore. O separador **Settings** contém o
perfil da conta e as suas opções; **Salvar** grava-os. O separador **Security** contém a
palavra-passe, a autenticação de dois fatores e as contas de início de sessão associadas, e o
separador **Roles** as funções globais do utilizador: ambos atuam de imediato, sem **Salvar**. O
que um utilizador pode fazer dentro de um projeto é definido na página **Equipe** desse projeto,
não aqui.

## Nome {#user-name}

O nome completo do utilizador, apresentado na barra lateral e nas listas de utilizadores e de
membros de projetos. O campo é obrigatório.

Para um utilizador **Externo**, o Semaphore obtém o nome do fornecedor de identidade ao iniciar
sessão, quando o fornecedor o envia, por isso uma alteração feita aqui pode não durar para além
do próximo início de sessão.

## Nome de usuário {#user-username}

O login do utilizador. O campo é obrigatório. O utilizador inicia sessão com o nome de utilizador
ou com o e-mail, e com a palavra-passe.

O nome de utilizador de um utilizador **Externo** não pode ser alterado: o servidor recusa
guardá-lo.

## Email {#user-email}

O endereço de e-mail do utilizador. O campo é obrigatório. O Semaphore envia-lhe os alertas de
tarefas do utilizador, e uma aplicação de autenticação apresenta-o como nome da conta. É privado:
só o utilizador e quem pode gerir utilizadores o veem.

Para um utilizador **Externo**, o e-mail provém do fornecedor de identidade: o campo é só de
leitura, e o Semaphore atualiza o endereço em cada início de sessão.

## Senha {#user-password}

Apresentado quando é criado um novo utilizador.

A palavra-passe com que o utilizador inicia sessão. O campo é obrigatório, exceto para um
utilizador **Externo**, que não tem palavra-passe no Semaphore. Para alterar a palavra-passe de
um utilizador existente, utilize **Change password** no separador **Security**.

## Enviar alertas {#user-alert}

Envia ao utilizador um e-mail quando uma tarefa falha num projeto em que o utilizador pode ver a
tarefa. O projeto tem de permitir alertas, e o administrador do servidor tem de ter ativado os
alertas por e-mail. Um modelo de tarefa pode impedir que as suas tarefas falhadas enviem alertas.

A opção não afeta os alertas de um projeto enviados para o Telegram, o Slack e os outros serviços
de mensagens.

## Usuário administrador {#user-admin}

Apresentado quando é administrador.

Um administrador pode fazer tudo no servidor: gerir os utilizadores, as funções globais e as
configurações do servidor, e abrir todos os projetos com todos os direitos.

Não pode retirar a opção à sua própria conta, e o servidor mantém pelo menos um administrador:
recusa retirar a opção ao último.

## Pro user {#user-pro}

Apresentado na edição Pro. Só um administrador pode alterar a opção.

Atribui ao utilizador um lugar da subscrição Pro. A subscrição tem um número fixo de utilizadores
Pro: quando é atingido, o servidor recusa guardar a opção. Sem uma subscrição ativa, a opção não
tem efeito.

## Externo {#user-external}

Apresentado quando é administrador. Só pode ser definido quando o utilizador é criado.

Um utilizador externo inicia sessão através de LDAP ou OpenID Connect, não com uma palavra-passe
do Semaphore: o campo da palavra-passe fica desativado, e um início de sessão com palavra-passe é
recusado. O nome e o e-mail são obtidos do fornecedor de identidade em cada início de sessão.
Deixe a caixa vazia para um utilizador que inicia sessão com palavra-passe.

## Change password {#user-changePassword}

Apresentado no separador **Security** de um utilizador existente que não é **Externo**, quando o
servidor permite iniciar sessão com palavra-passe.

Abre uma caixa de diálogo que define uma nova palavra-passe. Para alterar a sua própria
palavra-passe, tem de indicar a atual; para repor a palavra-passe de outro utilizador, não.

## Time-based one-time password {#user-totp}

Apresentado no separador **Security** quando o administrador ativou a autenticação de dois
fatores no servidor.

Ativa a autenticação de dois fatores para este utilizador. Quando liga o interruptor, o Semaphore
cria de imediato um segredo e apresenta-o como código QR: leia-o com uma aplicação de
autenticação. A partir daí, um início de sessão pede o código da aplicação depois da
palavra-passe.

Quando o servidor permite a recuperação, é apresentado um código de recuperação por baixo do
código QR. Só é apresentado uma vez: copie-o e guarde-o em segurança. Desligar o interruptor
remove o segredo de imediato.

## Fornecedor LDAP {#user-ldapProvider}

Apresentado no separador **Security** da sua própria conta quando o servidor tem um fornecedor
LDAP ao qual a sua conta ainda não está associada.

O servidor LDAP ao qual associar a sua conta, quando há vários.

## LDAP username {#user-ldapUsername}

Apresentado com **Fornecedor LDAP**.

O seu nome de utilizador no servidor LDAP. A associação permite-lhe também iniciar sessão nesta
conta através de LDAP.

## LDAP password {#user-ldapPassword}

Apresentado com **Fornecedor LDAP**.

A sua palavra-passe no servidor LDAP. O Semaphore utiliza-a uma vez, para verificar que a conta
LDAP é sua, e não a guarda.

## Link {#user-ldapLink}

Apresentado com **Fornecedor LDAP**.

Associa a sua conta à conta LDAP acima. A associação é recusada quando as credenciais estão
erradas, quando a conta LDAP está associada a outro utilizador, ou quando a sua conta já está
associada a esse fornecedor. Uma conta associada aparece em **Linked accounts**; feche o seu chip
para desfazer a associação.

## Global role {#global-role-select}

Apresentado no separador **Roles** de um utilizador existente, na edição Pro, quando pode gerir
funções globais.

Uma função global a atribuir ao utilizador, entre as funções que o utilizador ainda não tem. Uma
função global concede permissões sobre todo o servidor: gerir utilizadores, funções globais ou as
configurações do servidor, ou ler o registo de auditoria. As próprias funções globais são criadas
no separador **Funções globais** da página Usuários.

## Assign {#globalRoleAssign-assign}

Apresentado com **Global role**.

Atribui de imediato ao utilizador a função selecionada. As funções que o utilizador tem são
listadas abaixo, cada uma com um botão para a retirar. **Effective global permissions** mostra o
que o utilizador pode fazer afinal, e de onde vem cada permissão.
