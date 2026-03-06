package ikuaisdk

import (
	"context"
	"sync"
	"time"
)

type BatchRequest struct {
	FuncName string
	Action   string
	Param    interface{}
}

type BatchResult struct {
	FuncName string
	Data     interface{}
	Error    error
	Duration time.Duration
}

func (c *Client) BatchCall(ctx context.Context, requests []BatchRequest) []BatchResult {
	results := make([]BatchResult, len(requests))
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(idx int, request BatchRequest) {
			defer wg.Done()

			start := time.Now()
			var result interface{}

			err := c.Call(ctx, request.FuncName, request.Action, request.Param, &result)
			duration := time.Since(start)

			results[idx] = BatchResult{
				FuncName: request.FuncName,
				Data:     result,
				Error:    err,
				Duration: duration,
			}
		}(i, req)
	}

	wg.Wait()
	return results
}

func (c *Client) ParallelCall(ctx context.Context, funcs []string, action string) map[string]interface{} {
	var mu sync.Mutex
	results := make(map[string]interface{})
	var wg sync.WaitGroup

	for _, fn := range funcs {
		wg.Add(1)
		go func(funcName string) {
			defer wg.Done()

			var result interface{}
			err := c.Call(ctx, funcName, action, nil, &result)

			mu.Lock()
			if err != nil {
				results[funcName] = map[string]interface{}{
					"error": err.Error(),
				}
			} else {
				results[funcName] = result
			}
			mu.Unlock()
		}(fn)
	}

	wg.Wait()
	return results
}
