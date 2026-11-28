---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# 통합 편집

통합은 하나의 엔드포인트에서 시작되는 실행의 묶음입니다. 해당 엔드포인트의 URL로 온 요청이 인증을 통과하면, 매처가 요청을 받아들이는 통합의 모든 실행이 시작됩니다.

## 이름 {#integrationGroupForm-name}

'통합' 페이지에 표시되는 통합의 이름입니다. 필수 필드입니다.

## 엔드포인트 {#integrationGroupForm-endpointId}

이 통합의 실행을 시작하는 URL을 가진 엔드포인트입니다. 하나의 엔드포인트를 여러 통합이 사용할 수 있습니다. 필수 필드입니다.

## 인증 재정의 {#integrationGroupForm-authOverride}

꺼 두면 요청을 엔드포인트와 같은 방식으로 검사합니다. 켜면 아래에 나타나는 필드에서 설정한 자체 방식과 시크릿으로 이 통합의 요청을 검사하며, 이 경우 엔드포인트의 설정은 이
통합에 적용되지 않습니다.

## 인증 방식 {#integrationAuthForm-authMethod}

요청이 예상한 서비스에서 왔는지 Semaphore가 확인하는 방법입니다. 확인에 실패한 요청은 아무것도 시작하지
않으며, 웹후크는 다른 요청과 똑같이 응답합니다.

- **없음**: 확인하지 않습니다. URL을 아는 누구나 작업을 시작할 수 있습니다.
- **GitHub Webhooks**: GitHub가 보내는 것처럼, `X-Hub-Signature-256` 헤더에 시크릿으로 만든 본문의
  HMAC-SHA256 서명이 있어야 합니다.
- **Bitbucket Webhooks**: Bitbucket이 보내는 것처럼, `X-Hub-Signature` 헤더에 대해 같은 확인을 합니다.
- **Token**: **인증 헤더**에 지정한 헤더에 시크릿 자체가 있어야 합니다.
- **HMAC** 및 **HMAC (SHA-512)**: **인증 헤더**에 지정한 헤더에 시크릿으로 만든 본문의 HMAC-SHA256 또는
  HMAC-SHA512가 접두사(예: `sha256=`) 없이 소문자 16진수로 있어야 합니다.
- **BasicAuth**: 요청의 `Authorization` 헤더에 **인증 시크릿**에 지정한 키의 로그인과 비밀번호가 있어야
  합니다.

## 인증 헤더 {#integrationAuthForm-authHeader}

**인증 방식**가 **Token**, **HMAC** 또는 **HMAC (SHA-512)**일 때 표시됩니다.

토큰이나 서명을 담는 요청 헤더의 이름입니다. 예: `X-Webhook-Token`. 이름의 대소문자는 구분하지 않습니다.
이 헤더가 없는 요청은 거부됩니다.

## 인증 시크릿 {#integrationAuthForm-authSecretId}

**인증 방식**가 **없음**이 아닐 때 표시됩니다.

요청을 확인할 때 사용하는 키 저장소의 키입니다. **BasicAuth**에는 **비밀번호로 로그인** 유형의 키만
맞습니다. 다른 방법에는 **문자열** 유형의 키도 맞습니다. 확인에는 그 값이나 **비밀번호로 로그인** 키의
비밀번호가 사용됩니다.

**없음**을 제외한 모든 방법에 이 필드가 필요합니다. 키가 없거나 시크릿이 비어 있으면 모든 요청이 거부됩니다.
