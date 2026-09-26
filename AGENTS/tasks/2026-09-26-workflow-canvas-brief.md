# Workflow editor redesign on Drawflow — proposed brief for TASK@88179c9bc1

Workbench: task `TASK@88179c9bc1` (workspace «Semaphore UI»), research `RESEARCH@5bf4fc406f`.
The workbench refuses agent edits to a human-set brief, so the proposed fields live here.
Copy them into the task's context / constraints / criteria if you agree.

## Context

**Где мы сейчас.** Стек web/: Vue 2.7.16 (в package.json `^2.6.14`), Vuetify 2.7.2, vue-cli 5,
mocha + @vue/test-utils 1.x, eslint airbnb. Composition API в кодовой базе не используется.
Текущий канвас — `web/src/components/WorkflowGraph.vue` (741 строк), обёртка над **Drawflow 0.0.60**:
узлы рендерятся HTML-строками (`nodeHtml()` + ручной `escapeHtml`), канвас является источником
истины, модель экспортируется обходом внутренностей Drawflow. Страницы:
`web/src/views/project/WorkflowEditor.vue` (849 строк: палитра, панель свойств, Problems),
`WorkflowRun.vue` (341: read-only граф, поллинг 5 с, approvals), `Workflows.vue` (307: список),
`workflow/WorkflowRuns.vue`. Чистая логика уже вынесена в `web/src/lib/workflowGraph.js` и
`web/src/lib/workflowLayout.js` с тестами в `web/tests/unit/lib/`.

**Модель данных.** `db/Workflow.go:49` — `WorkflowNode {kind task|approval|delay|note, template_id,
task_params, convergence_mode, approval_timeout/message, delay_seconds, note, position_x, position_y}`,
`WorkflowEdge {source_node_id, destination_node_id, condition on_success|on_failure|always}`.
Валидация — `db.ValidateWorkflowTemplate`, клиентское зеркало — computed `problems` в
WorkflowEditor.vue. Бэкенд менять не нужно.

**Почему в прошлый раз взяли Drawflow:** `AGENTS/plans/2_19/graphical-workflow-editor.md:297`
(решение D4: Vue 2 + маленький размер).

**Docs:** `docs/docs/user-guide/workflows.md` (раздел Creating a workflow описывает текущие
взаимодействия) + `docs/static/assets/workflow-editor.webp`. Docs — сабмодуль с 10 локалями,
правила в `docs/CLAUDE.md`. i18n-ключи `workflow*` есть только в `web/src/lang/en.js`.

**Что не так с Drawflow.** Не поддерживается (последний релиз 2024-09, 273 открытых issue).
Нет fit view («Reset zoom» лишь сбрасывает zoom=1, граф уезжает за экран), minimap, undo/redo,
snap, multi-select, горячих клавиш, подписей на рёбрах, контекстного меню. Внутри узла нельзя
разместить Vue/Vuetify-компонент. Read-only режим держится на хаке mousedown/mouseup. Цвета
захардкожены, тёмная тема сломана. Компонент не тестируется.

**Исследование** — `RESEARCH@5bf4fc406f`. n8n и Kestra: Vue 3 + @vue-flow/core + dagre;
Airflow 3 и Dify: React Flow + elkjs; Windmill: Svelte Flow; Prefect: собственный рендерер.
Vue Flow работает только на Vue 3; собственный канвас отклонён заказчиком 2026-09-26.
**Решение: Drawflow 0.0.60 остаётся движком, переделывается визуальный слой и UX по образцу n8n.**
Что Drawflow умеет и чем пользуемся: `addNode` принимает пустой mount-point, в который мы сами
монтируем Vue-компонент карточки (`parent: this` → Vuetify/i18n наследуются); события
`nodeMoved`, `mouseMove`, `translate`, `zoom`, `connectionCreated/Removed/Selected`, `nodeSelected`,
`click`, `keydown`; `curvature` настраивается; `zoom`, `canvas_x/y`, `zoom_last_value` можно
выставлять напрямую (так уже делает `zoomReset`). Чего нет и что принимаем как ограничение:
multi-select/marquee, minimap, resize у note.

**UX-паттерны для переноса (n8n / Kestra / Airflow):** fit view при открытии; панель управления
слева внизу (zoom −/+, fit, tidy up); «+» на выходном порте узла → быстрый выбор следующего узла
с автопозицией и автосоединением; Tidy up (своя раскладка); snap к сетке; горячие
клавиши (Delete, Ctrl+Z / Ctrl+Shift+Z, Escape);
условие ребра как пилюля на его середине; анимированные рёбра активных путей в run view;
approve/reject прямо на узле; бейдж проблемы на узле; предупреждение о несохранённых изменениях.

