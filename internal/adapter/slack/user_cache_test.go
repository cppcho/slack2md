package slack

import (
	"sync"
	"testing"
)

func TestNewUserCache(t *testing.T) {
	cache := NewUserCache()

	if cache == nil {
		t.Fatal("Expected cache to be created, got nil")
	}
	if cache.cache == nil {
		t.Error("Expected cache map to be initialized")
	}
}

func TestUserCache_SetAndGet(t *testing.T) {
	cache := NewUserCache()

	// Set a user display name
	userID := "U123456"
	displayName := "John Doe"
	cache.Set(userID, displayName)

	// Get the user display name
	retrieved, ok := cache.Get(userID)

	if !ok {
		t.Error("Expected to find user in cache")
	}
	if retrieved != displayName {
		t.Errorf("Expected display name = %q, got %q", displayName, retrieved)
	}
}

func TestUserCache_GetNonExistent(t *testing.T) {
	cache := NewUserCache()

	// Try to get a non-existent user
	_, ok := cache.Get("U999999")

	if ok {
		t.Error("Expected not to find non-existent user in cache")
	}
}

func TestUserCache_OverwriteValue(t *testing.T) {
	cache := NewUserCache()

	userID := "U123456"

	// Set initial value
	cache.Set(userID, "Original Name")

	// Overwrite with new value
	cache.Set(userID, "Updated Name")

	// Verify new value is retrieved
	retrieved, ok := cache.Get(userID)

	if !ok {
		t.Error("Expected to find user in cache")
	}
	if retrieved != "Updated Name" {
		t.Errorf("Expected updated name = %q, got %q", "Updated Name", retrieved)
	}
}

func TestUserCache_MultipleUsers(t *testing.T) {
	cache := NewUserCache()

	// Set multiple users
	users := map[string]string{
		"U111": "Alice",
		"U222": "Bob",
		"U333": "Charlie",
	}

	for id, name := range users {
		cache.Set(id, name)
	}

	// Verify all users can be retrieved
	for id, expectedName := range users {
		retrieved, ok := cache.Get(id)
		if !ok {
			t.Errorf("Expected to find user %s in cache", id)
		}
		if retrieved != expectedName {
			t.Errorf("User %s: expected %q, got %q", id, expectedName, retrieved)
		}
	}
}

func TestUserCache_ThreadSafety(t *testing.T) {
	cache := NewUserCache()

	// Run concurrent operations
	var wg sync.WaitGroup
	concurrency := 100

	// Concurrent writes
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cache.Set("U123", "User")
		}(i)
	}

	// Concurrent reads
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Get("U123")
		}()
	}

	wg.Wait()

	// Verify cache is still functional
	cache.Set("U456", "Test User")
	name, ok := cache.Get("U456")
	if !ok {
		t.Error("Cache corrupted after concurrent operations")
	}
	if name != "Test User" {
		t.Errorf("Expected %q, got %q", "Test User", name)
	}
}

func TestUserCache_EmptyValues(t *testing.T) {
	cache := NewUserCache()

	// Set empty display name
	cache.Set("U123", "")

	// Should still be retrievable
	retrieved, ok := cache.Get("U123")

	if !ok {
		t.Error("Expected to find user with empty display name")
	}
	if retrieved != "" {
		t.Errorf("Expected empty string, got %q", retrieved)
	}
}
