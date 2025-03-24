// Package ghttp 提供了对 Gin 框架的封装，简化 HTTP 服务的开发和维护
// 集成了 gerror 进行错误处理、glog 进行日志记录和 gconf 进行配置管理
package ghttp

import (
	"time"

	"github.com/ntshibin/core/ghttp/internal/config"
	ctx "github.com/ntshibin/core/ghttp/internal/context"
	ht "github.com/ntshibin/core/ghttp/internal/health"
	mid "github.com/ntshibin/core/ghttp/internal/middleware"
	"github.com/ntshibin/core/ghttp/internal/middleware/ratelimit"
	"github.com/ntshibin/core/ghttp/internal/router"
	"github.com/ntshibin/core/ghttp/internal/server"
)

// HTTPConfig HTTP服务配置
type HTTPConfig = config.HTTPConfig

// Context 上下文
type Context = ctx.Context

// Result API响应结果
type Result = ctx.Result

// Router 路由器
type Router = router.Router

// RouterGroup 路由组
type RouterGroup = router.RouterGroup

// HandlerFunc 处理函数
type HandlerFunc = router.HandlerFunc

// Server 服务器
type Server = server.Server

// CORSConfig 跨域配置
type CORSConfig = mid.CORSConfig

// RateLimiter 限流接口
type RateLimiter = mid.RateLimiter

// HealthCheck 健康检查组件
type HealthCheck = ht.HealthCheck

// HealthStatus 健康状态
type HealthStatus = ht.Status

// HealthCheckFunc 健康检查函数
type HealthCheckFunc = ht.CheckFunc

// 健康状态常量
const (
	StatusUp       = ht.StatusUp
	StatusDown     = ht.StatusDown
	StatusDegraded = ht.StatusDegraded
)

