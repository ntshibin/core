package gerror_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ntshibin/core/gerror"
	"github.com/stretchr/testify/assert"
)

func TestIsCanceled(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "空错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "context.Canceled",
			err:      context.Canceled,
			expected: true,
		},
		{
			name:     "gerror.CodeGRPCCanceled",
			err:      gerror.New(gerror.CodeGRPCCanceled, "请求被取消"),
			expected: true,
		},
		{
			name:     "包含cancel关键词",
			err:      errors.New("operation was canceled by user"),
			expected: true,
		},
		{
			name:     "包含cancelled关键词(英式拼写)",
			err:      errors.New("request cancelled"),
			expected: true,
		},
		{
			name:     "其他错误",
			err:      errors.New("连接超时"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := gerror.IsCanceled(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsTimeout(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "空错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "context.DeadlineExceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "gerror.CodeGRPCDeadlineExceeded",
			err:      gerror.New(gerror.CodeGRPCDeadlineExceeded, "请求超时"),
			expected: true,
		},
		{
			name:     "gerror.CodeTimeout",
			err:      gerror.New(gerror.CodeTimeout, "操作超时"),
			expected: true,
		},
		{
			name:     "包含timeout关键词",
			err:      errors.New("operation timed out after 3s"),
			expected: true,
		},
		{
			name:     "包含deadline关键词",
			err:      errors.New("deadline exceeded"),
			expected: true,
		},
		{
			name:     "包含exceed关键词",
			err:      errors.New("request time limit exceeded"),
			expected: true,
		},
		{
			name:     "其他错误",
			err:      errors.New("连接失败"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := gerror.IsTimeout(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUnavailable(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "空错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "gerror.CodeGRPCUnavailable",
			err:      gerror.New(gerror.CodeGRPCUnavailable, "服务不可用"),
			expected: true,
		},
		{
			name:     "包含unavailable关键词",
			err:      errors.New("service unavailable"),
			expected: true,
		},
		{
			name:     "包含unreachable关键词",
			err:      errors.New("host unreachable"),
			expected: true,
		},
		{
			name:     "包含temporarily关键词",
			err:      errors.New("temporarily unavailable"),
			expected: true,
		},
		{
			name:     "其他错误",
			err:      errors.New("连接失败"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := gerror.IsUnavailable(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsInvalidArgument(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "空错误",
			err:      nil,
			expected: false,
		},
		{
			name:     "gerror.CodeGRPCInvalidArgument",
			err:      gerror.New(gerror.CodeGRPCInvalidArgument, "无效参数"),
			expected: true,
		},
		{
			name:     "gerror.CodeInvalidParam",
			err:      gerror.New(gerror.CodeInvalidParam, "参数错误"),
			expected: true,
		},
		{
			name:     "包含invalid关键词",
			err:      errors.New("invalid request parameter"),
			expected: true,
		},
		{
			name:     "包含argument关键词",
			err:      errors.New("bad argument"),
			expected: true,
		},
		{
			name:     "包含param关键词",
			err:      errors.New("missing required param"),
			expected: true,
		},
		{
			name:     "包含malform关键词",
			err:      errors.New("malformed request"),
			expected: true,
		},
		{
			name:     "其他错误",
			err:      errors.New("连接失败"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := gerror.IsInvalidArgument(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWrapRPCError(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		service      string
		method       string
		message      string
		expectedCode gerror.Code
	}{
		{
			name:         "空错误",
			err:          nil,
			service:      "UserService",
			method:       "GetUser",
			message:      "测试消息",
			expectedCode: 0, // 应该返回nil
		},
		{
			name:         "取消错误",
			err:          context.Canceled,
			service:      "UserService",
			method:       "GetUser",
			message:      "获取用户失败",
			expectedCode: gerror.CodeGRPCCanceled,
		},
		{
			name:         "超时错误",
			err:          context.DeadlineExceeded,
			service:      "UserService",
			method:       "GetUser",
			message:      "获取用户超时",
			expectedCode: gerror.CodeGRPCDeadlineExceeded,
		},
		{
			name:         "未找到错误",
			err:          errors.New("record not found"),
			service:      "UserService",
			method:       "GetUser",
			message:      "用户不存在",
			expectedCode: gerror.CodeGRPCNotFound,
		},
		{
			name:         "重复记录错误",
			err:          errors.New("already exists"),
			service:      "UserService",
			method:       "CreateUser",
			message:      "创建用户失败",
			expectedCode: gerror.CodeGRPCAlreadyExists,
		},
		{
			name:         "无效参数错误",
			err:          errors.New("invalid parameter"),
			service:      "UserService",
			method:       "UpdateUser",
			message:      "更新用户失败",
			expectedCode: gerror.CodeGRPCInvalidArgument,
		},
		{
			name:         "服务不可用错误",
			err:          errors.New("service unavailable"),
			service:      "UserService",
			method:       "GetUser",
			message:      "服务暂时不可用",
			expectedCode: gerror.CodeGRPCUnavailable,
		},
		{
			name:         "未知错误",
			err:          errors.New("unknown error"),
			service:      "UserService",
			method:       "GetUser",
			message:      "未知错误",
			expectedCode: gerror.CodeGRPCUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrappedErr := gerror.WrapRPCError(tc.err, tc.service, tc.method, tc.message)

			if tc.err == nil {
				assert.Nil(t, wrappedErr)
				return
			}

			assert.NotNil(t, wrappedErr)
			assert.Equal(t, tc.expectedCode, gerror.GetCode(wrappedErr))
			assert.Equal(t, tc.message, gerror.GetMessage(wrappedErr))

			// 检查上下文
			ctx := gerror.GetContext(wrappedErr)
			assert.Equal(t, tc.service, ctx["service"])
			assert.Equal(t, tc.method, ctx["method"])
		})
	}
}

func TestRPCNotFoundError(t *testing.T) {
	// 测试不带ID
	err1 := gerror.RPCNotFoundError("用户", nil)
	assert.Equal(t, gerror.CodeGRPCNotFound, gerror.GetCode(err1))
	assert.Contains(t, gerror.GetMessage(err1), "用户 未找到")

	// 测试带ID
	err2 := gerror.RPCNotFoundError("订单", 12345)
	assert.Equal(t, gerror.CodeGRPCNotFound, gerror.GetCode(err2))
	assert.Contains(t, gerror.GetMessage(err2), "订单(ID:12345) 未找到")

	// 测试空实体
	err3 := gerror.RPCNotFoundError("", nil)
	assert.Equal(t, gerror.CodeGRPCNotFound, gerror.GetCode(err3))
	assert.Contains(t, gerror.GetMessage(err3), "资源未找到")
}

func TestRPCTimeoutError(t *testing.T) {
	// 测试不带持续时间
	err1 := gerror.RPCTimeoutError("UserService", "GetUser", nil)
	assert.Equal(t, gerror.CodeGRPCDeadlineExceeded, gerror.GetCode(err1))
	assert.Contains(t, gerror.GetMessage(err1), "调用服务 UserService.GetUser 超时")

	// 测试带持续时间
	err2 := gerror.RPCTimeoutError("UserService", "GetUser", "3s")
	assert.Equal(t, gerror.CodeGRPCDeadlineExceeded, gerror.GetCode(err2))
	assert.Contains(t, gerror.GetMessage(err2), "调用服务 UserService.GetUser 超时，耗时: 3s")

	// 测试空服务和方法
	err3 := gerror.RPCTimeoutError("", "", nil)
	assert.Equal(t, gerror.CodeGRPCDeadlineExceeded, gerror.GetCode(err3))
	assert.Contains(t, gerror.GetMessage(err3), "RPC请求超时")
}

func TestRPCUnimplementedError(t *testing.T) {
	// 测试带功能名
	err1 := gerror.RPCUnimplementedError("文件上传")
	assert.Equal(t, gerror.CodeGRPCUnimplemented, gerror.GetCode(err1))
	assert.Contains(t, gerror.GetMessage(err1), "功能 文件上传 尚未实现")

	// 测试空功能名
	err2 := gerror.RPCUnimplementedError("")
	assert.Equal(t, gerror.CodeGRPCUnimplemented, gerror.GetCode(err2))
	assert.Contains(t, gerror.GetMessage(err2), "功能尚未实现")
}
