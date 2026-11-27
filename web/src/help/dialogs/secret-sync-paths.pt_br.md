---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Os caminhos de um armazenamento de segredos a partir dos quais o Semaphore importa segredos, e com
que frequência ele faz isso por conta própria. Aberta a partir de um armazenamento de segredos, a
importação preenche o Armazenamento de Chaves do projeto; aberta a partir de um grupo de
variáveis, ela preenche esse grupo. O que você altera aqui é mantido quando você clica em
**Fechar** e salvo com o armazenamento ou com o grupo, pelo **Salvar** dele.

## Auto-sync interval {#varGroupSync-syncInterval}

Exibido quando a caixa de diálogo é aberta a partir de um grupo de variáveis.

Com que frequência o grupo é importado por conta própria, em minutos. `0`, o padrão, desativa a
importação automática: o grupo então só é importado quando você clica no botão de sincronização
dele na página Grupos de Variáveis.

O servidor verifica uma vez por minuto e importa quando o intervalo já passou desde a última
tentativa, bem-sucedida ou não; a primeira verificação depois que você define um intervalo importa
imediatamente. Uma importação que falhou é tentada novamente após o próximo intervalo.

## Auto-sync interval {#secretStorageSync-syncInterval}

Exibido quando a caixa de diálogo é aberta a partir de um armazenamento de segredos.

Com que frequência o armazenamento é importado por conta própria, em minutos. `0`, o padrão,
desativa a importação automática: o armazenamento então só é importado quando você clica no botão
de sincronização dele na aba **Storages** do Armazenamento de Chaves.

O servidor verifica uma vez por minuto e importa quando o intervalo já passou desde a última
tentativa, bem-sucedida ou não; a primeira verificação depois que você define um intervalo importa
imediatamente. Uma importação que falhou é tentada novamente após o próximo intervalo.

## Caminho {#secretStorageSync-path}

Exibido em cada linha da lista. Adicione uma linha com **Add path**.

Um caminho no armazenamento a partir do qual importar os segredos, escrito como o armazenamento o
nomeia. Cada linha é um caminho; a importação lê todos eles.

## separator {#secretStorageSync-separator}

Exibido em cada linha da lista. Opcional.

Um separador que a importação deste caminho usa. Deixe-o vazio, a menos que os nomes no seu
armazenamento precisem de um.

## prefix {#secretStorageSync-prefix}

Exibido em cada linha da lista. Opcional.

Um prefixo que a importação deste caminho usa. Deixe-o vazio quando os segredos importados não
precisam de nenhum.

## Remover caminho {#secretStorageSync-removePath}

Exibido em cada linha da lista.

Remove o caminho da lista. A alteração entra em vigor quando o armazenamento ou o grupo é salvo.

## Add path {#secretStorageSync-addPath}

Adiciona uma linha vazia à lista. Enquanto não houver pelo menos um caminho, a caixa de diálogo
informa que nenhum caminho de sincronização está definido, e o botão de sincronização do
armazenamento ou do grupo permanece desativado.
