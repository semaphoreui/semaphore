# ui-features — zone index

Cross-cutting UI features of the Vue frontend and their backend halves: per-user preferences
stored on the server, server-side paging of task lists, and the built-in help panel.

| When you need it | Read |
| --- | --- |
| Adding a new per-user preference, or wondering why the side menu stores *unpinned* items in the `option` table | `user-options-unpinned-menu.md` |
| Touching the pin/unpin edit mode of the side menu or its i18n keys | `user-options-unpinned-menu.md` |
| Changing how History or the per-template task list page through tasks (`before`, `count`, `X-Has-Next`) | `history-keyset-pagination.md` |
| Tempted to add `COUNT(*)` or `OFFSET` to a task list query | `history-keyset-pagination.md` |
| Adding built-in help to a dialog or a page, or any help text / docs link to a form | `contextual-help-panel.md` |
| Changing the help panel, its host logic (`HelpDocsHost`), `EditDialog` or `HelpPage` around it | `contextual-help-panel.md` |
| Changing the format of help documents or the build that compiles and checks them (`web/vite/help`) | `contextual-help-panel.md` |
| A build that fails with `HelpDocError`, or a help panel laid out against its dialog | `contextual-help-panel.md` |
| Tempted to make a frontend build read the `docs/` clone | `contextual-help-panel.md` |
