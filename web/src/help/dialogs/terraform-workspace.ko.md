---
form: src/components/TerraformInventoryForm.vue
---

# 새로 만들기 Terraform Workspace

워크스페이스는 Terraform, OpenTofu 또는 Terragrunt 템플릿의 인벤토리입니다. 작업이 `init` 후에 선택하는
워크스페이스와, `init`이 비공개 모듈을 내려받을 때 사용하는 키를 지정합니다. 여기에서 만든 워크스페이스는 이
템플릿에 연결되어 그 상태 위에 표시됩니다. 템플릿의 작업은 체크 표시가 된 기본 워크스페이스에서 실행되며,
**Make default**로 기본 워크스페이스를 바꿉니다. 연필 버튼으로 열면 같은 대화 상자로 워크스페이스를 수정합니다.

## Workspace name {#tfWorkspace-inventory}

Terraform 워크스페이스의 이름입니다. 예: `production`. 이 필드는 필수입니다.

계획 전에 Semaphore는 이 이름으로 `workspace select -or-create=true`를 실행하므로, 워크스페이스가 아직 없으면
만들어집니다. 예를 들어 백엔드가 워크스페이스를 지원하지 않아 `workspace list`가 실패하면 Semaphore는 선택을
건너뜁니다.

Semaphore는 워크스페이스를 템플릿 이름과 이 이름을 합친 `TEMPLATE - NAME`으로 저장합니다. 인벤토리 페이지와
워크스페이스를 템플릿에 연결할 때 이 전체 이름이 표시됩니다. 워크스페이스 이름을 바꿔도 Terraform에서는 이름이
바뀌지 않습니다. 다음 작업이 새 이름의 워크스페이스를 선택하거나 만듭니다.

## SSH key for private modules {#tfWorkspace-sshKeyId}

`init`이 SSH를 통해 비공개 Git 리포지토리에서 모듈을 내려받을 때 사용하는 키입니다. 이 대화 상자에서는 이 필드가
필수입니다. 목록에는 키 저장소의 **SSH 키**, **비밀번호로 로그인**, **없음** 유형의 키가 표시됩니다.

**SSH 키**만 효과가 있습니다. Semaphore는 `init`이 진행되는 동안 이 키를 SSH 에이전트에 불러오며, Git은
`git::ssh://git@example.com/modules/vpc.git` 같은 모듈 소스에 이 키를 사용합니다. **비밀번호로 로그인** 유형의
키는 Git에 자격 증명을 전달하지 않습니다. 모듈에 키가 필요 없으면 **없음**을 선택하세요.
