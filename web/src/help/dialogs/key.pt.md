---
form: src/components/KeyForm.vue
---

# Editar Key

Uma chave do Armazenamento de Chaves guarda uma credencial que o projeto utiliza sem a mostrar: a
chave SSH ou a palavra-passe com que o Ansible inicia sessão nos hosts, as credenciais de um
repositório Git, a palavra-passe de um cofre do Ansible, o segredo de uma integração. Os
inventários, os repositórios, os modelos de tarefa e as integrações fazem referência a uma chave
pelo seu nome. Depois de guardado, o segredo nunca volta a ser enviado para o navegador.

## Nome da Chave {#key-name}

O nome com que a chave aparece no Armazenamento de Chaves e nos seletores de chave dos
inventários, repositórios, modelos de tarefa e integrações. O campo é obrigatório. Escolha um
nome que diga o que a chave abre, por exemplo `Deploy key of the web servers`.

## Origem {#key-source}

Onde é guardado o segredo da chave. Os separadores:

- **Local**: o segredo é escrito nesta caixa de diálogo e guardado na base de dados do Semaphore,
  cifrado quando o administrador definiu uma chave de cifragem para as chaves de acesso. É a
  opção predefinida.
- **Storage**: o segredo é guardado num armazenamento de segredos externo do projeto, escolhido
  em **Storage** abaixo. O separador só existe na edição Pro. Os próprios armazenamentos são
  configurados no separador **Storages** do Armazenamento de Chaves.
- **Env**: o segredo é lido de uma variável de ambiente do processo do servidor Semaphore.
  Escreva o nome da variável no campo abaixo dos separadores.
- **File**: o segredo é lido de um ficheiro no servidor Semaphore. Escreva o caminho absoluto do
  ficheiro no campo abaixo dos separadores. O ficheiro tem de estar dentro do diretório de
  segredos do servidor (`dirs.secrets` na sua configuração, `/tmp/semaphore` por predefinição);
  um caminho fora dele, ou que contenha `..`, falha quando a chave é utilizada. Uma quebra de
  linha no fim do ficheiro é descartada.

Com **Env** e **File**, o Semaphore nunca escreve o segredo: lê o valor sempre que uma tarefa
utiliza a chave, pelo que o segredo é alterado no sítio onde se encontra. O valor tem de
corresponder ao **Tipo** da chave. Uma chave **Texto** recebe o valor tal como está. Uma chave
**Chave SSH** ou **Login com senha** espera JSON com os campos desse tipo, por exemplo:

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

Quando edita uma chave, os separadores ficam bloqueados até que **Substituir** seja assinalado.
Uma chave sincronizada a partir de um armazenamento de segredos mantém a sua origem, e o seu nome
e tipo não podem ser alterados.

## Storage {#key-sourceStorageId}

Apresentado quando o separador **Storage** está escolhido.

O armazenamento de segredos do projeto que guarda o segredo: HashiCorp Vault, OpenBao,
Devolutions Server, AWS Secrets Manager ou Azure Key Vault, tal como configurado no separador
**Storages** do Armazenamento de Chaves.

- Um armazenamento que não é só de leitura recebe o segredo que escreve nesta caixa de diálogo:
  o Semaphore escreve-o lá quando guarda. Eliminar a chave elimina também o segredo do
  armazenamento. Uma chave guardada num armazenamento destes não pode ser movida mais tarde para
  outro armazenamento.
- Um armazenamento só de leitura é apenas lido. A caixa de diálogo não mostra então campos para
  o segredo, e o Semaphore lê-o do armazenamento sempre que uma tarefa utiliza a chave.

## Source Key {#key-sourceStorageKey}

Apresentado quando está selecionado um **Storage**.

Onde se encontra o segredo nesse armazenamento, por exemplo o caminho do segredo no HashiCorp
Vault.

## Tipo {#key-type}

O tipo de credencial que a chave guarda. O campo é obrigatório, e os campos abaixo dele mudam
consoante o tipo.

- **Chave SSH**: uma chave privada, com um nome de utilizador e uma frase-passe opcionais. O
  Ansible liga-se aos hosts com ela, e o Git clona com ela um repositório com um URL `ssh`.
- **Login com senha**: um nome de utilizador e uma palavra-passe. O Ansible liga-se aos hosts com
  eles ou torna-se outro utilizador com eles (**Credenciais Sudo** de um inventário); o Git clona
  com eles um repositório `https`; a palavra-passe abre um cofre do Ansible de um modelo de
  tarefa.
- **Texto**: um único valor secreto sem utilizador: um token, uma chave de API, uma palavra-passe.
  Utilize-o para o segredo de uma integração. Não pode servir de chave SSH, de credenciais sudo,
  de palavra-passe de um cofre nem de chave de acesso de um repositório.
