---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# 템플릿 수정 'Ansible Playbook'

작업 템플릿은 Semaphore가 무엇을 어떻게 실행할지 정합니다. 리포지토리와 그 안의 파일, 인벤토리와 변수, 그리고
작업을 시작할 때 사용자가 변경할 수 있는 옵션이 해당합니다. 모든 작업은 템플릿에서 시작됩니다. 대화 상자의
필드는 템플릿의 앱(Ansible, Terraform, OpenTofu, Terragrunt, Bash, Python, PowerShell 등)에 따라 다르며, 일부 앱에만
있는 필드에는 그렇게 적혀 있습니다. 변경 내용은 **저장** 후에 시작된 작업에 적용됩니다.

## 템플릿 유형 {#templateForm-typeTabs}

맨 위의 탭은 템플릿의 종류를 정합니다.

- **작업**: 코드를 실행합니다. 대부분의 템플릿이 이 종류입니다.
- **빌드**: 아티팩트를 만듭니다. 모든 빌드 작업은 **시작 버전**부터 세는 버전을 받습니다.
- **배포**: 빌드가 만든 아티팩트를 배포합니다. 시작할 때 배포할 버전의 빌드 작업을 선택합니다.

빌드 또는 배포 작업은 환경에 `SEMAPHORE_TASK_TYPE`, `SEMAPHORE_TASK_TARGET_VERSION`(빌드),
`SEMAPHORE_TASK_INCOMING_VERSION`(배포) 변수를 받습니다. Ansible은 같은 값을 추가 변수
`semaphore_vars.task_details`로 받습니다.

## 시작 버전 {#templateForm-startVersion}

유형이 **빌드**일 때 표시됩니다. 필수입니다.

첫 번째 빌드 작업의 버전입니다. 예: `1.0.0` 또는 `v0.1`. 이후의 각 빌드는 이전 버전의 마지막 숫자를 1 증가시키고,
그 주변의 텍스트와 앞자리 0 채우기를 유지합니다: `1.0.0`, `1.0.1`, `1.0.2`. 앞으로 건너뛰려면 여기의 숫자를 마지막
빌드보다 높이세요. 예: `2.0.0`. 숫자가 없는 버전은 바뀌지 않습니다.

## 빌드 템플릿 {#templateForm-buildTemplateId}

유형이 **배포**일 때 표시됩니다. 필수입니다.

이 템플릿이 배포하는 아티팩트를 작업으로 만드는 템플릿입니다. 목록에는 프로젝트의 빌드 템플릿이 있고, 그 아래에
배포 템플릿이 있습니다. 배포는 다른 배포 뒤에 올 수 있으며, 이 경우 그 배포가 빌드에서 받은 버전을 받습니다.

## 자동 실행 {#templateForm-autorun}

유형이 **배포**일 때 표시됩니다.

**빌드 템플릿**의 작업이 성공할 때마다 그 작업이 만든 버전으로 이 템플릿의 작업을 자동으로 시작합니다. 실패한
빌드는 아무것도 시작하지 않습니다.

## 이름 {#templateForm-name}

작업 템플릿 페이지에 표시되는 템플릿 이름입니다. 이 필드는 필수입니다.

## 리포지토리 {#templateForm-repositoryId}

실행할 코드를 보관하는 프로젝트의 리포지토리입니다. 이 필드는 필수입니다. 모든 작업 전에 Semaphore는 리포지토리
사본을 업데이트하고 그곳에서 코드를 실행합니다.

필드 오른쪽의 레이블은 작업이 체크아웃하는 브랜치를 보여 줍니다. 클릭하면 **브랜치** 필드가 열립니다.

## 브랜치 {#templateForm-gitBranch}

**리포지토리** 오른쪽의 브랜치 레이블을 클릭하면 표시됩니다.

리포지토리에 설정된 브랜치 대신 이 템플릿의 작업이 체크아웃하는 브랜치 또는 태그입니다. 리포지토리의 브랜치를
사용하려면 비워 두세요. Semaphore가 리포지토리의 브랜치를 읽을 수 있으면 필드에 브랜치 목록이 표시됩니다.

리포지토리가 모노레포이면 이 필드는 비활성화됩니다. 모든 템플릿이 리포지토리의 브랜치에서 하나의 체크아웃을
공유합니다.

## 플레이북 파일 경로 {#templateForm-playbook}

실행할 대상으로, 리포지토리 루트로부터의 경로입니다. 예: `deploy/site.yml`. 경로는 절대 경로일 수 없으며
리포지토리 밖으로 나갈 수 없습니다. 목록에 리포지토리의 파일이 표시되며, 목록에 없는 경로를 입력할 수도 있습니다.
필드 오른쪽의 버튼을 누르면 목록을 다시 읽습니다.

