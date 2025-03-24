package ratelimit

import (
	"time"

	"github.com/ntshibin/core/ghttp/internal/context"
)

// IPLimiter 基于IP的限流器
type IPLimiter struct {
	limiter Limiter // 底层限流器
}

// NewIPLimiter 创建基于IP的限流器
func NewIPLimiter(limiter Limiter) *IPLimiter {
	return &IPLimiter{
		limiter: limiter,
	}
}

// Allow 判断是否允许请求通过
func (ipl *IPLimiter) Allow(key string) bool {
	return ipl.limiter.Allow(key)
}

// Reset 重置指定key的限流记录
func (ipl *IPLimiter) Reset(key string) {
	ipl.limiter.Reset(key)
}

// NewIPRateLimiter 创建基于IP的令牌桶限流器
// rate: 令牌产生速率 (每秒)
// capacity: 桶容量
func NewIPRateLimiter(rate, capacity float64) *IPLimiter {
	return NewIPLimiter(NewTokenBucket(rate, capacity))
}

// NewIPFixedWindowLimiter 创建基于IP的固定窗口限流器
// limit: 时间窗口内的请求上限
// window: 时间窗口大小
func NewIPFixedWindowLimiter(limit int, window time.Duration) *IPLimiter {
	return NewIPLimiter(NewFixedWindow(limit, window))
}

// NewIPSlidingWindowLimiter 创建基于IP的滑动窗口限流器
// limit: 时间窗口内的请求上限
// window: 时间窗口大小
func NewIPSlidingWindowLimiter(limit int, window time.Duration) *IPLimiter {
	return NewIPLimiter(NewSlidingWindow(limit, window))
}

// GetClientIP 获取客户端IP地址的处理函数
func GetClientIP(c *context.Context) string {
	return c.ClientIP()
}
