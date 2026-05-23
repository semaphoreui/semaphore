# Implementation Plan — Store Unpinned Menu Items on the Backend

## Goal

The side-menu "pin / unpin" feature lets a user move navigation items into a **More**
sub-menu. The set of pinned items was previously persisted only in the browser's
`localStorage` (`nav__pinnedItems`). This means the preference was:

- lost when the user switched browser / device,
- lost when local storage was cleared,
- not shared across sessions of the same account.

The goal is to persist the pinned-items list **on the backend, per user**, so it
follows the account everywhere. `localStorage` is no longer used for this setting —
the backend is the sole source of truth. No backward compatibility with the legacy
`nav__pinnedItems` key is kept.

> The user explicitly asked to **use the existing `option` table** as the storage
> backend. This plan does exactly that — no new table is introduced.

## Design Summary

### Storage — reuse the `option` table, namespaced by user

The `option` table is a global key/value store:

```sql
create table `option` (
    `key`   varchar(255) primary key not null,
    `value` varchar(255) not null
);
```

It is already exposed through the `OptionsManager` interface (`db/store.go:218-224`)
and implemented for both SQL (`db/sql/option.go`) and Bolt (`db/bolt/option.go`).
`GetOptions` already supports **prefix filtering** (`key = filter OR key LIKE
filter.%`), which is exactly what we need to scope options to one user.

**Key convention:** per-user options use the key
`user<userID>.<setting>`. The pinned-items list is stored as:

| Key                              | Value                                  |
|----------------------------------|----------------------------------------|
| `user<userID>.nav.pinnedItems`  | JSON array, e.g. `["dashboard","history"]` |

`ValidateOptionKey` accepts `^[\w.]+$`, so `user42.nav.pinnedItems` is a valid key —
**no DB or store-layer change is required at all.** The entire `OptionsManager`
interface already does everything we need.

> **`varchar(255)` is sufficient.** The value is a JSON array of short navigation
> keys (~10–15 items, each a single word). A realistic worst case is well under
> 200 characters. This plan does **not** widen the column; a one-line note is added
> so a future setting with larger values reconsiders this.

### API — a new *non-admin* user-scoped options endpoint

The existing option handlers (`api/options.go`) are **admin-only** and operate on
arbitrary global keys. We must **not** reuse them — a regular user must be able to
read/write *only their own* options.

Add a new controller for **current-user options**. Every handler:

1. Reads the authenticated user from context (`helpers.GetFromContext(r, "user")`).
2. **Always** derives the storage key as `user<currentUser.ID>.<suffix>` — the
   user ID comes from the session, never from the request body. This makes it
   impossible to read or write another user's options or a global option.
3. Restricts the `<suffix>` to an **allowlist** of known user-setting keys
   (initially just `nav.pinnedItems`). Unknown suffixes are rejected with `400`.

This keeps the surface tight and intentional rather than a generic per-user KV API.

## Affected Areas (reference paths)

- `db/Option.go`, `db/sql/option.go`, `db/bolt/option.go`,
  `db/store.go:218-224` — existing option storage; **read-only reference, no change**.
- `db/sql/migrations/v2.9.62.sql` — original `option` table DDL (reference).
- `api/options.go` — existing admin option handlers (reference; not modified).
- `api/router.go:197-211` — route registration; `authenticatedAPI` /
  `tokenAPI` (`/user` subrouter) is where the new routes go.
- `api/user.go` — `UserController`; a good home for the new handlers, or a new
  `api/user_options.go` file.
- `api/users.go` — `deleteUser` handler; must clean up per-user options.
- `web/src/App.vue` — `data().pinnedNavKeys` (line ~988), `pinnedNavItemsList` /
  `unpinnedNavItems` computed (lines ~1131-1145), `loadData` / `loadUserInfo`
  (lines ~1349, ~1458), `togglePin` (line ~1321).

---

## Implementation Steps

### Phase 1 — Backend API

**1.1 New constant / allowlist**

In `api/user_options.go` (new file) define the set of permitted user-option keys:

```go
// keys a user is allowed to store via the per-user options API
var allowedUserOptionKeys = map[string]bool{
    "nav.pinnedItems": true,
}

func userOptionKey(userID int, suffix string) string {
    return fmt.Sprintf("user%d.%s", userID, suffix)
}
```

**1.2 `GET /api/user/options` — read the current user's options**

```go
func getUserOptions(w http.ResponseWriter, r *http.Request) {
    user := helpers.GetFromContext(r, "user").(*db.User)
    prefix := fmt.Sprintf("user%d.", user.ID)

    all, err := helpers.Store(r).GetOptions(db.RetrieveQueryParams{Filter: prefix})
    // err handling -> 500

    // strip the "user<id>." prefix so the client sees clean keys
    res := map[string]string{}
    for k, v := range all {
        res[strings.TrimPrefix(k, prefix)] = v
    }
    helpers.WriteJSON(w, http.StatusOK, res)
}
```

Returns e.g. `{"nav.pinnedItems": "[\"dashboard\",\"history\"]"}`. If the user has
no stored options, returns `{}`.

**1.3 `POST /api/user/options` — write one option**

```go
func setUserOption(w http.ResponseWriter, r *http.Request) {
    user := helpers.GetFromContext(r, "user").(*db.User)

    var opt db.Option
    if !helpers.Bind(w, r, &opt) { return }

    if !allowedUserOptionKeys[opt.Key] {
        helpers.WriteJSON(w, http.StatusBadRequest,
            map[string]string{"error": "unknown user option key"})
        return
    }

    err := helpers.Store(r).SetOption(userOptionKey(user.ID, opt.Key), opt.Value)
    // err handling -> 500

    helpers.WriteJSON(w, http.StatusOK, opt)
}
```

Note: the request body's `key` is the **suffix only** (`nav.pinnedItems`); the
handler prepends the user namespace. The body never controls the user ID.

**1.4 (optional) `DELETE /api/user/options/{key}` — reset a setting**

Useful for a future "reset menu layout" action. Validates the `{key}` against the
allowlist, then `DeleteOption(userOptionKey(user.ID, key))`. Can be deferred.

**1.5 Register routes** in `api/router.go`, on the existing `/user` subrouter
(`tokenAPI`, lines ~203-206) — it is already `authenticatedAPI`-scoped, so any
logged-in user (not just admins) can call it:

```go
tokenAPI.Path("/options").HandlerFunc(getUserOptions).Methods("GET", "HEAD")
tokenAPI.Path("/options").HandlerFunc(setUserOption).Methods("POST")
// optional:
tokenAPI.HandleFunc("/options/{key}", deleteUserOption).Methods("DELETE")
```

(Consider renaming `tokenAPI` → `userAPI` for clarity, or add a sibling subrouter;
cosmetic.)

**1.6 Clean up options on user deletion**

In `api/users.go`'s `deleteUser`, after the user is removed, call
`store.DeleteOptions(fmt.Sprintf("user%d", userID))` so deleted accounts don't
leave orphaned option rows. `DeleteOptions` already does prefix deletion.

### Phase 2 — Frontend (`web/src/App.vue`)

**2.1 Stop seeding `pinnedNavKeys` from `localStorage` directly**

Change the `data()` initializer (line ~988) to start as `null` and load the real
value asynchronously:

```js
pinnedNavKeys: null,
```

`null` keeps the current "everything pinned, nothing in More" default behaviour
(see `pinnedNavItemsList` / `unpinnedNavItems`).

**2.2 Load pinned items from the backend**

Add a `loadUserOptions()` method and call it from `loadData()` (after
`loadUserInfo()`, since it needs an authenticated session):

```js
async loadUserOptions() {
  const options = (await axios({
    method: 'get',
    url: '/api/user/options',
    responseType: 'json',
  })).data;

  if (options['nav.pinnedItems'] != null) {
    this.pinnedNavKeys = JSON.parse(options['nav.pinnedItems']);
  }
}
```

The legacy `nav__pinnedItems` `localStorage` key is **not** read. Users who pinned
items on the old client will see the default layout once and can re-pin as needed.

