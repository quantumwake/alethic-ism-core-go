package cache

import (
	"testing"
	"time"
)

type mockBackend struct {
	callCount int
}

func (m *mockBackend) GetData(id string) (string, error) {
	m.callCount++
	return "data-" + id, nil
}

func (m *mockBackend) GetList() ([]string, error) {
	m.callCount++
	return []string{"item1", "item2", "item3"}, nil
}

func TestCachedBackend_BuildCacheKey(t *testing.T) {
	backend := &mockBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()

	cb := NewCachedBackend(backend, cache, 1*time.Second)

	key1, err := cb.BuildCacheKey("GetData", "id1")
	if err != nil {
		t.Fatalf("Failed to build cache key: %v", err)
	}

	key2, err := cb.BuildCacheKey("GetData", "id2")
	if err != nil {
		t.Fatalf("Failed to build cache key: %v", err)
	}

	if key1 == key2 {
		t.Fatal("Expected different keys for different arguments")
	}

	key3, err := cb.BuildCacheKey("GetData", "id1")
	if err != nil {
		t.Fatalf("Failed to build cache key: %v", err)
	}

	if key1 != key3 {
		t.Fatal("Expected same key for same arguments")
	}
}

func TestCachedBackend_GetCached(t *testing.T) {
	ctx := t.Context()
	backend := &mockBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()

	cb := NewCachedBackend(backend, cache, 1*time.Second)

	fetchCount := 0
	fetchFunc := func() (any, error) {
		fetchCount++
		return backend.GetData("test")
	}

	result1, err := cb.GetCached(ctx, "test-key", fetchFunc)
	if err != nil {
		t.Fatalf("Failed to get cached value: %v", err)
	}

	if fetchCount != 1 {
		t.Fatalf("Expected fetch function to be called once, got %d", fetchCount)
	}

	result2, err := cb.GetCached(ctx, "test-key", fetchFunc)
	if err != nil {
		t.Fatalf("Failed to get cached value: %v", err)
	}

	if fetchCount != 1 {
		t.Fatalf("Expected fetch function to still be called once (cached), got %d", fetchCount)
	}

	if result1 != result2 {
		t.Fatal("Expected same result from cache")
	}
}

func TestCachedBackend_InvalidateCache(t *testing.T) {
	ctx := t.Context()

	backend := &mockBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()

	cb := NewCachedBackend(backend, cache, 1*time.Second)

	cache.Set(ctx, "key1", "value1", 1*time.Second)
	cache.Set(ctx, "key2", "value2", 1*time.Second)

	err := cb.InvalidateCache(ctx, "key1")
	if err != nil {
		t.Fatalf("Failed to invalidate cache: %v", err)
	}

	_, found1 := cache.Get(ctx, "key1")
	if found1 {
		t.Fatal("Expected key1 to be invalidated")
	}

	_, found2 := cache.Get(ctx, "key2")
	if !found2 {
		t.Fatal("Expected key2 to still be in cache")
	}

	err = cb.InvalidateCache(ctx, "*")
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	_, found2 = cache.Get(ctx, "key2")
	if found2 {
		t.Fatal("Expected all keys to be cleared")
	}
}

func TestCallCached(t *testing.T) {
	ctx := t.Context()
	backend := &mockBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()

	cb := NewCachedBackend(backend, cache, 1*time.Second)

	result1, err := CallCached(cb, ctx, "GetData", []any{"id1"},
		func() (string, error) {
			return backend.GetData("id1")
		})

	if err != nil {
		t.Fatalf("Failed to call cached function: %v", err)
	}

	if result1 != "data-id1" {
		t.Fatalf("Expected 'data-id1', got %s", result1)
	}

	initialCallCount := backend.callCount

	result2, err := CallCached(cb, ctx, "GetData", []any{"id1"},
		func() (string, error) {
			return backend.GetData("id1")
		})

	if err != nil {
		t.Fatalf("Failed to call cached function: %v", err)
	}

	if backend.callCount != initialCallCount {
		t.Fatal("Expected backend not to be called again (should use cache)")
	}

	if result1 != result2 {
		t.Fatal("Expected same result from cache")
	}
}

func TestCallCachedWithTTL(t *testing.T) {
	ctx := t.Context()
	backend := &mockBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()

	cb := NewCachedBackend(backend, cache, 1*time.Second)

	result1, err := CallCachedWithTTL(cb, ctx, "GetList", []any{}, 100*time.Millisecond,
		func() ([]string, error) {
			return backend.GetList()
		})

	if err != nil {
		t.Fatalf("Failed to call cached function: %v", err)
	}

	if len(result1) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(result1))
	}

	initialCallCount := backend.callCount

	_, err = CallCachedWithTTL(cb, ctx, "GetList", []any{}, 100*time.Millisecond,
		func() ([]string, error) {
			return backend.GetList()
		})

	if err != nil {
		t.Fatalf("Failed to call cached function: %v", err)
	}

	if backend.callCount != initialCallCount {
		t.Fatal("Expected backend not to be called again (should use cache)")
	}

	time.Sleep(150 * time.Millisecond)

	_, err = CallCachedWithTTL(cb, ctx, "GetList", []any{}, 100*time.Millisecond,
		func() ([]string, error) {
			return backend.GetList()
		})

	if err != nil {
		t.Fatalf("Failed to call cached function: %v", err)
	}

	if backend.callCount == initialCallCount {
		t.Fatal("Expected backend to be called again after TTL expiration")
	}
}
