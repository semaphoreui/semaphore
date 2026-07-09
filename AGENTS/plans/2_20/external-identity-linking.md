# External Identity Linking Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Link LDAP/OIDC users to accounts via a dedicated identity entity keyed by the provider's stable ID (OIDC `sub`, LDAP DN) instead of matching by email — eliminating the account-takeover class of bugs (Grafana CVE-2023-3128) while keeping upgrades seamless.

**Architecture:** New table `user__external_identity` (analog of Grafana's `user_auth`, with its known flaws fixed: UNIQUE index on `(provider, external_uid)`, wide `external_uid` column for long LDAP DNs, FK `ON DELETE CASCADE` instead of lazy stale-row cleanup, no OAuth-token columns mixed in). A single resolver `resolveExternalUser` is used by both LDAP and OIDC login flows: lookup by `(provider, external_uid)` first; fall back to legacy email/username matching only under a configurable mode (`auto` — default, self-closing migration path; `always`; `never`); otherwise create user + identity.

**Tech Stack:** Go, gorp (via existing `SqlDb` helpers), SQL migrations (MySQL/PostgreSQL/SQLite), testify.

## Grafana drawbacks deliberately excluded

| Grafana `user_auth` flaw | This design |
|---|---|
| `auth_id` varchar(190) breaks long LDAP DNs | `external_uid` varchar(700) (fits MySQL 3072-byte index limit with utf8mb4) |
| Non-unique index on `(auth_module, auth_id)`, "latest created wins" ambiguity | `UNIQUE (provider, external_uid)` |
| Stale rows referencing deleted users, cleaned lazily in login code | `FOREIGN KEY (user_id) ... ON DELETE CASCADE` |
| OAuth tokens mixed into the linking table | Not stored (YAGNI until a refresh-token feature exists) |
| Email-lookup toggle was a global breaking change ("user already exists" after upgrade) | `auto` mode: legacy `External` users without any identity row are matched by email once, then pinned — no flag day |
| Email lookup could adopt any account | Even in `always` mode only `External` accounts can be matched; local (password) accounts are never adopted |
| No API to view/unlink a single identity | `GET /api/users/{user_id}/identities` + `DELETE /api/users/{user_id}/identities/{provider}` (admin) |

## Global Constraints

- Do not use global variables (project rule); `util.Config` is the existing exception.
- Tests: `github.com/stretchr/testify/assert` / `require`, table-driven with `t.Run` (project rule).
- Migration SQL must run on MySQL, PostgreSQL and SQLite — follow the backtick style of `db/sql/migrations/v2.18.5.sql` (the loader preprocesses per dialect).
- Migration version here is **v2.20.0**; if another feature claims it first, take the next free version and update `db/Migration.go` accordingly.
- Provider key for LDAP is the literal string `"ldap"`; OIDC provider keys come from the `oidc_providers` map in config. An OIDC provider named `ldap` in config would collide — reject in resolver is NOT needed; instead validation is part of Task 4 resolver tests documentation (`"ldap"` is reserved, noted in config schema description).
- Existing behavior that must NOT change: local (non-External) users can never log in via LDAP/OIDC; OIDC still ignores username when matching (email only); LDAP legacy matching uses username OR email as today.

---

### Task 1: DB migration `v2.20.0` — table `user__external_identity`

**Files:**
- Create: `db/sql/migrations/v2.20.0.sql`
- Modify: `db/Migration.go` (append to `commonScripts`, currently ends with `{Version: "2.19.11"}`)
- Test: `db/sql/migration_2_20_0_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: table `user__external_identity` with columns `id`, `user_id`, `provider`, `external_uid`, `created`; unique index on `(provider, external_uid)`.

- [ ] **Step 1: Write the failing migration test**

```go
package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_2_20_0_CreatesExternalIdentityTable(t *testing.T) {
	// CreateTestStore (db/sql/SqlDb.go:124) sets util.Config, connects an
	// in-memory sqlite (PRAGMA foreign_keys=ON) and runs db.Migrate(store, nil).
	store := CreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	// Table exists and accepts a row.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, created) values (?, 'ldap', 'cn=admin,dc=example,dc=org', ?)",
		user.ID, user.Created)
	assert.NoError(t, err)

	// Unique (provider, external_uid) index rejects a duplicate.
	_, err = store.exec(
		"insert into user__external_identity (user_id, provider, external_uid, created) values (?, 'ldap', 'cn=admin,dc=example,dc=org', ?)",
		user.ID, user.Created)
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./db/sql/ -run TestMigration_2_20_0 -v -count=1`
Expected: FAIL (`no such table: user__external_identity` or migration version unknown)

- [ ] **Step 3: Create the migration SQL**

`db/sql/migrations/v2.20.0.sql`:

```sql
create table `user__external_identity` (
  `id` integer primary key autoincrement,
  `user_id` int not null,
  `provider` varchar(64) not null,
  `external_uid` varchar(700) not null,
  `created` datetime not null,

  foreign key (`user_id`) references `user`(`id`) on delete cascade
);

create unique index `user__external_identity__provider_uid`
  on `user__external_identity`(`provider`, `external_uid`);
create index `user__external_identity__user_id`
  on `user__external_identity`(`user_id`);
```

- [ ] **Step 4: Register the migration**

In `db/Migration.go`, in `commonScripts` after `{Version: "2.19.11"},` add:

```go
	{Version: "2.20.0"},
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./db/sql/ -run TestMigration_2_20_0 -v -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add db/sql/migrations/v2.20.0.sql db/Migration.go db/sql/migration_2_20_0_test.go
git commit -m "feat(auth): add user__external_identity table"
```

---

### Task 2: Model, store interface and SQL implementation

**Files:**
- Create: `db/UserExternalIdentity.go`
- Create: `db/sql/external_identity.go`
- Modify: `db/Store.go` (new `ExternalIdentityManager` interface, embed into `Store`, add `UserExternalIdentityProps`)
- Modify: `db/sql/SqlDb.go:97` (gorp table mapping, after the `TaskParams` line)
- Test: `db/sql/external_identity_test.go`

**Interfaces:**
- Consumes: table from Task 1.
- Produces (used by Tasks 4 and 7):

```go
type UserExternalIdentity struct {
	ID          int       `db:"id" json:"id"`
	UserID      int       `db:"user_id" json:"user_id"`
	Provider    string    `db:"provider" json:"provider"`
	ExternalUID string    `db:"external_uid" json:"external_uid"`
	Created     time.Time `db:"created" json:"created"`
}

type ExternalIdentityManager interface {
	GetExternalIdentity(provider string, externalUID string) (UserExternalIdentity, error)
	GetUserExternalIdentities(userID int) ([]UserExternalIdentity, error)
	CreateExternalIdentity(identity UserExternalIdentity) (UserExternalIdentity, error)
	DeleteExternalIdentity(userID int, provider string) error
}
```

- [ ] **Step 1: Write the failing store test**

`db/sql/external_identity_test.go`:

```go
package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalIdentityCRUD(t *testing.T) {
	store := CreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe",
		Name:     "John Doe",
		Email:    "jdoe@example.com",
		External: true,
	})
	require.NoError(t, err)

	// Not found before creation.
	_, err = store.GetExternalIdentity("ldap", "cn=jdoe,dc=example,dc=org")
	assert.ErrorIs(t, err, db.ErrNotFound)

	created, err := store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Provider:    "ldap",
		ExternalUID: "cn=jdoe,dc=example,dc=org",
	})
	require.NoError(t, err)
	assert.NotZero(t, created.Created)

	found, err := store.GetExternalIdentity("ldap", "cn=jdoe,dc=example,dc=org")
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.UserID)

	// Unique (provider, external_uid).
	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Provider:    "ldap",
		ExternalUID: "cn=jdoe,dc=example,dc=org",
	})
	assert.Error(t, err)

	// Same user, second provider — allowed.
	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Provider:    "keycloak",
		ExternalUID: "8b53f1e0-0000-0000-0000-000000000000",
	})
	require.NoError(t, err)

	list, err := store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	assert.Len(t, list, 2)

	require.NoError(t, store.DeleteExternalIdentity(user.ID, "keycloak"))
	list, err = store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Cascade on user deletion.
	require.NoError(t, store.DeleteUser(user.ID))
	_, err = store.GetExternalIdentity("ldap", "cn=jdoe,dc=example,dc=org")
	assert.ErrorIs(t, err, db.ErrNotFound)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./db/sql/ -run TestExternalIdentityCRUD -v -count=1`
Expected: compile error (`GetExternalIdentity` undefined)

- [ ] **Step 3: Create the model**

`db/UserExternalIdentity.go`:

```go
package db

