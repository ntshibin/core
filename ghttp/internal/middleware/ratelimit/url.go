package ratelimit

import (
	"time"

	"github.com/ntshibin/core/ghttp/internal/context"
)

// URLLimiter 基于URL的限流器
type URLLimiter struct {
	limiter Limiter // 底层限流器
}

// NewURLLimiter 创建基于URL的限流器
func NewURLLimiter(limiter Limiter) *URLLimiter {
	return &URLLimiter{
		limiter: limiter,
	}
}

// Allow 判断是否允许请求通过
func (ul *URLLimiter) Allow(key string) bool {
	return ul.limiter.Allow(key)
}

// Reset 重置指定key的限流记录
func (ul *URLLimiter) Reset(key string) {
	ul.limiter.Reset(key)
}

// NewURLRateLimiter 创建基于URL的令牌桶限流器
// rate: 令牌产生速率 (每秒)
// capacity: 桶容量
func NewURLRateLimiter(rate, capacity float64) *URLLimiter {
	return NewURLLimiter(NewTokenBucket(rate, capacity))
}

// NewURLFixedWindowLimiter 创建基于URL的固定窗口限流器
// limit: 时间窗口内的请求上限
// window: 时间窗口大小
func NewURLFixedWindowLimiter(limit int, window time.Duration) *URLLimiter {
	return NewURLLimiter(NewFixedWindow(limit, window))
}

// NewURLSlidingWindowLimiter 创建基于URL的滑动窗口限流器
// limit: 时间窗口内的请求上限
// window: 时间窗口大小
func NewURLSlidingWindowLimiter(limit int, window time.Duration) *URLLimiter {
	return NewURLLimiter(NewSlidingWindow(limit, window))
}

// GetRequestPath 获取请求路径的处理函数
func GetRequestPath(c *context.Context) string {
	return c.Request.URL.Path
}
