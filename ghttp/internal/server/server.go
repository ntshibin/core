// Package server 提供了HTTP服务器的管理
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ntshibin/core/ghttp/internal/config"
	"github.com/ntshibin/core/ghttp/internal/health"
	"github.com/ntshibin/core/ghttp/internal/middleware"
	"github.com/ntshibin/core/ghttp/internal/router"
	"github.com/ntshibin/core/glog"
)

// Server 包含HTTP服务和路由
type Server struct {
	Router      *router.Router
	HealthCheck *health.HealthCheck
}

// New 创建一个新的Server实例
func New(config config.HTTPConfig) *Server {
	// 设置Gin模式
	gin.SetMode(config.Mode)

	// 创建Gin引擎
	engine := gin.New()

	// 添加中间件
	engine.Use(
		gin.Recovery(),
		middleware.Logger(),
		middleware.ErrorHandler(),
	)

	// 设置受信任的代理
	if len(config.TrustedProxies) > 0 {
		engine.SetTrustedProxies(config.TrustedProxies)
	}

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      engine,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	// 获取服务名称和版本
	serviceName := getServiceName(config.ServiceName)
	serviceVersion := getServiceVersion(config.ServiceVersion)

	// 创建健康检查
	healthCheck := health.NewHealthCheck(serviceName, serviceVersion)

	// 创建服务实例
	s := &Server{
		Router: &router.Router{
			Engine: engine,
			Server: server,
			Config: config,
		},
		HealthCheck: healthCheck,
	}

	// 注册健康检查路由（默认启用，可通过配置禁用）
	if config.EnableHealthCheck {
		s.registerHealthEndpoints(config.HealthCheckPath)
	}

	return s
}

// Run 启动HTTP服务器
func (s *Server) Run() error {
	glog.Infof("HTTP服务启动于端口: %d, 模式: %s", s.Router.Config.Port, s.Router.Config.Mode)
	return s.Router.Server.ListenAndServe()
}

// Shutdown 优雅关闭HTTP服务器
func (s *Server) Shutdown() error {
	glog.Info("正在关闭HTTP服务...")

	ctx, cancel := context.WithTimeout(context.Background(), s.Router.Config.ShutdownTimeout)
	defer cancel()

	return s.Router.Server.Shutdown(ctx)
}

// registerHealthEndpoints 注册健康检查端点
func (s *Server) registerHealthEndpoints(basePath string) {
	if basePath == "" {
		basePath = "/health"
	}

	s.Router.GET(basePath, s.HealthCheck.Handler())
	s.Router.GET(basePath+"/liveness", s.HealthCheck.LivenessHandler())
	s.Router.GET(basePath+"/readiness", s.HealthCheck.ReadinessHandler())
	s.Router.GET(basePath+"/simple", s.HealthCheck.SimpleHandler())
}

// RegisterHealthEndpoints 注册健康检查端点（导出版本）
func (s *Server) RegisterHealthEndpoints(basePath string) {
	s.registerHealthEndpoints(basePath)
}

// getServiceName 获取服务名称
func getServiceName(configName string) string {
	if configName != "" {
		return configName
	}
	// 尝试从环境变量获取
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		return name
	}
	// 使用默认值
	return "ghttp-service"
}

// getServiceVersion 获取服务版本
func getServiceVersion(configVersion string) string {
	if configVersion != "" {
		return configVersion
	}
	// 尝试从环境变量获取
	if version := os.Getenv("SERVICE_VERSION"); version != "" {
		return version
	}
	// 使用默认值
	return "1.0.0"
}
