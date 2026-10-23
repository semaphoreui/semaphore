# ui-features — zone index

Cross-cutting UI features of the Vue frontend and their backend halves: per-user preferences
stored on the server, server-side paging of task lists, the integration call history, and the
not-yet-built in-app help panel.

| When you need it | Read |
| --- | --- |
| Adding a new per-user preference, or wondering why the side menu stores *unpinned* items in the `option` table | `user-options-unpinned-menu.md` |
| Touching the pin/unpin edit mode of the side menu or its i18n keys | `user-options-unpinned-menu.md` |
| Changing how History or the per-template task list page through tasks (`before`, `count`, `X-Has-Next`) | `history-keyset-pagination.md` |
| Tempted to add `COUNT(*)` or `OFFSET` to a task list query | `history-keyset-pagination.md` |
| Changing the Integrations History tab, or what a webhook call records | `integration-call-history.md` |
| Tempted to log integration calls into `audit_event`, or to store webhook payloads | `integration-call-history.md` |
| Picking up the contextual help panel, or adding any help text / docs link to a form | `contextual-help-panel.md` |
| Wiring the `docs/` submodule into a frontend build or CI job | `contextual-help-panel.md` |
