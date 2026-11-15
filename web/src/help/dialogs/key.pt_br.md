---
form: src/components/KeyForm.vue
---

# Editar Key

Uma chave do Armazenamento de Chaves guarda uma credencial que o projeto usa sem exibi-la: a chave
SSH ou a senha com que o Ansible faz login nos hosts, as credenciais de um repositório Git, a senha
de um vault do Ansible, o segredo de uma integração. Inventários, repositórios, modelos de tarefa e
integrações se referem a uma chave pelo nome dela. Depois de salvo, o segredo nunca é enviado de
volta ao navegador.

## Nome da Chave {#key-name}

O nome com que a chave aparece no Armazenamento de Chaves e nos seletores de chave de inventários,
repositórios, modelos de tarefa e integrações. O campo é obrigatório. Escolha um nome que diga o
que a chave abre, por exemplo `Deploy key of the web servers`.

## Origem {#key-source}

Onde o segredo da chave é guardado. As abas:

- **Local**: o segredo é digitado nesta caixa de diálogo e armazenado no banco de dados do
  Semaphore, criptografado quando o administrador definiu uma chave de criptografia para as chaves
  de acesso. Este é o padrão.
- **Storage**: o segredo é guardado em um armazenamento de segredos externo do projeto, escolhido
  em **Storage** abaixo. A aba existe apenas na edição Pro. Os próprios armazenamentos são
  configurados na aba **Storages** do Armazenamento de Chaves.
- **Env**: o segredo é lido de uma variável de ambiente do processo do servidor Semaphore. Digite o
  nome da variável no campo abaixo das abas.
- **File**: o segredo é lido de um arquivo no servidor Semaphore. Digite o caminho absoluto do
  arquivo no campo abaixo das abas. O arquivo precisa estar dentro do diretório de segredos do
  servidor (`dirs.secrets` na configuração dele, `/tmp/semaphore` por padrão); um caminho fora
  dele, ou com `..`, falha quando a chave é usada. Uma quebra de linha no final do arquivo é
  descartada.

Com **Env** e **File**, o Semaphore nunca grava o segredo: ele lê o valor toda vez que uma tarefa
usa a chave, então você altera o segredo onde ele fica. O valor precisa ser adequado ao **Tipo** da
chave. Uma chave **Texto** recebe o valor como ele é. Uma chave **Chave SSH** ou **Login com
senha** espera um JSON com os campos desse tipo, por exemplo:

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

Quando você edita uma chave, as abas ficam bloqueadas até que **Substituir** seja marcado. Uma
chave sincronizada a partir de um armazenamento de segredos mantém a sua origem, e o nome e o tipo
dela não podem ser alterados.

## Storage {#key-sourceStorageId}

Exibido quando a aba **Storage** é escolhida.

O armazenamento de segredos do projeto que guarda o segredo: HashiCorp Vault, OpenBao, Devolutions
Server, AWS Secrets Manager ou Azure Key Vault, conforme configurado na aba **Storages** do
Armazenamento de Chaves.

- Um armazenamento que não é somente leitura recebe o segredo que você digita nesta caixa de
  diálogo: o Semaphore o grava lá quando você salva. Excluir a chave também exclui o segredo do
  armazenamento. Uma chave guardada em um armazenamento assim não pode ser movida para outro
  armazenamento depois.
- Um armazenamento somente leitura é apenas lido. A caixa de diálogo então não mostra campos para
  o segredo, e o Semaphore o lê do armazenamento toda vez que uma tarefa usa a chave.

## Source Key {#key-sourceStorageKey}

Exibido quando um **Storage** está selecionado.

Onde o segredo está nesse armazenamento, por exemplo o caminho para o segredo no HashiCorp Vault.

## Tipo {#key-type}

Que tipo de credencial a chave guarda. O campo é obrigatório, e os campos abaixo dele mudam de
acordo com o tipo.

- **Chave SSH**: uma chave privada, com login e frase secreta opcionais. O Ansible se conecta aos
  hosts com ela, e o Git clona com ela um repositório com URL `ssh`.
- **Login com senha**: um login e uma senha. O Ansible se conecta aos hosts com eles ou assume
  outro usuário com eles (**Credenciais Sudo** de um inventário); o Git clona com eles um
  repositório `https`; a senha abre um vault do Ansible de um modelo de tarefa.
- **Texto**: um único valor secreto sem login: um token, uma chave de API, uma senha. Use-o para o
  segredo de uma integração. Não pode ser uma chave SSH, credenciais sudo, a senha de um vault ou a
  chave de acesso de um repositório.
