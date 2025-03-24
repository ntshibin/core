// Package gerror provides enhanced error handling capabilities.
package gerror

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// 定义gRPC相关错误码
const (
	CodeGRPCCanceled           Code = 12000 // 请求被取消
	CodeGRPCUnknown            Code = 12001 // 未知错误
	CodeGRPCInvalidArgument    Code = 12002 // 无效参数
	CodeGRPCDeadlineExceeded   Code = 12003 // 请求超时
	CodeGRPCNotFound           Code = 12004 // 资源未找到
	CodeGRPCAlreadyExists      Code = 12005 // 资源已存在
	CodeGRPCPermissionDenied   Code = 12006 // 权限拒绝
	CodeGRPCResourceExhausted  Code = 12007 // 资源耗尽
	CodeGRPCFailedPrecondition Code = 12008 // 前置条件失败
	CodeGRPCAborted            Code = 12009 // 操作中止
	CodeGRPCOutOfRange         Code = 12010 // 超出范围
	CodeGRPCUnimplemented      Code = 12011 // 未实现
	CodeGRPCInternal           Code = 12012 // 内部错误
	CodeGRPCUnavailable        Code = 12013 // 服务不可用
	CodeGRPCDataLoss           Code = 12014 // 数据丢失
	CodeGRPCUnauthenticated    Code = 12015 // 未认证
)

func init() {
	// 注册GRPC错误描述文本
	RegisterCodeText(CodeGRPCCanceled, "请求被取消")
	RegisterCodeText(CodeGRPCUnknown, "未知错误")
	RegisterCodeText(CodeGRPCInvalidArgument, "无效参数")
	RegisterCodeText(CodeGRPCDeadlineExceeded, "请求超时")
	RegisterCodeText(CodeGRPCNotFound, "资源未找到")
	RegisterCodeText(CodeGRPCAlreadyExists, "资源已存在")
	RegisterCodeText(CodeGRPCPermissionDenied, "权限拒绝")
	RegisterCodeText(CodeGRPCResourceExhausted, "资源耗尽")
	RegisterCodeText(CodeGRPCFailedPrecondition, "前置条件失败")
	RegisterCodeText(CodeGRPCAborted, "操作中止")
	RegisterCodeText(CodeGRPCOutOfRange, "超出范围")
	RegisterCodeText(CodeGRPCUnimplemented, "未实现")
	RegisterCodeText(CodeGRPCInternal, "内部错误")
	RegisterCodeText(CodeGRPCUnavailable, "服务不可用")
	RegisterCodeText(CodeGRPCDataLoss, "数据丢失")
	RegisterCodeText(CodeGRPCUnauthenticated, "未认证")
}

// 常见gRPC错误信息检测关键词
var (
	canceledKeywords      = []string{"canceled", "cancelled", "cancel"}
	timeoutKeywords       = []string{"timeout", "timed out", "deadline", "exceed"}
	notFoundKeywords      = []string{"not found", "notfound", "no such"}
	alreadyExistsKeywords = []string{"already exists", "exist"}
	permissionKeywords    = []string{"permission", "unauthorized", "denied", "access"}
	invalidArgKeywords    = []string{"invalid", "argument", "param", "malform"}
	unavailableKeywords   = []string{"unavailable", "unreachable", "temporarily"}
)

// IsCanceled 检查错误是否是取消类型
func IsCanceled(err error) bool {
	if err == nil {
		return false
	}

	// 检查标准库取消错误
	if errors.Is(err, context.Canceled) {
		return true
	}

	// 检查我们的错误类型
	if Is(err, New(CodeGRPCCanceled, "")) {
		return true
	}

	// 检查错误信息
	errMsg := strings.ToLower(err.Error())
	for _, kw := range canceledKeywords {
		if strings.Contains(errMsg, kw) {
			return true
		}
	}

	return false
}

