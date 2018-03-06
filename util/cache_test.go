package util

import (
	"testing"
)

func TestUserCache(t *testing.T) {
	var TestCache *VaultCache

	TestCache = NewVaultCache()

	t.Run("EmptyGet", func(t *testing.T) {
		_, err := TestCache.Get("nonexistent_user")
		if err == nil {
			t.Error("Succeeded to get password for a nonexistent user")
		}
	})

	t.Run("EmptyIncrease", func(t *testing.T) {
		err := TestCache.IncrementUsage("nonexistent_user")
		if err == nil {
			t.Error("Succeeded to increment usage for a nonexistent user")
		}
	})

	t.Run("EmptyDecrease", func(t *testing.T) {
		err := TestCache.DecrementUsage("nonexistent_user")
		if err == nil {
			t.Error("Succeeded to decrement usage for a nonexistent user")
		}
	})

	t.Run("EmptyLogout", func(t *testing.T) {
		err := TestCache.LogOut("nonexistent_user")
		if err == nil {
			t.Error("Succeeded to log out nonexistent user")
		}
	})

	t.Run("AddUser", func(t *testing.T) {
		err := TestCache.AddEntry("JaneDoe", "T0pS3cr3t!")
		if err != nil {
			t.Errorf("Failed to add user JaneDoe: %v", err)
		}
	})
	t.Run("UpdateUser", func(t *testing.T) {
		err := TestCache.AddEntry("JaneDoe", "123456")
		if err != nil {
			t.Errorf("Failed to re-add user JaneDoe: %v", err)
		}
		pass, err := TestCache.GetPassword("JaneDoe")
		if err != nil {
			t.Errorf("Failed to retrieve password: %v", err)
		}
		if pass != "123456" {
			t.Error("Failed to update password")
		}
	})
	t.Run("DeleteUserOnLogoutWithoutTasks", func(t *testing.T) {
		err := TestCache.LogOut("JaneDoe")
		if err != nil {
			t.Errorf("Failed to log out user JaneDoe: %v", err)
		}
		uc, err := TestCache.Get("JaneDoe")
		if err == nil {
			t.Errorf("User JaneDoe still exists after logging out. Usage count is %d", uc.UsageCount)
		}
	})
	t.Run("Increase", func(t *testing.T) {
		err := TestCache.AddEntry("JohnDoe", "T0pS3cr3t!")
		if err != nil {
			t.Errorf("Failed to add user JohnDoe: %v", err)
		}
		err = TestCache.IncrementUsage("JohnDoe")
		if err != nil {
			t.Errorf("Failed to increment usage count: %v", err)
		}
		uc, err := TestCache.Get("JohnDoe")
		if err != nil {
			t.Errorf("Failed to retrieve user data for JohnDoe: %v", err)
		}
		if uc.UsageCount != 1 {
			t.Errorf("Wrong usage count, expected 1, got %d", uc.UsageCount)
		}
	})

	t.Run("Decrease", func(t *testing.T) {
		err := TestCache.DecrementUsage("JohnDoe")
		if err != nil {
			t.Errorf("Failed to decrement usage count: %v", err)
		}
		uc, err := TestCache.Get("JohnDoe")
		if err != nil {
			t.Errorf("Failed to retrieve user data for JohnDoe: %v", err)
		}
		if uc.UsageCount != 0 {
			t.Errorf("Wrong usage count, expected 0, got %d", uc.UsageCount)
		}
	})

	t.Run("DecreaseBelowZero", func(t *testing.T) {
		err := TestCache.DecrementUsage("JohnDoe")
		if err == nil {
			uc, err := TestCache.Get("JohnDoe")
			if err != nil {
				t.Errorf("Failed to retrieve user data for JohnDoe: %v", err)
			}
			t.Errorf("Decrement usage count below zero, got %d", uc.UsageCount)
		}
	})

	t.Run("KeepUserWithTasksAfterLogout", func(t *testing.T) {
		err := TestCache.IncrementUsage("JohnDoe")
		if err != nil {
			t.Errorf("Failed to increment usage count: %v", err)
		}
		uc, err := TestCache.Get("JohnDoe")
		if err != nil {
			t.Errorf("Failed to retrieve user data for JohnDoe: %v", err)
		}
		if uc.UsageCount != 1 {
			t.Errorf("Wrong usage count, expected 1, got %d", uc.UsageCount)
		}
		err = TestCache.LogOut("JohnDoe")
		if err != nil {
			t.Errorf("Failed to log out JohnDoe: %v", err)
		}
		uc, err = TestCache.Get("JohnDoe")
		if err != nil {
			t.Errorf("Failed to get logged out user with tasks: %v", err)
		}
	})

	t.Run("RemoveUserOnZeroUsage", func(t *testing.T) {
		err := TestCache.DecrementUsage("JohnDoe")
		if err != nil {
			t.Errorf("Failed to decrement usage count: %v", err)
		}
		uc, err := TestCache.Get("JohnDoe")
		if err == nil {
			t.Errorf("Retrieved logged out user JohnDoe with expected usage count 0, got %d", uc.UsageCount)
		}
	})

}
