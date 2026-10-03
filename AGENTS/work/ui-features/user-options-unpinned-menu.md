# Per-user options and the unpinned side-menu items

The side menu lets a user move navigation items into a **More** group. That customisation is
stored per user on the backend in the existing global `option` key/value table, under a
`user<ID>.<setting>` key namespace, and exposed through a non-admin `/api/user/options` endpoint
pair. The same endpoint now also carries the user's language preference (`lang`), so it is the
general per-user preference store, not a menu-only feature.

Status (2026-09-27): implemented — `api/user_options.go` (+ `_test.go`), routes on the `/user`
subrouter in `api/router.go`, cleanup in `deleteUser` (`api/users.go`), `App.vue`
`loadUserOptions`/`togglePin`/`navEditMode`, `api-docs.yml` `/user/options`. Merged in PR #3883
(`feat/unpin_store_db`, e86d06e65, 2026-05-23); `lang` key added later in 51b02368e.

## Decisions and why

- **Store the unpinned set, not the pinned set.** Default is "everything pinned", so the unpinned
  list is the delta from default: it stays small (well inside `varchar(255)`) and a navigation
  item added in a later release is pinned automatically instead of silently landing in *More*.
- **Reuse the `option` table, no new table.** Explicit owner request. `OptionsManager` already
  supports prefix filtering (`key = f OR key LIKE 'f.%'`), so no store or migration change was
  needed. A dedicated `user__option` table with a real FK was rejected as unnecessary for now.
- **Own endpoint, never the admin option handlers.** `api/options.go` is admin-only and works on
  arbitrary keys. The user endpoint derives the key from the session user ID — the body never
  carries a user ID — and accepts only allow-listed suffixes (`allowedUserOptionKeys`); anything
  else is 400. This makes cross-user or global-key writes impossible by construction.
- **Values must be valid JSON** (`json.Valid`), a hard invariant so future settings can be
  parsed uniformly.
- **Edit mode for pin controls.** Pin/unpin buttons were always visible and caught accidental
  clicks. They now render only while `navEditMode` is on (toggle in the bottom strip of the
  drawer, `mdi-playlist-edit` → `mdi-check`). Edit mode is session-local and never persisted so
  every load starts in plain navigation mode. *More* is not auto-expanded in edit mode.
- **Optimistic UI.** The menu updates immediately; a failed save shows a snackbar.
- **Prefix-collision safety** relies on the literal `.` in `LIKE 'user1.%'`, so `user1` cannot
  match `user10.*`. `deleteUser` calls `DeleteOptions("user<id>")` to avoid orphan rows.
- HA: state lives in the shared DB, nothing node-local. Security: no new authz surface beyond
  the authenticated `/user` router.

## Open, broken or deferred

- ⚠️ 2026-09-27: `editMenu`/`finishEditingMenu` i18n keys exist only in `en.js`, `ru.js`, `cs.js`
  — the plan asked for every `web/src/lang/*.js`; the other 14 locales fall back to English.
- ⚠️ 2026-09-27: the store-level prefix-collision test (`user1` vs `user10`) the plan asked for
  was never written; the behaviour is only implied by the `LIKE` clause.
- The optional `DELETE /api/user/options/{key}` ("reset menu layout") was not implemented.
- Candidates to migrate from `localStorage` to this store: `darkMode`, `projectId`,
  `project<id>__lastVisitedViewId`. Extending `allowedUserOptionKeys` is all that is needed.
- `App.vue` still reads/writes `localStorage` for `lang` in addition to the backend option.

## Key files and references

- `api/user_options.go`, `api/user_options_test.go`, `api/router.go` (`tokenAPI` `/options`)
- `api/users.go` (`deleteUser` cleanup), `api/options.go` (admin handlers, untouched)
- `db/sql/option.go`, `db/bolt/option.go`, `db/Store.go` (`OptionsManager`)
- `web/src/App.vue` (`unpinnedNavKeys`, `navEditMode`, `loadUserOptions`, `togglePin`)
- `api-docs.yml` `/user/options`
- PR https://github.com/semaphoreui/semaphore/pull/3883
- Source plan: `AGENTS/plans/2_18/unpinned-menu-items-backend-storage.md`