레이블과 의미는 앱에 따라 바뀝니다.

- **Ansible**: 플레이북입니다. 필수이며, 목록에 `.yml` 및 `.yaml` 파일이 표시됩니다.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, 즉 코드가 있는 디렉터리입니다. 리포지토리
  루트에서 코드를 실행하려면 비워 두세요. 디렉터리의 내용을 나열하려면 `/`를 입력하세요.
- **Bash**, **Python**, **PowerShell** 및 기타 앱: **Script Filename**, 즉 실행할 스크립트입니다. 필수입니다.
  스크립트는 리포지토리 루트에서 실행됩니다.

## Different working directory {#templateForm-workingDirectoryToggle}

Ansible 템플릿에 표시됩니다.

리포지토리 루트가 아닌 다른 디렉터리에서 `ansible-playbook`을 실행합니다. 체크하면 **Working directory** 필드가
표시됩니다. 다시 루트에서 실행하려면 체크를 해제하세요.

## Working directory {#templateForm-workingDirectory}

**Different working directory**가 체크되어 있을 때 표시됩니다. 필수입니다.

`ansible-playbook`과 `ansible-galaxy`가 실행되는 디렉터리로, 리포지토리 루트로부터의 경로입니다. 예: `ansible`.
Ansible은 현재 디렉터리에서 그 파일을 읽으므로, 프로젝트의 `ansible.cfg`가 하위 디렉터리에 있을 때
사용하세요. 플레이북 경로는 여전히 리포지토리 루트를 기준으로 합니다. 경로는 절대 경로일 수 없으며 리포지토리
밖으로 나갈 수 없습니다.

## 인벤토리 {#templateForm-inventoryId}

Ansible, Terraform, OpenTofu, Terragrunt 템플릿에 표시됩니다.

- **Ansible**: 플레이북이 대상으로 삼는 인벤토리로, 호스트의 자격 증명을 포함합니다. 이 필드는 필수입니다.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, 즉 코드가 실행되는 워크스페이스입니다. Semaphore는
  계획 전에 워크스페이스를 선택하고, 없으면 만듭니다. 워크스페이스 없이 저장한 새 템플릿은 `default`라는
  워크스페이스를 받습니다.

목록에는 템플릿의 앱용으로 만든 프로젝트의 인벤토리가 있습니다.

## 변수 그룹 {#templateForm-environmentIds}

선택 사항입니다. 이 템플릿의 모든 작업이 받는 변수, 환경 변수, 비밀 값의 출처가 되는 변수 그룹입니다. 여러 개를
선택할 수 있습니다. 그룹은 필드의 순서대로 병합되며, 두 그룹이 같은 이름을 설정하면 나중 그룹이 우선합니다.

변수가 코드에 전달되는 방식은 앱에 따라 다릅니다. Ansible에는 추가 변수, Terraform과 OpenTofu에는 `-var`,
스크립트에는 `name=value` 인수입니다. 환경 변수는 모든 앱에 설정됩니다.

## 보기 {#templateForm-viewId}

선택 사항입니다. 템플릿이 표시되는 보기입니다. 보기는 작업 템플릿 페이지의 탭으로, 프로젝트의 템플릿을 묶습니다.
템플릿을 어떤 보기에도 넣지 않으려면 비워 두세요.

## 러너 태그 {#templateForm-runnerTag}

프로젝트 러너를 사용할 수 있을 때 표시됩니다.

이 템플릿의 작업은 서버가 직접 실행할 수 있는 경우에도 이 태그가 있는 러너에서만 실행됩니다. 이 태그는
인벤토리의 러너 태그보다 우선합니다. 인벤토리의 태그를 사용하거나 태그를 사용하지 않으려면 비워 두세요.

## Executor image {#templateForm-executorImage}

러너가 이 템플릿의 작업을 위해 시작하는 컨테이너 이미지입니다. 예: `semaphoreui/job:latest`. 러너 구성의
이미지를 사용하려면 비워 두세요.

Docker 및 Kubernetes 실행기만 컨테이너에서 작업을 실행하므로, 이 필드는 둘 중 하나를 사용할 수 있을 때만
활성화됩니다. 컨테이너 없이 실행되는 작업은 이 필드를 무시합니다.

## 설문 변수 {#surveyVars}