import "time"

// UserExternalIdentity links a Semaphore user to an identity at an external
// auth provider. Provider is "ldap" or a key of the oidc_providers config map.
// ExternalUID is the provider's stable user ID: the OIDC "sub" claim or the
// LDAP entry DN. Matching by this pair (instead of by email) prevents account
// takeover via reused/unverified emails.
type UserExternalIdentity struct {
	ID          int       `db:"id" json:"id"`
	UserID      int       `db:"user_id" json:"user_id"`
	Provider    string    `db:"provider" json:"provider"`
	ExternalUID string    `db:"external_uid" json:"external_uid"`
	Created     time.Time `db:"created" json:"created"`
}
```

- [ ] **Step 4: Declare interface, props and gorp mapping**

In `db/Store.go` next to `SessionManager`/`TokenManager`:

```go
type ExternalIdentityManager interface {
	GetExternalIdentity(provider string, externalUID string) (UserExternalIdentity, error)
	GetUserExternalIdentities(userID int) ([]UserExternalIdentity, error)
	CreateExternalIdentity(identity UserExternalIdentity) (UserExternalIdentity, error)
	DeleteExternalIdentity(userID int, provider string) error
}
```

Embed `ExternalIdentityManager` into the `Store` interface (next to `SessionManager`).

Next to `TokenProps` (db/Store.go ~line 707):

```go
var UserExternalIdentityProps = ObjectProps{
	TableName:         "user__external_identity",
	Type:              reflect.TypeOf(UserExternalIdentity{}),
	PrimaryColumnName: "id",
}
```

In `db/sql/SqlDb.go` after line 97 (`TaskParams` mapping):

```go
	d.sql.AddTableWithName(db.UserExternalIdentity{}, "user__external_identity").SetKeys(true, "id")
