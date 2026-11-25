---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Editar Secret Storage

Um armazenamento de segredos é um gestor de segredos externo no qual o Semaphore guarda segredos
e a partir do qual os lê: HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets Manager ou
Azure Key Vault. Um grupo de variáveis criado com um armazenamento guarda lá os seus segredos, e
um armazenamento com a sincronização ativada importa os seus segredos para o Armazenamento de
Chaves do projeto. O tipo do armazenamento é escolhido no menu **New Storage** e não pode ser
alterado nesta caixa de diálogo; os campos abaixo dependem dele.

## Nome {#secretStorage-name}

O nome com que o armazenamento aparece no separador **Storages** do Armazenamento de Chaves e no
seletor **Secret storage** de um novo grupo de variáveis. O campo é obrigatório.

## Server URL {#secretStorage-vaultURL}

Apresentado quando o armazenamento é um HashiCorp Vault, um OpenBao ou um Devolutions Server.

O endereço do servidor, com o respetivo esquema, por exemplo `https://vault.example.com:8200`. O
campo é obrigatório.

## Mount {#secretStorage-vaultMount}

Apresentado quando o armazenamento é um HashiCorp Vault ou um OpenBao. Opcional.

O caminho de montagem do motor de segredos que guarda os segredos. Deixe o campo vazio para
utilizar `secret`, o valor predefinido.

## Namespace {#secretStorage-vaultNamespace}

Apresentado quando o armazenamento é um HashiCorp Vault ou um OpenBao. Opcional.

O namespace em que se encontram os segredos. Os namespaces existem no Vault Enterprise e no HCP
Vault Dedicated, e no OpenBao a partir da versão 2.3. Deixe o campo vazio para um servidor sem
namespaces.

## Origem do segredo {#secretSource-options}

Apresentado para todos os tipos de armazenamento, acima do segredo com que o Semaphore se
autentica: o **Token** de um Vault ou de um OpenBao, o segredo da aplicação de um Devolutions
Server, a chave secreta da AWS, o segredo de cliente do Azure.

De onde o Semaphore obtém esse segredo:

- **Store in DB**: escreve o segredo nesta caixa de diálogo; o Semaphore cifra-o e guarda-o na
  sua base de dados.
- **From ENV**: escreve o nome de uma variável de ambiente do servidor Semaphore; o segredo é
  lido dela sempre que o armazenamento é utilizado.
- **From File**: escreve o caminho absoluto de um ficheiro no servidor Semaphore; o segredo é
  lido dele sempre, sem a quebra de linha final. O ficheiro tem de estar dentro do diretório de
  segredos do servidor (`dirs.secrets` na configuração, `/tmp/semaphore` por predefinição), e um
  caminho com `..` é recusado.

Mudar a origem limpa o campo abaixo dela.

## Token {#secretStorage-vaultToken}

Apresentado quando o armazenamento é um HashiCorp Vault ou um OpenBao e a origem do segredo é
**Store in DB**.

O token com que o Semaphore se autentica no servidor. Tem de ter permissão para ler os segredos
e, a não ser que o armazenamento seja **Read only**, para os escrever e eliminar.

O campo é obrigatório para um novo armazenamento. Quando abre um armazenamento existente, está
vazio: deixe-o vazio para manter o token guardado, escreva um novo para o substituir.

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

Apresentado quando o armazenamento é um HashiCorp Vault ou um OpenBao e a origem do segredo é
**From ENV** ou **From File**.

O nome da variável de ambiente, ou o caminho do ficheiro, que contém o token. O campo é
obrigatório para um novo armazenamento.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Apresentado quando o armazenamento é um Devolutions Server.

O Semaphore liga-se ao servidor sem verificar o respetivo certificado TLS. Qualquer pessoa entre
o Semaphore e o servidor pode então ler os segredos. Deixe a opção desativada e, em vez disso,
dê ao servidor um certificado em que o Semaphore confie.

## Vault ID {#secretStorage-dvlsKey}

Apresentado quando o armazenamento é um Devolutions Server.

O ID do cofre do Devolutions Server que guarda os segredos. O campo é obrigatório para um novo
armazenamento.

## App Key {#secretStorage-dvlsAppKey}

Apresentado quando o armazenamento é um Devolutions Server.

A chave de aplicação da identidade de aplicação com que o Semaphore se autentica, aquela cujo
segredo vai para **Secret**. O campo é obrigatório para um novo armazenamento.

## Secret {#secretStorage-dvlsSecret}

Apresentado quando o armazenamento é um Devolutions Server e a origem do segredo é
**Store in DB**.

O segredo da aplicação da identidade de aplicação de **App Key**. O campo é obrigatório para um
novo armazenamento. Quando abre um armazenamento existente, está vazio: deixe-o vazio para manter
o segredo guardado, escreva um novo para o substituir.

## Env var name / Path to the file {#secretStorage-dvlsEnv}

Apresentado quando o armazenamento é um Devolutions Server e a origem do segredo é **From ENV**
ou **From File**.

O nome da variável de ambiente, ou o caminho do ficheiro, que contém o segredo da aplicação. O
campo é obrigatório para um novo armazenamento.

## Region {#secretStorage-awsRegion}

