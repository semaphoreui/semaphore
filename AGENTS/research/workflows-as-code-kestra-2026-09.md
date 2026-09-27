# Workflows-as-Code: Kestra YAML → декларативные Workflows для Semaphore UI

Research MCP: `RESEARCH@bc356e4b9b` (group `GROUP@455e5d3336` «Semaphore UI workflows»), 2026-09-27.
Области: `AREA@f6ebec84b6` (Kestra flow YAML), `AREA@23588b75a0` (доставка YAML и EE),
`AREA@08c6718271` (конкуренты: AWX, GHA, GitLab, Argo, Windmill, Tekton), `AREA@0550459d15`
(Enterprise-требования). Полный текст с цитатами — в research MCP; это дайджест.

## Вывод

Запрос «YAML как в Kestra» = Git как источник истины + валидация в CI + аудит ревизий +
секреты вне файла + один редактор (граф ⇄ текст). Движок Workflows менять не нужно:
YAML — вторая проекция существующей модели (task/approval/delay/note, on_success/on_failure/always,
convergence all/any).

## Формат `semaphore/v1 Workflow`

- `apiVersion: semaphore/v1`, `kind: Workflow`, `metadata.name` — ключ upsert.
- `nodes[]` с человекочитаемым `id` (slug), `kind` (task по умолчанию), `template` по имени,
  `params` (= task_params, inventory по имени), `approval {timeout, message}`, `delay`, `text`.
- Зависимости на потомке: `depends_on: [build, {node: deploy, on: failure}]`; `converge: any|all`.
- Без координат, без числовых id БД, без секретов. Каноническая сериализация (топологический порядок).

## Архитектура

1. `pkg/workflowspec`: Parse / Render / Resolve (имена → id, placeholder-ID узлов) / Unresolve / JSON Schema.
2. API: `POST workflows/validate`, `GET/PUT workflows/{id}/source`, `POST workflows/apply?dry_run`,
   `GET/POST .../revisions`, `GET /api/schemas/workflow/v1.json`.
3. Таблица `project__workflow_revision` (yaml, hash, actor, origin ui|api|cli|git, commit); run хранит revision.
4. Вкладка Source в редакторе: сервер рендерит/валидирует, ошибки маппятся на узлы по `id`.
5. CLI `semaphore workflow validate|diff|apply|export` в API-режиме (`--server/--token`).
6. Managed by Git (Pro): repository + path/каталог, sync по schedule/webhook, `keep|delete|fail`,
   protected workflows, UI read-only + drift по hash.

## Фазы

1. Формат + Schema + validate/source/apply + Source-вкладка + ревизии + CLI + docs (миграция с Kestra).
2. Managed-by-Git, draft-ревизии, `workflow_template_id` у Schedule/Integration + секция `triggers`.
3. `inputs` (SurveyVar), `retry`/`timeout` на узле, sub-workflow, артефакты через remote runners,
   конвертер kestra → semaphore.

## Открытые вопросы

Уникальность имён шаблонов (slug у template?); git_managed на workflow или на проекте;
CLI API-режим; draft в v1 или v2.

## Gotcha

Fetch research MCP не скачал много страниц kestra.io / docs.github.com / docs.gitlab.com /
docs.ansible.com; непокрытые темы перечислены в разделе «Покрытие источников» research body.
