package ratelimit

import (
	"fmt"
	"time"

	"github.com/ntshibin/core/ghttp/internal/context"
)

// CombinedLimiter 组合多个限流器的限流器
type CombinedLimiter struct {
	limiters []Limiter // 限流器列表
}

// NewCombinedLimiter 创建组合限流器
func NewCombinedLimiter(limiters ...Limiter) *CombinedLimiter {
	return &CombinedLimiter{
		limiters: limiters,
	}
}

// Allow 判断是否允许请求通过 (所有限流器都允许才通过)
func (cl *CombinedLimiter) Allow(key string) bool {
	for _, limiter := range cl.limiters {
		if !limiter.Allow(key) {
			return false
		}
	}
	return true
}

// Reset 重置指定key的限流记录
func (cl *CombinedLimiter) Reset(key string) {
	for _, limiter := range cl.limiters {
		limiter.Reset(key)
	}
}

// IPAndURLLimiter IP和URL组合限流器
type IPAndURLLimiter struct {
	ipLimiter  Limiter // IP限流器
	urlLimiter Limiter // URL限流器
}

// NewIPAndURLLimiter 创建IP和URL组合限流器
func NewIPAndURLLimiter(ipLimiter, urlLimiter Limiter) *IPAndURLLimiter {
	return &IPAndURLLimiter{
		ipLimiter:  ipLimiter,
		urlLimiter: urlLimiter,
	}
}

// Allow 判断是否允许请求通过
func (iul *IPAndURLLimiter) Allow(key string) bool {
	// 拆分组合key
	ipKey, urlKey := splitIPAndURLKey(key)

	// IP和URL限流器都通过才允许
	return iul.ipLimiter.Allow(ipKey) && iul.urlLimiter.Allow(urlKey)
}

// Reset 重置指定key的限流记录
func (iul *IPAndURLLimiter) Reset(key string) {
	ipKey, urlKey := splitIPAndURLKey(key)
	iul.ipLimiter.Reset(ipKey)
	iul.urlLimiter.Reset(urlKey)
}

// NewIPAndURLRateLimiter 创建IP和URL组合的令牌桶限流器
func NewIPAndURLRateLimiter(ipRate, ipCapacity, urlRate, urlCapacity float64) *IPAndURLLimiter {
	return NewIPAndURLLimiter(
		NewTokenBucket(ipRate, ipCapacity),
		NewTokenBucket(urlRate, urlCapacity),
	)
}

// NewIPAndURLFixedWindowLimiter 创建IP和URL组合的固定窗口限流器
func NewIPAndURLFixedWindowLimiter(ipLimit int, ipWindow time.Duration, urlLimit int, urlWindow time.Duration) *IPAndURLLimiter {
	return NewIPAndURLLimiter(
		NewFixedWindow(ipLimit, ipWindow),
		NewFixedWindow(urlLimit, urlWindow),
	)
}

// NewIPAndURLSlidingWindowLimiter 创建IP和URL组合的滑动窗口限流器
func NewIPAndURLSlidingWindowLimiter(ipLimit int, ipWindow time.Duration, urlLimit int, urlWindow time.Duration) *IPAndURLLimiter {
	return NewIPAndURLLimiter(
		NewSlidingWindow(ipLimit, ipWindow),
		NewSlidingWindow(urlLimit, urlWindow),
	)
}

// GetIPAndURLKey 获取IP和URL组合的key
func GetIPAndURLKey(c *context.Context) string {
	return fmt.Sprintf("%s|%s", c.ClientIP(), c.Request.URL.Path)
}

// splitIPAndURLKey 拆分IP和URL组合的key
func splitIPAndURLKey(key string) (string, string) {
	// key格式: "ip|url"
	for i := 0; i < len(key); i++ {
		if key[i] == '|' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}
