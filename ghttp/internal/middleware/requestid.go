package middleware

import (
	"fmt"
	"time"

	"github.com/ntshibin/core/ghttp/internal/context"
)

// RequestID 请求ID中间件，为每个请求生成唯一ID
func RequestID() HandlerFunc {
	return func(c *context.Context) {
		// 获取请求头中的请求ID
		requestID := c.GetHeader("X-Request-ID")

		// 如果请求头中没有请求ID，则生成一个
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置请求ID
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}
