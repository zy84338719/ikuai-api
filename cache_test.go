package ikuaisdk

import (
	"testing"
	"time"
)

func TestResponseCache_GetSet(t *testing.T) {
	cache := NewResponseCache(1 * time.Minute)

	key := "test_key"
	data := map[string]string{"test": "data"}

	cache.Set(key, data)

	retrieved, exists := cache.Get(key)
	if !exists {
		t.Error("Expected item to exist in cache")
	}

	if retrieved == nil {
		t.Error("Expected non-nil data")
	}
}

func TestResponseCache_GetExpired(t *testing.T) {
	cache := NewResponseCache(100 * time.Millisecond)

	key := "test_key"
	data := "test_data"

	cache.Set(key, data)

	time.Sleep(150 * time.Millisecond)

	_, exists := cache.Get(key)
	if exists {
		t.Error("Expected expired item to not exist")
	}
}

func TestResponseCache_Clear(t *testing.T) {
	cache := NewResponseCache(1 * time.Minute)

	cache.Set("key1", "data1")
	cache.Set("key2", "data2")
	cache.Clear()

	if len(cache.items) != 0 {
		t.Errorf("Expected cache to be cleared, got %d items", len(cache.items))
	}
}

func TestClient_ResponseCache(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")

	if client.cache != nil {
		t.Error("Expected client cache to be nil by default")
	}

	client.EnableCache(5 * time.Minute)

	if client.cache == nil {
		t.Error("Expected client cache to be initialized")
	}

	key := "test_key"
	data := map[string]string{"test": "data"}

	client.SetCachedResponse(key, data)

	retrieved, exists := client.GetCachedResponse(key)
	if !exists {
		t.Error("Expected cached response to exist")
	}

	if retrieved == nil {
		t.Error("Expected non-nil cached response")
	}

	client.ClearResponseCache()

	_, exists = client.GetCachedResponse(key)
	if exists {
		t.Error("Expected cache to be cleared")
	}

	client.DisableCache()

	if client.cache != nil {
		t.Error("Expected cache to be disabled")
	}
}
