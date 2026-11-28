---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# 실행 수정

실행은 다른 서비스가 Semaphore의 webhook URL을 호출할 때(예: GitHub의 push 이벤트) 통합이 시작하는 것으로, 작업 템플릿의 작업 또는
workflow입니다. URL과 요청 검사는 실행이 아니라 통합의 엔드포인트에 속합니다. 실행 페이지에는 해당 실행의 매처와 요청에서 가져오는 값이 있습니다.

## 이름 {#integrationForm-name}

통합 페이지에 표시되는 통합의 이름입니다. 통합이 시작한 작업은 감사 로그에서 이 통합의 것으로 기록됩니다.
이 필드는 필수입니다.

## 통합 {#integrationForm-groupId}

실행이 속한 통합입니다. 실행은 해당 통합의 엔드포인트에 도달하고 인증을 통과한 요청으로 시작됩니다. 필수 필드입니다.

## 대상 {#integrationForm-targetToggle}

서버가 워크플로를 지원할 때 표시됩니다.

통합이 시작하는 대상입니다. **Template**은 작업 템플릿의 작업을 시작하고, **Workflow**는 워크플로 실행을
시작합니다. **Workflow**로 전환하면 아래의 작업 매개변수가 삭제됩니다.

## Workflow to run {#integrationForm-workflowId}

**대상**가 **Workflow**일 때 표시됩니다.

요청이 시작하는 워크플로입니다. 통합이 요청에서 가져오는 모든 값은 같은 이름의 워크플로 설문 변수에 대한
답으로 전달됩니다. 워크플로 통합에는 Pro 에디션이 필요합니다. Pro 에디션이 없으면 이 필드는 비활성화되고
서버는 변경을 거부합니다.

## 실행할 작업 템플릿 {#integrationForm-templateId}

**대상**가 **Template**일 때 표시됩니다.

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
