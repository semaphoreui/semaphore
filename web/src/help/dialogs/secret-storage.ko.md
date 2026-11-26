---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# 수정 Secret Storage

비밀 저장소는 Semaphore가 비밀 값을 보관하고 읽어 오는 외부 비밀 관리자입니다. HashiCorp Vault, OpenBao,
Devolutions Server, AWS Secrets Manager 또는 Azure Key Vault입니다. 저장소와 함께 만든 변수 그룹은 비밀 값을 그
저장소에 보관하며, 동기화가 켜진 저장소는 비밀 값을 프로젝트의 키 저장소로 가져옵니다. 저장소의 종류는
**New Storage** 메뉴에서 선택하며 이 대화 상자에서는 바꿀 수 없습니다. 아래의 필드는 종류에 따라 다릅니다.

## 이름 {#secretStorage-name}

키 저장소의 **Storages** 탭과 새 변수 그룹의 **Secret storage** 선택 목록에 표시되는 저장소 이름입니다. 이
필드는 필수입니다.

## Server URL {#secretStorage-vaultURL}

저장소가 HashiCorp Vault, OpenBao 또는 Devolutions Server일 때 표시됩니다.

스킴을 포함한 서버 주소입니다. 예: `https://vault.example.com:8200`. 이 필드는 필수입니다.

## Mount {#secretStorage-vaultMount}

저장소가 HashiCorp Vault 또는 OpenBao일 때 표시됩니다. 선택 사항입니다.

비밀 값을 보관하는 시크릿 엔진의 마운트 경로입니다. 기본값인 `secret`을 사용하려면 비워 두세요.

## Namespace {#secretStorage-vaultNamespace}

저장소가 HashiCorp Vault 또는 OpenBao일 때 표시됩니다. 선택 사항입니다.

비밀 값이 있는 네임스페이스입니다. 네임스페이스는 Vault Enterprise와 HCP Vault Dedicated, 그리고 버전 2.3부터의
OpenBao에 있습니다. 네임스페이스가 없는 서버라면 비워 두세요.

## Secret source {#secretSource-options}

모든 종류의 저장소에서, Semaphore가 로그인할 때 사용하는 비밀 값 위에 표시됩니다. Vault나 OpenBao의 **Token**,
Devolutions Server의 앱 시크릿, AWS의 비밀 키, Azure의 클라이언트 시크릿이 해당합니다.

Semaphore가 그 비밀 값을 가져오는 위치입니다.

- **Store in DB**: 이 대화 상자에 비밀 값을 입력합니다. Semaphore는 이를 암호화하여 데이터베이스에 보관합니다.
- **From ENV**: Semaphore 서버의 환경 변수 이름을 입력합니다. 저장소를 사용할 때마다 그 변수에서 비밀 값을
  읽습니다.
- **From File**: Semaphore 서버에 있는 파일의 절대 경로를 입력합니다. 매번 그 파일에서 마지막 줄 바꿈을 뺀 비밀
  값을 읽습니다. 파일은 서버의 비밀 디렉터리(구성의 `dirs.secrets`, 기본값 `/tmp/semaphore`) 안에 있어야 하며,
  `..`이 들어 있는 경로는 거부됩니다.

소스를 전환하면 그 아래의 필드가 지워집니다.

## Token {#secretStorage-vaultToken}

저장소가 HashiCorp Vault 또는 OpenBao이고 비밀 소스가 **Store in DB**일 때 표시됩니다.

Semaphore가 서버에 인증할 때 사용하는 토큰입니다. 이 토큰에는 비밀 값을 읽을 권한이 있어야 하며, 저장소가
**Read only**가 아니면 비밀 값을 쓰고 삭제할 권한도 있어야 합니다.

새 저장소에서는 이 필드가 필수입니다. 기존 저장소를 열면 이 필드는 비어 있습니다. 저장된 토큰을 유지하려면
비워 두고, 교체하려면 새 토큰을 입력하세요.

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

저장소가 HashiCorp Vault 또는 OpenBao이고 비밀 소스가 **From ENV** 또는 **From File**일 때 표시됩니다.

