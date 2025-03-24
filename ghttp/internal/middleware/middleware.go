// Package middleware 提供了各种HTTP请求中间件
package middleware

import (
	"github.com/ntshibin/core/ghttp/internal/context"
)

// HandlerFunc 定义了处理HTTP请求的函数类型
type HandlerFunc func(*context.Context)

// joinStrings 合并字符串数组，用逗号分隔
func joinStrings(arr []string) string {
	if len(arr) == 0 {
		return ""
	}

	result := arr[0]
	for i := 1; i < len(arr); i++ {
		result += ", " + arr[i]
	}

	return result
}
