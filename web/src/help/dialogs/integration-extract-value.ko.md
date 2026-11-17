---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

추출된 값은 들어오는 요청의 일부(예: 푸시 이벤트의 브랜치)를 가져와 통합이 시작하는 작업에 전달합니다.
통합의 모든 값은 작업을 시작하는 각 요청에서 가져옵니다.

## Name {#integrationValueForm-name}

통합 페이지에 표시되는 값의 이름입니다. 무엇을 가져오는지나 어디로 전달되는지는 바뀌지 않습니다. 이
필드는 필수입니다.

## Source of the Value {#integrationValueForm-valueSource}

값을 가져오는 요청의 부분입니다.

- **Header**: 요청 헤더의 값입니다.
- **Body**: 요청의 본문입니다.

## Data Type of Body {#integrationValueForm-bodyDataType}

**Source of the Value**가 **Body**일 때 표시됩니다.

본문을 읽는 방법입니다.

- **JSON**: 본문을 JSON으로 파싱하며, **Key**는 그 안의 한 값에 대한 경로입니다. 값은 JSON 유형을
  유지합니다. 숫자, 목록, 객체는 그대로 작업에 전달됩니다.
- **String**: 본문 전체를 텍스트로 가져오며, **Key**는 사용되지 않습니다.

## Key {#integrationValueForm-key}

요청에서 가져올 대상입니다. 이 필드는 필수입니다.

- **Header**의 경우: 헤더의 이름입니다. 예: `X-GitHub-Event`. 대소문자는 구분하지 않습니다. 요청에 없는
  헤더는 빈 값이 됩니다.
- **JSON** 형식의 **Body**의 경우: 값에 대한 경로로, 각 부분을 점으로 연결합니다. 배열의 항목은 `.[0]`처럼
  씁니다. 예: `repository.full_name` 또는 `commits.[0].id`. 본문에 없는 경로는 값을 만들지 않으며, 작업은
  통합에 설정된 값을 유지합니다.
- **String** 형식의 **Body**의 경우: 사용되지 않지만 양식에 값이 필요하므로 아무 값이나 입력하세요.

## Variable Usage {#integrationValueForm-variableType}

템플릿의 작업에서 값이 들어가는 위치입니다.

- **Variables**: 설문 변수의 답이 전달되는 방식과 같이, 작업의 추가 변수가 됩니다. 같은 이름의 답이나 변수
  그룹의 변수를 대체합니다.
- **Task Params**: 작업 대화 상자의 옵션이 저장되는 방식과 같이, 작업의 매개변수가 됩니다. 템플릿이
  브랜치 재정의를 허용하면 이름 `git_branch`는 작업의 브랜치를 설정합니다. 앞에 붙은 `refs/heads/`는
  제거됩니다.

워크플로를 실행하는 통합은 이 용도를 무시합니다. 모든 값은 **Variable**에 지정된 워크플로 설문 변수의 답이
됩니다.

## Variable {#integrationValueForm-variable}

값이 전달될 때의 이름입니다. 예: 플레이북이 `{{ branch }}`로 읽는 추가 변수라면 `branch`. 이 필드는
필수입니다.