토큰을 담고 있는 환경 변수의 이름 또는 파일의 경로입니다. 새 저장소에서는 이 필드가 필수입니다.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

저장소가 Devolutions Server일 때 표시됩니다.

Semaphore가 서버의 TLS 인증서를 확인하지 않고 연결합니다. 그러면 Semaphore와 서버 사이에 있는 누구나 비밀 값을
읽을 수 있습니다. 이 옵션은 꺼 두고, 대신 Semaphore가 신뢰하는 인증서를 서버에 설정하세요.

## Vault ID {#secretStorage-dvlsKey}

저장소가 Devolutions Server일 때 표시됩니다.

비밀 값을 보관하는 Devolutions Server 금고의 ID입니다. 새 저장소에서는 이 필드가 필수입니다.

## App Key {#secretStorage-dvlsAppKey}

저장소가 Devolutions Server일 때 표시됩니다.

Semaphore가 로그인할 때 사용하는 애플리케이션 ID의 애플리케이션 키로, 그 시크릿이 **Secret**에 들어갑니다. 새
저장소에서는 이 필드가 필수입니다.

## Secret {#secretStorage-dvlsSecret}

저장소가 Devolutions Server이고 비밀 소스가 **Store in DB**일 때 표시됩니다.

**App Key**의 애플리케이션 ID에 대한 앱 시크릿입니다. 새 저장소에서는 이 필드가 필수입니다. 기존 저장소를 열면
이 필드는 비어 있습니다. 저장된 시크릿을 유지하려면 비워 두고, 교체하려면 새 시크릿을 입력하세요.

## Env var name / Path to the file {#secretStorage-dvlsEnv}

저장소가 Devolutions Server이고 비밀 소스가 **From ENV** 또는 **From File**일 때 표시됩니다.

앱 시크릿을 담고 있는 환경 변수의 이름 또는 파일의 경로입니다. 새 저장소에서는 이 필드가 필수입니다.

## Region {#secretStorage-awsRegion}

저장소가 AWS Secrets Manager일 때 표시됩니다.

비밀 값이 있는 AWS 리전입니다. 예: `us-east-1`. 이 필드는 필수입니다.

## Endpoint URL {#secretStorage-awsEndpointURL}

저장소가 AWS Secrets Manager일 때 표시됩니다. 선택 사항입니다.

리전의 기본 엔드포인트 대신 호출할 Secrets Manager API의 주소입니다. 예: VPC 엔드포인트. 기본 AWS 엔드포인트를
사용하려면 비워 두세요.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

저장소가 AWS Secrets Manager일 때 표시됩니다.

켜면 Semaphore는 액세스 키를 저장하지 않고, 서버가 실행되는 환경(예: 인스턴스의 IAM 역할)에서 AWS 자격 증명을
가져옵니다. 이 경우 아래의 액세스 키 필드는 비활성화됩니다.

켜면 저장할 때 저장된 액세스 키가 제거됩니다. 다시 끄려면 새 **Secret Access Key**가 필요합니다.

## Access Key ID {#secretStorage-awsAccessKeyId}

저장소가 AWS Secrets Manager일 때 표시됩니다. **Use IAM Role / Instance Profile**이 켜져 있으면
비활성화됩니다.

Semaphore가 로그인할 때 사용하는 액세스 키의 ID입니다. 예: `AKIAIOSFODNN7EXAMPLE`. IAM 역할이 꺼져 있으면 이
필드는 필수입니다. ID는 비밀 값이 아니므로 그대로 저장되고 표시됩니다.

## Secret Access Key {#secretStorage-awsSecretKey}

저장소가 AWS Secrets Manager이고 비밀 소스가 **Store in DB**일 때 표시됩니다.
**Use IAM Role / Instance Profile**이 켜져 있으면 비활성화됩니다.

