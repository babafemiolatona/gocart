package middleware

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestRateLimiterAllowsWithinLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed within the limit", i+1)
		}
	}

	if rl.Allow("1.2.3.4") {
		t.Fatal("request beyond the limit should be blocked")
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)

	if !rl.Allow("1.2.3.4") {
		t.Fatal("first request should be allowed")
	}
	if rl.Allow("1.2.3.4") {
		t.Fatal("second request should be blocked within the window")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("1.2.3.4") {
		t.Fatal("request after the window expires should be allowed")
	}
}

func TestRateLimiterKeysAreIndependent(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	if !rl.Allow("1.2.3.4") {
		t.Fatal("first IP should be allowed")
	}
	if rl.Allow("1.2.3.4") {
		t.Fatal("first IP beyond limit should be blocked")
	}

	if !rl.Allow("5.6.7.8") {
		t.Fatal("a different IP should be unaffected")
	}
}

func TestRateLimiterConcurrentSafety(t *testing.T) {
	limit := 10
	rl := NewRateLimiter(limit, time.Minute)

	const goroutines = 50
	var wg sync.WaitGroup
	var allowed int32

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow("1.2.3.4") {
				allowed++
			}
		}()
	}
	wg.Wait()

	if allowed != int32(limit) {
		t.Fatalf("want exactly %d allowed under concurrency, got %d", limit, allowed)
	}
}

func TestRateLimitMiddlewareBlocks(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	r, called := newTestRouter(RateLimitMiddleware(rl))

	w := doRequest(r, nil)
	assertStatus(t, w, http.StatusOK)
	if !*called {
		t.Fatal("handler should run when under the limit")
	}

	doRequest(r, nil)
	assertStatus(t, w, http.StatusOK)

	blocked := doRequest(r, nil)
	assertStatus(t, blocked, http.StatusTooManyRequests)

	code, message := decodeError(t, blocked)
	if code != "too_many_requests" {
		t.Errorf("expected too_many_requests, got %q", code)
	}
	if message == "" {
		t.Error("expected a non-empty error message")
	}
}
