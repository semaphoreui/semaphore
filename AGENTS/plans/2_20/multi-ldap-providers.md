# Multiple LDAP Providers Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Support multiple LDAP directories (GitLab model): configured as a map in config (like `oidc_providers`), shown as tabs on the login page — first tab is the internal account login, following tabs are LDAP providers. The legacy flat `ldap_*` config keeps working and gets the tab right after the internal login.

**Architecture:** New config map `ldap_providers: map[string]LdapProvider` mirroring `oidc_providers`. The legacy flat `ldap_*` fields are converted at read time into a synthesized provider with the reserved ID `"ldap"` — no config migration, full backward compatibility. `user__external_identity` gains a `type` column (`ldap` | `oidc`) so an LDAP provider ID and an OIDC provider ID may share the same name without colliding; uniqueness becomes `(type, provider, external_uid)`. The login POST gains optional `method`/`provider` fields; requests without them behave exactly as today (legacy clients).

**Tech Stack:** Go, gorp (via existing `SqlDb` helpers), SQL migrations (MySQL/PostgreSQL/SQLite), testify, Vue 2 + Vuetify (`v-tabs`).

## Global Constraints

- Do not use global variables (project rule); `util.Config` is the existing exception. Sentinel errors and constants are allowed (existing pattern: `db.ErrNotFound`).
- Tests: `github.com/stretchr/testify/assert` / `require`, table-driven with `t.Run` (project rule).
- Migration SQL must run on MySQL, PostgreSQL and SQLite — backtick style, preprocessed per dialect (Go text/template with `.Mysql`/`.Postgresql`/`.Sqlite`, precedent `v2.16.8.sql`). The `.err.sql` companion is the ROLLBACK script (executed only by `TryRollbackMigration`), not an "errors tolerated" apply file.
- Migration version here is **v2.20.1** (current tail of `db/Migration.go` `commonScripts` is `{Version: "2.20.0"}`); if another feature claims it first, take the next free version.
- Provider ID rules (GitLab model): IDs are the keys of the `ldap_providers` map. The ID `"ldap"` is **reserved** for the legacy flat config and is skipped if present in the map. IDs are stored in `user__external_identity.provider` with `type='ldap'`.
- Identity `type` values are exactly `"ldap"` and `"oidc"` (constants `db.IdentityTypeLdap`, `db.IdentityTypeOidc`).
- Login page order: internal account tab first, then legacy `"ldap"` tab (if flat config enabled), then `ldap_providers` sorted by `order` (ties broken by ID).
- Existing behavior that must NOT change: login POST **without** the new `method` field behaves exactly as today (try legacy LDAP if enabled, fall back to password); OIDC flows; local accounts never adopted by external providers.
- After config struct changes, regenerate `config.schema.yaml` (see the `semaphore-config-schema` skill) — folded into Task 8.

## File Structure

| File | Responsibility |
|---|---|
| `util/LdapProvider.go` (new) | `LdapProvider` struct, `LdapProviderEntry`, defaults for mappings |
| `util/config.go` | `LdapProviders` map field; `ActiveLdapProviders()` / `GetLdapProvider()` |
| `db/sql/migrations/v2.20.1.sql` + `.err.sql` (new) | `type` column, backfill, re-keyed unique index |
| `db/Migration.go` | register `2.20.1` |
| `db/UserExternalIdentity.go`, `db/Store.go`, `db/sql/external_identity.go` | `Type` field, type-aware store methods |
| `api/login_identity.go` | `Type` in profile, type-aware resolve/link |
| `api/login.go` | provider-parameterized `tryFindLDAPUser`, `method`/`provider` in login POST, `ldap_providers` in login metadata |
| `api/user.go`, `api/users.go`, `api/router.go` | provider in LDAP link endpoint, type in unlink route |
| `api/admin_info.go` | `ldap_enabled` reflects any provider |
| `web/src/views/Auth.vue` | login tabs |
| `web/src/components/UserForm.vue` | provider select for LDAP linking, type-aware unlink |

---

### Task 1: Config — `ldap_providers` map + legacy merge

**Files:**
- Create: `util/LdapProvider.go`
- Create: `util/LdapProvider_test.go`
- Modify: `util/config.go` (LDAP settings block, around line 537)

**Interfaces:**
- Consumes: existing flat `Ldap*` fields, `LdapMappings` (util/config.go:71).
- Produces:
  - `type LdapProvider struct { DisplayName, Server string; NeedTLS bool; BindDN, BindPassword, SearchDN, SearchFilter string; Mappings *LdapMappings; Order int }`
  - `func (p LdapProvider) GetMappings() *LdapMappings` — never nil.
  - `type LdapProviderEntry struct { ID string; Provider LdapProvider }`
  - `func (conf *ConfigType) ActiveLdapProviders() []LdapProviderEntry` — login-page order.
  - `func (conf *ConfigType) GetLdapProvider(id string) (LdapProvider, bool)`

- [ ] **Step 1: Write the failing test**

`util/LdapProvider_test.go`:

