package api

import (
	"errors"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

// Sentinel errors let the OIDC redirect handler map link failures to
// user-visible error codes (see oidcRedirect).
var (
	errIdentityLinkedToAnother = errors.New("external identity is already linked to another account")
	errProviderAlreadyLinked   = errors.New("account already has an identity for this provider, unlink it first")
	errCannotUnlinkLastIdentity = errors.New("cannot unlink the last external identity")
)

// externalUserProfile is what an external auth flow (LDAP or OIDC) learned
// about the user. ExternalUID must be the provider's stable ID: LDAP DN or
// OIDC "sub" claim.
type externalUserProfile struct {
	Type            string // db.IdentityTypeLdap or db.IdentityTypeOidc
	Provider        string
	ExternalUID     string
	Username        string
	Name            string
	Email           string
	MatchByUsername bool // LDAP legacy behavior: also match by username
	// EmailVerified means the provider proved the user owns Email (OIDC
	// email_verified claim; for LDAP the directory is authoritative).
	// Unverified emails are never used to match existing accounts: an IdP
	// that lets users type in any email must not be able to adopt someone
	// else's account (Grafana CVE-2023-3128 class).
	EmailVerified bool
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

	if p.Type == "" {
		return db.User{}, errors.New("external identity: empty type")
	}

	identity, err := store.GetExternalIdentity(p.Type, p.Provider, p.ExternalUID)

	switch {
	case err == nil:
		user, uErr := store.GetUser(identity.UserID)
		if uErr != nil {
			return db.User{}, uErr
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
			Type:        p.Type,
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
		Type:        p.Type,
		Provider:    p.Provider,
		ExternalUID: p.ExternalUID,
	})
	if err != nil {
		// Roll back the just-created user, otherwise it stays orphaned and
		// (with email matching off) every retry dies on duplicate username.
		// Best-effort: the login already failed, DeleteUser error adds nothing.
		_ = store.DeleteUser(user.ID)
		return db.User{}, err
	}

	return user, nil
}

// matchExternalUserByEmail implements the legacy email/username matching,
// gated by the external_auth_email_matching config mode.
func matchExternalUserByEmail(store db.Store, p externalUserProfile) (db.User, error) {
	if util.Config.ExternalAuthEmailMatching == "never" {
		return db.User{}, db.ErrNotFound
	}

	email := ""
	if p.EmailVerified {
		email = p.Email
	}

	username := ""
	if p.MatchByUsername {
		username = p.Username
	}

	if email == "" && username == "" {
		return db.User{}, db.ErrNotFound
	}

	user, err := store.GetUserByLoginOrEmail(username, email)
	if err != nil {
		return db.User{}, err
	}

	// Local accounts are never adopted by external providers - this is the
	// takeover-protection invariant, independent of the mode.
	if !user.External {
		return db.User{}, db.ErrNotFound
	}

	if util.Config.ExternalAuthEmailMatching == "auto" {
		identities, idErr := store.GetUserExternalIdentities(user.ID)
		if idErr != nil {
			return db.User{}, idErr
		}
		if len(identities) > 0 {
			// Already pinned to some identity - email matching no longer applies.
			return db.User{}, db.ErrNotFound
		}
	}

	return user, nil
}

// ldapProfileMatchesSemaphoreUser checks that the LDAP directory profile
// belongs to the Semaphore account being linked. A successful bind only proves
// ownership of the LDAP credentials, not that they correspond to this account.
func ldapProfileMatchesSemaphoreUser(ldapUser, semaphoreUser db.User) bool {
	if ldapUser.Email != "" && strings.EqualFold(ldapUser.Email, semaphoreUser.Email) {
		return true
	}
	if ldapUser.Username != "" && strings.EqualFold(ldapUser.Username, semaphoreUser.Username) {
		return true
	}
	return false
}

// linkExternalIdentity attaches (idType, provider, externalUID) to user. Proof
// of ownership is the caller's job (active verified session + full auth flow
// at the provider) - email is never proof, see Grafana CVE-2023-3128.
func linkExternalIdentity(store db.Store, user db.User, idType string, provider string, externalUID string) error {
	if externalUID == "" {
		return errors.New("external identity: empty external UID")
	}

	existing, err := store.GetExternalIdentity(idType, provider, externalUID)
	switch {
	case err == nil:
		if existing.UserID == user.ID {
			return nil // already linked - idempotent
		}
		return errIdentityLinkedToAnother
	case !errors.Is(err, db.ErrNotFound):
		return err
	}

	identities, err := store.GetUserExternalIdentities(user.ID)
	if err != nil {
		return err
	}
	for _, identity := range identities {
		if identity.Type == idType && identity.Provider == provider {
			return errProviderAlreadyLinked
		}
	}

	_, err = store.CreateExternalIdentity(db.UserExternalIdentity{
		UserID:      user.ID,
		Type:        idType,
		Provider:    provider,
		ExternalUID: externalUID,
	})
	return err
}

// syncExternalUserAttrs updates name/email from the provider on each login,
// so an email change at the IdP is reflected instead of orphaning the account.
func syncExternalUserAttrs(store db.Store, user db.User, p externalUserProfile) (db.User, error) {
	// A linked local account keeps its local profile - the IdP is not
	// authoritative for non-External users.
	if !user.External {
		return user, nil
	}

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
