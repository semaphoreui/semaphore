// Package util component to handle the password caching.
package util

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

// Type CachedUser stores a user password in the in-Memory-Cache together with
// the number of concurrent usages and the info if the user is still logged in.
// These are used to determine expiration.
type CachedUser struct {
	Password   string
	UsageCount int
	LoggedIn   bool
}

// VaultCache serves as abstraction layer between these functions and the
// cache library we use
type VaultCache struct {
	store *cache.Cache
}

// UserVaultCache is the global variable holding the cache.
var UserVaultCache *VaultCache

// NewVaultCache creates/initializes a new cache.
func NewVaultCache() *VaultCache {
	vc := new(VaultCache)
	vc.store = cache.New(cache.NoExpiration, time.Minute*5)
	return vc
}

// Get retrieves CachedUser by its username.
func (vc *VaultCache) Get(User string) (*CachedUser, error) {
	var cu *CachedUser
	item, exists := vc.store.Get(User)
	if !exists {
		return nil, errors.New("User not cached")
	}
	cu = item.(*CachedUser)
	return cu, nil
}

// AddEntry adds a new CachedUser to the cache.
func (vc *VaultCache) AddEntry(User string, Password string) error {
	cu, err := vc.Get(User)
	if err == nil {
		if cu.Password != Password || !cu.LoggedIn {
			cu.Password = Password
			cu.LoggedIn = true
			err := vc.store.Replace(User, cu, cache.NoExpiration)
			if err != nil {
				return err
			}
		}
	} else {
		cu = new(CachedUser)
		cu.Password = Password
		cu.LoggedIn = true
		cu.UsageCount = 0
		err := vc.store.Add(User, cu, cache.NoExpiration)
		if err != nil {
			return err
		}
	}
	return nil
}

// IncrementUsage increments the usage count of the cached credentials for the
// given User. This information is required for the expiration handling.
func (vc *VaultCache) IncrementUsage(User string) error {
	cu, err := vc.Get(User)
	if err != nil {
		return err
	}
	if !cu.LoggedIn {
		return errors.New("User not logged in")
	}
	cu.UsageCount++
	return err
}

// DecrementUsage increments the usage count of the cached credentials for the
// given User. This information is required for the expiration handling.
func (vc *VaultCache) DecrementUsage(User string) error {
	cu, err := vc.Get(User)
	if err != nil {
		return err
	}
	if cu.UsageCount == 0 {
		return errors.New("No recorded usage")
	}
	if !cu.LoggedIn {
		vc.store.Delete(User)
	} else {
		cu.UsageCount--
	}
	return err
}

// LogOut marks the User as logged out, preventing further usage of the stored
// credentials. Running tasks can still access the credentials but starting new
// tasks with them is not allowed any more.
func (vc *VaultCache) LogOut(User string) error {
	cu, err := vc.Get(User)
	if err != nil {
		return err
	}
	if cu.UsageCount == 0 {
		vc.store.Delete(User)
	} else {
		cu.LoggedIn = false
	}
	return nil
}

// GetPassword retrieves the password for a given User
func (vc *VaultCache) GetPassword(User string) (string, error) {
	cu, err := vc.Get(User)
	if err != nil {
		return "", err
	}
	return cu.Password, nil
}
