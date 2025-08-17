package cache

import (
	"context"
	"testing"
	"time"
)

func TestInvalidateMethodPrefix(t *testing.T) {
	// Create a local cache and cached backend
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()
	
	backend := &testPrefixBackend{}
	cb := NewCachedBackend(backend, cache, 1*time.Minute)
	
	ctx := context.Background()
	
	// Test scenario: cache multiple calls with same prefix but different second args
	stateID := "state-123"
	
	// Cache key 1: FindStateFull("state-123", "flag1")
	key1, err := cb.BuildCacheKey("FindStateFull", stateID, "flag1")
	if err != nil {
		t.Fatal(err)
	}
	err = cache.Set(ctx, key1, "result1", 1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	
	// Cache key 2: FindStateFull("state-123", "flag2")
	key2, err := cb.BuildCacheKey("FindStateFull", stateID, "flag2")
	if err != nil {
		t.Fatal(err)
	}
	err = cache.Set(ctx, key2, "result2", 1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	
	// Cache key 3: FindStateFull("state-456", "flag1") - different state
	key3, err := cb.BuildCacheKey("FindStateFull", "state-456", "flag1")
	if err != nil {
		t.Fatal(err)
	}
	err = cache.Set(ctx, key3, "result3", 1*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify all keys exist
	if _, found := cache.Get(ctx, key1); !found {
		t.Error("key1 should exist before invalidation")
	}
	if _, found := cache.Get(ctx, key2); !found {
		t.Error("key2 should exist before invalidation")
	}
	if _, found := cache.Get(ctx, key3); !found {
		t.Error("key3 should exist before invalidation")
	}
	
	// Invalidate all entries for FindStateFull with stateID "state-123"
	err = cb.InvalidateMethodPrefix(ctx, "FindStateFull", stateID)
	if err != nil {
		t.Fatal(err)
	}
	
	// Check that state-123 entries are gone
	if _, found := cache.Get(ctx, key1); found {
		t.Error("key1 should be invalidated")
	}
	if _, found := cache.Get(ctx, key2); found {
		t.Error("key2 should be invalidated")
	}
	
	// Check that state-456 entry still exists
	if _, found := cache.Get(ctx, key3); !found {
		t.Error("key3 should still exist (different state)")
	}
}

// testPrefixBackend is a dummy backend for testing prefix invalidation
type testPrefixBackend struct{}