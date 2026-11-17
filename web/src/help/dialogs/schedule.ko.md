---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# 일정 수정

일정은 누군가 **실행**을 누르지 않아도 작업 템플릿의 작업을 시작합니다. 크론 표현식에 따라 반복해서 시작하거나,
정해진 시각에 한 번 시작합니다. 작업은 이 대화 상자에 저장된 매개변수를 받습니다. 모든 시간은 시간 필드 옆에
표시되는 Semaphore 서버의 시간대 기준입니다. 관리자가 `schedule.timezone` 또는 `SEMAPHORE_SCHEDULE_TIMEZONE`을
설정하지 않았다면 `UTC`입니다.

## Name {#scheduleForm-name}

일정 페이지에 표시되는 일정 이름입니다. 이 필드는 필수입니다. 일정이 하는 일에는 영향을 주지 않습니다.

## Template {#scheduleForm-templateId}

일정이 시작하는 작업 템플릿입니다. 이 필드는 필수입니다. 작업은 템플릿에서 시작한 것처럼 템플릿 자체의 인벤토리,
리포지토리, 변수 그룹 및 기타 설정으로 실행되며, 템플릿이 작업에서 변경을 허용하는 곳에는 아래의 매개변수가
적용됩니다.

## 작업 매개변수 {#scheduleForm-taskParams}

템플릿을 선택했을 때 표시됩니다.

일정이 시작하는 모든 작업의 매개변수입니다. 어떤 필드가 나타나는지는 템플릿에 따라 다릅니다. 템플릿의 설문
변수와, 템플릿이 작업에서 변경을 허용하는 옵션(예: 브랜치, 인벤토리, **제한**, **태그**, **태그 건너뛰기**,
**CLI args**), 그리고 Ansible 또는 Terraform, OpenTofu, Terragrunt의 옵션입니다.

서버는 일정이 실행될 때마다 템플릿을 확인합니다. 템플릿이 더 이상 작업에서 변경을 허용하지 않으면 여기에 저장된
값은 무시됩니다.

## 메시지 {#taskParams-message}

선택 사항입니다. 일정이 시작하는 모든 작업과 함께 저장되어 작업 기록에 표시되는 메모입니다. 예: `Nightly deploy`.
작업이 실행되는 방식에는 영향을 주지 않습니다.

## 설문 변수 {#taskParams-survey}

템플릿에 설문 변수가 있을 때 표시됩니다.

예약된 작업에 사용할 템플릿 설문 변수의 값입니다. 누군가 직접 설문을 채울 때와 같은 방식으로 작업에 전달됩니다.
필수 변수에는 값이 있어야 일정을 저장할 수 있습니다.

대화 상자가 열릴 때 템플릿의 기본값이 복사되어 일정과 함께 저장됩니다. 나중에 템플릿에서 기본값을 바꿔도 이전에
저장된 일정에는 반영되지 않습니다.

secret 유형 변수의 값은 일정과 함께 저장되지 않습니다. 예약된 작업은 그 값 없이 실행됩니다.

## 브랜치 {#taskParams-gitBranch}

템플릿이 작업에서 브랜치 재정의를 허용할 때 표시됩니다.

예약된 작업이 체크아웃하는 리포지토리의 브랜치 또는 태그입니다. 이 필드는 템플릿의 브랜치로 시작합니다. 서버는
템플릿이 재정의를 허용하는 동안에만 이 값을 사용합니다.

## 인벤토리 {#taskParams-inventoryId}

템플릿이 작업에서 인벤토리 선택을 허용할 때 표시됩니다.

예약된 작업이 템플릿의 인벤토리 대신 대상으로 삼는 인벤토리입니다. Terraform 또는 OpenTofu 템플릿에서는 이 필드의
이름이 **Workspace**이지만, 서버는 이 선택을 Ansible 템플릿에만 적용하고 그 밖의 템플릿에는 템플릿의 워크스페이스를
유지합니다.

## 디버그 {#taskParams-debug}

디버그를 허용하는 Ansible 템플릿에 표시됩니다.

