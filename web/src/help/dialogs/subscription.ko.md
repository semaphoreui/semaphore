---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

구독 키는 이 서버에서 Semaphore PRO 및 Enterprise의 기능을 활성화합니다. 이 대화 상자는 키를 활성화하고, 활성
키의 플랜을 보여 줍니다. 보유자, 상태, 만료 시점, 그리고 허용되는 Pro 사용자, 러너 및 기타 항목의 수와 사용 중인
수입니다. 관리자가 Pro 에디션의 사용자 메뉴에서 엽니다. 구독이 활성화되기 전까지 이 대화 상자의 이름은
**Upgrade to Semaphore PRO**입니다.

## Enter your PRO or EE key {#subscription-key}

PRO 또는 Enterprise 구독과 함께 받은 구독 키입니다. 여기에 붙여 넣거나 업로드 버튼으로 파일에서 불러온 다음
**Activate New Key**를 누르세요. 활성화하려면 이 필드가 필요합니다.

키가 서버 구성(`subscription.key` 또는 `subscription.key_file`, 또는 변수 `SEMAPHORE_SUBSCRIPTION_KEY` 및
`SEMAPHORE_SUBSCRIPTION_KEY_FILE`)에 설정되어 있으면 이 필드는 읽기 전용이며, 이 경우 키는 그곳에서만 변경할 수
있습니다.

## 키 파일 업로드 {#subscription-upload}

아직 구독이 활성화되지 않았을 때 표시됩니다.

컴퓨터에 있는 파일(`.txt`, `.key`, `.pem` 또는 `.lic`)에서 키를 읽어 위의 필드에 넣습니다. **Activate New Key**를
누를 때까지 키는 활성화되지 않습니다.

## 키 작업 {#subscription-actionsMenu}

구독이 활성화되어 있거나 만료되었을 때 표시됩니다.

서버의 키에 대한 작업 메뉴입니다.

- **Reload**: 구독 상태를 다시 가져옵니다. 예: 구독을 갱신한 후.
- **Upload**: 업로드 버튼처럼, 파일에서 새 키를 읽어 위의 필드에 넣습니다.
- **Reset**: 확인 없이 즉시 서버에서 키를 제거합니다.

## Activate New Key {#subscription-activate}

필드의 키를 이 서버에서 활성화합니다. 키가 유효하면 아래에 구독 플랜이 표시됩니다. 계정이 아직 Pro 사용자가
아니고 플랜에 빈 Pro 사용자 자리가 있으면, 대화 상자는 계정을 Pro 사용자로 만들지 묻습니다.

이미 활성화된 키는 경고로 보고됩니다. 키가 서버 구성에 설정되어 있으면 이 버튼은 비활성화됩니다.

## Buy Pro {#subscription-buy}

PRO 구독을 구매하는 Semaphore 포털 페이지를 브라우저의 새 탭에서 엽니다.

## Get 30-day free trial {#subscription-trial}

활성 구독이 없는 동안 표시됩니다.

30일 무료 평가판 키를 발급하는 Semaphore 포털 페이지를 브라우저의 새 탭에서 엽니다. 받은 키를 위의 필드에 붙여
넣으세요.