```

- [ ] **Step 5: Implement the SQL store**

`db/sql/external_identity.go`:

```go
package sql

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

func (d *SqlDb) CreateExternalIdentity(identity db.UserExternalIdentity) (db.UserExternalIdentity, error) {
	identity.Created = db.GetParsedTime(tz.Now())
	err := d.Sql().Insert(&identity)
	return identity, err
}

func (d *SqlDb) GetExternalIdentity(provider string, externalUID string) (identity db.UserExternalIdentity, err error) {
	err = d.selectOne(
		&identity,
		"select * from user__external_identity where provider=? and external_uid=?",
		provider, externalUID)
	return
}

func (d *SqlDb) GetUserExternalIdentities(userID int) (identities []db.UserExternalIdentity, err error) {
	_, err = d.selectAll(
		&identities,
		d.PrepareQuery("select * from user__external_identity where user_id=? order by created desc"),
		userID)
	return
}

func (d *SqlDb) DeleteExternalIdentity(userID int, provider string) error {
	_, err := d.exec(
		"delete from user__external_identity where user_id=? and provider=?",
		userID, provider)
	return err
}
```

(`selectOne` already maps `sql.ErrNoRows` → `db.ErrNotFound`, see `db/sql/SqlDb.go:439`.)

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./db/sql/ -run TestExternalIdentityCRUD -v -count=1`
Expected: PASS (`Connect()` executes `PRAGMA foreign_keys = ON` for sqlite — db/sql/SqlDb.go:100 — so the cascade assertion is reliable).

- [ ] **Step 7: Build everything (Store interface satisfied)**

Run: `go build ./...`
Expected: OK (any other `db.Store` implementers, e.g. in `pro_impl/`, must compile; if a pro wrapper embeds `db.Store` it inherits the methods automatically — verify).

- [ ] **Step 8: Commit**

```bash
git add db/UserExternalIdentity.go db/Store.go db/sql/SqlDb.go db/sql/external_identity.go db/sql/external_identity_test.go
git commit -m "feat(auth): external identity entity and store methods"
```

---

### Task 3: Config option `external_auth_email_matching`

**Files:**
- Modify: `util/config.go` (next to `PasswordLoginDisable`, ~line 577)
- Modify: `config.schema.yaml` — regenerate via the `semaphore-config-schema` skill after the code change
- Test: extend `util/config_test.go` only if a validation helper is added (none needed — plain string field)

