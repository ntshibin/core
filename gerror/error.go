// Package gerror 提供了一个增强的错误处理框架，支持错误包装、堆栈跟踪、错误码等高级特性。
// 该包设计用于在生产环境中提供更好的错误处理和调试能力。
package gerror

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// MaxStackDepth 限制堆栈跟踪的最大深度
	MaxStackDepth = 32
)

// Code 定义错误码类型
type Code int

// 预定义错误码
const (
	// 通用错误码
	CodeUnknown      Code = 10000 // 未知错误
	CodeInternal     Code = 10001 // 内部错误
	CodeInvalidParam Code = 10002 // 参数错误
	CodeUnauthorized Code = 10003 // 未授权
	CodeForbidden    Code = 10004 // 禁止访问
	CodeNotFound     Code = 10005 // 资源不存在
	CodeTimeout      Code = 10006 // 超时错误
	CodeConflict     Code = 10007 // 冲突错误
	CodeExhausted    Code = 10008 // 资源耗尽

	// 业务领域错误码 (11000-19999)
	// 可以根据不同业务领域进行扩展
)

// codeText 错误码对应的文本消息映射表
var codeText = map[Code]string{
	CodeUnknown:      "未知错误",
	CodeInternal:     "内部错误",
	CodeInvalidParam: "参数错误",
	CodeUnauthorized: "未授权",
	CodeForbidden:    "禁止访问",
	CodeNotFound:     "资源不存在",
	CodeTimeout:      "操作超时",
	CodeConflict:     "资源冲突",
	CodeExhausted:    "资源耗尽",
}

// RegisterCodeText 注册自定义错误码及对应文本
func RegisterCodeText(code Code, text string) {
	codeText[code] = text
}

// Text 返回错误码对应的文本描述
func (c Code) Text() string {
	if text, ok := codeText[c]; ok {
		return text
	}
	return "错误码(" + strconv.Itoa(int(c)) + ")"
}

// 定义自定义错误类型
type Error struct {
	code    Code        // 错误码
	message string      // 错误消息
	cause   error       // 原始错误
	stack   []uintptr   // 错误堆栈
	frames  []StackInfo // 解析后的堆栈信息
	context KV          // 额外上下文信息
	time    time.Time   // 错误发生时间
}

// StackInfo 保存堆栈帧信息
type StackInfo struct {
	File     string `json:"file"`     // 文件名
	Line     int    `json:"line"`     // 行号
	Function string `json:"function"` // 函数名
}

// KV 定义键值对类型，用于存储错误上下文
type KV map[string]interface{}

// New 创建一个新的错误，包含错误码和消息
func New(code Code, message string) error {
	return &Error{
		code:    code,
		message: message,
		time:    time.Now(),
		context: make(KV),
		stack:   captureStack(),
	}
}

// Newf 创建一个新的错误，支持格式化消息
func Newf(code Code, format string, args ...interface{}) error {
	return New(code, fmt.Sprintf(format, args...))
}

// Wrap 包装一个已有错误并添加额外信息
func Wrap(err error, code Code, message string) error {
	if err == nil {
		return nil
	}

	// 如果已经是我们的错误类型，保留原始错误码
	var gerr *Error
	if As(err, &gerr) && code == CodeUnknown {
		code = gerr.code
	}

	return &Error{
		code:    code,
		message: message,
		cause:   err,
		time:    time.Now(),
		context: make(KV),
		stack:   captureStack(),
	}
}

// Wrapf 包装一个已有错误并添加格式化的额外信息
func Wrapf(err error, code Code, format string, args ...interface{}) error {
	return Wrap(err, code, fmt.Sprintf(format, args...))
}

// WithStack 为普通错误添加堆栈跟踪
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	// 如果已经包含堆栈信息，则直接返回
	var gerr *Error
	if As(err, &gerr) {
		return err
	}

	return &Error{
		code:    CodeUnknown,
		message: err.Error(),
		cause:   err,
		time:    time.Now(),
		context: make(KV),
		stack:   captureStack(),
	}
}

