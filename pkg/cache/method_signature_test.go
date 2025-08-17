package cache

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// Test backend with various methods to test signature caching
type testBackend struct {
	callCount  int
	lastUserID string
}

func (t *testBackend) GetUser(userID string) (string, error) {
	t.callCount++
	t.lastUserID = userID
	return "user-" + userID, nil
}

func (t *testBackend) GetUserWithAge(userID string, age int) (string, error) {
	t.callCount++
	return "user-" + userID + "-age", nil
}

func (t *testBackend) GetAllUsers() ([]string, error) {
	t.callCount++
	return []string{"user1", "user2", "user3"}, nil
}

func (t *testBackend) UpdateUser(userID string, name string) error {
	t.callCount++
	return nil
}

func TestMethodSignatureCache(t *testing.T) {
	backend := &testBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	cb := NewCachedBackend(backend, cache, 1*time.Second)
	
	// Test method registration with a simple function
	simpleFunc := func(id string) (string, error) {
		return "result-" + id, nil
	}
	err := cb.RegisterMethod("SimpleFunc", simpleFunc, nil)
	if err != nil {
		t.Fatalf("Failed to register method: %v", err)
	}
	
	// Verify method signature was cached
	sig := cb.GetMethodSignature("SimpleFunc")
	if sig == nil {
		t.Fatal("Method signature not cached")
	}
	
	if sig.NumIn != 1 { // 1 parameter
		t.Fatalf("Expected 1 input, got %d", sig.NumIn)
	}
	
	if sig.NumOut != 2 { // string + error
		t.Fatalf("Expected 2 outputs, got %d", sig.NumOut)
	}
}

func TestAutoRegisterMethods(t *testing.T) {
	backend := &testBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	cb := NewCachedBackend(backend, cache, 1*time.Second)
	
	// Auto-registration should have happened in NewCachedBackend
	// Check if methods were registered
	methods := []string{"GetUser", "GetUserWithAge", "GetAllUsers", "UpdateUser"}
	
	for _, methodName := range methods {
		sig := cb.GetMethodSignature(methodName)
		if sig == nil {
			t.Errorf("Method %s was not auto-registered", methodName)
		}
	}
}

func TestExecuteWithCachedSignature(t *testing.T) {
	backend := &testBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	cb := NewCachedBackend(backend, cache, 1*time.Second)
	ctx := context.Background()
	
	// First execution - should register and cache method signature
	result1, err := cb.Execute(ctx, "GetUser", []interface{}{"user123"}, 
		func() (string, error) {
			return backend.GetUser("user123")
		})
	
	if err != nil {
		t.Fatalf("Failed to execute method: %v", err)
	}
	
	if result1 != "user-user123" {
		t.Fatalf("Expected 'user-user123', got %v", result1)
	}
	
	initialCallCount := backend.callCount
	
	// Second execution - should use cached signature and cached result
	result2, err := cb.Execute(ctx, "GetUser", []interface{}{"user123"},
		func() (string, error) {
			return backend.GetUser("user123")
		})
	
	if err != nil {
		t.Fatalf("Failed to execute method: %v", err)
	}
	
	// Should get cached result without calling backend
	if backend.callCount != initialCallCount {
		t.Fatal("Expected to use cached result, but backend was called")
	}
	
	if result1 != result2 {
		t.Fatal("Expected same result from cache")
	}
	
	// Verify method signature is still cached
	sig := cb.GetMethodSignature("GetUser")
	if sig == nil {
		t.Fatal("Method signature should still be cached")
	}
}

func TestMethodConfig(t *testing.T) {
	backend := &testBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	cb := NewCachedBackend(backend, cache, 1*time.Second)
	ctx := context.Background()
	
	// Configure UpdateUser method to not be cacheable
	cb.SetMethodConfig("UpdateUser", &MethodConfig{
		Cacheable: false,
	})
	
	// Configure GetAllUsers with custom TTL
	cb.SetMethodConfig("GetAllUsers", &MethodConfig{
		TTL:       100 * time.Millisecond,
		Cacheable: true,
	})
	
	// Test non-cacheable method
	_, err := cb.Execute(ctx, "UpdateUser", []interface{}{"user1", "newname"},
		func() error {
			return backend.UpdateUser("user1", "newname")
		})
	
	if err != nil {
		t.Fatalf("Failed to execute UpdateUser: %v", err)
	}
	
	initialCallCount := backend.callCount
	
	// Execute again - should call backend again since it's not cacheable
	_, err = cb.Execute(ctx, "UpdateUser", []interface{}{"user1", "newname"},
		func() error {
			return backend.UpdateUser("user1", "newname")
		})
	
	if err != nil {
		t.Fatalf("Failed to execute UpdateUser: %v", err)
	}
	
	if backend.callCount == initialCallCount {
		t.Fatal("Expected UpdateUser to be called again (not cached)")
	}
	
	// Test custom TTL method
	backend.callCount = 0
	result, err := cb.Execute(ctx, "GetAllUsers", []interface{}{},
		func() ([]string, error) {
			return backend.GetAllUsers()
		})
	
	if err != nil {
		t.Fatalf("Failed to execute GetAllUsers: %v", err)
	}
	
	users := result.([]string)
	if len(users) != 3 {
		t.Fatalf("Expected 3 users, got %d", len(users))
	}
	
	// Should be cached
	_, _ = cb.Execute(ctx, "GetAllUsers", []interface{}{},
		func() ([]string, error) {
			return backend.GetAllUsers()
		})
	
	if backend.callCount != 1 {
		t.Fatal("Expected result to be cached")
	}
	
	// Wait for custom TTL to expire
	time.Sleep(150 * time.Millisecond)
	
	// Should call backend again after TTL expiry
	_, _ = cb.Execute(ctx, "GetAllUsers", []interface{}{},
		func() ([]string, error) {
			return backend.GetAllUsers()
		})
	
	if backend.callCount != 2 {
		t.Fatal("Expected backend to be called after TTL expiry")
	}
}

