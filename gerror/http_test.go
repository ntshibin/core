package gerror_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ntshibin/core/gerror"
	"github.com/stretchr/testify/assert"
)

func TestGetHTTPStatus(t *testing.T) {
	testCases := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "内部错误",
			err:            gerror.New(gerror.CodeInternal, "服务器内部错误"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "未找到",
			err:            gerror.New(gerror.CodeNotFound, "资源不存在"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "参数错误",
			err:            gerror.New(gerror.CodeInvalidParam, "无效参数"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "未授权",
			err:            gerror.New(gerror.CodeUnauthorized, "未授权访问"),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "禁止访问",
			err:            gerror.New(gerror.CodeForbidden, "禁止访问"),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "超时错误",
			err:            gerror.New(gerror.CodeTimeout, "操作超时"),
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			name:           "冲突错误",
			err:            gerror.New(gerror.CodeConflict, "资源冲突"),
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "资源耗尽",
			err:            gerror.New(gerror.CodeExhausted, "资源耗尽"),
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "未知错误码",
			err:            gerror.New(gerror.Code(99999), "未知错误"),
			expectedStatus: http.StatusInternalServerError, // 默认值
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status := gerror.GetHTTPStatus(tc.err)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func TestGetHTTPResponse(t *testing.T) {
	// 测试基本错误
	err := gerror.New(gerror.CodeNotFound, "用户不存在")
	resp := gerror.GetHTTPResponse(err)

	assert.Equal(t, int(gerror.CodeNotFound), resp.Code)
	assert.Equal(t, "用户不存在", resp.Message)

	// 测试带上下文的错误
	errWithCtx := gerror.WithContext(err, "用户ID", 123)
	resp = gerror.GetHTTPResponse(errWithCtx)

	assert.Equal(t, int(gerror.CodeNotFound), resp.Code)
	assert.Equal(t, "用户不存在", resp.Message)
	assert.Equal(t, 123, resp.Details["用户ID"])
}

func TestWriteHTTPError(t *testing.T) {
	// 创建测试错误
	err := gerror.New(gerror.CodeForbidden, "访问被拒绝")
	err = gerror.WithContext(err, "资源", "/api/admin")

	// 创建 HTTP 响应记录器
	w := httptest.NewRecorder()

	// 写入错误响应
	gerror.WriteHTTPError(w, err)

	// 验证响应
	resp := w.Result()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// 关闭响应
	resp.Body.Close()
}

func TestRegisterHTTPStatus(t *testing.T) {
	// 自定义错误码
	customCode := gerror.Code(20001)

	// 注册自定义HTTP状态码
	gerror.RegisterHTTPStatus(customCode, http.StatusTeapot) // 418 I'm a teapot

	// 验证注册成功
	err := gerror.New(customCode, "我是一个茶壶")
	status := gerror.GetHTTPStatus(err)

	assert.Equal(t, http.StatusTeapot, status)
}