더 많은 출력과 함께 `ansible-playbook`을 실행합니다. `-v`가 **디버그 수준**에서 설정한 횟수만큼 반복됩니다.

## 디버그 수준 {#taskParams-debugLevel}

디버그를 허용하는 Ansible 템플릿에 표시되며, **디버그**가 체크되어 있을 때 활성화됩니다.

출력의 상세 수준으로, `1`(`-v`)부터 `6`(`-vvvvvv`)까지입니다. 기본값은 `4`입니다.

## 드라이 런 {#taskParams-dryRun}

템플릿이 숨기지 않는 한 Ansible 템플릿에 표시됩니다.

`--check`와 함께 플레이북을 실행합니다. Ansible은 무엇을 변경할지 보고하고 아무것도 변경하지 않습니다.

## 차이 {#taskParams-diff}

템플릿이 숨기지 않는 한 Ansible 템플릿에 표시됩니다.

`--diff`와 함께 플레이북을 실행합니다. Ansible은 파일과 템플릿에 가하는 변경 내용을 보여 줍니다.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

작업에서 변경을 허용하는 Ansible 템플릿에 표시됩니다.

플레이북을 실행하기 전에 `requirements.yml` 파일의 롤과 컬렉션 설치를 건너뜁니다. 요구 사항이 이미 설치되어 있고
작업이 Ansible Galaxy에 접근하지 않아야 할 때 체크하세요.

## Plan {#taskParams-plan}

Terraform, OpenTofu 또는 Terragrunt 템플릿에 표시됩니다.

`plan`만 실행합니다. 작업은 계획 후에 끝나며 아무것도 적용하지 않습니다.

## Destroy {#taskParams-destroy}

Terraform, OpenTofu 또는 Terragrunt 템플릿에 표시됩니다.

계획과 적용에 `-destroy`를 추가합니다. 작업은 워크스페이스의 리소스를 만들거나 업데이트하는 대신 삭제합니다.

## Auto Approve {#taskParams-autoApprove}

Terraform, OpenTofu 또는 Terragrunt 템플릿에 표시됩니다.

확인을 기다리지 않고 계획을 적용합니다. 템플릿이 자동 승인을 허용할 때만 작동하며, 모든 작업을 승인하도록 설정된
템플릿은 어느 경우든 계획을 적용합니다.

자동 승인이 없으면 예약된 작업은 계획 후에 멈추고, 누군가 작업 페이지에서 확인하거나 거부할 때까지 기다립니다.

## Upgrade {#taskParams-upgrade}

Terraform, OpenTofu 또는 Terragrunt 템플릿에 표시됩니다.

`-upgrade`와 함께 `init`을 실행합니다. 프로바이더와 모듈이 구성이 허용하는 최신 버전으로 업데이트됩니다.

## Reconfigure {#taskParams-reconfigure}

Terraform, OpenTofu 또는 Terragrunt 템플릿에 표시됩니다.

`-migrate-state` 대신 `-reconfigure`와 함께 `init`을 실행합니다. 백엔드가 새로 설정되며 기존 상태는
마이그레이션되지 않습니다.

## Run at {#scheduleForm-runAt}

한 번 실행되는 일정에 표시됩니다.

작업이 시작되는 날짜와 시간으로, 필드 끝에 표시되는 서버 시간대 기준입니다. 새 일정은 다음 정각으로 시작합니다.
시각은 미래여야 하며, 서버는 지난 시간을 거부합니다.

작업이 시작된 후 일정은 꺼지며, **Delete after run**이 체크되어 있으면 삭제됩니다. 그 시각에 Semaphore가 실행
중이 아니면 작업은 나중에 시작되지 않습니다. 다음에 시작될 때 일정은 실행 없이 꺼지거나 삭제됩니다.

## Delete after run {#scheduleForm-deleteAfterRun}

한 번 실행되는 일정에 표시됩니다.

작업이 시작된 후 일정을 삭제합니다. 일정을 꺼진 상태로 유지하여 새 시간을 설정하고 다시 켤 수 있게 하려면 체크하지
마세요.