이 템플릿의 작업이 시작될 때 사용자에게 요청하는 값입니다. 변수를 추가하려면 **+**를, 변경하려면 칩을, 제거하려면
칩의 X 표시를 클릭하세요. 질문 순서를 정하려면 칩을 끌어다 놓으세요.

답은 각 변수의 설정에 따라 변수 그룹의 변수처럼, 또는 환경 변수로 코드에 전달됩니다. 답은 같은 이름을 설정하는
변수 그룹보다 우선합니다.

## Allow parallel tasks {#templateForm-allowParallelTasks}

이 템플릿의 작업이 동시에 실행될 수 있게 합니다. 이 옵션이 없으면 템플릿의 작업은 실행 중인 작업이 끝날 때까지
대기열에서 기다립니다. 어느 경우든 프로젝트와 서버의 병렬 작업 한도가 적용됩니다.

## Issue JWT to task runner {#templateForm-jwtEnabled}

관리자가 서버에서 JWT 발급을 켰을 때 표시됩니다.

이 템플릿의 모든 작업에 Semaphore가 서명한 JSON Web Token을 환경 변수 `SEMAPHORE_JWT`로 제공합니다. 코드는 이
토큰을 다른 서비스에 보여 주어 자신이 어떤 작업인지 증명할 수 있습니다. 토큰에는 작업, 프로젝트, 템플릿, 작업을
시작한 사용자가 들어 있습니다. 서비스는 Semaphore 서버의 `/.well-known/jwks.json`에 있는 공개 키로 서명을
확인합니다.

## JWT audience {#templateForm-jwtAudience}

**Issue JWT to task runner**가 체크되어 있을 때 표시됩니다. 선택 사항입니다.

토큰이 대상으로 하는 서비스로, 토큰의 `aud` 클레임에 기록됩니다. 예: `https://vault.example.com`. 각 값을 입력한
후 Enter를 누르세요. 최대 32개의 값을 지정할 수 있으며, audience 없는 토큰을 원하면 비워 두세요.

## JWT TTL {#templateForm-jwtTtl}

**Issue JWT to task runner**가 체크되어 있을 때 표시됩니다. 선택 사항입니다.

작업 시작부터 토큰이 유효한 기간으로, `30s`, `15m` 또는 `1h`와 같은 기간으로 지정합니다. 서버의 기본값을
사용하려면 비워 두세요. 관리자가 다르게 설정하지 않았다면 1시간입니다. 값은 서버의 최댓값을 넘을 수 없으며, 따로
설정하지 않았다면 24시간입니다.

## 저는 특정 리포지토리의 새로운 커밋에 대해서만 크론으로 작업을 실행하고 싶습니다 {#templateForm-cronVisible}

리포지토리에 새 커밋이 나타나면 이 템플릿의 작업을 시작합니다. Semaphore는 선택한 간격마다 리포지토리에 마지막
커밋을 묻고, 마지막 확인 이후 커밋이 바뀌었을 때만 작업을 시작합니다.

상자를 체크하고 저장하면 확인 일정이 만들어지고, 체크를 해제하고 저장하면 그 일정이 삭제됩니다. 템플릿의 다른
일정은 유지됩니다.

## 리포지토리 {#templateForm-cronRepositoryId}

**저는 특정 리포지토리의 새로운 커밋에 대해서만 크론으로 작업을 실행하고 싶습니다**가 체크되어 있을 때 표시됩니다.
필수입니다.

새 커밋을 감시할 리포지토리입니다. Semaphore는 템플릿의 **브랜치**가 아니라 그 리포지토리에 설정된 브랜치를
감시합니다.

## 체크 간격 {#templateForm-cronFormat}

**저는 특정 리포지토리의 새로운 커밋에 대해서만 크론으로 작업을 실행하고 싶습니다**가 체크되어 있을 때 표시됩니다.

Semaphore가 리포지토리에서 새 커밋을 확인하는 빈도로, 매분부터 하루 한 번까지입니다. 기본값은 매분입니다.

## 성공 알림 억제 {#templateForm-suppressSuccessAlerts}

이 템플릿의 작업이 성공했을 때 알림을 보내지 않게 합니다. 실패한 작업의 알림은 계속 전송됩니다. 프로젝트에
알림이 켜져 있을 때만 효과가 있습니다.

## 오류 알림 억제 {#templateForm-suppressErrorAlerts}

이 템플릿의 작업이 실패했을 때 알림을 보내지 않게 합니다. 성공한 작업의 알림은 계속 전송됩니다. 프로젝트에
알림이 켜져 있을 때만 효과가 있습니다.

## CLI args {#templateForm-args}

