---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Os caminhos de um armazenamento de segredos a partir dos quais o Semaphore importa segredos, e a
frequência com que o faz automaticamente. Aberta a partir de um armazenamento de segredos, a
importação preenche o Armazenamento de Chaves do projeto; aberta a partir de um grupo de
variáveis, preenche esse grupo. O que alterar aqui é mantido quando prime **Fechar** e guardado
com o armazenamento ou o grupo, através do respetivo **Salvar**.

## Auto-sync interval {#varGroupSync-syncInterval}

Apresentado quando a caixa de diálogo é aberta a partir de um grupo de variáveis.

Com que frequência o grupo é importado automaticamente, em minutos. `0`, o valor predefinido,
desativa a importação automática: o grupo só é então importado quando prime o respetivo botão de
sincronização na página Grupos de Variáveis.

O servidor verifica uma vez por minuto e importa quando o intervalo tiver decorrido desde a
última tentativa, bem-sucedida ou não; a primeira verificação depois de definir um intervalo
importa de imediato. Uma importação falhada é tentada novamente após o intervalo seguinte.

## Auto-sync interval {#secretStorageSync-syncInterval}

Apresentado quando a caixa de diálogo é aberta a partir de um armazenamento de segredos.

Com que frequência o armazenamento é importado automaticamente, em minutos. `0`, o valor
predefinido, desativa a importação automática: o armazenamento só é então importado quando prime
o respetivo botão de sincronização no separador **Storages** do Armazenamento de Chaves.

O servidor verifica uma vez por minuto e importa quando o intervalo tiver decorrido desde a
última tentativa, bem-sucedida ou não; a primeira verificação depois de definir um intervalo
importa de imediato. Uma importação falhada é tentada novamente após o intervalo seguinte.

## Caminho {#secretStorageSync-path}

Apresentado em cada linha da lista. Adicione uma linha com **Add path**.

Um caminho no armazenamento a partir do qual importar os segredos, escrito tal como o
armazenamento o designa. Cada linha é um caminho; a importação lê todos eles.

## separator {#secretStorageSync-separator}

Apresentado em cada linha da lista. Opcional.

Um separador que a importação deste caminho utiliza. Deixe-o vazio, a não ser que os nomes no
seu armazenamento precisem de um.

## prefix {#secretStorageSync-prefix}

Apresentado em cada linha da lista. Opcional.

Um prefixo que a importação deste caminho utiliza. Deixe-o vazio quando os segredos importados
não precisam de nenhum.

## Remover caminho {#secretStorageSync-removePath}

Apresentado em cada linha da lista.

Remove o caminho da lista. A alteração produz efeito quando o armazenamento ou o grupo é
guardado.

## Add path {#secretStorageSync-addPath}

Adiciona uma linha vazia à lista. Enquanto não houver pelo menos um caminho, a caixa de diálogo
indica que não estão definidos caminhos de sincronização, e o botão de sincronização do
armazenamento ou do grupo permanece desativado.
