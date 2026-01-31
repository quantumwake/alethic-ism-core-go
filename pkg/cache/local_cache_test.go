package cache

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestLocalCache_SetAndGet(t *testing.T) {
	ctx := t.Context()

	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()

	cache.Set(ctx, "key1", "value1", 1*time.Second)
	value, found := cache.Get(ctx, "key1")
	if !found {
		t.Fatal("Expected to find cached value")
	}

	if value != "value1" {
		t.Fatalf("Expected 'value1', got %v", value)
	}
}

func TestLocalCache_Expiration(t *testing.T) {
	ctx := t.Context()

	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()
	cache.Set(ctx, "key1", "value1", 100*time.Millisecond)

	value, found := cache.Get(ctx, "key1")
	if !found {
		t.Fatal("Expected to find cached value immediately after setting")
	}
	if value != "value1" {
		t.Fatalf("Expected 'value1', got %v", value)
	}

	time.Sleep(150 * time.Millisecond)

	_, found = cache.Get(ctx, "key1")
	if found {
		t.Fatal("Expected cache entry to be expired")
	}
}

func TestLocalCache_Delete(t *testing.T) {
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()

	ctx := context.Background()

	cache.Set(ctx, "key1", "value1", 1*time.Second)
	foundKeyValue, ok := cache.Get(ctx, "key1")
	require.True(t, ok)
	require.Equal(t, "value1", foundKeyValue)
	cache.Delete(ctx, "key1")
	foundKeyValue, ok = cache.Get(ctx, "key1")
	require.False(t, ok)
	require.Nil(t, foundKeyValue)

	_, found := cache.Get(ctx, "key1")
	if found {
		t.Fatal("Expected cache entry to be deleted")
	}
}

func TestLocalCache_Clear(t *testing.T) {
	ctx := t.Context()
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()

	cache.Set(ctx, "key1", "value1", 1*time.Second)
	cache.Set(ctx, "key2", "value2", 1*time.Second)
	cache.Clear(ctx)

	_, found1 := cache.Get(ctx, "key1")
	_, found2 := cache.Get(ctx, "key2")

	if found1 || found2 {
		t.Fatal("Expected all cache entries to be cleared")
	}
}

func TestLocalCache_DefaultTTL(t *testing.T) {
	config := &Config{
		DefaultTTL:              200 * time.Millisecond,
		CleanupDurationInterval: 500 * time.Millisecond,
	}
	cache := NewLocalCache(config)
	defer cache.Close()

	ctx := context.Background()

	for i := 1; i <= 10; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		ttl := time.Duration(i) * time.Second
		cache.Set(ctx, key, value, ttl)
	}

	for i := 1; i <= 10; i++ {
		expectedKey := fmt.Sprintf("key%d", i)
		expectedValue := fmt.Sprintf("value%d", i)

		value, ok := cache.Get(ctx, expectedKey)
		require.True(t, ok)
		require.Equal(t, expectedValue, value)

		time.Sleep(1*time.Second + 50*time.Millisecond) // first item should expire.
		expectedLen := 10 - i                           // after i-th iteration, i items should have expired
		log.Printf("Sleeping to allow items to expire... expectedKey: %s, expectedValue: %s, expectedLen: %d\n", expectedKey, expectedValue, expectedLen)
		require.Equal(t, expectedLen, cache.Len())
		value, ok = cache.Get(ctx, expectedKey)
		require.False(t, ok)
		require.Nil(t, value)
	}
}

func TestLocalCache_ConcurrentAccess(t *testing.T) {
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Close()

	ctx := context.Background()
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			key := "key"
			value := i
			cache.Set(ctx, key, value, 1*time.Second)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cache.Get(ctx, "key")
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	<-done
	<-done
}
