---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

매처는 들어오는 요청에 대한 조건입니다. 매처는 프로젝트 별칭으로 들어온 요청이 어떤 통합을 시작할지
결정하므로, 하나의 URL로 여러 통합을 처리할 수 있습니다. 통합은 모든 매처가 충족될 때만 실행됩니다.
매처는 통합에 **프로젝트 별칭 사용**이 켜져 있을 때만 사용되며, 이 옵션이 켜져 있으면 통합에는 프로젝트
별칭으로만 접근할 수 있습니다.

프로젝트 별칭으로 찾은 통합에 매처가 없으면 절대 실행되지 않습니다. 매처를 하나 이상 추가하세요.

## Name {#integrationMatcherForm-name}

통합 페이지에 표시되는 매처의 이름입니다. 매처가 확인하는 내용은 바뀌지 않습니다. 이 필드는 필수입니다.

## Match on {#integrationMatcherForm-matchType}

매처가 읽는 요청의 부분입니다.

- **Header**: 요청 헤더의 값입니다.
- **Body**: 요청의 본문입니다.

## Body Data Format {#integrationMatcherForm-bodyDataType}

**Match on**이 **Body**일 때 표시됩니다.

본문을 읽는 방법입니다.

- **JSON**: 본문을 JSON으로 파싱하며, **Key**는 그 안의 한 값에 대한 경로입니다.
- **String**: 본문 전체를 텍스트로 비교하며, **Key**는 사용되지 않습니다.

## Key {#integrationMatcherForm-key}

요청에서 읽을 대상입니다. 이 필드는 필수입니다.

- **Header**의 경우: 헤더의 이름입니다. 예: `X-GitHub-Event`. 대소문자는 구분하지 않습니다.
- **JSON** 형식의 **Body**의 경우: 값에 대한 경로로, 각 부분을 점으로 연결합니다. 배열의 항목은 `.[0]`처럼
  씁니다. 예: `repository.full_name` 또는 `commits.[0].id`.
- **String** 형식의 **Body**의 경우: 사용되지 않지만 양식에 값이 필요하므로 아무 값이나 입력하세요.

## Comparison Method {#integrationMatcherForm-method}

요청에서 읽은 값을 **Value**와 비교하는 방법입니다.

- **==**: 두 값이 같습니다.
- **!=**: 두 값이 다릅니다. 요청에 없는 헤더나 키는 다른 것으로 간주됩니다.
- **Contains**: 요청에서 읽은 값에 **Value**가 포함되어 있습니다.

## Value {#integrationMatcherForm-value}

비교할 텍스트입니다. 예: `push` 또는 `refs/heads/main`. 비교는 정확히 일치해야 하며 대소문자를 구분합니다.
JSON 본문의 숫자나 불리언은 텍스트 형태로 비교됩니다: `42`, `true`. 이 필드는 필수입니다.