**2.3 Persist changes through the API**

Replace the `localStorage.setItem` call in `togglePin` (line ~1331) with a backend
write:

```js
async togglePin(key) {
  let pinned = this.pinnedNavKeys;
  if (pinned === null) {
    pinned = this.navItems.map((i) => i.key).filter((k) => k !== key);
  } else if (pinned.includes(key)) {
    pinned = pinned.filter((k) => k !== key);
  } else {
    pinned = [...pinned, key];
  }
  this.pinnedNavKeys = pinned;     // optimistic UI update
  await this.savePinnedNavKeys();
},

async savePinnedNavKeys() {
  try {
    await axios({
      method: 'post',
      url: '/api/user/options',
      responseType: 'json',
      data: { key: 'nav.pinnedItems', value: JSON.stringify(this.pinnedNavKeys) },
    });
  } catch (err) {
    EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
  }
},
```

The UI update is optimistic so the menu stays responsive; a failed save surfaces a
snackbar error. `localStorage` is not touched — the backend is the sole source of
truth.

### Phase 3 — Tests

**Backend** (`api/user_options_test.go`, using `net/http/httptest` per
`.claude/CLAUDE.md`):

- `setUserOption` with an allowlisted key stores `user<id>.nav.pinnedItems` and
  returns `200`.
- `setUserOption` with a non-allowlisted key returns `400` and writes nothing.
- `getUserOptions` returns only the current user's keys, with the `user<id>.`
  prefix stripped.
- Two different users do not see each other's options (namespacing isolation).
- `deleteUser` removes the user's option rows.

**Store layer** — already covered by `db/bolt/option_test.go`; add a case asserting
`GetOptions` prefix filtering does not match a longer sibling prefix
(`user-1` must not match `user10`). *(See note below — this is a real edge case.)*

> **Prefix-collision edge case.** `GetOptions`/`DeleteOptions` match
> `key = filter OR key LIKE filter.%`. With filter `user1`, the `LIKE user1.%`
> branch correctly excludes `user10.nav.*` because of the literal `.`. This is
> safe **as written**, but the test above pins the behaviour so a future refactor
> can't regress it.

### Phase 4 — Docs / API spec

- Add the two endpoints to `api-docs.yml` (`GET` / `POST /user/options`).
- Note in the PR description that the legacy `nav__pinnedItems` `localStorage` key
  is no longer read — existing users who had unpinned items will see the default
  layout once and can re-pin as desired.

---

## Rollout

- **No DB migration** — the `option` table is reused as-is.
- **No backward compatibility** — the frontend reads and writes the backend only.
  The legacy `nav__pinnedItems` `localStorage` entry is ignored; users who had it
  set will see the default (all items pinned) on first load and can re-pin.
- **Old frontend + new backend:** harmless — the old UI just keeps using
  `localStorage` and ignores the new endpoint.
- **New frontend + old backend:** unsupported — frontend and backend must be
  deployed together. The `GET /api/user/options` call will fail and the user will
  see the default layout.

## Risks & Notes

| Risk | Mitigation |
|------|------------|
| A user writing arbitrary global option keys | Handler always prepends `user<id>.` and rejects non-allowlisted suffixes; body never carries the user ID. |
| `varchar(255)` overflow | Pinned list is far smaller; documented for future settings. Consider widening only if a new user-option needs it. |
| Extra API round-trip on every pin toggle | Optimistic UI update; failure shows a snackbar. Toggling is rare. |
| Orphaned option rows after user deletion | `deleteUser` calls `DeleteOptions("user<id>")`. |

## Out of Scope (possible follow-ups)

- Migrating other `localStorage` preferences (`darkMode`, `lang`, `projectId`,
  `project<id>__lastVisitedViewId`) to the same per-user options mechanism — the
  endpoint added here is intentionally generic enough to absorb them later by
  extending `allowedUserOptionKeys`.
- A dedicated `user__option` table with a real `user_id` foreign key — cleaner than
  key-namespacing, but unnecessary now and explicitly not what was requested.
