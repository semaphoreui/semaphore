---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Semaphore가 비밀 값을 가져오는 비밀 저장소의 경로와, 자동으로 가져오는 빈도입니다. 비밀 저장소에서 열면 가져온
값은 프로젝트의 키 저장소를 채우고, 변수 그룹에서 열면 그 그룹을 채웁니다. 여기에서 변경한 내용은 **닫기**를
누를 때 유지되며, 저장소나 그룹의 **저장**으로 저장소나 그룹과 함께 저장됩니다.

## Auto-sync interval {#varGroupSync-syncInterval}

변수 그룹에서 대화 상자를 열었을 때 표시됩니다.

그룹을 자동으로 가져오는 빈도로, 분 단위입니다. 기본값인 `0`은 자동 가져오기를 끕니다. 이 경우 그룹은 변수 그룹
페이지에서 동기화 버튼을 누를 때만 가져옵니다.

서버는 1분마다 확인하며, 성공 여부와 관계없이 마지막 시도 이후 간격이 지나면 가져옵니다. 간격을 설정한 후 첫
번째 확인에서는 즉시 가져옵니다. 실패한 가져오기는 다음 간격 후에 다시 시도됩니다.

## Auto-sync interval {#secretStorageSync-syncInterval}

비밀 저장소에서 대화 상자를 열었을 때 표시됩니다.

저장소를 자동으로 가져오는 빈도로, 분 단위입니다. 기본값인 `0`은 자동 가져오기를 끕니다. 이 경우 저장소는 키
저장소의 **Storages** 탭에서 동기화 버튼을 누를 때만 가져옵니다.

서버는 1분마다 확인하며, 성공 여부와 관계없이 마지막 시도 이후 간격이 지나면 가져옵니다. 간격을 설정한 후 첫
번째 확인에서는 즉시 가져옵니다. 실패한 가져오기는 다음 간격 후에 다시 시도됩니다.

## 경로 {#secretStorageSync-path}

목록의 모든 행에 표시됩니다. **Add path**로 행을 추가하세요.

비밀 값을 가져올 저장소의 경로로, 저장소가 부르는 그대로 씁니다. 각 행이 하나의 경로이며, 가져오기는 모든 경로를
읽습니다.

## separator {#secretStorageSync-separator}

목록의 모든 행에 표시됩니다. 선택 사항입니다.

이 경로를 가져올 때 사용하는 구분자입니다. 저장소의 이름에 구분자가 필요한 경우가 아니면 비워 두세요.

## prefix {#secretStorageSync-prefix}

목록의 모든 행에 표시됩니다. 선택 사항입니다.

이 경로를 가져올 때 사용하는 접두사입니다. 가져온 비밀 값에 접두사가 필요 없으면 비워 두세요.

## 경로 제거 {#secretStorageSync-removePath}

목록의 모든 행에 표시됩니다.

목록에서 경로를 제거합니다. 변경 내용은 저장소나 그룹을 저장할 때 적용됩니다.

## Add path {#secretStorageSync-addPath}

목록에 빈 행을 추가합니다. 경로가 하나 이상 있을 때까지 대화 상자에는 정의된 동기화 경로가 없다고 표시되며,
저장소나 그룹의 동기화 버튼은 비활성화된 채로 있습니다.
