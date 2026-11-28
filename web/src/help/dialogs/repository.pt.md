---
form: src/components/RepositoryForm.vue
---

# Editar Repository

Um repositório é a origem dos playbooks, do código Terraform ou dos scripts de um modelo de
tarefa: um repositório Git ou um diretório na máquina que executa as tarefas. Antes de cada
tarefa, o Semaphore clona o repositório Git, ou atualiza a cópia que clonou anteriormente, faz
checkout do ramo e executa a tarefa nessa cópia.

## Nome {#repository-name}

O nome com que o repositório aparece na página Repositórios e no seletor de repositório de um
modelo de tarefa e de um inventário. O campo é obrigatório.

## URL ou caminho {#repository-gitUrl}

Onde está o código. O campo é obrigatório. Os chips por baixo dele indicam o tipo de endereço
que escreveu, e um clique num chip reescreve o endereço nesse tipo.

- **ssh**: `git@github.com:org/repo.git` ou `ssh://git@github.com/org/repo.git`. O Git inicia
  sessão com o utilizador do endereço e a **Chave SSH** escolhida em **Chave de Acesso**.
- **https** e **http**: `https://github.com/org/repo.git`. Um repositório privado precisa de uma
  chave **Login com senha**; um público precisa de **Nenhum**. Evite `http`: o nome de
  utilizador e a palavra-passe circulam então sem encriptação.
- **git**: `git://host/org/repo.git`, o protocolo Git só de leitura, que não tem início de
  sessão.
- **file**: `file:///srv/git/repo.git`, um repositório Git no disco da máquina que executa a
  tarefa, clonado como um repositório remoto.
- **caminho absoluto**: um caminho absoluto como `/srv/playbooks` ou `D:\playbooks`. O Semaphore
  não o clona: a tarefa é executada diretamente nesse diretório, que tem de existir na máquina
  que executa a tarefa (o servidor Semaphore ou um executor). O ramo não se aplica.

Alterar o endereço remove as cópias que o servidor Semaphore clonou a partir do endereço antigo.

## Ramo {#repository-gitBranch}

O ramo que o Semaphore clona e mantém atualizado, por exemplo `main`. O campo é obrigatório,
exceto para um caminho absoluto, em que fica bloqueado.

Um modelo de tarefa pode indicar um ramo próprio, que tem prioridade sobre este, e um modelo pode
permitir que as suas tarefas escolham um ramo quando são iniciadas. Uma tag também funciona, mas
não é possível fazer pull de uma tag, por isso o Semaphore volta a clonar o repositório antes de
cada tarefa.

## Single checkout for all templates {#repository-monorepo}

Desativado por predefinição: cada modelo de tarefa recebe a sua própria cópia do repositório.
Assinale a caixa para clonar o repositório uma única vez e partilhar essa cópia entre todos os
modelos que o utilizam, no caso de um repositório demasiado grande para manter uma cópia por
modelo.

A cópia partilhada fica no ramo deste repositório: os modelos que a utilizam não podem ter um
ramo próprio, e as suas tarefas não podem escolher outro ramo ou commit. Bloqueado para um
caminho absoluto.

## Chave de Acesso {#repository-sshKeyId}

A chave do Armazenamento de Chaves com que o Git inicia sessão no repositório. O campo é
obrigatório. Propõe chaves de três tipos, e a escolha depende do URL:

- **Chave SSH** para um URL `ssh`. O Semaphore carrega a chave privada num agente SSH durante o
  clone ou a atualização. O utilizador é o do URL, como `git` em
  `git@github.com:org/repo.git`.
- **Login com senha** para um URL `https`: o nome de utilizador e a palavra-passe, ou o token de
  acesso do servidor Git. Com um nome de utilizador vazio, a palavra-passe é enviada sozinha,
  como nome de utilizador, que é o que alguns servidores esperam de um token. As mesmas
  credenciais são fornecidas ao `ansible-galaxy` para as roles e coleções no mesmo servidor.
- **Nenhum** para um repositório público, um URL `git` ou `file`, ou um caminho absoluto.

Uma chave que não corresponda ao URL não dá credenciais ao Git, e o clone de um repositório
privado falha nesse caso.
