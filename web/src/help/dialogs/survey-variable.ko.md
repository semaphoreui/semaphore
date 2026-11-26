---
form: src/components/SurveyVars.vue
---

# 설문 변수

설문 변수는 템플릿을 실행할 때마다 작업 대화 상자가 묻는 질문입니다. 답은 변수로 작업에 전달됩니다. Ansible에는
추가 변수, Terraform과 OpenTofu에는 `-var`, 스크립트에는 인수, 또는 환경 변수로 전달됩니다. 변수는 템플릿 대화
상자를 저장할 때 템플릿과 함께 저장됩니다. 같은 대화 상자로 워크플로의 설문도 편집하며, 이 경우 **Secret** 유형은
제공되지 않습니다.

## 이름 {#surveyVarDialog-name}

작업이 답을 받는 이름입니다. 추가 변수, Terraform 변수, 스크립트의 `name=value` 인수 또는 환경 변수의 이름입니다.
양 끝의 공백은 제거됩니다. 이 필드는 필수입니다.

설문의 다른 변수에 없는 이름이면서 도구가 받아들이는 이름을 사용하세요. 예: `app_version`. 답은 템플릿 변수
그룹에 있는 같은 이름의 변수를 대체합니다.

## 제목 {#surveyVarDialog-title}

작업 대화 상자에 표시되는 필드의 레이블입니다. 예: `Version to deploy`. 템플릿 대화 상자에 있는 변수 칩의
텍스트이기도 합니다. 이 필드는 필수입니다.

## 설명 {#surveyVarDialog-description}

선택 사항입니다. 작업 대화 상자의 필드에 포커스가 있는 동안 그 아래에 표시되는 힌트로, 무엇을 입력할지나 예를
알려 줍니다.

## 유형 {#surveyVarDialog-type}

작업 대화 상자가 요청하는 답의 종류입니다.

- **String**: 한 줄의 텍스트입니다. 기본값입니다.
- **Integer**: 숫자만 입력할 수 있으며, 작업 대화 상자는 그 밖의 모든 것을 거부합니다. 빼기 기호도 거부됩니다.
  답은 그래도 텍스트로 작업에 전달되므로, 플레이북에서 변환하세요. 예: `| int`.
- **Secret**: 작업 대화 상자가 가려서 표시하는 텍스트입니다. 답은 작업과 함께 저장되지 않으므로 재실행할 때 다시
  묻고, Ansible은 이를 vault로 암호화된 파일로 받습니다. secret에는 기본값이 없습니다.
- **Enum**: 아래에서 정의한 목록에서 하나의 값을 선택합니다.
- **Text**: 여러 줄의 텍스트입니다.
- **Select**: 아래에서 정의한 목록에서 여러 값을 선택합니다. 작업은 이를 목록으로 받습니다. 인수, `-var` 또는
  환경 변수에서는 목록이 JSON으로 기록됩니다. 예: `["a","b"]`.

유형을 바꾸면 기본값이 지워지고, 목록이 없는 유형으로 바꾸면 값도 지워집니다.

## Pass variable as {#surveyVarDialog-target}

답이 작업에 전달되는 방식입니다.

- **Extra variable**: 도구의 일반적인 방식입니다. Ansible은 추가 변수로, Terraform과 OpenTofu는
  `-var name=value`로, Bash, PowerShell 또는 Python 스크립트는 스크립트 뒤의 `name=value` 인수로 받습니다.
- **Environment variable**: 모든 종류의 템플릿에서 프로세스의 환경 변수 `name=value`로 전달됩니다. 이 경우
  변수는 다른 방식으로는 전달되지 않습니다.

## 키 {#surveyVarDialog-valueName}

유형이 **Enum** 또는 **Select**일 때 표시됩니다.

작업 대화 상자가 목록에서 이 값에 대해 표시하는 텍스트입니다. 모든 값에는 키가 필요하며 키는 서로 달라야 합니다.
그렇지 않으면 대화 상자가 변수 저장을 거부합니다.

## 값 {#surveyVarDialog-valueValue}

유형이 **Enum** 또는 **Select**일 때 표시됩니다.

이 값을 선택했을 때 작업이 받는 것입니다. 키와 다를 수 있습니다. 키 `Production`이 `prod`를 전달할 수 있습니다.

## Add Value {#surveyVarDialog-addValue}

유형이 **Enum** 또는 **Select**일 때 표시됩니다.

값 목록에 빈 행을 추가합니다. **Enum**이나 **Select**에는 값이 하나 이상 필요합니다. 행의 휴지통 아이콘을 누르면
그 행이 제거됩니다.

## Default value {#surveyVarDialog-defaultValue}

**Secret**을 제외한 모든 유형에 표시됩니다.

작업 대화 상자가 열릴 때 채우는 답으로, 그곳에서 변경할 수 있습니다. **Enum**에서는 값 중 하나를, **Select**에서는
원하는 개수의 값을 선택하세요. 빈 답으로 시작하려면 이 필드를 비워 두세요.

기본값은 작업 대화 상자만 채웁니다. 다른 방식(예: API)으로 시작한 작업은 기본값을 받지 않습니다.

## 필수 {#surveyVarDialog-required}

변수에 답이 있을 때까지 작업 대화 상자는 작업을 시작하지 않습니다. 기본값과 마찬가지로 이 확인은 서버가 아니라
작업 대화 상자에서만 이루어집니다.
