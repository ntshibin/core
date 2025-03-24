package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ntshibin/core/ghttp/internal/context"
)

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowOrigins     []string      // 允许的源
	AllowMethods     []string      // 允许的方法
	AllowHeaders     []string      // 允许的头部
	ExposeHeaders    []string      // 暴露的头部
	AllowCredentials bool          // 是否允许凭证
	MaxAge           time.Duration // 预检请求缓存时间
}

// DefaultCORSConfig 默认跨域配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

// CORS 创建跨域中间件
func CORS(config CORSConfig) HandlerFunc {
	return func(c *context.Context) {
		origin := c.Request.Header.Get("Origin")

		// 设置允许的源
		if len(config.AllowOrigins) == 0 || config.AllowOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			allowed := false
			for _, allowedOrigin := range config.AllowOrigins {
				if allowedOrigin == origin {
					c.Header("Access-Control-Allow-Origin", origin)
					allowed = true
					break
				}
			}
			if !allowed {
				c.Header("Access-Control-Allow-Origin", config.AllowOrigins[0])
			}
		}

		// 设置允许的方法
		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", joinStrings(config.AllowMethods))
		}

		// 设置允许的头部
		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders))
		}

		// 设置暴露的头部
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders))
		}

		// 设置是否允许凭证
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求缓存时间
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", int(config.MaxAge.Seconds())))
		}

		// 如果是预检请求，则直接返回
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
