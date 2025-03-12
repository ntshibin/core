package errorx

import (
	"fmt"
	"runtime"
	"strings"
)

// Error 自定义错误结构体
type Error struct {
	code    int
	message string
	cause   error
	stack   string
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Code 获取错误码
func (e *Error) Code() int {
	return e.code
}

// Cause 获取原始错误
func (e *Error) Cause() error {
	return e.cause
}

// Stack 获取堆栈信息
func (e *Error) Stack() string {
	return e.stack
}

// New 创建新的错误
func New(code int, message string) error {
	return &Error{
		code:    code,
		message: message,
		stack:   getStack(),
	}
}

// Wrap 包装错误
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	code := HTTPCodeError
	if e, ok := err.(*Error); ok {
		code = e.code
	}

	return &Error{
		code:    code,
		message: message,
		cause:   err,
		stack:   getStack(),
	}
}

// WrapWithCode 包装错误并指定错误码
func WrapWithCode(err error, code int, message string) error {
	if err == nil {
		return nil
	}

	return &Error{
		code:    code,
		message: message,
		cause:   err,
		stack:   getStack(),
	}
}

// 获取堆栈信息
func getStack() string {
	var stack strings.Builder
	for i := 2; i < 15; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		stack.WriteString(fmt.Sprintf("\n\t%s:%d %s", file, line, fn.Name()))
	}
	return stack.String()
}