모든 작업의 명령에 추가되는 인수입니다. 칩 하나가 인수 하나이며, 셸 없이 그대로 전달됩니다. `--forks=10`을 칩
하나에 넣거나, `--forks`와 `10`을 칩 두 개에 넣으세요. 순서를 바꾸려면 칩을 끌어다 놓으세요. 목록 오른쪽의
버튼을 누르면 인수가 JSON 배열로 열립니다.

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: 플레이북 앞에 `ansible-playbook`에 추가됩니다.
- **Terraform**, **OpenTofu**, **Terragrunt**: `init`이 아니라 `plan`과 `apply`에 추가됩니다.
- 스크립트: 스크립트 이름 뒤에 추가됩니다.

## CLI 인수 {#templateForm-allowOverrideArgsInTask}

**프롬프트** 아래에 있습니다. 작업을 시작할 때 사용자가 자신의 인수를 추가할 수 있게 합니다. 이 인수는 템플릿의
인수 뒤에 추가됩니다.

## 브랜치 {#templateForm-allowOverrideBranchInTask}

**프롬프트** 아래에 있습니다. 작업을 시작할 때 사용자가 템플릿의 브랜치 대신 다른 브랜치, 태그 또는 커밋을 선택할
수 있게 합니다. 이 옵션이 없으면 작업은 항상 템플릿의 브랜치를 체크아웃합니다.

리포지토리가 모노레포이면 이 상자는 비활성화됩니다.

## 인벤토리 {#templateForm-allowOverrideInventory}

**프롬프트** 아래에 있습니다. Ansible, Terraform, OpenTofu, Terragrunt 템플릿에 표시됩니다.

작업을 시작할 때 사용자가 다른 인벤토리를 선택할 수 있게 합니다. 그곳에서 선택한 인벤토리는 Ansible 템플릿만
사용합니다. Terraform, OpenTofu 또는 Terragrunt 작업은 상자가 체크되어 있어도 템플릿의 워크스페이스에서
실행됩니다.

## 제한 {#templateForm-limit}

Ansible 템플릿에 표시됩니다.

인벤토리 중에서 플레이북을 실행할 호스트나 그룹으로, `--limit`으로 전달됩니다. 칩 하나가 패턴 하나입니다. 예:
`web` 또는 `db01`. 모든 호스트에서 실행하려면 목록을 비워 두세요.

## 태그 {#templateForm-tags}

Ansible 템플릿에 표시됩니다.

이 태그가 있는 작업과 롤만 실행하며, `--tags`로 전달됩니다. 모든 것을 실행하려면 목록을 비워 두세요.

## 태그 건너뛰기 {#templateForm-skipTags}

Ansible 템플릿에 표시됩니다.

이 태그가 있는 작업과 롤을 건너뛰며, `--skip-tags`로 전달됩니다.

## 금고 {#templateVaults}

Ansible 템플릿에 표시됩니다.

플레이북이 파일을 복호화하는 데 필요한 Ansible Vault 비밀번호입니다. 추가하려면 **+**를, 변경하려면 칩을,
제거하려면 칩의 X 표시를 클릭하세요. 각 금고에는 이름, 즉 vault ID가 있으며, 키 저장소의 키나 리포지토리의
클라이언트 스크립트에서 비밀번호를 제공합니다.

## 자동 승인 {#templateForm-autoApprove}

Terraform, OpenTofu, Terragrunt 템플릿에 표시됩니다.

묻지 않고 계획을 적용합니다. 이 옵션이 없으면 작업은 계획 후에 멈추고 사용자가 확인하거나 거부할 때까지
기다립니다. 변경 내용이 없는 계획은 적용할 것 없이 작업을 끝냅니다.

## Override backend settings {#templateForm-overrideBackend}

Semaphore가 Terraform 상태를 보관할 수 있을 때 Terraform, OpenTofu, Terragrunt 템플릿에 표시됩니다.

코드의 상태를 Semaphore에 보관합니다. `init` 전에 Semaphore는 `http` 백엔드를 설정하는 파일을 코드 디렉터리에
기록하고, 작업 후에 삭제합니다.

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

**Override backend settings**와 함께 표시되며, 그 상자가 체크되어 있을 때 활성화됩니다. 선택 사항입니다.

Semaphore가 백엔드를 기록하는 파일의 이름입니다. 예: `semaphore_backend.tf`. 기본값은 `backend.tf`입니다. 이름은
`.tf`로 끝나야 하며 문자, 숫자, `_`, `-`, `.`만 포함할 수 있습니다. 코드에 없는 파일 이름을 고르세요.
Semaphore는 그 파일을 덮어쓰고 작업 후에 삭제합니다.