**Interfaces:**
- Produces: `util.Config.ExternalAuthEmailMatching` (string: `""`/`"auto"` | `"always"` | `"never"`), consumed by Task 4.

- [ ] **Step 1: Add the field**

In `util/config.go`, next to `PasswordLoginDisable`:

```go
	// ExternalAuthEmailMatching controls whether an LDAP/OIDC login may be
	// linked to an existing user by email when no external identity record
	// exists yet:
	//   "auto" (default) - only external users without any linked identity
	//                      (one-time adoption of pre-2.20 accounts);
	//   "always"         - any external user (needed when the same person
	//                      logs in via several providers);
	//   "never"          - identities are matched strictly by provider ID.
	// Local (password) accounts are never matched regardless of the mode.
	ExternalAuthEmailMatching string `json:"external_auth_email_matching,omitempty" env:"SEMAPHORE_EXTERNAL_AUTH_EMAIL_MATCHING" rule:"^(auto|always|never)?$" default:"auto"`
```

Note: check how `rule:`/`default:` tags are handled in `util/config.go` config loading; if `default:` is not applied for strings automatically, treat empty string as `auto` in code (Task 4 does exactly that — do NOT rely on the tag).

- [ ] **Step 2: Build**

Run: `go build ./util/`
Expected: OK

- [ ] **Step 3: Regenerate config schema**

Invoke the `semaphore-config-schema` skill (it derives `config.schema.yaml` from `util.ConfigType`). In the field description mention that provider key `ldap` is reserved for the LDAP integration.

- [ ] **Step 4: Commit**

```bash
git add util/config.go config.schema.yaml
git commit -m "feat(auth): external_auth_email_matching config option"
```

---

### Task 4: Resolver `resolveExternalUser`

**Files:**
- Create: `api/login_identity.go`
- Test: `api/login_identity_test.go`

**Interfaces:**
- Consumes: `ExternalIdentityManager` methods (Task 2), `util.Config.ExternalAuthEmailMatching` (Task 3), existing `store.GetUserByLoginOrEmail`, `store.CreateUserWithoutPassword`, `store.UpdateUser`, `store.GetUser`.
- Produces (used by Tasks 5–6):

```go
type externalUserProfile struct {
	Provider    string // "ldap" or oidc provider key
	ExternalUID string // LDAP DN or OIDC sub
	Username    string
	Name        string
	Email       string
	MatchByUsername bool // LDAP legacy behavior: also match by username
}

func resolveExternalUser(store db.Store, p externalUserProfile) (db.User, error)
```

- [ ] **Step 1: Write the failing tests**

`api/login_identity_test.go`:

```go
package api

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentityTest(t *testing.T, emailMatching string) db.Store {
	t.Helper()
	// IMPORTANT: CreateTestStore() overwrites util.Config (db/sql/SqlDb.go:124),
	// so the option must be set AFTER creating the store.
	store := sql.CreateTestStore()
	util.Config.ExternalAuthEmailMatching = emailMatching
	return store
}

func ldapProfile(uid string, email string) externalUserProfile {
	return externalUserProfile{
		Provider:        "ldap",
		ExternalUID:     uid,
		Username:        "jdoe",
		Name:            "John Doe",
		Email:           email,
		MatchByUsername: true,
	}
}

func TestResolveExternalUser_CreatesUserAndIdentity(t *testing.T) {
	store := setupIdentityTest(t, "auto")

	user, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	require.NoError(t, err)
	assert.True(t, user.External)

	ids, err := store.GetUserExternalIdentities(user.ID)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, "cn=jdoe,dc=x", ids[0].ExternalUID)
}

func TestResolveExternalUser_FindsByIdentityAfterEmailChange(t *testing.T) {
	store := setupIdentityTest(t, "auto")

	first, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	require.NoError(t, err)

	// Email changed at the IdP: same identity, attributes synced, no new user.
	second, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "john@example.com"))
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "john@example.com", second.Email)
}

func TestResolveExternalUser_AutoAdoptsLegacyExternalUser(t *testing.T) {
	store := setupIdentityTest(t, "auto")

	// Pre-2.20 external user: no identity rows.
	legacy, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	user, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	require.NoError(t, err)
	assert.Equal(t, legacy.ID, user.ID)

	ids, err := store.GetUserExternalIdentities(legacy.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
}

func TestResolveExternalUser_AutoDoesNotAdoptPinnedUser(t *testing.T) {
	store := setupIdentityTest(t, "auto")

	// User already pinned to a different provider identity.
	pinned, err := resolveExternalUser(store, externalUserProfile{
		Provider: "keycloak", ExternalUID: "sub-1",
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)

	// Same email arrives from another provider: must NOT reuse the account.
	other, err := resolveExternalUser(store, externalUserProfile{
		Provider: "okta", ExternalUID: "sub-2",
		Username: "jdoe2", Name: "John Doe", Email: "jdoe@example.com",
	})
	if err == nil {
		assert.NotEqual(t, pinned.ID, other.ID)
	} else {
		// Unique email constraint may forbid the second user - also acceptable,
		// but it must NOT be a silent merge.
		assert.Error(t, err)
	}
}

func TestResolveExternalUser_AlwaysLinksSecondProvider(t *testing.T) {
	store := setupIdentityTest(t, "always")

	first, err := resolveExternalUser(store, externalUserProfile{
		Provider: "keycloak", ExternalUID: "sub-1",
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)

	second, err := resolveExternalUser(store, externalUserProfile{
		Provider: "okta", ExternalUID: "sub-2",
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)

	ids, err := store.GetUserExternalIdentities(first.ID)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestResolveExternalUser_NeverSkipsEmailMatching(t *testing.T) {
	store := setupIdentityTest(t, "never")

	_, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	// Legacy user is NOT adopted; creation of a duplicate-email user must fail
	// rather than silently merge.
	_, err = resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	assert.Error(t, err)
}

func TestResolveExternalUser_NeverAdoptsLocalUser(t *testing.T) {
	store := setupIdentityTest(t, "always")

	local, err := store.CreateUser(db.UserWithPwd{
		Pwd: "verystrongpassword1",
		User: db.User{
			Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
		},
	})
	require.NoError(t, err)

	res, err := resolveExternalUser(store, ldapProfile("cn=jdoe,dc=x", "jdoe@example.com"))
	if err == nil {
		assert.NotEqual(t, local.ID, res.ID)
	} else {
		assert.Error(t, err)
	}

	// The local account must remain identity-free.
	ids, err := store.GetUserExternalIdentities(local.ID)
	require.NoError(t, err)
	assert.Empty(t, ids)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./api/ -run TestResolveExternalUser -v -count=1`
Expected: compile error (`resolveExternalUser` undefined)

- [ ] **Step 3: Implement the resolver**

`api/login_identity.go`:

```go
package api

import (
	"errors"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

// externalUserProfile is what an external auth flow (LDAP or OIDC) learned
// about the user. ExternalUID must be the provider's stable ID: LDAP DN or
// OIDC "sub" claim.
type externalUserProfile struct {
	Provider        string
	ExternalUID     string
	Username        string
	Name            string
	Email           string
	MatchByUsername bool // LDAP legacy behavior: also match by username
}

// resolveExternalUser maps an external identity to a Semaphore user:
//  1. by (provider, external_uid) — the only trusted key;
//  2. by email/username — only under external_auth_email_matching mode,
//     only External users (local accounts are never adopted);
//  3. otherwise a new user is created and linked.
func resolveExternalUser(store db.Store, p externalUserProfile) (db.User, error) {
	if p.ExternalUID == "" {
		return db.User{}, errors.New("external identity: empty external UID")
	}

	identity, err := store.GetExternalIdentity(p.Provider, p.ExternalUID)

	switch {
	case err == nil:
		user, err2 := store.GetUser(identity.UserID)
		if err2 != nil {
			return db.User{}, err2
		}
		return syncExternalUserAttrs(store, user, p)
	case !errors.Is(err, db.ErrNotFound):
		return db.User{}, err
	}

	user, err := matchExternalUserByEmail(store, p)

	switch {
	case err == nil:
		if _, err = store.CreateExternalIdentity(db.UserExternalIdentity{
			UserID:      user.ID,
			Provider:    p.Provider,
			ExternalUID: p.ExternalUID,
		}); err != nil {
			return db.User{}, err
		}
		return syncExternalUserAttrs(store, user, p)
	case !errors.Is(err, db.ErrNotFound):
		return db.User{}, err
	}

	user, err = store.CreateUserWithoutPassword(db.User{
		Username: p.Username,
		Name:     p.Name,
		Email:    p.Email,
		External: true,
	})
	if err != nil {
		return db.User{}, err
	}

	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Provider:    p.Provider,
		ExternalUID: p.ExternalUID,
	})
	return user, err
}

// matchExternalUserByEmail implements the legacy email/username matching,
// gated by the external_auth_email_matching config mode.
func matchExternalUserByEmail(store db.Store, p externalUserProfile) (db.User, error) {
	mode := util.Config.ExternalAuthEmailMatching
	if mode == "" {
		mode = "auto" // ponytail: default applied here, not via config tag
	}

	if mode == "never" {
		return db.User{}, db.ErrNotFound
	}

	login := ""
	if p.MatchByUsername {
		login = p.Username
	}

	user, err := store.GetUserByLoginOrEmail(login, p.Email)
	if err != nil {
		return db.User{}, err
	}

	// Local accounts are never adopted by external providers - this is the
	// takeover-protection invariant, independent of the mode.
	if !user.External {
		return db.User{}, db.ErrNotFound
	}

	if mode == "auto" {
		identities, err2 := store.GetUserExternalIdentities(user.ID)
		if err2 != nil {
			return db.User{}, err2
		}
		if len(identities) > 0 {
			// Already pinned to some identity - email matching no longer applies.
			return db.User{}, db.ErrNotFound
		}
	}

	return user, nil
}

// syncExternalUserAttrs updates name/email from the provider on each login,
// so an email change at the IdP is reflected instead of orphaning the account.
func syncExternalUserAttrs(store db.Store, user db.User, p externalUserProfile) (db.User, error) {
	changed := false
	if p.Email != "" && user.Email != p.Email {
		user.Email = p.Email
		changed = true
	}
	if p.Name != "" && user.Name != p.Name {
		user.Name = p.Name
		changed = true
	}
	if !changed {
		return user, nil
	}
	err := store.UpdateUser(db.UserWithPwd{User: user})
	return user, err
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./api/ -run TestResolveExternalUser -v -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add api/login_identity.go api/login_identity_test.go
git commit -m "feat(auth): external user resolver with identity linking"
```

---

### Task 5: Wire the LDAP flow

**Files:**
- Modify: `api/login.go:46-149` (`tryFindLDAPUser` — return DN), `api/login.go:241-258` (`loginByLDAP`), `api/login.go:343-361` (call site in `login`)

**Interfaces:**
- Consumes: `resolveExternalUser` (Task 4).
- Produces: `tryFindLDAPUser(username, password string) (*db.User, string, error)` — third return is the entry DN.

- [ ] **Step 1: Change `tryFindLDAPUser` to return the DN**

Signature: `func tryFindLDAPUser(username, password string) (*db.User, string, error)`.
All existing `return nil, err` → `return nil, "", err`; `return nil, nil` (user not found) → `return nil, "", nil`.
The DN is already computed at `api/login.go:94` (`userDN := sr.Entries[0].DN`); final return becomes:

```go
	return &ldapUser, userDN, nil
```

- [ ] **Step 2: Replace `loginByLDAP` body**

```go
func loginByLDAP(store db.Store, ldapUser db.User, userDN string) (db.User, error) {
	return resolveExternalUser(store, externalUserProfile{
		Provider:        "ldap",
		ExternalUID:     userDN,
		Username:        ldapUser.Username,
		Name:            ldapUser.Name,
		Email:           ldapUser.Email,
		MatchByUsername: true,
	})
}
```

(The old `!user.External` check moved into `matchExternalUserByEmail` — behavior preserved: LDAP login into a local account is still rejected.)

- [ ] **Step 3: Update the call site in `login()`**

```go
	var ldapUser *db.User
	var ldapUserDN string

	if util.Config.LdapEnable {
		ldapUser, ldapUserDN, err = tryFindLDAPUser(login.Auth, login.Password)
		...
	}

	...
	} else {
		user, err = loginByLDAP(helpers.Store(r), *ldapUser, ldapUserDN)
	}
```

(Keep the surrounding error handling exactly as it is today.)

- [ ] **Step 4: Run the api and db tests**

Run: `go test ./api/... ./db/... -count=1`
Expected: PASS (including existing `api/login_test.go`)

