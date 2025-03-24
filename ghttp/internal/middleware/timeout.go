package middleware

import (
	stdctx "context"
	"net/http"
	"time"

	"github.com/ntshibin/core/ghttp/internal/context"
	"github.com/ntshibin/core/glog"
)

// Timeout 超时中间件，设置请求超时时间
func Timeout(timeout time.Duration) HandlerFunc {
	return func(c *context.Context) {
		// 创建一个带超时的上下文
		ctx, cancel := stdctx.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 设置请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 使用 channel 控制超时
		done := make(chan bool, 1)

		go func() {
			// 处理请求
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			// 请求处理完成
			return
		case <-ctx.Done():
			// 请求超时
			if ctx.Err() == stdctx.DeadlineExceeded {
				glog.Warnf("请求超时: %s %s", c.Request.Method, c.Request.URL.Path)
				c.Abort()
				c.Fail(http.StatusRequestTimeout, "请求处理超时")
			}
		}
	}
}
