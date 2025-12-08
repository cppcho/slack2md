package slack

import "testing"

func TestNewClient(t *testing.T) {
	client := NewClient("xoxb-test-token", "xapp-test-token")
	if client == nil {
		t.Error("Expected client to be created")
	}
	if client.api == nil {
		t.Error("Expected API client to be initialized")
	}
}

func TestNewClient_WithoutAppToken(t *testing.T) {
	client := NewClient("xoxb-test-token", "")
	if client == nil {
		t.Error("Expected client to be created")
	}
	if client.api == nil {
		t.Error("Expected API client to be initialized")
	}
}

func TestUserCache_GetSet(t *testing.T) {
	cache := NewUserCache()

	// Test Get on empty cache
	_, ok := cache.Get("user123")
	if ok {
		t.Error("Expected cache miss for non-existent user")
	}

	// Test Set and Get
	cache.Set("user123", "John Doe")
	name, ok := cache.Get("user123")
	if !ok {
		t.Error("Expected cache hit after setting value")
	}
	if name != "John Doe" {
		t.Errorf("Expected 'John Doe', got '%s'", name)
	}
}

func TestUserCache_Concurrent(t *testing.T) {
	cache := NewUserCache()

	// Test concurrent access
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("user1", "Name1")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Get("user1")
		}
		done <- true
	}()

	// Wait for completion
	<-done
	<-done
}