- **Nenhum**: nenhuma credencial. Utilize-o para um repositório que não precisa de início de
  sessão (um repositório `https` público, um URL `file` ou um caminho local) e para um inventário
  cujos hosts não precisam de credenciais ou definem eles próprios a ligação.

## Nome de usuário {#key-loginPasswordLogin}

Apresentado quando o tipo é **Login com senha**. Opcional.

O nome de utilizador da credencial.

- O Ansible liga-se aos hosts como este utilizador (`--user`); como **Credenciais Sudo** de um
  inventário, é o utilizador em que o Ansible se torna (`--become-user`). Quando o campo está
  vazio, o Ansible utiliza o utilizador indicado pelo inventário ou pela sua própria configuração.
- Para um repositório `https`, o Git inicia sessão com este utilizador e a palavra-passe. Deixe-o
  vazio para um token que o servidor Git aceita sozinho: a palavra-passe é então enviada como
  nome de utilizador.

## Senha {#key-loginPasswordPassword}

Apresentado quando o tipo é **Login com senha**. Obrigatório.

A palavra-passe da credencial. O Semaphore responde com ela aos pedidos de palavra-passe do
Ansible: a palavra-passe SSH (`--ask-pass`) para as credenciais do utilizador, a palavra-passe de
`become` (`--ask-become-pass`) para as credenciais sudo e a palavra-passe do cofre para um cofre
do Ansible. Para um repositório `https`, é a palavra-passe ou o token de acesso do servidor Git.

## Valor {#key-stringValue}

Apresentado quando o tipo é **Texto**. Obrigatório.

O próprio valor secreto, por exemplo o token com que uma integração verifica os pedidos que
recebe. É guardado tal como é escrito.

## Nome de usuário {#key-sshLogin}

Apresentado quando o tipo é **Chave SSH**. Opcional.

O utilizador com que o Ansible se liga aos hosts (`--user`). Quando o campo está vazio, o Ansible
utiliza o utilizador indicado pelo inventário ou pela sua própria configuração.

Para um repositório Git, coloque antes o utilizador no URL do repositório, como em
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Apresentado quando o tipo é **Chave SSH**. Opcional.

A frase-passe com que a chave privada está cifrada. Deixe o campo vazio para uma chave sem
frase-passe. Uma frase-passe errada faz falhar, no momento em que a chave é carregada, todas as
tarefas que utilizam a chave. Uma chave gerada pelo Semaphore não tem frase-passe: o que este
campo contiver é então descartado.

## Generate SSH Key {#key-generateSshKey}

Apresentado quando o tipo é **Chave SSH**.

O Semaphore cria um novo par de chaves quando guarda: uma chave privada RSA de 2048 bits, que
fica na chave, e a respetiva chave pública, que é mostrada uma vez numa caixa de diálogo depois
de guardar e, a seguir, em **Public key** sempre que edita a chave. Adicione a chave pública a
`~/.ssh/authorized_keys` nos hosts, ou como chave de implementação do repositório Git.

O campo da chave privada fica bloqueado enquanto a caixa estiver assinalada. Numa chave
existente, assinale primeiro **Substituir**: o par de chaves antigo é substituído. Não é possível
gerar uma chave cujo segredo é lido de **Env**, de **File** ou de um armazenamento só de leitura,
pois o Semaphore não tem onde o escrever.

## Chave Privada {#key-sshPrivateKey}

Apresentado quando o tipo é **Chave SSH**. Obrigatório, exceto quando **Generate SSH Key** está
assinalado.

A chave privada, no formato OpenSSH ou PEM, incluindo as linhas `BEGIN` e `END`. Para uma tarefa,
o Semaphore carrega-a num agente SSH próprio, que o Ansible e o Git utilizam durante a tarefa.

## Public key {#key-publicKey}

Apresentado quando edita uma chave SSH gerada pelo Semaphore, enquanto **Substituir** não está
assinalado.

A metade pública do par de chaves gerado. **Show** expande-a, o botão de cópia coloca-a na área
de transferência. Adicione-a a `~/.ssh/authorized_keys` nos hosts, ou como chave de implementação
do repositório Git.

É a chave pública do último par gerado pelo Semaphore. Depois de substituir manualmente a chave
privada com **Substituir**, a chave aqui apresentada deixa de lhe corresponder.

## Substituir {#key-overrideSecret}

Apresentado quando edita uma chave existente.

O Semaphore nunca volta a mostrar um segredo guardado, por isso os campos do segredo estão vazios
e bloqueados. Assinale a caixa para substituir o segredo: a origem, o tipo e os campos do segredo
ficam disponíveis, e tudo o que contêm é guardado de novo, pelo que deve introduzir o segredo
completo outra vez, não só a parte que muda. Sem ela, guardar altera apenas o nome da chave.

Guardar uma chave remove também as cópias que o servidor Semaphore mantém dos repositórios que a
utilizam: a tarefa seguinte volta a cloná-los.
