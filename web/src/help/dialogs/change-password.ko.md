---
form: src/components/ChangePasswordForm.vue
---

# 비밀번호 변경

사용자 이름과 비밀번호로 Semaphore에 로그인할 때 사용할 새 비밀번호를 설정합니다. 새 비밀번호는
**저장**을 누르면 적용됩니다. LDAP 또는 OpenID Connect로 로그인하는 외부 사용자에게는 여기에서 변경할
비밀번호가 없습니다.

## Current password {#changePassword-currentPassword}

자신의 비밀번호를 변경할 때 표시됩니다.

현재 사용 중인 비밀번호입니다. 이 필드는 필수입니다. Semaphore는 새 비밀번호를 설정하기 전에 이를
확인하므로, 탈취된 세션이나 API 토큰으로 비밀번호를 바꿀 수 없습니다. 비밀번호가 일치하지 않으면 서버는
변경을 거부합니다.

다른 사용자의 비밀번호를 재설정하는 관리자나 사용자 관리자에게는 이 필드가 표시되지 않습니다.

## 비밀번호 {#changePassword-password}

새 비밀번호입니다. 이 필드는 필수입니다. 입력한 내용을 보려면 눈 아이콘을 클릭하세요.

Semaphore는 비밀번호의 해시만 저장합니다.
