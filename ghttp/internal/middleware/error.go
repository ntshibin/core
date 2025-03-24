package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ntshibin/core/gerror"
	"github.com/ntshibin/core/ghttp/internal/context"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 先处理请求

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last().Err

			// 检查是否为 gerror 类型的错误
			var gErr *gerror.Error
			if gerror.As(err, &gErr) {
				c.JSON(http.StatusInternalServerError, context.Result{
					Code:    int(gErr.Code()),
					Message: gErr.Message(),
					Data:    nil,
				})
			} else {
				// 其他错误类型
				c.JSON(http.StatusInternalServerError, context.Result{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
					Data:    nil,
				})
			}

			// 已处理错误，不再继续传播
			c.Abort()
		}
	}
}
