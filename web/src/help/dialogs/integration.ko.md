---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# 통합 수정

통합은 다른 서비스가 Semaphore의 웹후크 URL을 호출할 때(예: GitHub의 푸시 이벤트) 작업 템플릿이나 워크플로를
시작합니다. URL은 별칭으로, `/api/integrations/`와 임의의 코드로 이루어집니다. 프로젝트의 별칭은 통합
페이지에 있고, 개별 통합의 별칭은 그 통합의 페이지에 있습니다. 그 페이지에는 매처와 통합이 요청에서
가져오는 값도 있습니다.

## Name {#integrationForm-name}

통합 페이지에 표시되는 통합의 이름입니다. 통합이 시작한 작업은 감사 로그에서 이 통합의 것으로 기록됩니다.
이 필드는 필수입니다.

## Runs {#integrationForm-targetToggle}

서버가 워크플로를 지원할 때 표시됩니다.

통합이 시작하는 대상입니다. **Template**은 작업 템플릿의 작업을 시작하고, **Workflow**는 워크플로 실행을
시작합니다. **Workflow**로 전환하면 아래의 작업 매개변수가 삭제됩니다.

## Workflow to run {#integrationForm-workflowId}

**Runs**가 **Workflow**일 때 표시됩니다.

요청이 시작하는 워크플로입니다. 통합이 요청에서 가져오는 모든 값은 같은 이름의 워크플로 설문 변수에 대한
답으로 전달됩니다. 워크플로 통합에는 Pro 에디션이 필요합니다. Pro 에디션이 없으면 이 필드는 비활성화되고
서버는 변경을 거부합니다.

## Task Template to run {#integrationForm-templateId}

**Runs**가 **Template**일 때 표시됩니다.

요청이 작업을 시작할 작업 템플릿입니다. 이 필드는 필수입니다. 템플릿을 선택하면 그 아래에 해당 작업의
매개변수가 나타납니다.

## 작업 매개변수 {#integrationForm-taskParams}

작업 템플릿을 선택했을 때 표시됩니다.

이 통합이 시작하는 모든 작업이 받는 매개변수로, 템플릿에서 직접 작업을 시작한 것과 같습니다. 템플릿이
작업에서 재정의를 허용하는 옵션만 여기에 있습니다. 요청에서 가져온 값이 이 값보다 우선합니다. **Variables**
유형의 값은 같은 이름의 답을, **Task Params** 유형의 값은 같은 이름의 매개변수를 재정의합니다.

템플릿이 작업에서 CLI 인수를 허용하면 **CLI args**에 작업 명령줄에 추가할 인수를 입력합니다.

## 메시지 {#taskParams-message}

선택 사항입니다. 통합이 시작하는 모든 작업의 메시지로, 작업 기록에 표시됩니다.

## 설문 변수 {#taskParams-survey}

템플릿에 설문 변수가 있을 때 표시됩니다.

템플릿 설문에 대한 답으로, 모든 작업에 동일하게 사용됩니다. 요청에서 가져온 **Variables** 유형의 값은 같은
이름의 답을 대체합니다.

secret 유형 변수의 답은 통합과 함께 저장되지 않습니다. 비밀 값은 대신 템플릿의 변수 그룹을 통해
전달하세요.

## 브랜치 {#taskParams-gitBranch}

템플릿이 작업에서 브랜치 재정의를 허용할 때 표시됩니다.

작업이 체크아웃하는 리포지토리의 브랜치 또는 태그입니다. 템플릿의 브랜치를 사용하려면 그대로 두세요.
요청에서 가져온 **Task Params** 유형의 `git_branch`라는 값이 이 값을 재정의하며, 그 값의 앞에 붙은
`refs/heads/`는 제거됩니다.

## 인벤토리 {#taskParams-inventoryId}

템플릿이 작업에서 인벤토리 재정의를 허용할 때 표시됩니다.

작업이 템플릿의 인벤토리 대신 대상으로 삼는 인벤토리이며, Terraform의 경우 워크스페이스입니다.

## Ansible 옵션 {#taskParams-ansible}

템플릿이 Ansible 플레이북을 실행할 때 표시됩니다.

모든 작업에 적용되는 `ansible-playbook` 옵션으로, 각 옵션은 템플릿이 허용할 때 표시됩니다. **제한**,
**태그**, **태그 건너뛰기**, **디버그**, **드라이 런**(`--check`), **차이**, **Skip Galaxy install**입니다.

## Terraform 옵션 {#taskParams-terraform}

템플릿이 Terraform, OpenTofu 또는 Terragrunt를 실행할 때 표시됩니다.

- **Plan**: 작업은 계획만 실행하고 멈추며, 아무것도 적용되지 않습니다.
- **Destroy**: 계획과 적용이 리소스를 삭제합니다.
- **Auto Approve**: 템플릿이 자동 승인을 허용하면 계획이 확인 없이 적용됩니다. 이 옵션이 없으면 작업은
  누군가 Semaphore에서 확인할 때까지 기다립니다.
- **Upgrade** 및 **Reconfigure**: `init`이 `-upgrade` 및 `-reconfigure`와 함께 실행됩니다.

## Auth method {#integrationForm-authMethod}

요청이 예상한 서비스에서 왔는지 Semaphore가 확인하는 방법입니다. 확인에 실패한 요청은 아무것도 시작하지
않으며, 웹후크는 다른 요청과 똑같이 응답합니다.

- **None**: 확인하지 않습니다. URL을 아는 누구나 작업을 시작할 수 있습니다.
- **GitHub Webhooks**: GitHub가 보내는 것처럼, `X-Hub-Signature-256` 헤더에 시크릿으로 만든 본문의
  HMAC-SHA256 서명이 있어야 합니다.
- **Bitbucket Webhooks**: Bitbucket이 보내는 것처럼, `X-Hub-Signature` 헤더에 대해 같은 확인을 합니다.
- **Token**: **Auth header**에 지정한 헤더에 시크릿 자체가 있어야 합니다.
- **HMAC** 및 **HMAC (SHA-512)**: **Auth header**에 지정한 헤더에 시크릿으로 만든 본문의 HMAC-SHA256 또는
  HMAC-SHA512가 접두사(예: `sha256=`) 없이 소문자 16진수로 있어야 합니다.
- **BasicAuth**: 요청의 `Authorization` 헤더에 **인증 시크릿**에 지정한 키의 로그인과 비밀번호가 있어야
  합니다.

## Auth header {#integrationForm-authHeader}

**Auth method**가 **Token**, **HMAC** 또는 **HMAC (SHA-512)**일 때 표시됩니다.

토큰이나 서명을 담는 요청 헤더의 이름입니다. 예: `X-Webhook-Token`. 이름의 대소문자는 구분하지 않습니다.
이 헤더가 없는 요청은 거부됩니다.

## 인증 시크릿 {#integrationForm-authSecretId}

**Auth method**가 **None**이 아닐 때 표시됩니다.

요청을 확인할 때 사용하는 키 저장소의 키입니다. **BasicAuth**에는 **비밀번호로 로그인** 유형의 키만
맞습니다. 다른 방법에는 **문자열** 유형의 키도 맞습니다. 확인에는 그 값이나 **비밀번호로 로그인** 키의
비밀번호가 사용됩니다.

**None**을 제외한 모든 방법에 이 필드가 필요합니다. 키가 없거나 시크릿이 비어 있으면 모든 요청이 거부됩니다.