```go
package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActiveLdapProviders_Order(t *testing.T) {
	conf := &ConfigType{
		LdapEnable: true,
		LdapServer: "legacy.example.com:389",
		LdapProviders: map[string]LdapProvider{
			"corp":   {DisplayName: "Corp AD", Server: "corp.example.com:389", Order: 2},
			"berlin": {DisplayName: "Berlin", Server: "berlin.example.com:389", Order: 1},
			// Reserved ID must be skipped: it would collide with the legacy provider.
			"ldap": {DisplayName: "Impostor", Server: "evil.example.com:389"},
		},
	}

	providers := conf.ActiveLdapProviders()

	require.Len(t, providers, 3)
	// Legacy flat config always first, right after the internal login tab.
	assert.Equal(t, "ldap", providers[0].ID)
	assert.Equal(t, "legacy.example.com:389", providers[0].Provider.Server)
	assert.Equal(t, "berlin", providers[1].ID)
	assert.Equal(t, "corp", providers[2].ID)
}

func TestActiveLdapProviders_NewConfigOnly(t *testing.T) {
	conf := &ConfigType{
		LdapProviders: map[string]LdapProvider{
			"corp": {DisplayName: "Corp AD", Server: "corp.example.com:389"},
		},
	}

	providers := conf.ActiveLdapProviders()

	require.Len(t, providers, 1)
	assert.Equal(t, "corp", providers[0].ID)
}

func TestActiveLdapProviders_Empty(t *testing.T) {
	conf := &ConfigType{}
	assert.Empty(t, conf.ActiveLdapProviders())
}

func TestGetLdapProvider(t *testing.T) {
	conf := &ConfigType{
		LdapEnable: true,
		LdapServer: "legacy.example.com:389",
		LdapProviders: map[string]LdapProvider{
			"corp": {Server: "corp.example.com:389"},
		},
	}

	p, ok := conf.GetLdapProvider("corp")
	require.True(t, ok)
	assert.Equal(t, "corp.example.com:389", p.Server)

	p, ok = conf.GetLdapProvider("ldap")
	require.True(t, ok)
	assert.Equal(t, "legacy.example.com:389", p.Server)

	_, ok = conf.GetLdapProvider("nope")
	assert.False(t, ok)
}

func TestLdapProvider_GetMappings_DefaultsWhenNil(t *testing.T) {
	p := LdapProvider{}
	m := p.GetMappings()
	require.NotNil(t, m)
	assert.Equal(t, "dn", m.DN)
	assert.Equal(t, "mail", m.Mail)
	assert.Equal(t, "uid", m.UID)
	assert.Equal(t, "cn", m.CN)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./util/ -run 'TestActiveLdapProviders|TestGetLdapProvider|TestLdapProvider_GetMappings' -v -count=1`
Expected: FAIL (compile error: `LdapProvider` undefined)

- [ ] **Step 3: Implement**

`util/LdapProvider.go`:

```go
package util

import "sort"

// LdapProvider is one LDAP directory configured under `ldap_providers`
// (analog of OidcProvider for `oidc_providers`). Field JSON names mirror
// the legacy flat ldap_* config options.
type LdapProvider struct {
	DisplayName  string        `json:"display_name"`
	Server       string        `json:"server"`
	NeedTLS      bool          `json:"needtls"`
	BindDN       string        `json:"binddn"`
	BindPassword string        `json:"bindpassword"`
	SearchDN     string        `json:"searchdn"`
	SearchFilter string        `json:"searchfilter"`
	Mappings     *LdapMappings `json:"mappings"`
	Order        int           `json:"order"`
}

// GetMappings returns the attribute mappings, falling back to the same
// defaults the legacy flat config uses (see LdapMappings `default:` tags).
func (p LdapProvider) GetMappings() *LdapMappings {
	if p.Mappings != nil {
		return p.Mappings
	}
	return &LdapMappings{DN: "dn", Mail: "mail", UID: "uid", CN: "cn"}
}

// LdapProviderEntry couples a provider with its stable ID (the config map
// key). The ID is stored in user__external_identity.provider with type='ldap'.
type LdapProviderEntry struct {
	ID       string
	Provider LdapProvider
}

// ActiveLdapProviders returns configured LDAP providers in login-page order:
// the legacy flat ldap_* config first (reserved ID "ldap"), then entries of
// `ldap_providers` sorted by Order (ties broken by ID). A `ldap_providers`
// entry keyed "ldap" is skipped: it would collide with the legacy provider.
func (conf *ConfigType) ActiveLdapProviders() (res []LdapProviderEntry) {
	if conf.LdapEnable {
		res = append(res, LdapProviderEntry{
			ID: "ldap",
			Provider: LdapProvider{
				DisplayName:  "LDAP",
				Server:       conf.LdapServer,
				NeedTLS:      conf.LdapNeedTLS,
				BindDN:       conf.LdapBindDN,
				BindPassword: conf.LdapBindPassword,
				SearchDN:     conf.LdapSearchDN,
				SearchFilter: conf.LdapSearchFilter,
				Mappings:     conf.LdapMappings,
			},
		})
	}

	var rest []LdapProviderEntry
	for id, p := range conf.LdapProviders {
		if id == "ldap" {
			continue
		}
		rest = append(rest, LdapProviderEntry{ID: id, Provider: p})
	}
	sort.Slice(rest, func(i, j int) bool {
		if rest[i].Provider.Order != rest[j].Provider.Order {
			return rest[i].Provider.Order < rest[j].Provider.Order
		}
		return rest[i].ID < rest[j].ID
	})

	return append(res, rest...)
}

// GetLdapProvider resolves a provider ID coming from a login or link request.
func (conf *ConfigType) GetLdapProvider(id string) (LdapProvider, bool) {
	for _, entry := range conf.ActiveLdapProviders() {
		if entry.ID == id {
			return entry.Provider, true
		}
	}
	return LdapProvider{}, false
}
```

