package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/ntshibin/core/ghttp/internal/context"
	"github.com/ntshibin/core/glog"
)

// Recovery 恢复中间件，捕获 panic 并返回500错误
func Recovery() HandlerFunc {
	return func(c *context.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 打印堆栈信息
				stack := string(debug.Stack())
				glog.Errorf("Panic 恢复: %v\n%s", err, stack)

				// 返回500错误
				c.Abort()
				if errStr, ok := err.(string); ok {
					c.Fail(http.StatusInternalServerError, errStr)
				} else {
					c.Fail(http.StatusInternalServerError, "内部服务器错误")
				}
			}
		}()

		c.Next()
	}
}
