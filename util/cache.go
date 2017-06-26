// Handle the password caching
package util

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

type CachedUser struct {
	Password   string
	UsageCount int
	LoggedIn   bool
}

type VaultCache struct {
	store *cache.Cache
}

var UserVaultCache *VaultCache

func NewVaultCache() *VaultCache {
	var vc *VaultCache
	vc = new(VaultCache)
	vc.store = cache.New(cache.NoExpiration, time.Minute*5)
	return vc
}

func (vc *VaultCache) Get(User string) (*CachedUser, error) {
	var cu *CachedUser
	item, exists := vc.store.Get(User)
	if exists == false {
		return nil, errors.New("User not cached")
	}
	cu = item.(*CachedUser)
	return cu, nil
}

func (vc *VaultCache) AddEntry(User string, Password string) error {
	cu, err := vc.Get(User)
	if err == nil {
		if cu.Password != Password || cu.LoggedIn == false {
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

func (vc *VaultCache) IncrementUsage(User string) error {
	cu, err := vc.Get(User)
	if err != nil {
		return err
	}
	if cu.LoggedIn == false {
		return errors.New("User not logged in")
	}
	cu.UsageCount++
	return err
}

func (vc *VaultCache) DecrementUsage(User string) error {
	cu, err := vc.Get(User)
	if err != nil {
		return err
	}
	if cu.UsageCount == 0 {
		return errors.New("No recorded usage")
	}
	if cu.LoggedIn == false {
		vc.store.Delete(User)
	} else {
		cu.UsageCount--
	}
	return err
}

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

func (vc *VaultCache) GetPassword(User string) (string, error) {
	cu, err := vc.Get(User)
	if err != nil {
		return "", err
	}
	return cu.Password, nil
}
