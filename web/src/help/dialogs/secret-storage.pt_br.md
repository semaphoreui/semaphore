---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Editar Secret Storage

Um armazenamento de segredos é um gerenciador de segredos externo no qual o Semaphore guarda
segredos e do qual os lê: HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets Manager ou
Azure Key Vault. Um grupo de variáveis criado com um armazenamento guarda os segredos dele lá, e um
armazenamento com a sincronização ativada importa os seus segredos para o Armazenamento de Chaves
do projeto. O tipo do armazenamento é escolhido no menu **New Storage** e não pode ser alterado
nesta caixa de diálogo; os campos abaixo dependem dele.

## Nome {#secretStorage-name}

O nome com que o armazenamento aparece na aba **Storages** do Armazenamento de Chaves e no seletor
**Secret storage** de um novo grupo de variáveis. O campo é obrigatório.

## Server URL {#secretStorage-vaultURL}

Exibido quando o armazenamento é um HashiCorp Vault, um OpenBao ou um Devolutions Server.

O endereço do servidor, com o esquema, por exemplo `https://vault.example.com:8200`. O campo é
obrigatório.

## Mount {#secretStorage-vaultMount}

Exibido quando o armazenamento é um HashiCorp Vault ou um OpenBao. Opcional.

O caminho de montagem do secrets engine que guarda os segredos. Deixe o campo vazio para usar
`secret`, o padrão.

## Namespace {#secretStorage-vaultNamespace}

Exibido quando o armazenamento é um HashiCorp Vault ou um OpenBao. Opcional.

O namespace em que os segredos ficam. Os namespaces existem no Vault Enterprise e no HCP Vault
Dedicated, e no OpenBao a partir da versão 2.3. Deixe o campo vazio para um servidor sem
namespaces.

## Origem do segredo {#secretSource-options}

Exibido para todos os tipos de armazenamento, acima do segredo com que o Semaphore se autentica: o
**Token** de um Vault ou de um OpenBao, o app secret de um Devolutions Server, a secret key da AWS,
o client secret do Azure.

De onde o Semaphore obtém esse segredo:

- **Store in DB**: você digita o segredo nesta caixa de diálogo; o Semaphore o criptografa e o
  guarda no seu banco de dados.
- **From ENV**: você digita o nome de uma variável de ambiente do servidor Semaphore; o segredo é
  lido dela toda vez que o armazenamento é usado.
- **From File**: você digita o caminho absoluto de um arquivo no servidor Semaphore; o segredo é
  lido dele toda vez, sem a quebra de linha final. O arquivo precisa estar dentro do diretório de
  segredos do servidor (`dirs.secrets` na configuração, `/tmp/semaphore` por padrão), e um caminho
  com `..` é recusado.

Trocar a origem limpa o campo abaixo dela.

## Token {#secretStorage-vaultToken}

Exibido quando o armazenamento é um HashiCorp Vault ou um OpenBao e a origem do segredo é
**Store in DB**.

O token com que o Semaphore se autentica no servidor. Ele precisa ter permissão para ler os
segredos e, a menos que o armazenamento seja **Read only**, para gravá-los e excluí-los.

O campo é obrigatório para um novo armazenamento. Quando você abre um armazenamento existente, ele
fica vazio: deixe-o vazio para manter o token armazenado, digite um novo para substituí-lo.

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

Exibido quando o armazenamento é um HashiCorp Vault ou um OpenBao e a origem do segredo é
**From ENV** ou **From File**.

O nome da variável de ambiente, ou o caminho do arquivo, que guarda o token. O campo é obrigatório
para um novo armazenamento.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Exibido quando o armazenamento é um Devolutions Server.

O Semaphore se conecta ao servidor sem verificar o certificado TLS dele. Qualquer pessoa entre o
Semaphore e o servidor pode então ler os segredos. Deixe a opção desativada e, em vez disso, dê ao
servidor um certificado em que o Semaphore confie.

## Vault ID {#secretStorage-dvlsKey}

Exibido quando o armazenamento é um Devolutions Server.

O ID do cofre do Devolutions Server que guarda os segredos. O campo é obrigatório para um novo
armazenamento.

## App Key {#secretStorage-dvlsAppKey}

Exibido quando o armazenamento é um Devolutions Server.

A chave de aplicativo da identidade de aplicativo com que o Semaphore se autentica, aquela cujo
segredo vai em **Secret**. O campo é obrigatório para um novo armazenamento.

## Secret {#secretStorage-dvlsSecret}

Exibido quando o armazenamento é um Devolutions Server e a origem do segredo é **Store in DB**.

O app secret da identidade de aplicativo de **App Key**. O campo é obrigatório para um novo
armazenamento. Quando você abre um armazenamento existente, ele fica vazio: deixe-o vazio para
manter o segredo armazenado, digite um novo para substituí-lo.

## Env var name / Path to the file {#secretStorage-dvlsEnv}

Exibido quando o armazenamento é um Devolutions Server e a origem do segredo é **From ENV** ou
**From File**.

O nome da variável de ambiente, ou o caminho do arquivo, que guarda o app secret. O campo é
obrigatório para um novo armazenamento.