// IsTimeout 检查错误是否是超时类型
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	// 检查标准库超时错误
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// 检查我们的错误类型
	if Is(err, New(CodeGRPCDeadlineExceeded, "")) || Is(err, New(CodeTimeout, "")) {
		return true
	}

	// 检查错误信息
	errMsg := strings.ToLower(err.Error())
	for _, kw := range timeoutKeywords {
		if strings.Contains(errMsg, kw) {
			return true
		}
	}

	return false
}

// IsUnavailable 检查错误是否表示服务不可用
func IsUnavailable(err error) bool {
	if err == nil {
		return false
	}

	// 检查我们的错误类型
	if Is(err, New(CodeGRPCUnavailable, "")) {
		return true
	}

	// 检查错误信息
	errMsg := strings.ToLower(err.Error())
	for _, kw := range unavailableKeywords {
		if strings.Contains(errMsg, kw) {
			return true
		}
	}

	return false
}

// IsInvalidArgument 检查错误是否表示无效参数
func IsInvalidArgument(err error) bool {
	if err == nil {
		return false
	}

	// 检查我们的错误类型
	if Is(err, New(CodeGRPCInvalidArgument, "")) || Is(err, New(CodeInvalidParam, "")) {
		return true
	}

	// 检查错误信息
	errMsg := strings.ToLower(err.Error())
	for _, kw := range invalidArgKeywords {
		if strings.Contains(errMsg, kw) {
			return true
		}
	}

	return false
}

// WrapRPCError 包装微服务/RPC调用中的错误
func WrapRPCError(err error, service string, method string, message string) error {
	if err == nil {
		return nil
	}

	var code Code

	// 根据错误类型选择适当的错误码
	switch {
	case IsCanceled(err):
		code = CodeGRPCCanceled
	case IsTimeout(err):
		code = CodeGRPCDeadlineExceeded
	case IsNotFound(err):
		code = CodeGRPCNotFound
	case IsDuplicate(err):
		code = CodeGRPCAlreadyExists
	case IsInvalidArgument(err):
		code = CodeGRPCInvalidArgument
	case IsUnavailable(err):
		code = CodeGRPCUnavailable
	default:
		code = CodeGRPCUnknown
	}

	// 构造错误消息
	if message == "" {
		if service != "" && method != "" {
			message = "调用服务 " + service + "." + method + " 失败"
		} else {
			message = "RPC调用失败"
		}
	}

	// 使用上下文信息丰富错误
	wrappedErr := Wrap(err, code, message)
	if service != "" {
		wrappedErr = WithContext(wrappedErr, "service", service)
	}
	if method != "" {
		wrappedErr = WithContext(wrappedErr, "method", method)
	}

	return wrappedErr
}

// RPCNotFoundError 创建一个RPC资源未找到错误
func RPCNotFoundError(resource string, id interface{}) error {
	msg := "资源未找到"
	if resource != "" {
		if id != nil {
			msg = resource + "(ID:" + fmt.Sprintf("%v", id) + ") 未找到"
		} else {
			msg = resource + " 未找到"
		}
	}
	return New(CodeGRPCNotFound, msg)
}

// RPCTimeoutError 创建一个RPC超时错误
func RPCTimeoutError(service string, method string, duration interface{}) error {
	var msg string
	if service != "" && method != "" {
		if duration != nil {
			msg = "调用服务 " + service + "." + method + " 超时，耗时: " + fmt.Sprintf("%v", duration)
		} else {
			msg = "调用服务 " + service + "." + method + " 超时"
		}
	} else {
		msg = "RPC请求超时"
	}
	return New(CodeGRPCDeadlineExceeded, msg)
}

// RPCUnimplementedError 创建一个未实现功能的错误
func RPCUnimplementedError(feature string) error {
	var msg string
	if feature != "" {
		msg = "功能 " + feature + " 尚未实现"
	} else {
		msg = "功能尚未实现"
	}
	return New(CodeGRPCUnimplemented, msg)
}
