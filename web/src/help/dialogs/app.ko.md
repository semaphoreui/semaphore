---
form: src/components/AppForm.vue
---

# Edit App

앱은 Semaphore가 작업을 실행할 때 사용하는 도구입니다. Ansible, Terraform, OpenTofu, Terragrunt, Bash,
Python, PowerShell 또는 직접 지정한 명령이 있습니다. 모든 작업 템플릿은 하나의 앱용으로 만들어지며, 이
대화 상자는 서버가 앱을 시작하는 방법과 인터페이스에 앱이 표시되는 방식을 설정합니다.

설정은 데이터베이스에 저장되며, 설정을 저장한 서버에서는 즉시 적용됩니다. 클러스터의 다른 노드는 다시
시작할 때 설정을 읽습니다. 비운 필드는 서버가 다시 시작될 때까지 이전 값을 유지합니다.

## ID {#appForm-id}

필수입니다. 앱이 저장되는 키로, 작업 템플릿은 이 키로 자신의 앱을 참조합니다. 기본 제공 앱은
`ansible`, `terraform`, `tofu`, `terragrunt`, `pulumi`, `bash`, `python`, `powershell`입니다. 그 밖의
ID를 지정하면 직접 만든 앱이 추가됩니다. Semaphore는 이 앱을 스크립트처럼 실행하며, **Path**에 지정한
명령을 사용하거나 **Path**가 비어 있으면 ID와 같은 이름의 명령을 사용합니다.

점이 없는 ID를 사용하세요. 기존 앱의 ID를 바꿔도 앱의 이름이 바뀌지는 않습니다. 설정은 새 ID로 새 앱으로
저장되고, 이전 앱은 그대로 남습니다.

## Icon {#appForm-icon}

선택 사항입니다. Material Design Icons 세트에 있는 아이콘의 이름으로, `mdi-` 접두사 없이 입력합니다.
예: `language-python`. 앱의 기본 제공 아이콘을 유지하려면 비워 두세요. 이 경우 직접 만든 앱에는 물음표가
표시됩니다.

## 이름 {#appForm-title}

선택 사항입니다. 애플리케이션 목록과 **새 템플릿** 메뉴에 표시되는 앱 이름입니다. 기본 제공 이름(예:
`Ansible Playbook`)을 유지하려면 비워 두세요. 이 경우 직접 만든 앱에는 이름이 없으므로 이름을 지정하세요.

## Path {#appForm-path}

선택 사항입니다. Semaphore가 기본 명령 대신 실행하는 명령으로, 서버의 `PATH`에서 찾을 이름이나 전체
경로를 지정합니다. 예: `/usr/local/bin/terraform-1.9`. 기본값은 `terraform`, `tofu`, `terragrunt`,
`bash`, `python3`, `powershell`이며, 직접 만든 앱은 ID와 같은 이름의 명령을 실행합니다.

Ansible은 이 필드를 무시합니다. Semaphore는 항상 `ansible-playbook`과 `ansible-galaxy`를 실행합니다.

## Priority {#appForm-priority}

애플리케이션 목록과 **새 템플릿** 메뉴에서 앱의 위치입니다. 숫자가 큰 앱이 먼저 표시됩니다. 기본 제공
앱은 Ansible의 1000부터 Python의 500까지입니다. 우선순위가 `0` 이하인 기본 제공 앱은 서버가 시작될 때
기본 제공 우선순위로 되돌아갑니다.

## 인수 {#appForm-args}

선택 사항입니다. 명령 바로 뒤에 명령줄에 추가되는 인수로, 칩 하나에 인수 하나입니다. 인수를 추가하려면
**+**를 클릭하고, 변경하려면 칩을 클릭하고, 순서를 바꾸려면 칩을 끌어다 놓으세요.

- Terraform, OpenTofu, Terragrunt: 이 인수는 Semaphore가 모든 명령에 추가하는 인수(예: `init`, `apply`)보다
  앞에 옵니다.
- 스크립트 앱: 이 인수는 템플릿의 스크립트보다 앞에 오며 기본 인수를 대체합니다. PowerShell은 기본적으로
  `-File`과 함께 실행되므로 목록의 마지막에 `-File`을 넣으세요.
- Ansible은 이 필드를 무시합니다.

## Active {#appForm-active}

활성 앱은 **새 템플릿** 메뉴에 표시됩니다. 그 메뉴에서 앱을 숨기려면 체크를 해제하세요. 이미 이 앱으로
만든 템플릿은 앱을 유지하며 계속 실행됩니다.

새 앱은 체크가 해제되어 있습니다. 앱을 사용할 수 있게 하려면 체크하세요.
