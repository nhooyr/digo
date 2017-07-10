package discgo

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type rateLimiter struct {
	sync.RWMutex
	pathRateLimiters map[string]*pathRateLimiter
	resetAfter       time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		pathRateLimiters: make(map[string]*pathRateLimiter),
	}
}

func (rl *rateLimiter) getPathRateLimiter(path string) *pathRateLimiter {
	rl.RLock()
	prl, ok := rl.pathRateLimiters[path]
	rl.RUnlock()
	if !ok {
		rl.Lock()
		prl, ok = rl.pathRateLimiters[path]
		if !ok {
			prl = &pathRateLimiter{
				rl: rl,
			}
			rl.pathRateLimiters[path] = prl
		}
		rl.Unlock()
	}
	return prl
}

func (rl *rateLimiter) setResetAfter(resetAfter time.Time) {
	rl.Lock()
	rl.resetAfter = resetAfter
	rl.Unlock()
}

// TODO allow concurrent requests to the same rate limited endpoint
type pathRateLimiter struct {
	mu         sync.Mutex
	remaining  int
	resetAfter time.Time
	rl         *rateLimiter
}

func (prl *pathRateLimiter) lock() {
	prl.mu.Lock()
	now := time.Now()
	if prl.remaining < 1 && prl.resetAfter.After(now) {
		time.Sleep(prl.resetAfter.Sub(now))
	}
	prl.remaining--

	prl.rl.RLock()
	now = time.Now()
	if prl.rl.resetAfter.After(now) {
		time.Sleep(prl.rl.resetAfter.Sub(now))
	}
	prl.rl.RUnlock()
}

func (prl *pathRateLimiter) unlock(h http.Header) (err error) {
	defer prl.mu.Unlock()

	if h == nil {
		return nil
	}

	globalHeader := h.Get("X-RateLimit-Global")
	resetHeader := h.Get("X-RateLimit-Reset")

	if globalHeader != "" {
		retryAfterHeader := h.Get("Retry-After")
		var parsedRetryAfter int64
		parsedRetryAfter, err = strconv.ParseInt(retryAfterHeader, 10, 64)
		if err != nil {
			return err
		}
		retryAfter := time.Now().Add(time.Duration(parsedRetryAfter) * time.Millisecond)
		prl.rl.setResetAfter(retryAfter)
	} else if resetHeader != "" {
		var parsedResetAfter int64
		parsedResetAfter, err = strconv.ParseInt(resetHeader, 10, 64)
		if err != nil {
			return err
		}
		prl.resetAfter = time.Unix(parsedResetAfter, 0)
	}

	remainingHeader := h.Get("X-RateLimit-Remaining")
	if remainingHeader != "" {
		prl.remaining, err = strconv.Atoi(remainingHeader)
	}

	return err
}
