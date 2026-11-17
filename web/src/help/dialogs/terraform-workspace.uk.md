---
form: src/components/TerraformInventoryForm.vue
---

# Новий Terraform Workspace

Workspace — це інвентар шаблону Terraform, OpenTofu або Terragrunt: він указує workspace, який
завдання вибирає після `init`, і ключ, з яким `init` завантажує приватні модулі. Workspace,
створений тут, прикріплюється до цього шаблону і відображається над його станом. Завдання шаблону
виконуються в його workspace за замовчуванням, позначеному галочкою; **Make default** змінює
його. Той самий діалог редагує workspace, коли ви відкриваєте його кнопкою з олівцем.

## Workspace name {#tfWorkspace-inventory}

Назва workspace Terraform, наприклад `production`. Поле обов’язкове.

Перед планом Semaphore запускає `workspace select -or-create=true` з цією назвою, тож workspace
створюється, якщо його ще немає. Коли `workspace list` завершується з помилкою, наприклад тому, що
бекенд не підтримує workspace, Semaphore пропускає вибір.

Semaphore зберігає workspace як `TEMPLATE - NAME`, тобто назву шаблону і цю назву: це повне ім’я
ви бачите на сторінці «Інвентар» і коли прикріплюєте workspace до шаблону. Перейменування
workspace не перейменовує його в Terraform: наступне завдання вибирає або створює workspace з
новою назвою.

## SSH key for private modules {#tfWorkspace-sshKeyId}

Ключ, який `init` використовує для завантаження модулів із приватних Git-репозиторіїв через SSH.
У цьому діалозі поле обов’язкове. Список пропонує ключі типу **SSH-ключ**, **Логін з паролем** і
**Немає** зі «Сховища ключів».

Діє лише **SSH-ключ**: Semaphore завантажує його в SSH-агент на час `init`, і Git використовує
його для джерел модулів на зразок `git::ssh://git@example.com/modules/vpc.git`. Ключ типу
**Логін з паролем** не передає Git жодних облікових даних. Виберіть **Немає**, коли модулям не
потрібен ключ.
