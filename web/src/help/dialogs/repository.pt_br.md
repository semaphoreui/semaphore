---
form: src/components/RepositoryForm.vue
---

# Editar Repository

Um repositório é de onde vêm os playbooks, o código Terraform ou os scripts de um modelo de
tarefa: um repositório Git ou um diretório na máquina que executa as tarefas. Antes de cada
tarefa, o Semaphore clona o repositório Git, ou atualiza a cópia que clonou antes, faz o checkout
do ramo e executa a tarefa nessa cópia.

## Nome {#repository-name}

O nome com que o repositório aparece na página Repositórios e no seletor de repositório de um
modelo de tarefa e de um inventário. O campo é obrigatório.

## URL ou caminho {#repository-gitUrl}

Onde está o código. O campo é obrigatório. Os chips abaixo dele mostram o tipo de endereço que
você digitou, e um clique em um chip reescreve o endereço para esse tipo.

- **ssh**: `git@github.com:org/repo.git` ou `ssh://git@github.com/org/repo.git`. O Git faz login
  com o usuário do endereço e a **Chave SSH** escolhida em **Chave de Acesso**.
- **https** e **http**: `https://github.com/org/repo.git`. Um repositório privado precisa de uma
  chave **Login com senha**; um público precisa de **Nenhum**. Evite `http`: o login e a senha
  trafegam então sem criptografia.
- **git**: `git://host/org/repo.git`, o protocolo Git somente leitura, que não tem login.
- **file**: `file:///srv/git/repo.git`, um repositório Git no disco da máquina que executa a
  tarefa, clonado como um remoto.
- **caminho absoluto**: um caminho absoluto como `/srv/playbooks` ou `D:\playbooks`. O Semaphore
  não o clona: a tarefa é executada diretamente nesse diretório, que precisa existir na máquina
  que executa a tarefa (o servidor Semaphore ou um executor). O ramo não se aplica a ele.

Alterar o endereço remove as cópias que o servidor Semaphore clonou a partir do antigo.

## Ramo {#repository-gitBranch}

O ramo que o Semaphore clona e mantém atualizado, por exemplo `main`. O campo é obrigatório, exceto
para um caminho absoluto, para o qual fica bloqueado.

Um modelo de tarefa pode indicar um ramo próprio, que tem precedência sobre este, e um modelo pode
permitir que as suas tarefas escolham um ramo ao serem iniciadas. Uma tag também funciona, mas uma
tag não pode receber pull, então o Semaphore clona o repositório de novo antes de cada tarefa.

## Single checkout for all templates {#repository-monorepo}

Desativado por padrão: cada modelo de tarefa recebe a sua própria cópia do repositório. Marque a
caixa para clonar o repositório uma vez e compartilhar essa cópia entre todos os modelos que o
usam, para um repositório grande demais para manter uma cópia por modelo.

A cópia compartilhada permanece no ramo deste repositório: os modelos que a usam não podem ter um
ramo próprio, e as tarefas deles não podem escolher outro ramo ou commit. Bloqueado para um caminho
absoluto.

## Chave de Acesso {#repository-sshKeyId}

A chave do Armazenamento de Chaves com a qual o Git faz login no repositório. O campo é
obrigatório. Ele oferece chaves de três tipos, e a escolha depende da URL:

- **Chave SSH** para uma URL `ssh`. O Semaphore carrega a chave privada em um agente SSH durante
  o clone ou a atualização. O usuário é o da URL, como `git` em `git@github.com:org/repo.git`.
- **Login com senha** para uma URL `https`: o login e a senha, ou o token de acesso do servidor
  Git. Com um login vazio, a senha é enviada sozinha, como nome de usuário, que é o que alguns
  servidores esperam de um token. As mesmas credenciais são passadas ao `ansible-galaxy` para
  roles e coleções no mesmo servidor.
- **Nenhum** para um repositório público, uma URL `git` ou `file`, ou um caminho absoluto.

Uma chave que não corresponde à URL não fornece credenciais ao Git, e um repositório privado então
falha ao ser clonado.