액세스 키의 비밀 값입니다. 새 저장소에서, 그리고 IAM 역할을 끈 후에는 이 필드가 필수입니다. 그 밖의 경우
저장소를 열면 이 필드는 비어 있습니다. 저장된 키를 유지하려면 비워 두고, 교체하려면 새 키를 입력하세요.

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

저장소가 AWS Secrets Manager이고 비밀 소스가 **From ENV** 또는 **From File**일 때 표시됩니다.
**Use IAM Role / Instance Profile**이 켜져 있으면 비활성화됩니다.

비밀 액세스 키를 담고 있는 환경 변수의 이름 또는 파일의 경로입니다. 이 필드는 **Secret Access Key**와 같은
경우에 필수입니다.

## Vault URL {#secretStorage-azureVaultURL}

저장소가 Azure Key Vault일 때 표시됩니다.

키 자격 증명 모음의 주소입니다. 예: `https://my-vault.vault.azure.net`. 이 필드는 필수입니다.

## Tenant ID {#secretStorage-azureTenantId}

저장소가 Azure Key Vault일 때 표시됩니다.

Semaphore가 로그인하는 애플리케이션이 속한 Microsoft Entra 테넌트의 ID입니다. 이 필드는 필수입니다.

## Client ID {#secretStorage-azureClientId}

저장소가 Azure Key Vault일 때 표시됩니다.

그 애플리케이션의 애플리케이션(클라이언트) ID입니다. 이 애플리케이션에는 키 자격 증명 모음의 비밀 값에 대한
액세스 권한이 필요합니다. 이 필드는 필수입니다.

## Client Secret {#secretStorage-azureClientSecret}

저장소가 Azure Key Vault이고 비밀 소스가 **Store in DB**일 때 표시됩니다.

애플리케이션의 클라이언트 시크릿입니다. 새 저장소에서는 이 필드가 필수입니다. 기존 저장소를 열면 이 필드는
비어 있습니다. 저장된 시크릿을 유지하려면 비워 두고, 교체하려면 새 시크릿을 입력하세요.

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

저장소가 Azure Key Vault이고 비밀 소스가 **From ENV** 또는 **From File**일 때 표시됩니다.

클라이언트 시크릿을 담고 있는 환경 변수의 이름 또는 파일의 경로입니다. 새 저장소에서는 이 필드가 필수입니다.

## Read only {#secretStorage-readonly}

켜면 Semaphore는 저장소에서 읽기만 합니다. 변수 그룹의 비밀 값을 그 안에 만들거나 변경할 수 없으며, 저장하면
실패합니다. Semaphore에서 비밀 값, 키 또는 변수 그룹을 삭제해도 그 비밀 값은 저장소에 남습니다.

끄면 Semaphore는 저장소를 사용하는 변수 그룹의 비밀 값을 저장소에 기록하고, 그룹과 함께 저장소에서 삭제합니다.

## Sync keys enabled {#secretStorage-syncEnabled}

켜면 **Sync paths**에 설정한 경로에서 저장소의 비밀 값을 프로젝트의 키 저장소로 가져옵니다. 가져오기는
**Storages** 탭에서 저장소의 동기화 버튼을 누를 때와, **Sync paths**에 설정한 간격마다 자동으로 실행됩니다.

가져온 키는 동기화된 것으로 표시됩니다. 이름과 유형을 바꿀 수 없으며, Semaphore에서 삭제해도 비밀 값은 저장소에
남습니다. 동기화가 켜진 상태에서 저장소를 삭제하면 저장소가 가져온 키도 함께 삭제됩니다.

## Sync paths {#secretStorage-syncPaths}

**Sync keys enabled**가 켜져 있을 때 표시됩니다.

가져올 저장소의 경로와 자동 가져오기 간격이 있는 대화 상자를 엽니다. 버튼의 숫자는 경로의 개수입니다. 경로가
하나 이상 있을 때까지 저장소의 동기화 버튼은 비활성화된 채로 있습니다.

그 대화 상자에는 **닫기**만 있습니다. 그 안에서 변경한 내용은 이 대화 상자의 **저장**(새 저장소의 경우
**생성**)으로 저장소와 함께 저장됩니다.
