---
form: src/components/InventoryGroupForm.vue
---

# 그룹 수정

관리형 인벤토리의 호스트 그룹입니다. 플레이북은 그룹 이름으로 호스트의 일부에서 실행되며, 그룹의 변수는 그룹의 모든 호스트에 한 번에 설정됩니다. 호스트는 Hosts 탭, 호스트 대화 상자 또는 선택한 호스트에 대한 작업으로 그룹에 넣습니다.

## 이름 {#inventoryGroupForm-name}

Ansible이 그룹을 식별하는 이름이며, 예를 들어 플레이북의 `hosts: web`에 쓰입니다. 이 필드는 필수입니다. 이름은 인벤토리 안에서 고유하고 대소문자를 구분합니다. `web`과 `Web`은 서로 다른 그룹입니다. 공백, 쉼표, 대괄호, 콜론은 쓸 수 없고 길이는 최대 250자이며 `all`과 `ungrouped`는 쓸 수 없습니다. 문자, 숫자, 밑줄 이외의 문자가 있는 이름(예: `web-eu`)에 대해 Ansible이 경고하지만 이름은 그대로 동작합니다. `*`, `?`, `!`, `&`, `~`, `@` 문자도 쓸 수 없습니다. Ansible이 이를 패턴으로 읽기 때문입니다.

## 설명 {#inventoryGroupForm-description}

선택 사항입니다. 사람을 위한 메모이며 그룹 목록에 표시됩니다. Ansible에는 전달되지 않습니다.

## Child groups {#inventoryGroupForm-children}

선택 사항입니다. 이 그룹 안에 있는 그룹입니다. 하위 그룹의 호스트는 이 그룹에도 속합니다. 그룹은 자기 자신의 상위 그룹이 될 수 없습니다.

## 변수 {#inventoryGroupForm-variables}

선택 사항입니다. 이 그룹의 변수를 YAML 또는 JSON 매핑으로 작성합니다.

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

그룹의 모든 호스트가 이 변수를 받습니다. 호스트의 변수는 같은 이름의 그룹 변수보다 우선하고, 그룹의 변수는 인벤토리 변수보다 우선합니다. YAML 앵커, 별칭, 병합 키는 허용되지 않습니다.