- [ ] **Step 5: Commit**

```bash
git add api/login.go
git commit -m "feat(auth): link LDAP logins via external identity (DN)"
```

---

### Task 6: Wire the OIDC flow (`sub` claim)

**Files:**
- Modify: `api/login.go:571-575` (`claimResult` — add `sub`), `api/login.go:652-672` (`claimOidcUserInfo`, `claimOidcToken` — fill `sub`), `api/login.go:752-810` (`oidcRedirect` — use resolver)

**Interfaces:**
- Consumes: `resolveExternalUser` (Task 4).
- Produces: nothing new.

- [ ] **Step 1: Capture `sub`**

```go
type claimResult struct {
	sub      string
	username string
	name     string
	email    string
}
```

In `claimOidcToken` (`api/login.go:663`), after parsing claims:

```go
	res, err = parseClaims(claims, &provider)
	res.sub = idToken.Subject
	return
```

In `claimOidcUserInfo` (`api/login.go:652`), same pattern:

```go
	res, err = parseClaims(claims, &provider)
	res.sub = userInfo.Subject
	return
```

In the `oidcRedirect` userinfo fallback branch (`api/login.go:765-782`), where `claims.email`/`claims.name` are set directly from `userInfo`, also set:

```go
	claims.sub = userInfo.Subject
```

- [ ] **Step 2: Replace the email-lookup block in `oidcRedirect`**

Replace `api/login.go:790-810` (from `user, err := helpers.Store(r).GetUserByLoginOrEmail(...)` through the `!user.External` check) with:

```go
	if claims.sub == "" {
		log.Error(fmt.Errorf("oidc provider %s returned no sub claim", pid))
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	user, err := resolveExternalUser(helpers.Store(r), externalUserProfile{
		Provider:    pid,
		ExternalUID: claims.sub,
		Username:    claims.username,
		Name:        claims.name,
		Email:       claims.email,
		// MatchByUsername stays false: OIDC matches by email only
		// (username matching "creates a lot of problems" - see old comment).
	})
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}
```

- [ ] **Step 3: Run tests**

Run: `go test ./api/... -count=1`
Expected: PASS

- [ ] **Step 4: Manual smoke test (optional but recommended)**

Configure a local Keycloak/Authentik OIDC provider, log in twice, change the user's email at the IdP between logins:
- 1st login: user created, row in `user__external_identity` with `provider=<pid>`, `external_uid=<sub>`.
- 2nd login: same user ID, email updated, still one identity row.

- [ ] **Step 5: Commit**

```bash
git add api/login.go
git commit -m "feat(auth): link OIDC logins via external identity (sub claim)"
```

---

### Task 7: Admin API — view and unlink identities

**Files:**
- Modify: `api/users.go` (two handlers on the users controller)
- Modify: `api/router.go:266-270` (routes under `userAPI`, which already applies `GetUserMiddleware`; admin enforcement as in `UpdateUser`/`DeleteUser` handlers)
- Test: `api/users_identity_test.go`

**Interfaces:**
- Consumes: `GetUserExternalIdentities`, `DeleteExternalIdentity` (Task 2).
- Produces: `GET /api/users/{user_id}/identities` → `[]db.UserExternalIdentity`; `DELETE /api/users/{user_id}/identities/{provider}` → 204.

- [ ] **Step 1: Write the failing test**

