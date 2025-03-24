package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ntshibin/core/glog"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 获取状态
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		// 记录日志
		if raw != "" {
			path = path + "?" + raw
		}

		// 根据状态码选择日志级别
		switch {
		case status >= 500:
			glog.Errorf("| %3d | %12v | %15s | %s | %s |", status, latency, clientIP, method, path)
		case status >= 400:
			glog.Warnf("| %3d | %12v | %15s | %s | %s |", status, latency, clientIP, method, path)
		case status >= 300:
			glog.Infof("| %3d | %12v | %15s | %s | %s |", status, latency, clientIP, method, path)
		default:
			glog.Infof("| %3d | %12v | %15s | %s | %s |", status, latency, clientIP, method, path)
		}
	}
}
