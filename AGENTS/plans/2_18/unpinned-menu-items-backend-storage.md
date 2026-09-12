# Implementation Plan — Store Unpinned Menu Items on the Backend

## Goal

The side-menu "pin / unpin" feature lets a user move navigation items into a **More**
sub-menu. The customization must persist **per user on the backend**, so it
follows the account across browsers, devices, and sessions.

### Why we store *unpinned* items, not pinned items

The default state for every user is "everything pinned, nothing in **More**". The
*unpinned* set is therefore typically the **smaller** list, and is exactly the
delta from the default. Storing this delta instead of the full pinned list has
two concrete benefits:

1. **Storage stays compact and well under the `varchar(255)` limit** — most users
   have an empty list; even heavy customizers store only a few keys.
2. **New navigation items added in future releases automatically appear in the
   user's pinned section** rather than being silently moved into **More** because
   they're "not on the stored pinned list". Users opt items *out*, never *in*.

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
`user<userID>.<setting>`. The list of items the user has moved into **More** is
stored as:

| Key                                | Value                                  |
|------------------------------------|----------------------------------------|
| `user<userID>.nav.unpinnedItems`   | JSON array, e.g. `["environment","keys"]` |

The value is the set of navigation keys the user has explicitly **unpinned**.
Default (every item pinned, nothing in **More**) is represented by an empty array
or by no row at all.

`ValidateOptionKey` accepts `^[\w.]+$`, so `user42.nav.unpinnedItems` is a valid key —
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
   (initially just `nav.unpinnedItems`). Unknown suffixes are rejected with `400`.
4. Requires `value` to be **valid JSON** (`json.Valid([]byte(opt.Value))`). The
   handler stores the raw JSON string; non-JSON bodies are rejected with `400`.
   This is a hard invariant — all user options are JSON-encoded, full stop.

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
- `web/src/App.vue` — `data().unpinnedNavKeys`, `pinnedNavItemsList` /
  `unpinnedNavItems` computed properties, `loadUserOptions`, `togglePin`,
  `saveUnpinnedNavKeys`.

---

## Implementation Steps

### Phase 1 — Backend API

**1.1 New constant / allowlist**

In `api/user_options.go` (new file) define the set of permitted user-option keys:

```go
// keys a user is allowed to store via the per-user options API
var allowedUserOptionKeys = map[string]bool{
    "nav.unpinnedItems": true,
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

Returns e.g. `{"nav.unpinnedItems": "[\"dashboard\",\"history\"]"}`. If the user has
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

    if !json.Valid([]byte(opt.Value)) {
        helpers.WriteJSON(w, http.StatusBadRequest,
            map[string]string{"error": "user option value must be valid JSON"})
        return
    }

    err := helpers.Store(r).SetOption(userOptionKey(user.ID, opt.Key), opt.Value)
    // err handling -> 500

    helpers.WriteJSON(w, http.StatusOK, opt)
}
```

Note: the request body's `key` is the **suffix only** (`nav.unpinnedItems`); the
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

**2.1 State: track unpinned keys, not pinned keys**

The component holds the user's customization as the list of keys they have
unpinned. Default is an empty array — "nothing unpinned, everything visible in
the main section".

```js
unpinnedNavKeys: [],
```

The computed properties derive both lists from `navItems` and `unpinnedNavKeys`:

```js
pinnedNavItemsList() {
  return this.navItems.filter((item) => !this.unpinnedNavKeys.includes(item.key));
},
unpinnedNavItems() {
  return this.navItems.filter((item) =>  this.unpinnedNavKeys.includes(item.key));
},
```

A nav item that is unknown to the current `unpinnedNavKeys` array (e.g. introduced
in a later release) is automatically pinned — exactly the desired behaviour.

**2.2 Load the user's unpinned items from the backend**

Add a `loadUserOptions()` method and call it from `loadData()` (after
`loadUserInfo()`, since it needs an authenticated session):

```js
async loadUserOptions() {
  const options = (await axios({
    method: 'get',
    url: '/api/user/options',
    responseType: 'json',
  })).data;

  if (options['nav.unpinnedItems'] != null) {
    this.unpinnedNavKeys = JSON.parse(options['nav.unpinnedItems']);
  }
}
```

If the user has no stored value, they start with the default layout (every item
pinned).

**2.3 Persist changes through the API**

`togglePin` mutates `unpinnedNavKeys` and pushes it to the backend:

```js
async togglePin(key) {
  if (this.unpinnedNavKeys.includes(key)) {
    this.unpinnedNavKeys = this.unpinnedNavKeys.filter((k) => k !== key);
  } else {
    this.unpinnedNavKeys = [...this.unpinnedNavKeys, key];
  }
  await this.saveUnpinnedNavKeys();
},

async saveUnpinnedNavKeys() {
  try {
    await axios({
      method: 'post',
      url: '/api/user/options',
      responseType: 'json',
      data: { key: 'nav.unpinnedItems', value: JSON.stringify(this.unpinnedNavKeys) },
    });
  } catch (err) {
    EventBus.$emit('i-snackbar', { color: 'error', text: getErrorMessage(err) });
  }
},
```