In `util/config.go`, after the `LdapNeedTLS` field (line 545) add:

```go
	// LdapProviders configures multiple LDAP directories (like OidcProviders
	// for OIDC). The key is the provider ID shown in identity records; the
	// ID "ldap" is reserved for the legacy flat ldap_* config above.
	LdapProviders map[string]LdapProvider `json:"ldap_providers,omitempty" env:"SEMAPHORE_LDAP_PROVIDERS"`
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./util/ -run 'TestActiveLdapProviders|TestGetLdapProvider|TestLdapProvider_GetMappings' -v -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add util/LdapProvider.go util/LdapProvider_test.go util/config.go
git commit -m "feat(config): ldap_providers map with legacy ldap_* merge"
```

---

### Task 2: DB migration `v2.20.1` — `type` column on `user__external_identity`

**Files:**
- Create: `db/sql/migrations/v2.20.1.sql`
- Create: `db/sql/migrations/v2.20.1.err.sql`
- Modify: `db/Migration.go` (append after `{Version: "2.20.0"},` — line 134)
- Create: `db/sql/migration_2_20_1_test.go`

**Interfaces:**
- Consumes: table `user__external_identity` from migration v2.20.0.
- Produces: column `type` varchar(16) not null default `'oidc'`; unique index `(type, provider, external_uid)`; legacy rows with `provider='ldap'` backfilled to `type='ldap'`.

- [ ] **Step 1: Write the failing migration test**

`db/sql/migration_2_20_1_test.go`:

```go
package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_1_AddsIdentityTypeColumn(t *testing.T) {
	store := CreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	// The type column exists and accepts both values.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, type, created) values (?, 'corp', 'cn=jdoe,dc=example,dc=org', 'ldap', ?)",
		user.ID, user.Created)
	require.NoError(t, err)

	// Same (provider, external_uid) under a different type is allowed:
	// an OIDC provider may share a name with an LDAP provider.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, type, created) values (?, 'corp', 'cn=jdoe,dc=example,dc=org', 'oidc', ?)",
		user.ID, user.Created)
	assert.NoError(t, err)

	// Duplicate (type, provider, external_uid) is rejected.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, type, created) values (?, 'corp', 'cn=jdoe,dc=example,dc=org', 'ldap', ?)",
		user.ID, user.Created)
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./db/sql/ -run TestMigration_2_20_1 -v -count=1`
Expected: FAIL (`no such column: type`)

- [ ] **Step 3: Create the migration SQL**

`db/sql/migrations/v2.20.1.sql` — `type` is varchar(4) ('ldap'/'oidc' are exactly 4 chars): with utf8mb4 the unique key is (4+64+700)×4 = 3072 bytes, exactly InnoDB's limit; varchar(16) would exceed it. The old index drop is dialect-forked because bare `drop index` is invalid MySQL:

```sql
alter table `user__external_identity` add `type` varchar(4) not null default 'oidc';

update `user__external_identity` set `type` = 'ldap' where `provider` = 'ldap';

{{if .Mysql}}
alter table `user__external_identity` drop index `user__external_identity__provider_uid`;
{{else}}
drop index `user__external_identity__provider_uid`;
{{end}}

create unique index `user__external_identity__type_provider_uid`
  on `user__external_identity`(`type`, `provider`, `external_uid`);
```

`db/sql/migrations/v2.20.1.err.sql` — rollback (undo) script:

```sql
{{if .Mysql}}
alter table `user__external_identity` drop index `user__external_identity__type_provider_uid`;
{{else}}
drop index `user__external_identity__type_provider_uid`;
{{end}}

alter table `user__external_identity` drop column `type`;

create unique index `user__external_identity__provider_uid`
  on `user__external_identity`(`provider`, `external_uid`);
```

- [ ] **Step 4: Register the migration**

In `db/Migration.go`, after `{Version: "2.20.0"},` add:

```go
	{Version: "2.20.1"},
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./db/sql/ -run TestMigration_2_20_1 -v -count=1`
Expected: PASS

Also run the previous migration test to make sure nothing regressed:

Run: `go test ./db/sql/ -run TestMigration_2_20_0 -v -count=1`
Expected: PASS (its duplicate-insert assertion inserts identical rows, which now also violate the new `(type, provider, external_uid)` index — still an error). If it asserts on the *old* index name or omits `type` in an insert column list, update that test to include `type`.

- [ ] **Step 6: Commit**

```bash
git add db/sql/migrations/v2.20.1.sql db/sql/migrations/v2.20.1.err.sql db/Migration.go db/sql/migration_2_20_1_test.go
git commit -m "feat(db): type column (ldap|oidc) on user__external_identity"
```

