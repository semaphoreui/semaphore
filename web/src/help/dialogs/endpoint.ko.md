---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# 엔드포인트 편집

엔드포인트는 webhook의 진입점으로, 다른 서비스가 호출하는 URL과 그 요청을 검사하는 방식을 가집니다. 통합은 엔드포인트에 연결되어 엔드포인트가 수락한 요청으로 실행을
시작합니다. 엔드포인트의 URL(별칭)은 '엔드포인트' 탭에서 추가합니다.

## 이름 {#endpointForm-name}

'엔드포인트' 탭과 통합 양식에 표시되는 엔드포인트의 이름입니다. 필수 필드입니다.

## 설명 {#endpointForm-description}

선택 사항입니다. 엔드포인트의 용도에 대한 메모이며 이름 옆에 표시됩니다.

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