// WithContext 为错误添加上下文信息
func WithContext(err error, key string, value interface{}) error {
	if err == nil {
		return nil
	}

	// 复制原始错误
	var gerr *Error
	if !As(err, &gerr) {
		// 如果不是我们的错误类型，先用WithStack转换
		err = WithStack(err)
		As(err, &gerr)
	}

	// 创建新的错误实例，复制原始信息
	newErr := &Error{
		code:    gerr.code,
		message: gerr.message,
		cause:   gerr.cause,
		time:    gerr.time,
		stack:   gerr.stack,
		context: make(KV),
	}

	// 复制已有上下文
	for k, v := range gerr.context {
		newErr.context[k] = v
	}

	// 添加新的上下文
	newErr.context[key] = value

	return newErr
}

// WithContextMap 为错误添加多个上下文信息
func WithContextMap(err error, ctx map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// 复制原始错误
	var gerr *Error
	if !As(err, &gerr) {
		// 如果不是我们的错误类型，先用WithStack转换
		err = WithStack(err)
		As(err, &gerr)
	}

	// 创建新的错误实例，复制原始信息
	newErr := &Error{
		code:    gerr.code,
		message: gerr.message,
		cause:   gerr.cause,
		time:    gerr.time,
		stack:   gerr.stack,
		context: make(KV),
	}

	// 复制已有上下文
	for k, v := range gerr.context {
		newErr.context[k] = v
	}

	// 添加新的上下文
	for k, v := range ctx {
		newErr.context[k] = v
	}

	return newErr
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %s", e.code, e.message, e.cause.Error())
	}
	return fmt.Sprintf("[%d] %s", e.code, e.message)
}

// Code 返回错误码
func (e *Error) Code() Code {
	return e.code
}

// Message 返回错误消息
func (e *Error) Message() string {
	return e.message
}

// Cause 返回原始错误
func (e *Error) Cause() error {
	return e.cause
}

// Unwrap 实现errors.Unwrap接口
func (e *Error) Unwrap() error {
	return e.cause
}

// Context 获取错误的上下文信息
func (e *Error) Context() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range e.context {
		result[k] = v
	}
	return result
}

// StackTrace 返回格式化的堆栈跟踪信息
func (e *Error) StackTrace() []StackInfo {
	if e.frames == nil {
		e.frames = parseStack(e.stack)
	}
	return e.frames
}

// Is 实现errors.Is接口
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.code == t.code
}

// captureStack 捕获当前的调用堆栈
func captureStack() []uintptr {
	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(3, stack[:])
	return stack[:length]
}

// parseStack 解析堆栈信息
func parseStack(stack []uintptr) []StackInfo {
	frames := runtime.CallersFrames(stack)
	result := make([]StackInfo, 0, len(stack))

	for {
		frame, more := frames.Next()
		// 过滤掉gerror包内的帧
		if !strings.Contains(frame.File, "gerror") || strings.Contains(frame.File, "_test.go") {
			result = append(result, StackInfo{
				File:     cleanPath(frame.File),
				Line:     frame.Line,
				Function: cleanFuncName(frame.Function),
			})
		}
		if !more {
			break
		}
	}

	return result
}

// cleanPath 简化文件路径
func cleanPath(filePath string) string {
	_, file := path.Split(filePath)
	return file
}

// cleanFuncName 简化函数名
func cleanFuncName(funcName string) string {
	// 获取包名/函数名
	if i := strings.LastIndex(funcName, "/"); i >= 0 {
		funcName = funcName[i+1:]
	}
	if i := strings.Index(funcName, "."); i >= 0 {
		funcName = funcName[i+1:]
	}
	return funcName
}

// FormatError 格式化错误信息，支持多种格式（文本、JSON）
func FormatError(err error, format string) string {
	if err == nil {
		return ""
	}

	var gerr *Error
	if !As(err, &gerr) {
		// 如果不是我们的错误类型，简单返回错误信息
		return err.Error()
	}

	switch format {
	case "json":
		return formatJSON(gerr)
	default:
		return formatText(gerr)
	}
}