---

### Task 3: Model + store — type-aware external identities

**Files:**
- Modify: `db/UserExternalIdentity.go`
- Modify: `db/Store.go` (`ExternalIdentityManager`, line 400)
- Modify: `db/sql/external_identity.go`
- Modify: `db/sql/external_identity_test.go` (adapt existing tests to new signatures)

**Interfaces:**
- Consumes: migration from Task 2.
- Produces:
  - `db.UserExternalIdentity` gains `Type string \`db:"type" json:"type"\``.
  - Constants `db.IdentityTypeLdap = "ldap"`, `db.IdentityTypeOidc = "oidc"`.
  - `GetExternalIdentity(idType string, provider string, externalUID string) (UserExternalIdentity, error)`
  - `DeleteExternalIdentity(userID int, idType string, provider string) error`
  - `CreateExternalIdentity(identity UserExternalIdentity)` — unchanged signature, `Type` comes in the struct.
  - `GetUserExternalIdentities(userID int)` — unchanged.

- [ ] **Step 1: Update the model**

`db/UserExternalIdentity.go`:

```go
const (
	IdentityTypeLdap = "ldap"
	IdentityTypeOidc = "oidc"
)

type UserExternalIdentity struct {
	ID          int       `db:"id" json:"id"`
	UserID      int       `db:"user_id" json:"user_id"`
	Type        string    `db:"type" json:"type"`
	Provider    string    `db:"provider" json:"provider"`
	ExternalUID string    `db:"external_uid" json:"external_uid"`
	Created     time.Time `db:"created" json:"created"`
}
```

- [ ] **Step 2: Update the interface**

`db/Store.go` (line 400):

```go
type ExternalIdentityManager interface {
	GetExternalIdentity(idType string, provider string, externalUID string) (UserExternalIdentity, error)
	GetUserExternalIdentities(userID int) ([]UserExternalIdentity, error)
	CreateExternalIdentity(identity UserExternalIdentity) (UserExternalIdentity, error)
	DeleteExternalIdentity(userID int, idType string, provider string) error
}
```

- [ ] **Step 3: Update the SQL implementation**

`db/sql/external_identity.go`:

```go
func (d *SqlDb) GetExternalIdentity(idType string, provider string, externalUID string) (identity db.UserExternalIdentity, err error) {
	err = d.selectOne(
		&identity,
		"select * from user__external_identity where type=? and provider=? and external_uid=?",
		idType, provider, externalUID)
	return
}

func (d *SqlDb) DeleteExternalIdentity(userID int, idType string, provider string) error {
	_, err := d.exec(
		"delete from user__external_identity where user_id=? and type=? and provider=?",
		userID, idType, provider)
	return err
}
```

`CreateExternalIdentity` and `GetUserExternalIdentities` stay as they are (the `Type` field travels in the struct via gorp).

- [ ] **Step 4: Adapt existing store tests**

In `db/sql/external_identity_test.go` update every call to the new signatures and set `Type: db.IdentityTypeLdap` / `db.IdentityTypeOidc` in created identities. Add one assertion that two identities differing only in `Type` coexist:

```go
func TestExternalIdentity_TypeSeparatesNamespaces(t *testing.T) {
	store := CreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID: user.ID, Type: db.IdentityTypeLdap, Provider: "corp", ExternalUID: "uid-1",
	})
	require.NoError(t, err)
	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID: user.ID, Type: db.IdentityTypeOidc, Provider: "corp", ExternalUID: "uid-1",
	})
	require.NoError(t, err)

	found, err := store.GetExternalIdentity(db.IdentityTypeLdap, "corp", "uid-1")
	require.NoError(t, err)
	assert.Equal(t, db.IdentityTypeLdap, found.Type)
}
```

- [ ] **Step 5: Build and run store tests**

Run: `go build ./... && go test ./db/... -count=1`
Expected: `db` packages PASS. `api` package does NOT compile yet (callers use old signatures) — that is Task 4; do not commit `api` changes here.

- [ ] **Step 6: Commit**

```bash
git add db/UserExternalIdentity.go db/Store.go db/sql/external_identity.go db/sql/external_identity_test.go
git commit -m "feat(db): type-aware external identity store methods"
```

---

### Task 4: Resolver and linker — identity type threading

**Files:**
- Modify: `api/login_identity.go`
- Modify: `api/login_identity_test.go`
- Modify: `api/login.go` (`loginByLDAP` line ~241, `oidcRedirect` resolve/link calls)
- Modify: `api/user.go` (`linkLdapIdentity` call site — full rework comes in Task 6, here only the extra argument)
- Modify: `api/users.go` + `api/router.go` (unlink route gains `{type}`)
- Modify: `api/users_identity_test.go` (adapt to new route/signatures)

**Interfaces:**
- Consumes: Task 3 store methods and constants.
- Produces:
  - `externalUserProfile` gains `Type string` (must be `db.IdentityTypeLdap` or `db.IdentityTypeOidc`).
  - `resolveExternalUser(store db.Store, p externalUserProfile) (db.User, error)` — same name, type-aware inside.
  - `linkExternalIdentity(store db.Store, user db.User, idType string, provider string, externalUID string) error`
  - Unlink route becomes `DELETE /api/users/{user_id}/identities/{type}/{provider}`.