The UI update is optimistic so the menu stays responsive; a failed save surfaces a
snackbar error.

**2.4 Edit-mode toggle for the side menu**

Currently the pin / unpin button (`.nav-pin-wrap`, `web/src/App.vue:274` and
`web/src/App.vue:317`) is always visible in every nav row, which makes accidental
clicks (intending to navigate, hitting the pin icon instead) easy and frequent.
Hide these controls behind an explicit **edit mode** that the user opts into.

**Where the toggle lives.** Add a new `v-btn icon` in the bottom action strip of
the side menu (`web/src/App.vue:333`, the `v-list-item` inside the `append` slot),
**between the Light/Dark mode `v-switch`** and **the language flag `v-menu`**:

```html
<v-list-item>
  <v-switch class="DarkModeSwitch" v-model="darkMode" ... />

  <v-spacer />

  <!-- NEW: edit-mode toggle -->
  <v-btn
    icon
    :color="navEditMode ? 'primary' : undefined"
    :title="navEditMode ? $t('finishEditingMenu') : $t('editMenu')"
    @click="navEditMode = !navEditMode"
  >
    <v-icon>{{ navEditMode ? 'mdi-check' : 'mdi-pencil-outline' }}</v-icon>
  </v-btn>

  <v-menu top min-width="150" max-width="235" ...> <!-- language picker --> </v-menu>
</v-list-item>
```

Icon choice: `mdi-pencil-outline` when off, `mdi-check` (or `mdi-pencil`) when on.
The active state is also color-highlighted so it's obvious the menu is in edit
mode.

**State.** Add `navEditMode: false` to `data()` (near `pinnedNavKeys`). This is a
session-local UI flag and is intentionally **not** persisted — every page load
starts in normal (non-edit) mode so navigation is the default behaviour.

**Show pin buttons only in edit mode.** Wrap the existing `.nav-pin-wrap` blocks
with `v-if="navEditMode"` so the unpin / pin `v-btn` only renders while the user
is editing:

```html
<div class="nav-pin-wrap" v-if="navEditMode && navItems.length > 1">
  <v-btn icon @click.stop.prevent="togglePin(item.key)" :title="$t('unpin')">
    <v-icon small>mdi-pin-off-outline</v-icon>
  </v-btn>
</div>
```

And for the "More" group (`web/src/App.vue:317`), the same `v-if="navEditMode"`
guard. The rows themselves still navigate normally; only the pin/unpin controls
appear/disappear.

**"More" visibility while editing.** The "More" section is **not** auto-expanded
in edit mode. The user opens it explicitly via the existing chevron if they want
to pin currently-unpinned items.

**i18n.** Add two new strings:

- `editMenu` — "Edit menu" (tooltip + aria-label when off)
- `finishEditingMenu` — "Done" (tooltip when on)

Add to every `web/src/lang/*.js` file alongside existing `pin` / `unpin` keys.

**Result for the user.**

- Normal use: no pin buttons visible anywhere — rows are pure navigation, zero
  accidental unpinning.
- To customize: click the pencil icon, pin/unpin freely (open the "More" group
  manually via the chevron if needed), click the check icon to finish.

### Phase 3 — Tests

**Backend** (`api/user_options_test.go`, using `net/http/httptest` per
`.claude/CLAUDE.md`):

- `setUserOption` with an allowlisted key and a valid JSON value stores
  `user<id>.nav.unpinnedItems` and returns `200`.
- `setUserOption` with a non-allowlisted key returns `400` and writes nothing.
- `setUserOption` with a non-JSON value (e.g. `"not json"`, trailing garbage,
  empty string) returns `400` and writes nothing.
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

---

## Rollout

- **No DB migration** — the `option` table is reused as-is.
- **Frontend + backend must be deployed together** — the frontend depends on
  `GET`/`POST /api/user/options`. If the backend lacks the endpoint, the call
  fails and the user sees the default layout (every item pinned).

## Risks & Notes

| Risk | Mitigation |
|------|------------|
| A user writing arbitrary global option keys | Handler always prepends `user<id>.` and rejects non-allowlisted suffixes; body never carries the user ID. |
| `varchar(255)` overflow | Pinned list is far smaller; documented for future settings. Consider widening only if a new user-option needs it. |
| Extra API round-trip on every pin toggle | Optimistic UI update; failure shows a snackbar. Toggling is rare. |
| Orphaned option rows after user deletion | `deleteUser` calls `DeleteOptions("user<id>")`. |

## Out of Scope (possible follow-ups)

- Migrating other client-only preferences (`darkMode`, `lang`, `projectId`,
  `project<id>__lastVisitedViewId`) to the same per-user options mechanism — the
  endpoint added here is intentionally generic enough to absorb them later by
  extending `allowedUserOptionKeys`.
- A dedicated `user__option` table with a real `user_id` foreign key — cleaner than
  key-namespacing, but unnecessary now and explicitly not what was requested.
