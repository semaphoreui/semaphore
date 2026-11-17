---
form: src/components/TerraformInventoryForm.vue
---

# Новый Terraform Workspace

Workspace — это инвентарь шаблона Terraform, OpenTofu или Terragrunt: он задает workspace,
который задача выбирает после `init`, и ключ, с которым `init` скачивает приватные модули.
Workspace, созданный здесь, привязывается к этому шаблону и показывается над его состоянием.
Задачи шаблона выполняются в его workspace по умолчанию — том, что отмечен галочкой; **Make
default** меняет его. Этот же диалог изменяет workspace, когда вы открываете его кнопкой с
карандашом.

## Workspace name {#tfWorkspace-inventory}

Имя workspace Terraform, например `production`. Поле обязательное.

Перед планом Semaphore выполняет `workspace select -or-create=true` с этим именем, поэтому
workspace создается, если его еще нет. Если `workspace list` завершается с ошибкой, например
потому что backend не поддерживает workspace, Semaphore пропускает выбор.

Semaphore хранит workspace как `TEMPLATE - NAME` — имя шаблона и это имя: полное имя видно на
странице «Инвентарь» и при привязке workspace к шаблону. Переименование workspace не
переименовывает его в Terraform: следующая задача выбирает или создает workspace с новым именем.

## SSH key for private modules {#tfWorkspace-sshKeyId}

Ключ, с которым `init` скачивает модули из приватных Git-репозиториев по SSH. В этом диалоге поле
обязательное. Список предлагает ключи типа **SSH ключ**, **Логин с паролем** и **Нет** из раздела
«Хранилище ключей».

Действует только **SSH ключ**: Semaphore загружает его в SSH-агент на время `init`, и Git
использует его для источников модулей, например `git::ssh://git@example.com/modules/vpc.git`.
Ключ типа **Логин с паролем** не передает Git учетных данных. Выберите **Нет**, если модулям не
нужен ключ.