## 제한 {#templateForm-allowOverrideLimit}

**Ansible 프롬프트** 아래에 있습니다. 작업을 시작할 때 사용자가 템플릿의 **제한** 대신 실행할 호스트를 설정할 수
있게 합니다.

## 태그 {#templateForm-allowOverrideTags}

**Ansible 프롬프트** 아래에 있습니다. 작업을 시작할 때 사용자가 템플릿의 **태그** 대신 태그를 설정할 수 있게
합니다.

## 태그 건너뛰기 {#templateForm-allowOverrideSkipTags}

**Ansible 프롬프트** 아래에 있습니다. 작업을 시작할 때 사용자가 템플릿의 **태그 건너뛰기** 대신 건너뛸 태그를
설정할 수 있게 합니다.

## 디버그 {#templateForm-allowDebug}

**Ansible 프롬프트** 아래에 있습니다. 작업을 시작할 때 사용자가 `-v`부터 `-vvvvvv`까지 상세 출력을 켤 수 있게
합니다.

## 드라이 런 {#templateForm-showDryRun}

**Ansible 프롬프트** 아래에 있습니다. 기본적으로 체크되어 있습니다. 작업을 시작할 때 **드라이 런** 옵션을
제공하며, 이 옵션은 `--check`와 함께 플레이북을 실행합니다. Ansible은 무엇을 변경할지 보고하고 아무것도 변경하지
않습니다. 이 옵션을 없애려면 체크를 해제하세요.

## 차이 {#templateForm-showDiff}

**Ansible 프롬프트** 아래에 있습니다. 기본적으로 체크되어 있습니다. 작업을 시작할 때 **차이** 옵션을 제공하며,
이 옵션은 `--diff`와 함께 플레이북을 실행하여 파일에 가해진 변경 내용을 보여 줍니다. 이 옵션을 없애려면 체크를
해제하세요.

## 자동 승인 {#templateForm-allowAutoApprove}

Terraform, OpenTofu, Terragrunt 프롬프트 아래에 있습니다. 작업을 시작할 때 사용자가 **Auto Approve**를 체크하여
확인 없이 계획을 적용할 수 있게 합니다. 작업 대화 상자는 어느 경우든 그 옵션을 제공하지만, 이 상자가 없으면
Semaphore는 그 옵션을 무시합니다. 템플릿의 **자동 승인**이 체크되어 있지 않으면 작업은 계획 후에 확인을
기다립니다.

## Galaxy install options {#templateForm-galaxy}

Ansible 템플릿에 표시됩니다. 클릭하면 Galaxy 설치 옵션이 표시되며, 레이블에는 기본값과 다른 옵션의 수가
표시됩니다.

플레이북이 실행되기 전에 Semaphore는 리포지토리가 요구하는 롤과 컬렉션을 `ansible-galaxy`로 설치합니다.
플레이북 옆과 리포지토리 루트에서 `requirements.yml`, `roles/requirements.yml`, `collections/requirements.yml`을
찾으며, 마지막 작업 이후 파일이 바뀌었을 때만 설치를 실행합니다.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

**Galaxy install options** 아래에 있습니다. Galaxy 설치를 건너뜁니다. 플레이북은 이전에 설치된 롤과 컬렉션으로
실행됩니다. 요구 사항이 다른 방법으로 설치되거나, 설치가 느리고 필요하지 않을 때 사용하세요.

## Role install args {#templateForm-galaxyRoleArgs}

**Galaxy install options** 아래에 있습니다. `ansible-galaxy role install`에 추가되는 인수로, 칩 하나에 인수
하나입니다. 예: `--ignore-certs` 또는 `--server=https://galaxy.example.com`. Semaphore는 전달해도 안전한 설치
플래그만 받아들입니다. 알 수 없는 플래그, `--token`, `--api-key`, `-p` 또는 `--roles-path`가 있으면 템플릿 저장이
실패합니다.

## Collection install args {#templateForm-galaxyCollectionArgs}

**Galaxy install options** 아래에 있습니다. `ansible-galaxy collection install`에 추가되는 인수로, 칩 하나에 인수
하나입니다. 예: `--pre` 또는 `--upgrade`. **Role install args**와 같은 규칙이 적용됩니다. `--token`, `--api-key`,
`-p`, `--collections-path`는 허용되지 않습니다.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

**Galaxy install options** 아래의 **Prompts**에 있습니다. 작업을 시작할 때 사용자가 Galaxy 설치를 건너뛸지 선택할
수 있게 합니다. 작업의 선택이 템플릿의 **Skip Galaxy install**보다 우선합니다.