- [ ] **Step 1: Update tests to the new contract**

In `api/login_identity_test.go`:
- Add `Type: db.IdentityTypeLdap` (LDAP cases) or `Type: db.IdentityTypeOidc` (OIDC cases) to every `externalUserProfile` literal.
- Change `linkExternalIdentity(store, user, "keycloak", "sub-1")` calls to `linkExternalIdentity(store, user, db.IdentityTypeOidc, "keycloak", "sub-1")`.
- Add one new test:

```go
func TestLinkExternalIdentity_SameProviderNameDifferentType(t *testing.T) {
	store := setupIdentityTest(t, "never")

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "alice", Name: "Alice", Email: "alice@example.com", External: true,
	})
	require.NoError(t, err)

	// "corp" LDAP and "corp" OIDC are different providers.
	require.NoError(t, linkExternalIdentity(store, user, db.IdentityTypeLdap, "corp", "cn=alice,dc=example,dc=org"))
	require.NoError(t, linkExternalIdentity(store, user, db.IdentityTypeOidc, "corp", "sub-alice"))

	ids, err := store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./api/ -run 'TestLinkExternalIdentity|TestResolveExternalUser' -v -count=1`
Expected: FAIL (compile errors — new signatures not implemented yet)

- [ ] **Step 3: Implement in `api/login_identity.go`**

```go
type externalUserProfile struct {
	Type            string // db.IdentityTypeLdap or db.IdentityTypeOidc
	Provider        string
	ExternalUID     string
	Username        string
	Name            string
	Email           string
	MatchByUsername bool
}
```

In `resolveExternalUser`:
- guard: `if p.Type == "" { return db.User{}, errors.New("external identity: empty type") }`
- `store.GetExternalIdentity(p.Type, p.Provider, p.ExternalUID)`
- both `CreateExternalIdentity` calls get `Type: p.Type`.

In `linkExternalIdentity(store, user, idType, provider, externalUID)`:
- `store.GetExternalIdentity(idType, provider, externalUID)`
- the per-provider duplicate check compares both fields: `if identity.Type == idType && identity.Provider == provider`
- `CreateExternalIdentity` gets `Type: idType`.

- [ ] **Step 4: Update callers**

`api/login.go`:
- `loginByLDAP` gains the provider ID (used by Task 5; for now thread the legacy constant):

```go
func loginByLDAP(store db.Store, ldapUser db.User, userDN string, providerID string) (db.User, error) {
	return resolveExternalUser(store, externalUserProfile{
		Type:            db.IdentityTypeLdap,
		Provider:        providerID,
		ExternalUID:     userDN,
		Username:        ldapUser.Username,
		Name:            ldapUser.Name,
		Email:           ldapUser.Email,
		MatchByUsername: true,
	})
}
```
Call site in `login()`: `loginByLDAP(helpers.Store(r), *ldapUser, ldapUserDN, "ldap")`.
- OIDC resolve in `oidcRedirect`: add `Type: db.IdentityTypeOidc`.
- OIDC link in `oidcRedirect`: `linkExternalIdentity(helpers.Store(r), sessionUser, db.IdentityTypeOidc, pid, claims.sub)`.

`api/user.go` (`linkLdapIdentity`): `linkExternalIdentity(helpers.Store(r), *currentUser, db.IdentityTypeLdap, "ldap", userDN)`.

`api/router.go` line 280 — route gains the type segment:

```go
	userPasswordAPI.Path("/identities/{type}/{provider}").HandlerFunc(usersController.DeleteUserIdentity).Methods("DELETE")
```

`api/users.go` (`DeleteUserIdentity`): read both vars and validate the type:

```go
	idType := mux.Vars(r)["type"]
	provider := mux.Vars(r)["provider"]
	if idType != db.IdentityTypeLdap && idType != db.IdentityTypeOidc {
		helpers.WriteErrorStatus(w, "Invalid identity type", http.StatusBadRequest)
		return
	}
	// ... existing logic, with:
	err := helpers.Store(r).DeleteExternalIdentity(user.ID, idType, provider)
```

Adapt `api/users_identity_test.go` to the new route and signatures (set `Type` in fixtures, hit `/identities/oidc/...` URLs).

- [ ] **Step 5: Run tests**

Run: `go build ./... && go test ./api/ -run 'TestLinkExternalIdentity|TestResolveExternalUser|TestSyncExternalUserAttrs' -v -count=1 && go test ./api/ -run Identities -v -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add api/login_identity.go api/login_identity_test.go api/login.go api/user.go api/users.go api/router.go api/users_identity_test.go
git commit -m "feat(auth): thread identity type (ldap|oidc) through resolver and linker"
```

---

### Task 5: LDAP auth — provider-parameterized bind + login POST/metadata

**Files:**
- Modify: `api/login.go` (`tryFindLDAPUser` line ~46, `login()` line ~279, `loginMetadata` line ~271)
- Create: `api/login_ldap_test.go` (metadata + method routing tests; no live LDAP)

