package middleware

import (
	"github.com/ntshibin/core/gerror"
	"github.com/ntshibin/core/ghttp/internal/context"
)

// Auth 认证中间件，验证请求是否包含有效的认证信息
func Auth(authFunc func(*context.Context) error) HandlerFunc {
	return func(c *context.Context) {
		if err := authFunc(c); err != nil {
			// 认证失败
			var gErr *gerror.Error
			if gerror.As(err, &gErr) && gErr.Code() == gerror.CodeUnauthorized {
				c.Unauthorized(err)
			} else {
				c.Unauthorized(gerror.New(gerror.CodeUnauthorized, "未授权访问"))
			}
			c.Abort()
			return
		}

		// 认证成功，继续处理请求
		c.Next()
	}
}
