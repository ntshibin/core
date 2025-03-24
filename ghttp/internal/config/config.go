// Package config 提供了HTTP服务的配置管理
package config

import (
	"time"

	"github.com/ntshibin/core/gconf"
	"github.com/ntshibin/core/gerror"
)

// 配置相关错误码
const (
	ErrConfigLoad gerror.Code = 30000 // 配置加载错误
)

// HTTPConfig 提供了HTTP服务器的配置选项
type HTTPConfig struct {
	Mode              string        `json:"mode" yaml:"mode" env:"HTTP_MODE" default:"release"`                                      // 运行模式: debug, release, test
	Port              int           `json:"port" yaml:"port" env:"HTTP_PORT" default:"8080"`                                         // 监听端口
	ReadTimeout       time.Duration `json:"read_timeout" yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" default:"10s"`                  // 请求读取超时
	WriteTimeout      time.Duration `json:"write_timeout" yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" default:"10s"`               // 响应写入超时
	ShutdownTimeout   time.Duration `json:"shutdown_timeout" yaml:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" default:"5s"`       // 优雅关闭超时
	TrustedProxies    []string      `json:"trusted_proxies" yaml:"trusted_proxies" env:"HTTP_TRUSTED_PROXIES"`                       // 受信任的代理
	ServiceName       string        `json:"service_name" yaml:"service_name" env:"SERVICE_NAME" default:"ghttp-service"`             // 服务名称
	ServiceVersion    string        `json:"service_version" yaml:"service_version" env:"SERVICE_VERSION" default:"1.0.0"`            // 服务版本
	EnableHealthCheck bool          `json:"enable_health_check" yaml:"enable_health_check" env:"ENABLE_HEALTH_CHECK" default:"true"` // 是否启用健康检查
	HealthCheckPath   string        `json:"health_check_path" yaml:"health_check_path" env:"HEALTH_CHECK_PATH" default:"/health"`    // 健康检查路径前缀
}

// LoadConfig 从文件加载HTTP配置
func LoadConfig(path string) (HTTPConfig, error) {
	var config HTTPConfig
	err := gconf.Load(path, &config)
	if err != nil {
		return DefaultConfig(), gerror.Wrapf(err, ErrConfigLoad, "加载HTTP配置失败: %s", path)
	}
	return config, nil
}

// MustLoadConfig 从文件加载HTTP配置，出错时panic
func MustLoadConfig(path string) HTTPConfig {
	var config HTTPConfig
	gconf.MustLoad(path, &config)
	return config
}

// DefaultConfig 返回默认的HTTP配置
func DefaultConfig() HTTPConfig {
	var config HTTPConfig
	// 只加载默认值，不处理环境变量
	if err := gconf.LoadDefaultsFromStruct(&config); err != nil {
		// 如果出错则返回硬编码默认值
		return HTTPConfig{
			Mode:              "release",
			Port:              8080,
			ReadTimeout:       time.Second * 10,
			WriteTimeout:      time.Second * 10,
			ShutdownTimeout:   time.Second * 5,
			TrustedProxies:    []string{},
			ServiceName:       "ghttp-service",
			ServiceVersion:    "1.0.0",
			EnableHealthCheck: true,
			HealthCheckPath:   "/health",
		}
	}
	return config
}

// LoadConfigFromEnv 从环境变量加载HTTP配置
func LoadConfigFromEnv() HTTPConfig {
	var config HTTPConfig
	// 先加载默认值
	gconf.LoadDefaultsFromStruct(&config)
	// 再加载环境变量
	gconf.LoadEnvFromStruct(&config)
	return config
}