- **Nenhum**: nenhuma credencial. Use-o para um repositório que não precisa de login (um
  repositório `https` público, uma URL `file` ou um caminho local) e para um inventário cujos hosts
  não precisam de credenciais ou definem a conexão por conta própria.

## Nome de usuário {#key-loginPasswordLogin}

Exibido quando o tipo é **Login com senha**. Opcional.

O nome de usuário da credencial.

- O Ansible se conecta aos hosts como este usuário (`--user`); como **Credenciais Sudo** de um
  inventário, é o usuário que o Ansible assume (`--become-user`). Quando o campo está vazio, o
  Ansible usa o usuário indicado pelo inventário ou pela própria configuração dele.
- Para um repositório `https`, o Git faz login com este usuário e a senha. Deixe-o vazio para um
  token que o servidor Git aceita sozinho: a senha é então enviada como nome de usuário.

## Senha {#key-loginPasswordPassword}

Exibido quando o tipo é **Login com senha**. Obrigatório.

A senha da credencial. O Semaphore responde com ela aos prompts de senha do Ansible: a senha SSH
(`--ask-pass`) para credenciais de usuário, a senha do `become` (`--ask-become-pass`) para
credenciais sudo e a senha do vault para um vault do Ansible. Para um repositório `https`, é a
senha ou o token de acesso do servidor Git.

## Valor {#key-stringValue}

Exibido quando o tipo é **Texto**. Obrigatório.

O próprio valor secreto, por exemplo o token com que uma integração confere as requisições que
recebe. Ele é armazenado exatamente como digitado.

## Nome de usuário {#key-sshLogin}

Exibido quando o tipo é **Chave SSH**. Opcional.

O usuário com que o Ansible se conecta aos hosts (`--user`). Quando o campo está vazio, o Ansible
usa o usuário indicado pelo inventário ou pela própria configuração dele.

Para um repositório Git, coloque o usuário na URL do repositório, como em
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Exibido quando o tipo é **Chave SSH**. Opcional.

A frase secreta com que a chave privada está criptografada. Deixe-a vazia para uma chave sem frase
secreta. Uma frase secreta errada faz falhar, no carregamento da chave, toda tarefa que usa a
chave. Uma chave gerada pelo Semaphore não tem frase secreta: o que este campo contém é então
descartado.

## Generate SSH Key {#key-generateSshKey}

Exibido quando o tipo é **Chave SSH**.

O Semaphore cria um novo par de chaves quando você salva: uma chave privada RSA de 2048 bits, que
fica na chave, e a chave pública correspondente, que é exibida uma vez em uma caixa de diálogo
depois de salvar e depois em **Public key** sempre que você edita a chave. Adicione a chave
pública ao `~/.ssh/authorized_keys` nos hosts, ou como deploy key do repositório Git.

O campo da chave privada fica bloqueado enquanto a caixa está marcada. Em uma chave existente,
marque **Substituir** primeiro: o par de chaves antigo é substituído. Uma chave cujo segredo é lido
de **Env**, **File** ou de um armazenamento somente leitura não pode ser gerada, pois o Semaphore
não tem onde gravá-la.

## Chave Privada {#key-sshPrivateKey}

Exibido quando o tipo é **Chave SSH**. Obrigatório, a menos que **Generate SSH Key** esteja
marcado.

A chave privada, no formato OpenSSH ou PEM, incluindo as linhas `BEGIN` e `END`. Para uma tarefa,
o Semaphore a carrega em um agente SSH próprio, que o Ansible e o Git usam durante a tarefa.

## Public key {#key-publicKey}

Exibido quando você edita uma chave SSH que o Semaphore gerou, enquanto **Substituir** não está
marcado.

A metade pública do par de chaves gerado. **Show** a expande, o botão de copiar a coloca na área
de transferência. Adicione-a ao `~/.ssh/authorized_keys` nos hosts, ou como deploy key do
repositório Git.

É a chave pública do último par que o Semaphore gerou. Depois que você substitui a chave privada
manualmente com **Substituir**, a chave exibida aqui não corresponde mais a ela.

## Substituir {#key-overrideSecret}

Exibido quando você edita uma chave existente.

O Semaphore nunca mostra novamente um segredo salvo, então os campos do segredo ficam vazios e
bloqueados. Marque a caixa para substituir o segredo: a origem, o tipo e os campos do segredo se
abrem, e tudo neles é salvo de novo, então informe o segredo inteiro novamente, não apenas a parte
que muda. Sem ela, salvar altera apenas o nome da chave.

Salvar uma chave também remove as cópias que o servidor Semaphore mantém dos repositórios que a
usam: a próxima tarefa os clona novamente.