## Constraints

**Всегда**
- Vue 2.7 Options API, Vuetify 2, стиль как в `web/src` (eslint airbnb, BEM `Component__el--mod`,
  тексты через `$t()` с ключами в `web/src/lang/en.js`).
- Узлы, рёбра, метки — Vue-компоненты. Никакого `v-html` / `innerHTML` / HTML-строк с данными workflow.
- Тёмная тема через Vuetify theme (`$vuetify.theme.dark`, `--v-*`, `theme--dark`), без захардкоженных фонов.
- Модель API без изменений; бэкенд, БД и `pro_impl` не трогать — фича HA-нейтральна.
- Drawflow остаётся движком; вся визуальная часть — в наших компонентах поверх него
  (`web/src/components/workflow/*`). Один `WorkflowGraph.vue` для редактора и run view.
- Обратная совместимость: сохранённые workflow открываются с теми же позициями; авто-раскладка
  только при всех нулевых позициях (`needsAutoLayout`).
- Права как сейчас: `manageProjectResources` для редактирования, `runProjectTasks` для approvals/stop.
- Чистая логика в `web/src/lib/*.js` со spec в `web/tests/unit/lib/`; компоненты через `@vue/test-utils`.
- Обновить `docs/docs/user-guide/workflows.md` и скриншот по правилам `docs/CLAUDE.md`.

**Сначала спросить**
- Любая новая npm-зависимость (план обходится без новых; dagre не нужен — раскладка своя).
- Патчи самого Drawflow (fork / patch-package) — только если без них не обойтись.
- Любое изменение API/БД (например, хранение viewport).
- Второй рантайм Vue 3 внутри приложения.
- Удаление палитры (quick-add через «+» только дополняет её).

**Никогда**
- Не писать собственный канвас и не подключать `@vue-flow/core` / `@vue2-flow/core` / Rete.js.
- Не удалять и не ослаблять `workflowGraph.spec.js`, `workflowLayout.spec.js`.
- Не менять смысл клиентской валидации — она зеркало `db.ValidateWorkflowTemplate`.
- Не делать автосохранение: Save остаётся явным.

## Criteria

1. В `web/src/components/WorkflowGraph.vue` нет `nodeHtml`, `innerHTML`, `outerHTML` и
   `escapeHtml`; узлы — смонтированные Vue-компоненты; в `web/` зелёные `npm run lint`,
   `npm run build`, `npm run test:unit`.
2. Новые spec в `web/tests/unit/lib/` покрывают: fit view (bbox с отступом при zoom 1 и 0.5),
   zoom к курсору (точка под курсором неподвижна), snap к сетке 20 px, undo→redo round-trip
   на 20 шагах, tidy up тестового DAG (4 колонки, без наложений).
3. Редактор (чек-лист в PR): узел добавляется через «+» на выходном порте (соединён, стоит справа),
   drag из палитры и кнопкой «+» на канвасе; ребро — перетаскиванием; клик по пилюле ребра меняет
   condition; Delete удаляет выделенное; Ctrl/Cmd+Z и Ctrl/Cmd+Shift+Z на 20 шагов; Escape снимает
   выделение и закрывает панель свойств.
4. Канвас: при открытии граф целиком в кадре; колесо — pan, Ctrl/Cmd+колесо — zoom к курсору;
   панель слева внизу: zoom −/+, fit, tidy up; snap 20 px; два существующих workflow демо-стенда
   открываются с прежними позициями.
5. Run view на том же канвасе: статусы узлов, анимированные рёбра активных путей, live countdown на
   delay, клик по task-узлу открывает лог, approve/reject на approval-узле; поллинг не сбрасывает pan/zoom.
6. Тёмная тема читаема; к PR приложены скриншоты light и dark редактора и run view.
7. `grep -rn "v-html\|innerHTML\|outerHTML" web/src/components/workflow web/src/components/WorkflowGraph.vue web/src/views/project/Workflow*.vue`
   пуст; шаблон с именем `<img src=x onerror=alert(1)>` и заметка с `<script>` рендерятся как текст.
8. Ошибки валидации показаны бейджем на узле и в сводке; Save заблокирован при ошибках; уход с
   несохранёнными изменениями требует подтверждения.
9. Docs: раздел «Creating a workflow» описывает новые взаимодействия, `workflow-editor.webp`
   заменён, `node scripts/check-docs.mjs` зелёный.
