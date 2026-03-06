package utils

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"time"
)

var (
	macRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	ipRegex  = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}$`)
)

func IsValidMAC(mac string) bool {
	return macRegex.MatchString(mac)
}

func IsValidIPv4(ip string) bool {
	return ipRegex.MatchString(ip)
}

func IsValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

func IsValidPort(port int) bool {
	return port >= 0 && port <= 65535
}

func Retry(ctx context.Context, maxAttempts int, delay time.Duration, fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := fn(); err != nil {
			lastErr = err
			if attempt < maxAttempts {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
					continue
				}
			}
			continue
		}

		return nil
	}

	return fmt.Errorf("retry failed after %d attempts: %w", maxAttempts, lastErr)
}

func RetryWithBackoff(ctx context.Context, maxAttempts int, initialDelay time.Duration, fn func() error) error {
	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := fn(); err != nil {
			lastErr = err
			if attempt < maxAttempts {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
					delay = delay * 2
					if delay > 30*time.Second {
						delay = 30 * time.Second
					}
					continue
				}
			}
			continue
		}

		return nil
	}

	return fmt.Errorf("retry with backoff failed after %d attempts: %w", maxAttempts, lastErr)
}

type RateLimiter struct {
	ticker   *time.Ticker
	stopChan chan struct{}
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	interval := time.Second / time.Duration(requestsPerSecond)
	return &RateLimiter{
		ticker:   time.NewTicker(interval),
		stopChan: make(chan struct{}),
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.ticker.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.stopChan:
		return fmt.Errorf("rate limiter stopped")
	}
}

func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
	close(rl.stopChan)
}

func ValidateIPRange(ip string) error {
	if !IsValidIPv4(ip) {
		return fmt.Errorf("invalid IPv4 address: %s", ip)
	}
	return nil
}

func ValidateMACAddress(mac string) error {
	if !IsValidMAC(mac) {
		return fmt.Errorf("invalid MAC address: %s", mac)
	}
	return nil
}

func ValidatePort(port int) error {
	if !IsValidPort(port) {
		return fmt.Errorf("invalid port number: %d (must be 0-65535)", port)
	}
	return nil
}
