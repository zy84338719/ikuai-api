package ikuaisdk

import (
	"testing"
	"time"
)

func TestResponseCache(t *testing.T) {
	cache := NewResponseCache(1 * time.Second)

	// Test Set and Get
	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Get() = %v, %v; want value1, true", val, ok)
	}

	// Test non-existent key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for non-existent key")
	}
}

func TestResponseCacheExpiration(t *testing.T) {
	cache := NewResponseCache(100 * time.Millisecond)

	cache.Set("expire_key", "value")

	// Should exist immediately
	_, ok := cache.Get("expire_key")
	if !ok {
		t.Error("Key should exist immediately after Set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, ok = cache.Get("expire_key")
	if ok {
		t.Error("Key should be expired")
	}
}

func TestResponseCacheDelete(t *testing.T) {
	cache := NewResponseCache(1 * time.Second)

	cache.Set("delete_key", "value")
	cache.Delete("delete_key")

	_, ok := cache.Get("delete_key")
	if ok {
		t.Error("Key should be deleted")
	}
}

func TestResponseCacheClear(t *testing.T) {
	cache := NewResponseCache(1 * time.Second)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Clear()

	_, ok1 := cache.Get("key1")
	_, ok2 := cache.Get("key2")
	if ok1 || ok2 {
		t.Error("All keys should be cleared")
	}
}

func TestClientCacheMethods(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")

	// Test cache disabled by default
	_, ok := client.GetCachedResponse("key")
	if ok {
		t.Error("Cache should be disabled by default")
	}

	// Enable cache
	client.EnableCache(1 * time.Second)
	client.SetCachedResponse("key", "value")
	val, ok := client.GetCachedResponse("key")
	if !ok || val != "value" {
		t.Errorf("GetCachedResponse() = %v, %v; want value, true", val, ok)
	}

	// Disable cache
	client.DisableCache()
	_, ok = client.GetCachedResponse("key")
	if ok {
		t.Error("Cache should be disabled after DisableCache()")
	}
}
