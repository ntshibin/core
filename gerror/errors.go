// Package gerror provides enhanced error handling capabilities.
package gerror

import (
	"errors"
)

func init() {
	// 设置实际实现
	AsFunc = errors.As
	IsFunc = errors.Is
	UnwrapFunc = errors.Unwrap
}

// Join 合并多个错误为一个错误。
// 如果没有错误或所有错误都是nil，返回nil。
func Join(errs ...error) error {
	return errors.Join(errs...)
}
