---
form: src/components/InventoryHostForm.vue
---

# 호스트 수정

관리형 인벤토리의 호스트입니다. 이 인벤토리를 사용하는 작업은 호스트 목록에서 켜져 있는 호스트에서 실행됩니다.

## 이름 {#inventoryHostForm-name}

Ansible이 호스트를 식별하는 이름이며, 변수에서 `ansible_host`를 설정하지 않으면 연결할 주소로도 쓰입니다. 이 필드는 필수입니다. 이름은 인벤토리 안에서 고유하고 대소문자를 구분합니다. `web1`과 `Web1`은 서로 다른 호스트입니다. 공백, 쉼표, 대괄호, 콜론은 쓸 수 없으며 길이는 최대 250자입니다.

## 설명 {#inventoryHostForm-description}

선택 사항입니다. 사람을 위한 메모이며 호스트 목록에 표시됩니다. Ansible에는 전달되지 않습니다.

## 변수 {#inventoryHostForm-variables}

선택 사항입니다. 이 호스트의 변수를 YAML 또는 JSON 매핑으로 작성합니다.

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

호스트의 변수는 같은 이름의 인벤토리 변수보다 우선합니다. YAML 앵커, 별칭, 병합 키는 허용되지 않습니다.