// formatText 以文本格式格式化错误
func formatText(err *Error) string {
	var sb strings.Builder

	// 基本错误信息
	sb.WriteString(fmt.Sprintf("错误码: %d (%s)\n", err.code, err.code.Text()))
	sb.WriteString(fmt.Sprintf("消息: %s\n", err.message))
	sb.WriteString(fmt.Sprintf("时间: %s\n", err.time.Format(time.RFC3339)))

	// 上下文信息
	if len(err.context) > 0 {
		sb.WriteString("上下文:\n")
		for k, v := range err.context {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}

	// 堆栈信息
	stackInfo := err.StackTrace()
	if len(stackInfo) > 0 {
		sb.WriteString("堆栈:\n")
		for i, frame := range stackInfo {
			sb.WriteString(fmt.Sprintf("  %d: %s:%d [%s]\n", i, frame.File, frame.Line, frame.Function))
		}
	}

	// 原始错误
	if err.cause != nil {
		sb.WriteString(fmt.Sprintf("原因: %s\n", err.cause.Error()))
	}

	return sb.String()
}

// formatJSON 以JSON格式格式化错误
func formatJSON(err *Error) string {
	type jsonError struct {
		Code      int                    `json:"code"`
		CodeText  string                 `json:"code_text"`
		Message   string                 `json:"message"`
		Time      string                 `json:"time"`
		Context   map[string]interface{} `json:"context,omitempty"`
		Stack     []StackInfo            `json:"stack,omitempty"`
		Cause     string                 `json:"cause,omitempty"`
		CauseType string                 `json:"cause_type,omitempty"`
	}

	je := jsonError{
		Code:     int(err.code),
		CodeText: err.code.Text(),
		Message:  err.message,
		Time:     err.time.Format(time.RFC3339),
		Context:  err.Context(),
		Stack:    err.StackTrace(),
	}

	if err.cause != nil {
		je.Cause = err.cause.Error()
		je.CauseType = fmt.Sprintf("%T", err.cause)
	}

	bytes, e := json.MarshalIndent(je, "", "  ")
	if e != nil {
		return fmt.Sprintf("错误序列化失败: %s", e.Error())
	}

	return string(bytes)
}

// Func 错误包装函数类型，用于简化中间件和包装器的实现
type Func func(error) error

// Chain 链接多个错误处理函数
func Chain(funcs ...Func) Func {
	return func(err error) error {
		for _, f := range funcs {
			err = f(err)
		}
		return err
	}
}

// As 是errors.As的包装
func As(err error, target interface{}) bool {
	return AsFunc(err, target)
}

// Is 是errors.Is的包装
func Is(err, target error) bool {
	return IsFunc(err, target)
}

// 默认实现，可以被测试覆盖
var (
	AsFunc func(err error, target interface{}) bool = func(err error, target interface{}) bool {
		return false // 实际实现会在init中设置
	}
	IsFunc func(err, target error) bool = func(err, target error) bool {
		return false // 实际实现会在init中设置
	}
)

// WriteStackTrace 将错误的堆栈跟踪写入到指定的writer
func WriteStackTrace(err error, w io.Writer) {
	var gerr *Error
	if !As(err, &gerr) {
		fmt.Fprintf(w, "错误不包含堆栈信息: %s\n", err.Error())
		return
	}

	for i, frame := range gerr.StackTrace() {
		fmt.Fprintf(w, "%d: %s:%d [%s]\n", i, frame.File, frame.Line, frame.Function)
	}
}

// RootCause 获取错误链中的根本原因
func RootCause(err error) error {
	for err != nil {
		unwrapped := Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
	return nil
}

// GetCode 从错误中获取错误码
func GetCode(err error) Code {
	var gerr *Error
	if As(err, &gerr) {
		return gerr.Code()
	}
	return CodeUnknown
}

// GetMessage 从错误中获取错误消息
func GetMessage(err error) string {
	if err == nil {
		return ""
	}

	var gerr *Error
	if As(err, &gerr) {
		return gerr.Message()
	}
	return err.Error()
}

// GetContext 从错误中获取上下文
func GetContext(err error) map[string]interface{} {
	var gerr *Error
	if As(err, &gerr) {
		return gerr.Context()
	}
	return nil
}

// Unwrap 是errors.Unwrap的包装
func Unwrap(err error) error {
	return UnwrapFunc(err)
}

// 默认实现，可以被测试覆盖
var UnwrapFunc func(err error) error = func(err error) error {
	return nil // 实际实现会在init中设置
}