Apresentado quando o armazenamento é um AWS Secrets Manager.

A região da AWS em que se encontram os segredos, por exemplo `us-east-1`. O campo é obrigatório.

## Endpoint URL {#secretStorage-awsEndpointURL}

Apresentado quando o armazenamento é um AWS Secrets Manager. Opcional.

O endereço da API do Secrets Manager a chamar em vez do endpoint predefinido da região, por
exemplo um endpoint de VPC. Deixe o campo vazio para utilizar o endpoint predefinido da AWS.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Apresentado quando o armazenamento é um AWS Secrets Manager.

Quando ativado, o Semaphore não guarda nenhuma chave de acesso e obtém as credenciais da AWS do
ambiente em que o servidor é executado, como a função IAM da sua instância. Os campos da chave de
acesso abaixo ficam então desativados.

Ativá-lo remove a chave de acesso guardada quando guarda. Desativá-lo novamente exige uma nova
**Secret Access Key**.

## Access Key ID {#secretStorage-awsAccessKeyId}

Apresentado quando o armazenamento é um AWS Secrets Manager. Desativado quando
**Use IAM Role / Instance Profile** está ativado.

O ID da chave de acesso com que o Semaphore se autentica, por exemplo `AKIAIOSFODNN7EXAMPLE`. O
campo é obrigatório quando a função IAM está desativada. O ID não é um segredo: é guardado e
apresentado tal como está.

## Secret Access Key {#secretStorage-awsSecretKey}

Apresentado quando o armazenamento é um AWS Secrets Manager e a origem do segredo é
**Store in DB**. Desativado quando **Use IAM Role / Instance Profile** está ativado.

O segredo da chave de acesso. O campo é obrigatório para um novo armazenamento e depois de a
função IAM ser desativada. Caso contrário, está vazio quando abre o armazenamento: deixe-o vazio
para manter a chave guardada, escreva uma nova para a substituir.

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

Apresentado quando o armazenamento é um AWS Secrets Manager e a origem do segredo é **From ENV**
ou **From File**. Desativado quando **Use IAM Role / Instance Profile** está ativado.

O nome da variável de ambiente, ou o caminho do ficheiro, que contém a chave de acesso secreta. O
campo é obrigatório nos mesmos casos que **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Apresentado quando o armazenamento é um Azure Key Vault.

O endereço do cofre de chaves, por exemplo `https://my-vault.vault.azure.net`. O campo é
obrigatório.

## Tenant ID {#secretStorage-azureTenantId}

Apresentado quando o armazenamento é um Azure Key Vault.

O ID do inquilino do Microsoft Entra da aplicação com que o Semaphore se autentica. O campo é
obrigatório.

## Client ID {#secretStorage-azureClientId}

Apresentado quando o armazenamento é um Azure Key Vault.

O ID da aplicação (cliente) dessa aplicação. Esta precisa de acesso aos segredos do cofre de
chaves. O campo é obrigatório.

## Client Secret {#secretStorage-azureClientSecret}

Apresentado quando o armazenamento é um Azure Key Vault e a origem do segredo é **Store in DB**.

O segredo de cliente da aplicação. O campo é obrigatório para um novo armazenamento. Quando abre
um armazenamento existente, está vazio: deixe-o vazio para manter o segredo guardado, escreva um
novo para o substituir.

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

Apresentado quando o armazenamento é um Azure Key Vault e a origem do segredo é **From ENV** ou
**From File**.

O nome da variável de ambiente, ou o caminho do ficheiro, que contém o segredo de cliente. O
campo é obrigatório para um novo armazenamento.

## Read only {#secretStorage-readonly}

Quando ativado, o Semaphore apenas lê do armazenamento. Um segredo de um grupo de variáveis não
pode ser criado nem alterado nele: guardá-lo falha. Eliminar um segredo, uma chave ou um grupo de
variáveis no Semaphore deixa o respetivo segredo no armazenamento.

Quando desativado, o Semaphore escreve no armazenamento os segredos dos grupos de variáveis que o
utilizam e elimina-os dele juntamente com os grupos.

## Sync keys enabled {#secretStorage-syncEnabled}

Ative-o para importar segredos do armazenamento para o Armazenamento de Chaves do projeto, a
partir dos caminhos definidos em **Sync paths**. A importação é executada quando prime o botão de
sincronização do armazenamento no separador **Storages**, e automaticamente no intervalo
definido em **Sync paths**.

Uma chave importada fica marcada como sincronizada: o seu nome e tipo não podem ser alterados, e
eliminá-la no Semaphore deixa o segredo no armazenamento. Eliminar o armazenamento enquanto a
sincronização está ativada elimina também as chaves que ele importou.

## Sync paths {#secretStorage-syncPaths}

Apresentado quando **Sync keys enabled** está ativado.

Abre a caixa de diálogo com os caminhos do armazenamento a partir dos quais importar e o
intervalo da importação automática. O número no botão é o número de caminhos. O botão de
sincronização do armazenamento permanece desativado até que tenha pelo menos um caminho.

A caixa de diálogo tem apenas **Fechar**: o que alterar nela é guardado com o armazenamento,
através de **Salvar** desta caixa de diálogo (**Criar** para um novo armazenamento).
