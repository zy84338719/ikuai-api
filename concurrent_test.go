package ikuaisdk

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestBatchCall_ConcurrentExecution(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")

	requests := []BatchRequest{
		{FuncName: "api1", Action: "show"},
		{FuncName: "api2", Action: "show"},
		{FuncName: "api3", Action: "show"},
	}

	start := time.Now()
	results := client.BatchCall(context.Background(), requests)
	duration := time.Since(start)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Duration == 0 {
			t.Error("Expected duration to be recorded")
		}
	}

	t.Logf("BatchCall completed in %v", duration)
}

func TestParallelCall_ConcurrentExecution(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")

	funcs := []string{"api1", "api2", "api3", "api4", "api5"}

	start := time.Now()
	results := client.ParallelCall(context.Background(), funcs, "show")
	duration := time.Since(start)

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	for _, fn := range funcs {
		if _, exists := results[fn]; !exists {
			t.Errorf("Expected result for %s", fn)
		}
	}

	t.Logf("ParallelCall completed in %v", duration)
}

func TestBatchCall_ErrorHandling(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")

	requests := []BatchRequest{
		{FuncName: "valid_api", Action: "show"},
		{FuncName: "invalid_api", Action: "show"},
	}

	results := client.BatchCall(context.Background(), requests)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error != nil {
			t.Logf("API %s returned error (expected): %v", result.FuncName, result.Error)
		}
	}
}

func TestConcurrentCalls_ThreadSafety(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")

	var wg sync.WaitGroup
	numGoroutines := 10
	results := make([][]BatchResult, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			requests := []BatchRequest{
				{FuncName: "api1", Action: "show"},
				{FuncName: "api2", Action: "show"},
			}

			results[id] = client.BatchCall(context.Background(), requests)
		}(i)
	}

	wg.Wait()

	for i, res := range results {
		if len(res) != 2 {
			t.Errorf("Goroutine %d: Expected 2 results, got %d", i, len(res))
		}
	}
}

func TestParallelCall_ContextCancellation(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	funcs := []string{"api1", "api2", "api3"}
	results := client.ParallelCall(ctx, funcs, "show")

	for _, fn := range funcs {
		if result, exists := results[fn]; exists {
			if errMap, ok := result.(map[string]interface{}); ok {
				if errMsg, hasError := errMap["error"]; hasError {
					t.Logf("API %s returned error (expected with cancelled context): %v", fn, errMsg)
				}
			}
		}
	}
}