## Region {#secretStorage-awsRegion}

Exibido quando o armazenamento é um AWS Secrets Manager.

A região da AWS em que os segredos estão, por exemplo `us-east-1`. O campo é obrigatório.

## Endpoint URL {#secretStorage-awsEndpointURL}

Exibido quando o armazenamento é um AWS Secrets Manager. Opcional.

O endereço da API do Secrets Manager a chamar no lugar do endpoint padrão da região, por exemplo
um endpoint de VPC. Deixe o campo vazio para usar o endpoint padrão da AWS.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Exibido quando o armazenamento é um AWS Secrets Manager.

Quando ativado, o Semaphore não armazena nenhuma chave de acesso e obtém as credenciais da AWS do
ambiente em que o servidor é executado, como a IAM role da instância dele. Os campos da chave de
acesso abaixo ficam então desativados.

Ativá-lo remove a chave de acesso armazenada quando você salva. Desativá-lo novamente exige uma
nova **Secret Access Key**.

## Access Key ID {#secretStorage-awsAccessKeyId}

Exibido quando o armazenamento é um AWS Secrets Manager. Desativado quando
**Use IAM Role / Instance Profile** está ativado.

O ID da chave de acesso com que o Semaphore se autentica, por exemplo `AKIAIOSFODNN7EXAMPLE`. O
campo é obrigatório quando a IAM role está desativada. O ID não é um segredo: ele é armazenado e
exibido como é.

## Secret Access Key {#secretStorage-awsSecretKey}

Exibido quando o armazenamento é um AWS Secrets Manager e a origem do segredo é **Store in DB**.
Desativado quando **Use IAM Role / Instance Profile** está ativado.

O segredo da chave de acesso. O campo é obrigatório para um novo armazenamento e depois que a IAM
role é desativada. Caso contrário, ele fica vazio quando você abre o armazenamento: deixe-o vazio
para manter a chave armazenada, digite uma nova para substituí-la.

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

Exibido quando o armazenamento é um AWS Secrets Manager e a origem do segredo é **From ENV** ou
**From File**. Desativado quando **Use IAM Role / Instance Profile** está ativado.

O nome da variável de ambiente, ou o caminho do arquivo, que guarda a secret access key. O campo é
obrigatório nos mesmos casos que **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Exibido quando o armazenamento é um Azure Key Vault.

O endereço do key vault, por exemplo `https://my-vault.vault.azure.net`. O campo é obrigatório.

## Tenant ID {#secretStorage-azureTenantId}

Exibido quando o armazenamento é um Azure Key Vault.

O ID do tenant do Microsoft Entra do aplicativo como o qual o Semaphore se autentica. O campo é
obrigatório.

## Client ID {#secretStorage-azureClientId}

Exibido quando o armazenamento é um Azure Key Vault.

O ID do aplicativo (cliente) desse aplicativo. Ele precisa ter acesso aos segredos do key vault. O
campo é obrigatório.

## Client Secret {#secretStorage-azureClientSecret}

Exibido quando o armazenamento é um Azure Key Vault e a origem do segredo é **Store in DB**.

O client secret do aplicativo. O campo é obrigatório para um novo armazenamento. Quando você abre
um armazenamento existente, ele fica vazio: deixe-o vazio para manter o segredo armazenado, digite
um novo para substituí-lo.

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

Exibido quando o armazenamento é um Azure Key Vault e a origem do segredo é **From ENV** ou
**From File**.

O nome da variável de ambiente, ou o caminho do arquivo, que guarda o client secret. O campo é
obrigatório para um novo armazenamento.

## Read only {#secretStorage-readonly}

Quando ativado, o Semaphore apenas lê do armazenamento. Um segredo de um grupo de variáveis não
pode ser criado nem alterado nele: salvá-lo falha. Excluir um segredo, uma chave ou um grupo de
variáveis no Semaphore deixa o segredo correspondente no armazenamento.

Quando desativado, o Semaphore grava nele os segredos dos grupos de variáveis que usam o
armazenamento e os exclui dele junto com os grupos.

## Sync keys enabled {#secretStorage-syncEnabled}

Ative para importar segredos do armazenamento para o Armazenamento de Chaves do projeto, a partir
dos caminhos definidos em **Sync paths**. A importação é executada quando você clica no botão de
sincronização do armazenamento na aba **Storages**, e por conta própria no intervalo definido em
**Sync paths**.

Uma chave importada é marcada como sincronizada: o nome e o tipo dela não podem ser alterados, e
excluí-la no Semaphore deixa o segredo no armazenamento. Excluir o armazenamento enquanto a
sincronização está ativada também exclui as chaves que ele importou.

## Sync paths {#secretStorage-syncPaths}

Exibido quando **Sync keys enabled** está ativado.

Abre a caixa de diálogo com os caminhos do armazenamento a partir dos quais importar e o intervalo
da importação automática. O número no botão é o número de caminhos. O botão de sincronização do
armazenamento permanece desativado até que ele tenha pelo menos um caminho.

A caixa de diálogo tem apenas **Fechar**: o que você altera nela é salvo com o armazenamento, pelo
**Salvar** desta caixa de diálogo (**Criar** para um novo armazenamento).
