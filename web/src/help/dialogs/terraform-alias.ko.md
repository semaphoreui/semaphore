---
form: src/components/TerraformAliasForm.vue
---

# 수정 Key

별칭은 Semaphore가 이 워크스페이스를 위해 제공하는 HTTP 백엔드의 주소입니다. 이 주소를 가리키는 Terraform 또는
OpenTofu는 Semaphore에 있는 워크스페이스 상태를 읽고, 쓰고, 잠그고, 잠금 해제합니다. 주소는 별칭을 만들 때
Semaphore가 생성하며, 이 대화 상자에서는 주소를 보호하는 키만 선택합니다. 별칭은 Semaphore PRO의 기능입니다.

## Auth key {#tfAlias-authKeyId}

별칭을 보호하는 키입니다. 목록에는 프로젝트 키 저장소의 **비밀번호로 로그인** 유형 키만 표시됩니다. 목록이 비어
있으면 먼저 그곳에서 키를 만드세요.

별칭의 클라이언트는 이 키의 로그인과 비밀번호를 보냅니다. 저장한 후 별칭 목록에서 주소를 복사하여, 키의 로그인 및
비밀번호와 함께 구성의 `http` 백엔드에 넣으세요.

```hcl
terraform {
  backend "http" {
    address        = "ALIAS_URL"
    lock_address   = "ALIAS_URL"
    unlock_address = "ALIAS_URL"
    username       = "LOGIN"
    password       = "PASSWORD"
  }
}
```

이 주소는 `http` 백엔드의 기본 메서드에 응답하므로, 하나의 주소로 상태, 잠금, 잠금 해제를 모두 처리합니다. 기존
별칭에 다른 키를 선택해도 주소는 유지됩니다. 이 경우 클라이언트에는 새 키의 로그인과 비밀번호가 필요합니다.
