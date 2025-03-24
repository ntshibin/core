package middleware

import (
	"net/http"

	"github.com/ntshibin/core/ghttp/internal/context"
	"github.com/ntshibin/core/ghttp/internal/middleware/ratelimit"
)

// RateLimiter 限流接口
type RateLimiter interface {
	Allow(key string) bool
	Reset(key string)
}

// RateLimit 限流中间件，限制请求频率
func RateLimit(limiter RateLimiter, keyFunc func(*context.Context) string) HandlerFunc {
	return func(c *context.Context) {
		key := keyFunc(c)
		if !limiter.Allow(key) {
			// 限流
			c.Fail(http.StatusTooManyRequests, "请求过于频繁，请稍后重试")
			c.Abort()
			return
		}

		// 未限流，继续处理请求
		c.Next()
	}
}

// DefaultRateLimiter 创建默认的令牌桶限流器 (每秒1个请求，最多5个请求的突发流量)
func DefaultRateLimiter() RateLimiter {
	return ratelimit.NewTokenBucket(1, 5)
}

// IPRateLimiter 创建基于IP的令牌桶限流器 (每秒1个请求，最多5个请求的突发流量)
func IPRateLimiter() RateLimiter {
	return ratelimit.NewIPRateLimiter(1, 5)
}

// URLRateLimiter 创建基于URL的令牌桶限流器 (每秒10个请求，最多20个请求的突发流量)
func URLRateLimiter() RateLimiter {
	return ratelimit.NewURLRateLimiter(10, 20)
}

// IPAndURLRateLimiter 创建IP和URL组合限流器
func IPAndURLRateLimiter() RateLimiter {
	return ratelimit.NewIPAndURLRateLimiter(1, 5, 10, 20)
}

// GetClientIPKey 获取客户端IP地址的key
func GetClientIPKey(c *context.Context) string {
	return ratelimit.GetClientIP(c)
}

// GetRequestPathKey 获取请求路径的key
func GetRequestPathKey(c *context.Context) string {
	return ratelimit.GetRequestPath(c)
}

// GetIPAndURLKey 获取IP和URL组合的key
func GetIPAndURLKey(c *context.Context) string {
	return ratelimit.GetIPAndURLKey(c)
}
