// Package httpclient provides a shared HTTP client with sane defaults:
// timeouts, basic retry on 5xx/429, and pooled connections.
package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"
)

// Default returns an *http.Client configured for outbound calls to third-party APIs.
func Default() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
		},
	}
}

// DoWithRetry executes the request retrying on transient failures (5xx, 429, network errors).
// Up to maxAttempts (including the first try). Caller is responsible for closing the response body.
func DoWithRetry(ctx context.Context, c *http.Client, req *http.Request, maxAttempts int) (*http.Response, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Clone request: body must be readable on each attempt.
		clone := req.Clone(ctx)
		resp, err := c.Do(clone)
		if err == nil && !shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		// Drain and close body if present so we can reuse the connection.
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
		}

		if attempt == maxAttempts {
			break
		}

		// Exponential backoff with jitter (200ms, 400ms, 800ms ...).
		base := time.Duration(200*(1<<(attempt-1))) * time.Millisecond
		jitter := time.Duration(rand.Int64N(int64(base / 2)))
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(base + jitter):
		}
	}

	if lastErr == nil {
		lastErr = errors.New("retry: unknown error")
	}
	return nil, lastErr
}

func shouldRetry(status int) bool {
	return status == http.StatusTooManyRequests ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusBadGateway ||
		status == http.StatusGatewayTimeout
}