func TestExecuteWithArgs(t *testing.T) {
	backend := &testBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	cb := NewCachedBackend(backend, cache, 1*time.Second)
	ctx := context.Background()
	
	// Register method
	err := cb.RegisterMethod("GetUserWithAge", backend.GetUserWithAge, nil)
	if err != nil {
		t.Fatalf("Failed to register method: %v", err)
	}
	
	// Create reflect.Value arguments
	funcArgs := []reflect.Value{
		reflect.ValueOf("user456"),
		reflect.ValueOf(25),
	}
	
	// Execute with arguments
	result, err := cb.ExecuteWithArgs(ctx, "GetUserWithAge", 
		[]interface{}{"user456", 25}, // cache key args
		funcArgs,                       // function args
		backend.GetUserWithAge)
	
	if err != nil {
		t.Fatalf("Failed to execute with args: %v", err)
	}
	
	if result != "user-user456-age" {
		t.Fatalf("Expected 'user-user456-age', got %v", result)
	}
	
	initialCallCount := backend.callCount
	
	// Execute again - should use cache
	result2, err := cb.ExecuteWithArgs(ctx, "GetUserWithAge",
		[]interface{}{"user456", 25},
		funcArgs,
		backend.GetUserWithAge)
	
	if err != nil {
		t.Fatalf("Failed to execute with args: %v", err)
	}
	
	if backend.callCount != initialCallCount {
		t.Fatal("Expected to use cached result")
	}
	
	if result != result2 {
		t.Fatal("Expected same result from cache")
	}
}

func TestRegisterMethods(t *testing.T) {
	backend := &testBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	cb := NewCachedBackend(backend, cache, 1*time.Second)
	
	// Register multiple methods at once
	methods := map[string]interface{}{
		"Method1": func() string { return "result1" },
		"Method2": func() string { return "result2" },
		"Method3": func() string { return "result3" },
	}
	
	err := cb.RegisterMethods(methods)
	if err != nil {
		t.Fatalf("Failed to register methods: %v", err)
	}
	
	// Verify all methods were registered
	for name := range methods {
		sig := cb.GetMethodSignature(name)
		if sig == nil {
			t.Errorf("Method %s was not registered", name)
		}
	}
}

func TestMethodSignaturePerformance(t *testing.T) {
	backend := &testBackend{}
	cache := NewLocalCache(NewDefaultConfig())
	defer cache.Stop()
	
	cb := NewCachedBackend(backend, cache, 1*time.Second)
	ctx := context.Background()
	
	// Measure time for first execution (with reflection)
	start := time.Now()
	for i := 0; i < 100; i++ {
		cb.Execute(ctx, "PerfTest", []interface{}{i},
			func() string {
				return "result"
			})
	}
	firstDuration := time.Since(start)
	
	// Clear cache but keep method signatures
	cache.Clear(ctx)
	
	// Measure time for subsequent executions (using cached signatures)
	start = time.Now()
	for i := 0; i < 100; i++ {
		cb.Execute(ctx, "PerfTest", []interface{}{i + 100},
			func() string {
				return "result"
			})
	}
	secondDuration := time.Since(start)
	
	// Second run should be faster or similar (cached signatures)
	// This is a rough test - actual performance depends on many factors
	t.Logf("First run (with initial reflection): %v", firstDuration)
	t.Logf("Second run (cached signatures): %v", secondDuration)
	
	// Verify signature is still cached
	sig := cb.GetMethodSignature("PerfTest")
	if sig == nil {
		t.Fatal("Method signature should be cached after performance test")
	}
}