**Interfaces:**
- Consumes: `util.Config.ActiveLdapProviders()` / `GetLdapProvider()` (Task 1), `loginByLDAP` (Task 4).
- Produces:
  - `func tryFindLDAPUser(provider util.LdapProvider, username, password string) (*db.User, string, error)`
  - Login POST body gains optional `method` (`""` | `"password"` | `"ldap"`) and `provider` (LDAP provider ID, default `"ldap"`).
  - Login GET metadata gains `ldap_providers: [{id, name}]`; `login_with_ldap` is true when any provider exists.

- [ ] **Step 1: Write the failing tests**

`api/login_ldap_test.go`:

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLoginConfig() {
	util.Config = &util.ConfigType{
		LdapEnable: true,
		LdapServer: "legacy.example.com:389",
		LdapProviders: map[string]util.LdapProvider{
			"corp": {DisplayName: "Corp AD", Server: "corp.example.com:389", Order: 1},
		},
	}
}

func TestLoginMetadata_LdapProviders(t *testing.T) {
	setupLoginConfig()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	w := httptest.NewRecorder()

	login(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"ldap_providers":[{"id":"ldap","name":"LDAP"},{"id":"corp","name":"Corp AD"}]`)
	assert.Contains(t, body, `"login_with_ldap":true`)
}

func TestLogin_UnknownLdapProvider(t *testing.T) {
	setupLoginConfig()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"auth":"jdoe","password":"x","method":"ldap","provider":"nope"}`))
	w := httptest.NewRecorder()

	login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./api/ -run 'TestLoginMetadata_LdapProviders|TestLogin_UnknownLdapProvider' -v -count=1`
Expected: FAIL (`ldap_providers` missing; unknown provider not rejected)

- [ ] **Step 3: Implement**

`tryFindLDAPUser` — provider-parameterized (delete the `util.Config.LdapEnable` guard; the caller resolves the provider):

```go
func tryFindLDAPUser(provider util.LdapProvider, username, password string) (*db.User, string, error) {
	var l *ldap.Conn
	var err error
	if provider.NeedTLS {
		// SECURITY: InsecureSkipVerify=true is pre-existing behavior carried
		// over from the flat config (api/login.go:54); it allows MITM on the
		// LDAP connection. Do not extend it further; see the out-of-scope
		// note about a per-provider tls_skip_verify option (default false).
		l, err = ldap.DialTLS("tcp", provider.Server, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		l, err = ldap.Dial("tcp", provider.Server)
	}
	// ... body identical to today, with these substitutions:
	//   util.Config.LdapBindDN       -> provider.BindDN
	//   util.Config.LdapBindPassword -> provider.BindPassword
	//   util.Config.LdapSearchDN     -> provider.SearchDN
	//   util.Config.LdapSearchFilter -> provider.SearchFilter
	//   util.Config.LdapMappings     -> provider.GetMappings()  (both attribute list and parseClaims)
}
```

`loginMetadata` (add field + fill in the GET branch of `login()`):

```go
type loginMetadataLdapProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type loginMetadata struct {
	OidcProviders     []loginMetadataOidcProvider `json:"oidc_providers"`
	LdapProviders     []loginMetadataLdapProvider `json:"ldap_providers"`
	LoginWithPassword bool                        `json:"login_with_password"`
	LoginWithLdap     bool                        `json:"login_with_ldap"`
	AuthMethods       LoginAuthMethods            `json:"auth_methods"`
}
```

GET branch:

```go
	ldapProviders := util.Config.ActiveLdapProviders()
	config.LdapProviders = make([]loginMetadataLdapProvider, 0, len(ldapProviders))
	for _, entry := range ldapProviders {
		name := entry.Provider.DisplayName
		if name == "" {
			name = entry.ID
		}
		config.LdapProviders = append(config.LdapProviders, loginMetadataLdapProvider{ID: entry.ID, Name: name})
	}
	config.LoginWithLdap = len(ldapProviders) > 0
```

POST branch — method routing (replaces the current LDAP-then-password block):

```go
	var login struct {
		Auth     string `json:"auth" binding:"required"`
		Password string `json:"password" binding:"required"`
		Method   string `json:"method"`   // "", "password" or "ldap"
		Provider string `json:"provider"` // LDAP provider ID when method == "ldap"
	}
	if !helpers.Bind(w, r, &login) {
		return
	}

	login.Auth = strings.ToLower(login.Auth)

	var err error
	var user db.User

	switch login.Method {
	case "password":
		if util.Config.PasswordLoginDisable {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user, err = loginByPassword(helpers.Store(r), login.Auth, login.Password)

	case "ldap":
		providerID := login.Provider
		if providerID == "" {
			providerID = "ldap"
		}
		provider, ok := util.Config.GetLdapProvider(providerID)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var ldapUser *db.User
		var ldapUserDN string
		ldapUser, ldapUserDN, err = tryFindLDAPUser(provider, login.Auth, login.Password)
		if err != nil || ldapUser == nil {
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"context":  "ldap",
					"provider": providerID,
					"auth":     login.Auth,
				}).Warn("Failed to find user in LDAP")
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user, err = loginByLDAP(helpers.Store(r), *ldapUser, ldapUserDN, providerID)

	default:
		// Legacy clients without the method field: previous behavior —
		// try the legacy flat LDAP first, fall back to password.
		var ldapUser *db.User
		var ldapUserDN string

		if legacy, ok := util.Config.GetLdapProvider("ldap"); ok {
			ldapUser, ldapUserDN, err = tryFindLDAPUser(legacy, login.Auth, login.Password)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"context": "ldap",
					"auth":    login.Auth,
				}).Warn("Failed to find user in LDAP")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		if ldapUser == nil {
			user, err = loginByPassword(helpers.Store(r), login.Auth, login.Password)
		} else {
			user, err = loginByLDAP(helpers.Store(r), *ldapUser, ldapUserDN, "ldap")
		}
	}

	// ... existing error handling (ErrNotFound -> 401, etc.) unchanged
