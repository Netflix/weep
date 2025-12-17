package session

import (
	"testing"
	"time"
)

func TestRandomString(t *testing.T) {
	t.Run("generates string of correct length", func(t *testing.T) {
		length := 64
		result := randomString(length)
		if len(result) != length {
			t.Errorf("expected length %d, got %d", length, len(result))
		}
	})

	t.Run("generates different strings on multiple calls", func(t *testing.T) {
		result1 := randomString(64)
		result2 := randomString(64)
		if result1 == result2 {
			t.Error("expected different strings but got the same")
		}
	})

	t.Run("only contains valid characters", func(t *testing.T) {
		result := randomString(100)
		for _, char := range result {
			found := false
			for _, validChar := range letters {
				if char == validChar {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("invalid character %c in random string", char)
			}
		}
	})

	t.Run("handles zero length", func(t *testing.T) {
		result := randomString(0)
		if len(result) != 0 {
			t.Errorf("expected empty string, got length %d", len(result))
		}
	})
}

func TestCreateCache(t *testing.T) {
	t.Run("creates a new cache", func(t *testing.T) {
		cache := createCache()
		if cache == nil {
			t.Error("expected non-nil cache")
		}
	})
}

func TestTokenCache_Set(t *testing.T) {
	t.Run("sets a token with attributes", func(t *testing.T) {
		cache := &tokenCache{}
		token := "test-token"
		role := "test-role"
		ttl := 3600

		cache.Set(token, role, ttl)

		attr, err := cache.Get(token)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if attr.Role != role {
			t.Errorf("expected role %s, got %s", role, attr.Role)
		}
		if attr.InitialTtl != ttl {
			t.Errorf("expected ttl %d, got %d", ttl, attr.InitialTtl)
		}
		if !attr.Expiration.After(time.Now()) {
			t.Error("expected expiration to be in the future")
		}
	})

	t.Run("sets expiration time correctly", func(t *testing.T) {
		cache := &tokenCache{}
		token := "test-token"
		role := "test-role"
		ttl := 10

		before := time.Now()
		cache.Set(token, role, ttl)
		after := time.Now()

		attr, err := cache.Get(token)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expectedMin := before.Add(time.Duration(ttl) * time.Second)
		expectedMax := after.Add(time.Duration(ttl) * time.Second)

		if attr.Expiration.Before(expectedMin) {
			t.Errorf("expiration %v is before expected minimum %v", attr.Expiration, expectedMin)
		}
		if attr.Expiration.After(expectedMax) {
			t.Errorf("expiration %v is after expected maximum %v", attr.Expiration, expectedMax)
		}
	})
}

func TestTokenCache_Get(t *testing.T) {
	t.Run("retrieves existing token", func(t *testing.T) {
		cache := &tokenCache{}
		token := "test-token"
		role := "test-role"
		ttl := 3600

		cache.Set(token, role, ttl)

		attr, err := cache.Get(token)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if attr == nil {
			t.Error("expected non-nil attributes")
		}
		if attr.Role != role {
			t.Errorf("expected role %s, got %s", role, attr.Role)
		}
		if attr.InitialTtl != ttl {
			t.Errorf("expected ttl %d, got %d", ttl, attr.InitialTtl)
		}
	})

	t.Run("returns error for non-existent token", func(t *testing.T) {
		cache := &tokenCache{}

		attr, err := cache.Get("non-existent-token")
		if err == nil {
			t.Error("expected error for non-existent token")
		}
		if attr != nil {
			t.Error("expected nil attributes for non-existent token")
		}
	})

	t.Run("handles concurrent access", func(t *testing.T) {
		cache := &tokenCache{}
		token := "test-token"
		role := "test-role"
		ttl := 3600

		cache.Set(token, role, ttl)

		done := make(chan bool)
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			go func() {
				attr, err := cache.Get(token)
				if err != nil {
					errors <- err
				} else if attr.Role != role {
					errors <- err
				}
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		close(errors)
		for err := range errors {
			if err != nil {
				t.Errorf("concurrent access error: %v", err)
			}
		}
	})
}

func TestTokenCache_GenerateToken(t *testing.T) {
	t.Run("generates token with correct length", func(t *testing.T) {
		cache := &tokenCache{}
		role := "test-role"
		ttl := 3600

		token := cache.generateToken(role, ttl)
		if len(token) != 64 {
			t.Errorf("expected token length 64, got %d", len(token))
		}
	})

	t.Run("stores token in cache", func(t *testing.T) {
		cache := &tokenCache{}
		role := "test-role"
		ttl := 3600

		token := cache.generateToken(role, ttl)

		attr, err := cache.Get(token)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if attr.Role != role {
			t.Errorf("expected role %s, got %s", role, attr.Role)
		}
		if attr.InitialTtl != ttl {
			t.Errorf("expected ttl %d, got %d", ttl, attr.InitialTtl)
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		cache := &tokenCache{}
		role := "test-role"
		ttl := 3600

		token1 := cache.generateToken(role, ttl)
		token2 := cache.generateToken(role, ttl)

		if token1 == token2 {
			t.Error("expected unique tokens but got duplicates")
		}
	})
}

func TestTokenCache_CheckToken(t *testing.T) {
	t.Run("validates existing valid token", func(t *testing.T) {
		cache := &tokenCache{}
		token := "test-token"
		role := "test-role"
		ttl := 3600

		cache.Set(token, role, ttl)

		valid, remainingTtl := cache.checkToken(token)
		if !valid {
			t.Error("expected token to be valid")
		}
		if remainingTtl < 3595 || remainingTtl > 3600 {
			t.Errorf("expected remaining TTL to be slightly less than 3600, got %d", remainingTtl)
		}
	})

	t.Run("rejects non-existent token", func(t *testing.T) {
		cache := &tokenCache{}

		valid, remainingTtl := cache.checkToken("non-existent-token")
		if valid {
			t.Error("expected token to be invalid")
		}
		if remainingTtl != 0 {
			t.Errorf("expected remainingTtl 0, got %d", remainingTtl)
		}
	})

	t.Run("rejects expired token", func(t *testing.T) {
		cache := &tokenCache{}
		token := "test-token"
		role := "test-role"
		ttl := -1 // Expired immediately

		cache.Set(token, role, ttl)
		time.Sleep(10 * time.Millisecond) // Ensure it's expired

		valid, remainingTtl := cache.checkToken(token)
		if valid {
			t.Error("expected token to be invalid")
		}
		if remainingTtl != 0 {
			t.Errorf("expected remainingTtl 0, got %d", remainingTtl)
		}
	})
}

func TestTokenCache_Clean(t *testing.T) {
	t.Run("removes expired tokens", func(t *testing.T) {
		cache := &tokenCache{}
		expiredToken := "expired-token"
		validToken := "valid-token"

		cache.Set(expiredToken, "role1", -1) // Already expired
		cache.Set(validToken, "role2", 3600) // Valid for an hour

		time.Sleep(10 * time.Millisecond) // Ensure expiration check works

		cache.clean()

		// Expired token should be removed
		_, err := cache.Get(expiredToken)
		if err == nil {
			t.Error("expected expired token to be removed")
		}

		// Valid token should still exist
		attr, err := cache.Get(validToken)
		if err != nil {
			t.Errorf("unexpected error for valid token: %v", err)
		}
		if attr == nil {
			t.Error("expected valid token to still exist")
		}
	})

	t.Run("removes tokens with invalid attributes", func(t *testing.T) {
		cache := &tokenCache{}
		token := "test-token"

		// Store invalid data
		cache.Store(token, "not-a-pointer")

		cache.clean()

		// Token should be removed
		_, err := cache.Get(token)
		if err == nil {
			t.Error("expected token with invalid attributes to be removed")
		}
	})

	t.Run("handles empty cache", func(t *testing.T) {
		cache := &tokenCache{}

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("clean panicked on empty cache: %v", r)
			}
		}()

		cache.clean()
	})
}

func TestTokenCache_Concurrency(t *testing.T) {
	t.Run("handles concurrent Set and Get operations", func(t *testing.T) {
		cache := &tokenCache{}
		done := make(chan bool)

		// Concurrent writes
		for i := 0; i < 10; i++ {
			go func(id int) {
				cache.Set("token-"+string(rune(id)), "role", 3600)
				done <- true
			}(i)
		}

		// Wait for writes
		for i := 0; i < 10; i++ {
			<-done
		}

		// Concurrent reads
		for i := 0; i < 10; i++ {
			go func(id int) {
				cache.Get("token-" + string(rune(id)))
				done <- true
			}(i)
		}

		// Wait for reads
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("handles concurrent clean operations", func(t *testing.T) {
		cache := &tokenCache{}
		done := make(chan bool)

		// Add some tokens
		for i := 0; i < 10; i++ {
			cache.Set("token-"+string(rune(i)), "role", 3600)
		}

		// Run clean concurrently
		for i := 0; i < 5; i++ {
			go func() {
				cache.clean()
				done <- true
			}()
		}

		// Wait for all cleans
		for i := 0; i < 5; i++ {
			<-done
		}
	})
}