`api/users_identity_test.go` (follow the request-builder pattern of `api/user_options_test.go` — context values `store` and `user`; check how existing users handlers read the target user from context via `GetUserMiddleware` and mirror that setup):

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAndDeleteUserIdentities(t *testing.T) {
	util.Config = &util.ConfigType{}
	store := sql.CreateTestStore()

	admin, err := store.CreateUser(db.UserWithPwd{
		Pwd:  "verystrongpassword1",
		User: db.User{Username: "root", Name: "Root", Email: "root@example.com", Admin: true},
	})
	require.NoError(t, err)

	target, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com", External: true,
	})
	require.NoError(t, err)

	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID: target.ID, Provider: "ldap", ExternalUID: "cn=jdoe,dc=x",
	})
	require.NoError(t, err)

	// GET list
	r := httptest.NewRequest(http.MethodGet, "/api/users/2/identities", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &admin)
	r = helpers.SetContextValue(r, "_user", target) // GetUserMiddleware sets "_user" as db.User value (api/users.go:169)
	w := httptest.NewRecorder()
	usersController := NewUsersController(nil)
	usersController.GetUserIdentities(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "cn=jdoe,dc=x")

	// DELETE unlink
	r = httptest.NewRequest(http.MethodDelete, "/api/users/2/identities/ldap", nil)
	r = helpers.SetContextValue(r, "store", store)
	r = helpers.SetContextValue(r, "user", &admin)
	r = helpers.SetContextValue(r, "_user", target)
	r = muxSetVars(r, map[string]string{"provider": "ldap"}) // helper: mux.SetURLVars
	w = httptest.NewRecorder()
	usersController.DeleteUserIdentity(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)

	ids, err := store.GetUserExternalIdentities(target.ID)
	require.NoError(t, err)
	assert.Empty(t, ids)
}
```

Context keys verified: `GetUserMiddleware` (api/users.go:143-172) puts the target as `"_user"` (`db.User` value) and enforces admin-or-self itself; the editor is `"user"` (`*db.User`). `muxSetVars` is `mux.SetURLVars` from gorilla/mux.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./api/ -run TestGetAndDeleteUserIdentities -v -count=1`
Expected: compile error (`GetUserIdentities` undefined)

- [ ] **Step 3: Implement handlers**

In `api/users.go`, following the style of the neighboring handlers (same middleware-provided target user, same admin guard as `UpdateUser`):

```go
func (c *UsersController) GetUserIdentities(w http.ResponseWriter, r *http.Request) {
	targetUser := helpers.GetFromContext(r, "_user").(db.User)

	identities, err := helpers.Store(r).GetUserExternalIdentities(targetUser.ID)
	if err != nil {
		helpers.WriteErrorStatus(w, "Failed to get identities", http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, identities)
}

func (c *UsersController) DeleteUserIdentity(w http.ResponseWriter, r *http.Request) {
	targetUser := helpers.GetFromContext(r, "_user").(db.User)
	provider := mux.Vars(r)["provider"]

	if err := helpers.Store(r).DeleteExternalIdentity(targetUser.ID, provider); err != nil {
		helpers.WriteErrorStatus(w, "Failed to delete identity", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Register routes**

In `api/router.go`, next to the `userPasswordAPI` block (line 272, which already uses `GetUserMiddleware`):

```go
	userPasswordAPI.Path("/identities").HandlerFunc(usersController.GetUserIdentities).Methods("GET", "HEAD")
	userPasswordAPI.Path("/identities/{provider}").HandlerFunc(usersController.DeleteUserIdentity).Methods("DELETE")
```

`GetUserMiddleware` enforces admin-or-self (api/users.go:160), so a user may also view/unlink their own identities — intended. If product decision changes to admin-only, add an `editor.Admin` check in the handlers like `UpdateUser` does for role changes.

- [ ] **Step 5: Run tests**

Run: `go test ./api/ -run TestGetAndDeleteUserIdentities -v -count=1`
Expected: PASS

- [ ] **Step 6: Full test run and commit**

Run: `go test ./api/... ./db/... ./util/... -count=1`
Expected: PASS

```bash
git add api/users.go api/router.go api/users_identity_test.go
git commit -m "feat(auth): admin API to list and unlink external identities"
```

---

## Out of scope (deliberate, YAGNI)

- Storing IdP tokens in the identity row — add together with a refresh-token feature.
- UI for identities (frontend) — API first; UI is a separate plan.
- Group/role mapping — separate plan (see research note "Дизайн маппинга ролей из LDAP/SSO"); this table is its prerequisite (mappings become per-provider).
- Backfill migration that guesses identities for existing users — impossible to do safely (no stored sub/DN); `auto` mode handles it lazily and self-closes.

## Verification checklist (after all tasks)

- [ ] `go build ./...` and `go test ./... -count=1` green (including `pro_impl` if checked out).
- [ ] Fresh install: LDAP + OIDC login create identity rows; second login reuses them.
- [ ] Upgrade simulation: create user with `External=true` and no identity rows, log in via LDAP with matching email → adopted, identity created; repeat with `external_auth_email_matching: never` → login fails (no silent merge).
- [ ] Email change at IdP: same account, updated email.
- [ ] Local account with same email: LDAP/OIDC login never enters it in any mode.