## Show cron format {#scheduleForm-rawCron}

크론 일정에 표시됩니다.

크론 표현식을 직접 입력하는 방식(켬)과 **Timing** 및 시간 체크박스로 만드는 방식(끔) 사이를 전환합니다. 두
방식 모두 같은 표현식을 편집합니다. 브라우저는 이 선택을 기억합니다.

표현식을 체크박스로 표시할 수 없을 때는 스위치가 켜진 상태로 잠깁니다. 표현식이 유효하지 않거나 `@daily` 또는
`CRON_TZ=` 같은 형식을 사용할 때입니다.

## Cron {#scheduleForm-cronFormat}

**Show cron format**이 켜져 있을 때 크론 일정에 표시됩니다.

다섯 필드로 된 크론 표현식입니다. 분, 시, 일, 월, 요일(`0`은 일요일)입니다. 이 필드는 필수이며, 입력하는 동안
서버가 표현식을 확인합니다. 새 일정은 매분을 뜻하는 `* * * * *`로 시작합니다.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

서버는 `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly`, `@every 1h30m`과, 다른 시간대를 위한
`CRON_TZ=Europe/Berlin` 같은 접두사도 받아들입니다. 아래의 미리 보기는 이런 표현식의 다음 실행을 표시할 수
없습니다.

## Timing {#scheduleForm-timing}

**Show cron format**이 꺼져 있을 때 크론 일정에 표시됩니다.

작업이 실행되는 빈도입니다. 그에 따라 아래의 체크박스가 바뀝니다.

- **Hourly**: **Minutes**.
- **Daily**: **Hours** 및 **Minutes**.
- **Weekly**: **Weekdays**, **Hours** 및 **Minutes**.
- **Monthly**: **Days**, **Hours** 및 **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** 및 **Minutes**.

## Months {#scheduleForm-months}

**Timing**이 **Yearly**일 때 표시됩니다.

작업이 실행되는 월입니다. 여러 개를 체크하면 각 월에 실행됩니다. 아무것도 체크하지 않으면 매월 실행됩니다.

## Weekdays {#scheduleForm-weekdays}

**Timing**이 **Weekly**일 때 표시됩니다.

작업이 실행되는 요일입니다. 아무것도 체크하지 않으면 매일 실행됩니다.

## Days {#scheduleForm-days}

**Timing**이 **Monthly** 또는 **Yearly**일 때 표시됩니다.

작업이 실행되는 날짜입니다. 그 날짜가 없는 달은 건너뜁니다. 31일로 설정된 작업은 4월에 실행되지 않습니다. 아무것도
체크하지 않으면 매일 실행됩니다.

## Hours {#scheduleForm-hours}

**Timing**이 **Daily**, **Weekly**, **Monthly** 또는 **Yearly**일 때 표시됩니다.

작업이 실행되는 시로, 오른쪽에 표시된 시간대 기준입니다. 아무것도 체크하지 않으면 매시 실행됩니다.

## Minutes {#scheduleForm-minutes}

**Show cron format**이 꺼져 있을 때 크론 일정에 표시됩니다.

작업이 실행되는 분으로, 5분 단위입니다. 아무것도 체크하지 않으면 작업은 매분 실행됩니다. 한 시간에 한 번이나
하루에 한 번 실행되어야 하는 작업이라면 하나 이상(예: `:00`)을 체크하세요.

## Next run time {#scheduleForm-nextRun}

양식에 있는 현재 일정에 따라 다음 작업이 시작될 시각으로, 서버 시간대와 브라우저의 현지 시간으로 표시됩니다.
대시는 표현식이 유효하지 않거나, 미리 보기가 읽을 수 없는 표현식(예: `@daily`)이라는 뜻입니다.

## 활성화됨 {#scheduleForm-active}

일정이 작업을 시작할지 여부입니다. 일정을 실행하지 않고 유지하려면 체크를 해제하세요. 일정 목록의 스위치도 같은
역할을 합니다. 새 일정은 활성화되어 있습니다.
