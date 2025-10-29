package ratelimit

import (
	"sync"
	"time"
)

type RateLimiter struct {
	limits map[string]*userLimit
	mu     sync.RWMutex
}

type userLimit struct {
	count    int
	window   time.Time
	isProAcc bool
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*userLimit),
	}
}

func (r *RateLimiter) AllowLog(clientID string, isProAccount bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	limit, exists := r.limits[clientID]

	if !exists || now.Sub(limit.window) >= time.Minute {
		r.limits[clientID] = &userLimit{
			count:    1,
			window:   now,
			isProAcc: isProAccount,
		}
		return true
	}

	if !limit.isProAcc && limit.count >= 100 {
		return false
	}

	limit.count++
	return true
}