```

- [ ] **Step 4: Run tests**

Run: `go test ./api/ -run 'TestLoginMetadata_LdapProviders|TestLogin_UnknownLdapProvider' -v -count=1 && go build ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add api/login.go api/login_ldap_test.go
git commit -m "feat(auth): multi-provider LDAP login with method routing"
```

---

### Task 6: LDAP identity linking — provider parameter

**Files:**
- Modify: `api/user.go` (`linkLdapIdentity`, line ~51)

**Interfaces:**
- Consumes: `util.Config.GetLdapProvider()` (Task 1), `linkExternalIdentity` (Task 4), `tryFindLDAPUser` (Task 5).
- Produces: `POST /api/user/identities/ldap` body gains optional `provider` (LDAP provider ID, default `"ldap"`).

- [ ] **Step 1: Implement**

```go
// linkLdapIdentity attaches an LDAP identity to the current account.
// Proof of ownership is a successful bind with the user's own LDAP credentials.
func linkLdapIdentity(w http.ResponseWriter, r *http.Request) {
	currentUser := helpers.GetFromContext(r, "user").(*db.User)

	var creds struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Provider string `json:"provider"` // LDAP provider ID, default "ldap"
	}
	if !helpers.Bind(w, r, &creds) {
		return
	}

	providerID := creds.Provider
	if providerID == "" {
		providerID = "ldap"
	}

	provider, ok := util.Config.GetLdapProvider(providerID)
	if !ok {
		helpers.WriteErrorStatus(w, "LDAP provider not found", http.StatusBadRequest)
		return
	}

	ldapUser, userDN, err := tryFindLDAPUser(provider, creds.Username, creds.Password)
	if err != nil || ldapUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := linkExternalIdentity(helpers.Store(r), *currentUser, db.IdentityTypeLdap, providerID, userDN); err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

Note: the `util.Config.LdapEnable` guard is replaced by `GetLdapProvider` (which already covers both legacy and map providers).

- [ ] **Step 2: Build and run api tests**

Run: `go build ./... && go test ./api/ -count=1`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add api/user.go
git commit -m "feat(auth): link LDAP identity for a specific provider"
```

---

### Task 7: UI — login method switcher (v-btn-toggle)

**Files:**
- Modify: `web/src/views/Auth.vue`

**Interfaces:**
- Consumes: `GET /api/auth/login` → `ldap_providers: [{id, name}]`, `login_with_password`; `POST /api/auth/login` with `method`/`provider` (Task 5).
- Produces: a `v-btn-toggle` method switcher on the login card — internal account first, then LDAP providers in server order. OIDC buttons stay below, visible for every selection.

- [ ] **Step 1: Add tab state and metadata**

In `data()` add:

```js
      ldapProviders: [],
      loginTab: 0,
```

In `loadLoginData()` (the method that fills `oidcProviders` / `loginWithPassword` from `/api/auth/login`) add:

```js
      this.ldapProviders = data.ldap_providers || [];
```

Add computed properties:

```js
    loginTabs() {
      const tabs = [];
      if (this.loginWithPassword || this.isPortal) {
        tabs.push({ id: null, name: this.$t('signIn') });
      }
      this.ldapProviders.forEach((p) => tabs.push({ id: p.id, name: p.name, ldap: true }));
      return tabs;
    },

    activeLoginTab() {
      return this.loginTabs[this.loginTab] || this.loginTabs[0];
    },
```

- [ ] **Step 2: Add the method switcher to the template**

In the `<div v-else>` login block (Auth.vue line ~191), above the password form insert (`v-btn-toggle`, not `v-tabs` — user's explicit choice; `mandatory` keeps one option always selected, `v-model` is the index into `loginTabs`):

```html
              <v-btn-toggle
                v-if="loginTabs.length > 1"
                v-model="loginTab"
                mandatory
                class="mb-4 d-flex"
              >
                <v-btn
                  v-for="tab in loginTabs"
                  :key="tab.id || 'local'"
                  small
                  class="flex-grow-1"
                >
                  {{ tab.name }}
                </v-btn>
              </v-btn-toggle>
```

Change the username/password form condition so LDAP tabs always show it, and the email-portal branch only applies on the internal tab:

```html
              <div v-if="(activeLoginTab && activeLoginTab.ldap) || loginWithPassword">
                <!-- existing username/password fields + Sign In button, unchanged -->
              </div>

              <div v-else-if="isPortal">
                <!-- existing email branch, unchanged -->
              </div>