// New 创建一个新的服务器
func New(config HTTPConfig) *Server {
	return server.New(config)
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (HTTPConfig, error) {
	return config.LoadConfig(path)
}

// MustLoadConfig 从文件加载配置，失败时panic
func MustLoadConfig(path string) HTTPConfig {
	return config.MustLoadConfig(path)
}

// DefaultConfig 返回默认配置
func DefaultConfig() HTTPConfig {
	return config.DefaultConfig()
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() HTTPConfig {
	return config.LoadConfigFromEnv()
}

// CORS 跨域中间件
func CORS(config CORSConfig) HandlerFunc {
	return router.HandlerFunc(mid.CORS(config))
}

// Recovery 恢复中间件
func Recovery() HandlerFunc {
	return router.HandlerFunc(mid.Recovery())
}

// RequestID 请求ID中间件
func RequestID() HandlerFunc {
	return router.HandlerFunc(mid.RequestID())
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) HandlerFunc {
	return router.HandlerFunc(mid.Timeout(timeout))
}

// Auth 认证中间件
func Auth(authFunc func(*Context) error) HandlerFunc {
	middlewareAuthFunc := func(c *ctx.Context) error {
		return authFunc(c)
	}
	return router.HandlerFunc(mid.Auth(middlewareAuthFunc))
}

// RateLimit 限流中间件
func RateLimit(limiter RateLimiter, keyFunc func(*Context) string) HandlerFunc {
	middlewareKeyFunc := func(c *ctx.Context) string {
		return keyFunc(c)
	}
	return router.HandlerFunc(mid.RateLimit(limiter, middlewareKeyFunc))
}

// DefaultCORSConfig 返回默认CORS配置
func DefaultCORSConfig() CORSConfig {
	return mid.DefaultCORSConfig()
}

// DefaultRateLimiter 创建默认的令牌桶限流器
func DefaultRateLimiter() RateLimiter {
	return mid.DefaultRateLimiter()
}

// IPRateLimiter 创建基于IP的令牌桶限流器
func IPRateLimiter() RateLimiter {
	return mid.IPRateLimiter()
}

// URLRateLimiter 创建基于URL的令牌桶限流器
func URLRateLimiter() RateLimiter {
	return mid.URLRateLimiter()
}

// IPAndURLRateLimiter 创建IP和URL组合的限流器
func IPAndURLRateLimiter() RateLimiter {
	return mid.IPAndURLRateLimiter()
}

// GetClientIPKey 获取客户端IP的限流键值
func GetClientIPKey(c *Context) string {
	return mid.GetClientIPKey(c)
}

// GetRequestPathKey 获取请求路径的限流键值
func GetRequestPathKey(c *Context) string {
	return mid.GetRequestPathKey(c)
}

// GetIPAndURLKey 获取IP和URL组合的限流键值
func GetIPAndURLKey(c *Context) string {
	return mid.GetIPAndURLKey(c)
}

// NewTokenBucket 创建令牌桶限流器
func NewTokenBucket(rate, capacity float64) RateLimiter {
	return ratelimit.NewTokenBucket(rate, capacity)
}

// NewFixedWindow 创建固定窗口限流器
func NewFixedWindow(limit int, window time.Duration) RateLimiter {
	return ratelimit.NewFixedWindow(limit, window)
}

// NewSlidingWindow 创建滑动窗口限流器
func NewSlidingWindow(limit int, window time.Duration) RateLimiter {
	return ratelimit.NewSlidingWindow(limit, window)
}

// GetHealthCheck 获取服务器的健康检查组件
func GetHealthCheck(s *Server) *HealthCheck {
	return s.HealthCheck
}

// AddHealthCheck 添加健康检查项
func AddHealthCheck(s *Server, name string, check HealthCheckFunc) {
	s.HealthCheck.AddCheck(name, check)
}

// RemoveHealthCheck 移除健康检查项
func RemoveHealthCheck(s *Server, name string) {
	s.HealthCheck.RemoveCheck(name)
}

// SetHealthStatus 设置服务健康状态
func SetHealthStatus(s *Server, status HealthStatus) {
	s.HealthCheck.SetStatus(status)
}

// DBHealthCheck 创建数据库健康检查
func DBHealthCheck(ping func() error) HealthCheckFunc {
	return ht.DBCheck(ping)
}

// HTTPHealthCheck 创建HTTP服务健康检查
func HTTPHealthCheck(url string, timeout time.Duration) HealthCheckFunc {
	return ht.HTTPCheck(url, timeout)
}

// RedisHealthCheck 创建Redis健康检查
func RedisHealthCheck(ping func() error) HealthCheckFunc {
	return ht.RedisCheck(ping)
}

// DiskSpaceHealthCheck 创建磁盘空间健康检查
func DiskSpaceHealthCheck(path string, warningThreshold, criticalThreshold float64) HealthCheckFunc {
	return ht.DiskSpaceCheck(path, warningThreshold, criticalThreshold)
}

// MemoryHealthCheck 创建内存使用健康检查
func MemoryHealthCheck(warningThreshold, criticalThreshold float64) HealthCheckFunc {
	return ht.MemoryCheck(warningThreshold, criticalThreshold)
}

// CustomHealthCheck 创建自定义健康检查
func CustomHealthCheck(check func() (bool, string, map[string]interface{})) HealthCheckFunc {
	return ht.CustomCheck(check)
}

// EnableHealthCheck 启用健康检查
func EnableHealthCheck(s *Server) {
	if !s.Router.Config.EnableHealthCheck {
		s.Router.Config.EnableHealthCheck = true
		// 通过重新创建服务实例中的健康检查路由
		RegisterHealthEndpoints(s, s.Router.Config.HealthCheckPath)
	}
}

// DisableHealthCheck 禁用健康检查
func DisableHealthCheck(s *Server) {
	s.Router.Config.EnableHealthCheck = false
}

// IsHealthCheckEnabled 检查健康检查是否已启用
func IsHealthCheckEnabled(s *Server) bool {
	return s.Router.Config.EnableHealthCheck
}

// RegisterHealthEndpoints 注册健康检查端点
func RegisterHealthEndpoints(s *Server, basePath string) {
	if basePath == "" {
		basePath = "/health"
	}

	s.Router.GET(basePath, s.HealthCheck.Handler())
	s.Router.GET(basePath+"/liveness", s.HealthCheck.LivenessHandler())
	s.Router.GET(basePath+"/readiness", s.HealthCheck.ReadinessHandler())
	s.Router.GET(basePath+"/simple", s.HealthCheck.SimpleHandler())
}
