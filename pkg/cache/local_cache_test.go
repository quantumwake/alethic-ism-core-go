package cache

import (
	"context"
	"testing"
	"time"
)

func TestLocalCache_SetAndGet(t *testing.T) {
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	ctx := context.Background()
	
	err := cache.Set(ctx, "key1", "value1", 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	
	value, found := cache.Get(ctx, "key1")
	if !found {
		t.Fatal("Expected to find cached value")
	}
	
	if value != "value1" {
		t.Fatalf("Expected 'value1', got %v", value)
	}
}

func TestLocalCache_Expiration(t *testing.T) {
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	ctx := context.Background()
	
	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	
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
	defer cache.Stop()
	
	ctx := context.Background()
	
	err := cache.Set(ctx, "key1", "value1", 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	
	err = cache.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("Failed to delete cache entry: %v", err)
	}
	
	_, found := cache.Get(ctx, "key1")
	if found {
		t.Fatal("Expected cache entry to be deleted")
	}
}

func TestLocalCache_Clear(t *testing.T) {
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	ctx := context.Background()
	
	err := cache.Set(ctx, "key1", "value1", 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	
	err = cache.Set(ctx, "key2", "value2", 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	
	err = cache.Clear(ctx)
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}
	
	_, found1 := cache.Get(ctx, "key1")
	_, found2 := cache.Get(ctx, "key2")
	
	if found1 || found2 {
		t.Fatal("Expected all cache entries to be cleared")
	}
}

func TestLocalCache_DefaultTTL(t *testing.T) {
	config := &CacheConfig{
		DefaultTTL: 200 * time.Millisecond,
	}
	cache := NewLocalCache(config)
	defer cache.Stop()
	
	ctx := context.Background()
	
	err := cache.Set(ctx, "key1", "value1", 0)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	
	_, found := cache.Get(ctx, "key1")
	if !found {
		t.Fatal("Expected to find cached value")
	}
	
	time.Sleep(250 * time.Millisecond)
	
	_, found = cache.Get(ctx, "key1")
	if found {
		t.Fatal("Expected cache entry to be expired using default TTL")
	}
}

func TestLocalCache_ConcurrentAccess(t *testing.T) {
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	ctx := context.Background()
	done := make(chan bool)
	
	go func() {
		for i := 0; i < 100; i++ {
			key := "key"
			value := i
			_ = cache.Set(ctx, key, value, 1*time.Second)
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