```

- [ ] **Step 3: Send method/provider on submit**

In the `signIn()` method, extend the POST body:

```js
      const body = {
        auth: this.username,
        password: this.password,
      };
      if (this.activeLoginTab && this.activeLoginTab.ldap) {
        body.method = 'ldap';
        body.provider = this.activeLoginTab.id;
      } else if (this.loginTabs.length > 1) {
        // Internal tab explicitly requests password auth so a typo never
        // hits the LDAP directory (legacy single-form behavior).
        body.method = 'password';
      }
```

(keep the existing axios call, just pass `body` as `data`). When there are no LDAP providers (`loginTabs.length <= 1`) the body stays exactly as today — legacy server compatibility.

- [ ] **Step 4: Lint**

Run: `cd web && npx eslint src/views/Auth.vue`
Expected: no output

- [ ] **Step 5: Manual check**

Config for a quick manual pass (two providers + legacy):

```json
{
  "ldap_enable": true,
  "ldap_server": "localhost:389",
  "ldap_providers": {
    "corp": { "display_name": "Corp AD", "server": "corp.example.com:389", "order": 1 }
  }
}
```

Expected: a button toggle `Sign In | LDAP | Corp AD`; OIDC buttons (if configured) below the form for every selection; with no LDAP config — no toggle at all (current look).

- [ ] **Step 6: Commit**

```bash
git add web/src/views/Auth.vue
git commit -m "feat(ui): login method switcher for internal account and LDAP providers"
```

---

### Task 8: UI linking, admin info, config schema

**Files:**
- Modify: `web/src/components/UserForm.vue` (Security tab, LDAP link form line ~393, unlink call)
- Modify: `api/admin_info.go` (line 36)
- Modify: `config.schema.yaml`

**Interfaces:**
- Consumes: `ldap_providers` from login metadata, `provider` in `POST /api/user/identities/ldap` (Task 6), `DELETE .../identities/{type}/{provider}` (Task 4).
- Produces: provider select in the LDAP link form; type-aware unlink; correct admin info; up-to-date config schema.

- [ ] **Step 1: UserForm — provider select + type-aware unlink**

In `UserForm.vue`:
- Load `ldap_providers` together with the existing login-metadata fetch (the component already loads OIDC providers from `/api/auth/login` — extend that handler with `this.ldapProviders = data.ldap_providers || [];`, add `ldapProviders: []` to `data()`).
- LDAP link form: when `ldapProviders.length > 1`, show a select above the credentials fields and send its value:

```html
          <v-select
            v-if="ldapProviders.length > 1"
            v-model="ldapLinkProvider"
            :items="ldapProviders"
            item-text="name"
            item-value="id"
            :label="$t('provider')"
            dense
          />
```

```js
      // data():
      ldapLinkProvider: null,

      // linkLdapIdentity():
      await axios.post('/api/user/identities/ldap', {
        username: this.ldapUsername,
        password: this.ldapPassword,
        provider: this.ldapLinkProvider || undefined,
      });
```

- Identity list: display `identity.type` next to the provider name (e.g. `Corp AD (ldap)`); resolve LDAP display names from `ldapProviders` the same way OIDC names are resolved via `oidcProvider.name` (line ~386).
- Unlink call: change the DELETE URL to include the type:

```js
      await axios.delete(`/api/user/identities/${identity.type}/${identity.provider}`);
```

(If the current code uses the admin route `/api/users/${userId}/identities/...`, apply the same `{type}/{provider}` change there.)

- [ ] **Step 2: admin_info**

`api/admin_info.go` line 36:

```go
	authInfo["ldap_enabled"] = len(util.Config.ActiveLdapProviders()) > 0
```

- [ ] **Step 3: Config schema**

Regenerate `config.schema.yaml` for the new `ldap_providers` field — follow the `semaphore-config-schema` skill (map of `LdapProvider` objects: `display_name`, `server`, `needtls`, `binddn`, `bindpassword`, `searchdn`, `searchfilter`, `mappings`, `order`; note in the description that the key `ldap` is reserved for the legacy flat config).

- [ ] **Step 4: Lint + build**

Run: `cd web && npx eslint src/components/UserForm.vue && cd .. && go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add web/src/components/UserForm.vue api/admin_info.go config.schema.yaml
git commit -m "feat(ui): multi-LDAP linking, admin info and config schema"
```

---

## Out of scope (YAGNI)

- Per-provider `hosts` list for redundancy/failover (GitLab has it; add when someone asks).
- Admin UI for managing LDAP providers (config-file only, same as OIDC).
- Migrating legacy flat config to the map automatically (the synthesized `"ldap"` entry covers it forever at zero cost).
- Per-provider colors/icons on tabs (OIDC-style `color`/`icon` can be added to `LdapProvider` later without breaking anything).

## Follow-up worth doing (security)

- TLS verification for LDAPS is currently disabled unconditionally (`InsecureSkipVerify: true`, pre-existing). With per-provider config now available, add `tls_skip_verify bool` to `LdapProvider` — **default false** (verify) for new providers, `true` only for the synthesized legacy `"ldap"` entry to avoid breaking existing installs. Separate change, not part of this plan.
