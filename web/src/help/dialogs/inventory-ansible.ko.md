---
form: src/components/InventoryForm.vue
---

# 수정 Ansible Inventory

인벤토리는 Ansible이 플레이북을 실행할 호스트를 알려 주고, Semaphore가 그 호스트에 연결할 때 사용하는
자격 증명을 보관합니다. 작업 템플릿은 인벤토리를 참조하며, 해당 템플릿의 모든 작업은 그 인벤토리를
대상으로 실행됩니다. 인벤토리 형식은
[Ansible 인벤토리 가이드](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html)에
설명되어 있습니다.

## 이름 {#inventoryForm-name}

인벤토리 페이지와 작업 템플릿의 인벤토리 선택 목록에 표시되는 인벤토리 이름입니다. 호스트를
구분할 수 있는 이름을 지정하세요. 예: `Production web servers`.

## 러너 태그 {#inventoryForm-runnerTag}

프로젝트 러너를 사용할 수 있을 때 표시됩니다.

이 인벤토리를 사용하는 작업은 이 태그가 있는 러너에서 실행됩니다. 작업 템플릿에 설정된 러너 태그가
인벤토리의 태그보다 우선합니다. 어떤 러너든 작업을 가져갈 수 있게 하려면 이 필드를 비워 두세요.

## 사용자 자격 증명 {#inventoryForm-sshKeyId}

Ansible이 호스트에 로그인할 때 사용하는 키 저장소의 키입니다. 이 필드는 필수이며 세 가지 유형의
키를 사용할 수 있습니다.

- **SSH 키**: Ansible은 개인 키와, 키에 로그인이 있는 경우 그 로그인으로 연결합니다.
- **비밀번호로 로그인**: Ansible은 로그인과 비밀번호로 연결합니다.
- **없음**: Semaphore는 자격 증명을 전달하지 않습니다. `localhost`처럼 자격 증명이 필요 없는
  호스트이거나, 인벤토리 자체에 연결 방법이 정의되어 있을 때 사용합니다.

## Sudo 자격 증명 {#inventoryForm-becomeKeyId}

선택 사항입니다. 키 저장소에 있는 **비밀번호로 로그인** 유형의 키로, 플레이가 `become`으로 권한을
상승할 때 사용됩니다. 키의 로그인은 Ansible이 전환할 사용자이고, 키의 비밀번호는 Ansible이
프롬프트에 응답할 때 입력하는 비밀번호입니다.

플레이가 권한을 상승하지 않거나, 위의 사용자가 권한 상승에 비밀번호가 필요 없는 경우에는 이 필드를
비워 두세요.

## 유형 {#inventoryForm-type}

인벤토리를 가져오는 위치입니다.

- **Static**: 인벤토리를 이 대화 상자에 INI 형식으로 입력합니다.
- **Static YAML**: 인벤토리를 이 대화 상자에 YAML 형식으로 입력합니다. 인벤토리 플러그인 구성에도
  사용합니다.
- **File**: 인벤토리는 Git 리포지토리에 있는 파일입니다.
- **Managed**: 호스트와 그 변수를 레코드로 저장합니다. 수정하려면 목록에서 인벤토리를 엽니다.

유형에 따라 그 아래의 필드가 바뀝니다.

## 인벤토리 파일 경로 {#inventoryForm-inventoryPath}

유형이 **File**일 때 표시됩니다.

리포지토리 루트를 기준으로 한 인벤토리 파일 경로입니다. 예: `inventory/production.yml`.
리포지토리는 리포지토리 필드에서 선택한 것이며, 그 필드가 비어 있으면 이 인벤토리를 사용하는 작업
템플릿의 리포지토리입니다.

## 리포지토리 {#inventoryForm-repositoryId}

유형이 **File**일 때 표시됩니다. 선택 사항입니다.

인벤토리 파일이 들어 있는 리포지토리입니다. Semaphore는 작업을 실행할 때마다 리포지토리 사본을
업데이트합니다. 파일이 플레이북과 같은 리포지토리에 있으면 이 필드를 비워 두세요. 이 경우 경로는
작업 템플릿의 리포지토리에서 찾습니다.

## 인벤토리 {#inventoryForm-inventoryCode}

유형이 **Static** 또는 **Static YAML**일 때 표시됩니다.

인벤토리의 텍스트입니다. Semaphore는 이를 저장하고 작업마다 임시 파일에 기록합니다. INI 형식의
정적 인벤토리 예:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

같은 호스트를 **Static YAML**로 작성한 예:

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
