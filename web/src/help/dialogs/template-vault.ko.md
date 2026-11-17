---
form: src/components/TemplateVaults.vue
---

# 금고

금고는 플레이북 리포지토리에서 Ansible Vault로 암호화된 파일과 변수의 비밀번호를 Ansible에 제공합니다. 템플릿의
각 금고는 `ansible-playbook`의 `--vault-id` 옵션 하나가 됩니다. 금고는 템플릿 대화 상자를 저장할 때 템플릿과 함께
저장됩니다.

## 금고 ID {#templateVaultDialog-name}

선택 사항입니다. 비밀번호가 해당하는 vault ID로, 내용을 암호화할 때 사용한 레이블입니다
(`ansible-vault encrypt --vault-id prod@prompt`). vault ID `default`를 사용하려면 비워 두세요.

ID는 소문자로 저장됩니다. 템플릿의 각 금고에는 고유한 ID가 필요하며, 비어 있을 수 있는 것은 하나뿐입니다.

## 유형 {#templateVaultDialog-type}

비밀번호를 가져오는 곳입니다.

- **비밀번호**: 키 저장소의 키입니다. Ansible이 비밀번호를 요청하면 Semaphore가 그 키의 비밀번호를 입력하므로,
  비밀번호는 파일에 기록되지 않습니다.
- **클라이언트 스크립트**: 비밀번호를 출력하는 vault 비밀번호 클라이언트 스크립트로, Ansible이 실행합니다.

## 금고 비밀번호 {#templateVaultDialog-vaultKeyId}

유형이 **비밀번호**일 때 표시됩니다.

vault 비밀번호를 보관하는 키입니다. **비밀번호로 로그인** 유형의 키만 나열되며, Semaphore는 그 비밀번호를 사용하고
로그인은 무시합니다. 이 필드는 필수입니다.

## 스크립트 경로 {#templateVaultDialog-script}

유형이 **클라이언트 스크립트**일 때 표시됩니다.

스크립트의 경로입니다. 예: `scripts/vault-pass-client.py`. 상대 경로는 Ansible이 실행되는 디렉터리, 즉 리포지토리
루트 또는 템플릿에 작업 디렉터리가 있으면 그 디렉터리를 기준으로 합니다. 이 필드는 필수입니다.

파일 이름은 확장자 유무와 관계없이 `-client`로 끝나야 합니다. 그래야만 Ansible이 파일을 클라이언트 스크립트로
실행하고 `--vault-id`로 vault ID를 전달합니다. 스크립트는 작업을 실행하는 머신에서 Semaphore가 실행되는 사용자로
실행됩니다.
