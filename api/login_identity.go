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
