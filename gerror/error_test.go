package gerror_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ntshibin/core/gerror"
	"github.com/stretchr/testify/assert"
)

// TestNew 测试创建新错误
func TestNew(t *testing.T) {
	// 测试基本错误创建
	err := gerror.New(gerror.CodeInvalidParam, "无效参数")
	assert.NotNil(t, err)
	assert.Equal(t, gerror.CodeInvalidParam, gerror.GetCode(err))
	assert.Equal(t, "无效参数", gerror.GetMessage(err))
	assert.Contains(t, err.Error(), "无效参数")
	assert.Contains(t, err.Error(), "[10002]") // CodeInvalidParam = 10002

	// 测试格式化错误创建
	err = gerror.Newf(gerror.CodeNotFound, "未找到ID为%d的用户", 123)
	assert.NotNil(t, err)
	assert.Equal(t, gerror.CodeNotFound, gerror.GetCode(err))
	assert.Equal(t, "未找到ID为123的用户", gerror.GetMessage(err))
	assert.Contains(t, err.Error(), "未找到ID为123的用户")
}

// TestWrap 测试错误包装
func TestWrap(t *testing.T) {
	// 创建原始错误
	originalErr := errors.New("原始错误")

	// 包装错误
	wrappedErr := gerror.Wrap(originalErr, gerror.CodeInternal, "内部处理失败")
	assert.NotNil(t, wrappedErr)
	assert.Equal(t, gerror.CodeInternal, gerror.GetCode(wrappedErr))
	assert.Equal(t, "内部处理失败", gerror.GetMessage(wrappedErr))
	assert.Contains(t, wrappedErr.Error(), "内部处理失败")
	assert.Contains(t, wrappedErr.Error(), "原始错误")

	// 测试格式化包装
	wrappedErr = gerror.Wrapf(originalErr, gerror.CodeTimeout, "操作超时，尝试次数: %d", 3)
	assert.NotNil(t, wrappedErr)
	assert.Equal(t, gerror.CodeTimeout, gerror.GetCode(wrappedErr))
	assert.Equal(t, "操作超时，尝试次数: 3", gerror.GetMessage(wrappedErr))
	assert.Contains(t, wrappedErr.Error(), "操作超时，尝试次数: 3")
	assert.Contains(t, wrappedErr.Error(), "原始错误")

	// 包装nil错误应该返回nil
	assert.Nil(t, gerror.Wrap(nil, gerror.CodeInternal, "不应该出现"))
	assert.Nil(t, gerror.Wrapf(nil, gerror.CodeInternal, "不应该出现"))
}

// TestWithStack 测试为普通错误添加堆栈跟踪
func TestWithStack(t *testing.T) {
	// 创建普通错误
	originalErr := errors.New("普通错误")

	// 添加堆栈跟踪
	stackErr := gerror.WithStack(originalErr)
	assert.NotNil(t, stackErr)
	assert.Equal(t, gerror.CodeUnknown, gerror.GetCode(stackErr))
	assert.Equal(t, "普通错误", gerror.GetMessage(stackErr))

	// 格式化错误，查看是否有堆栈信息
	formatted := gerror.FormatError(stackErr, "text")
	assert.Contains(t, formatted, "普通错误")
	assert.Contains(t, formatted, "堆栈:")
	assert.Contains(t, formatted, "error_test.go") // 应包含当前文件名

	// 对nil错误调用WithStack应返回nil
	assert.Nil(t, gerror.WithStack(nil))

	// 对已经有堆栈的错误调用WithStack应该返回原错误
	stackErr2 := gerror.WithStack(stackErr)
	assert.Equal(t, stackErr, stackErr2)
}

// TestWithContext 测试为错误添加上下文信息
func TestWithContext(t *testing.T) {
	// 创建基本错误
	err := gerror.New(gerror.CodeInvalidParam, "参数验证失败")

	// 添加单个上下文
	err = gerror.WithContext(err, "参数名", "用户ID")
	assert.NotNil(t, err)

	// 验证上下文是否被添加
	ctx := gerror.GetContext(err)
	assert.NotNil(t, ctx)
	assert.Equal(t, "用户ID", ctx["参数名"])

	// 添加更多上下文
	err = gerror.WithContext(err, "期望值", "正整数")
	err = gerror.WithContext(err, "实际值", -1)

	// 验证多个上下文
	ctx = gerror.GetContext(err)
	assert.Equal(t, "用户ID", ctx["参数名"])
	assert.Equal(t, "正整数", ctx["期望值"])
	assert.Equal(t, -1, ctx["实际值"])

	// 使用map添加上下文
	err = gerror.WithContextMap(err, map[string]interface{}{
		"请求ID": "req-123",
		"用户IP": "192.168.1.1",
	})

	// 验证map添加的上下文
	ctx = gerror.GetContext(err)
	assert.Equal(t, "req-123", ctx["请求ID"])
	assert.Equal(t, "192.168.1.1", ctx["用户IP"])
	assert.Equal(t, "用户ID", ctx["参数名"]) // 确保原始上下文保留

	// 对nil错误调用WithContext应返回nil
	assert.Nil(t, gerror.WithContext(nil, "key", "value"))
	assert.Nil(t, gerror.WithContextMap(nil, map[string]interface{}{"key": "value"}))
}

// TestErrorFormat 测试错误格式化
func TestErrorFormat(t *testing.T) {
	// 创建一个带堆栈和上下文的错误
	originalErr := errors.New("数据库查询失败")
	err := gerror.Wrap(originalErr, gerror.CodeInternal, "处理用户请求时出错")
	err = gerror.WithContext(err, "用户ID", 42)
	err = gerror.WithContext(err, "请求路径", "/api/users")

	// 测试文本格式
	textFormat := gerror.FormatError(err, "text")
	assert.Contains(t, textFormat, "错误码: 10001")
	assert.Contains(t, textFormat, "内部错误") // CodeInternal的文本
	assert.Contains(t, textFormat, "处理用户请求时出错")
	assert.Contains(t, textFormat, "用户ID: 42")
	assert.Contains(t, textFormat, "请求路径: /api/users")
	assert.Contains(t, textFormat, "堆栈:")
	assert.Contains(t, textFormat, "error_test.go")
	assert.Contains(t, textFormat, "原因: 数据库查询失败")

	// 测试JSON格式
	jsonFormat := gerror.FormatError(err, "json")
	assert.Contains(t, jsonFormat, `"code": 10001`)
	assert.Contains(t, jsonFormat, `"code_text": "内部错误"`)
	assert.Contains(t, jsonFormat, `"message": "处理用户请求时出错"`)
	assert.Contains(t, jsonFormat, `"用户ID": 42`)
	assert.Contains(t, jsonFormat, `"请求路径": "/api/users"`)
	assert.Contains(t, jsonFormat, `"stack":`)
	assert.Contains(t, jsonFormat, `"file": "error_test.go"`)
	assert.Contains(t, jsonFormat, `"cause": "数据库查询失败"`)

	// 对普通错误进行格式化
	plainErr := errors.New("普通错误")
	plainFormat := gerror.FormatError(plainErr, "text")
	assert.Equal(t, "普通错误", plainFormat)

	// 对nil错误进行格式化
	nilFormat := gerror.FormatError(nil, "text")
	assert.Equal(t, "", nilFormat)
}

// TestIs 测试错误类型比较
func TestIs(t *testing.T) {
	// 创建两个错误
	err1 := gerror.New(gerror.CodeNotFound, "用户不存在")
	err2 := gerror.New(gerror.CodeNotFound, "产品不存在")
	err3 := gerror.New(gerror.CodeInvalidParam, "参数无效")

	// 相同错误码的错误应该匹配
	assert.True(t, gerror.Is(err1, err2))

	// 不同错误码的错误不应该匹配
	assert.False(t, gerror.Is(err1, err3))

	// 包装后的错误应该能识别原始错误码
	wrappedErr := gerror.Wrap(err1, gerror.CodeUnknown, "包装错误")
	assert.True(t, gerror.Is(wrappedErr, err1))
}

// TestRootCause 测试获取错误链的根本原因
func TestRootCause(t *testing.T) {
	// 创建错误链
	rootErr := errors.New("原始数据库错误")
	dbErr := gerror.Wrap(rootErr, gerror.CodeDBConnection, "数据库连接失败")
	svcErr := gerror.Wrap(dbErr, gerror.CodeInternal, "服务处理失败")
	apiErr := gerror.Wrap(svcErr, gerror.CodeInvalidParam, "API请求处理失败")

	// 获取根本原因
	cause := gerror.RootCause(apiErr)
	assert.NotNil(t, cause)
	assert.Equal(t, rootErr.Error(), cause.Error())
}

// TestCodeText 测试错误码文本
func TestCodeText(t *testing.T) {
	// 测试预定义错误码文本
	assert.Equal(t, "内部错误", gerror.CodeInternal.Text())
	assert.Equal(t, "参数错误", gerror.CodeInvalidParam.Text())
	assert.Equal(t, "资源不存在", gerror.CodeNotFound.Text())

	// 测试自定义错误码
	customCode := gerror.Code(99999)
	assert.Equal(t, "错误码(99999)", customCode.Text())

	// 注册自定义错误码文本
	gerror.RegisterCodeText(customCode, "自定义错误")
	assert.Equal(t, "自定义错误", customCode.Text())
}

// TestJoin 测试错误合并
func TestJoin(t *testing.T) {
	// 创建多个错误
	err1 := gerror.New(gerror.CodeInvalidParam, "参数1无效")
	err2 := gerror.New(gerror.CodeInvalidParam, "参数2无效")
	err3 := gerror.New(gerror.CodeInvalidParam, "参数3无效")

	// 合并错误
	joined := gerror.Join(err1, err2, err3)
	assert.NotNil(t, joined)

	// 检查合并的错误是否包含所有原始错误信息
	errStr := joined.Error()
	assert.Contains(t, errStr, "参数1无效")
	assert.Contains(t, errStr, "参数2无效")
	assert.Contains(t, errStr, "参数3无效")

	// 合并包含nil的错误列表
	joined = gerror.Join(err1, nil, err3)
	assert.NotNil(t, joined)
	errStr = joined.Error()
	assert.Contains(t, errStr, "参数1无效")
	assert.Contains(t, errStr, "参数3无效")

	// 合并全为nil的错误列表
	joined = gerror.Join(nil, nil)
	assert.Nil(t, joined)

	// 空列表
	joined = gerror.Join()
	assert.Nil(t, joined)
}

// TestStackTrace 测试堆栈跟踪功能
func TestStackTrace(t *testing.T) {
	// 创建一个带堆栈的错误
	err := gerror.New(gerror.CodeInternal, "内部错误")

	// 获取错误的堆栈跟踪
	var gerr *gerror.Error
	assert.True(t, gerror.As(err, &gerr))

	// 验证堆栈信息
	stackInfo := gerr.StackTrace()
	assert.NotEmpty(t, stackInfo)

	// 检查堆栈帧信息
	found := false
	for _, frame := range stackInfo {
		// 应该包含当前测试函数的信息
		if strings.Contains(frame.Function, "TestStackTrace") {
			found = true
			assert.Contains(t, frame.File, "error_test.go")
			break
		}
	}
	assert.True(t, found, "堆栈应包含当前测试函数")
}

// TestWriteStackTrace 测试将堆栈跟踪写入io.Writer
func TestWriteStackTrace(t *testing.T) {
	// 创建一个带堆栈的错误
	err := gerror.New(gerror.CodeInternal, "内部错误")

	// 写入到字符串
	var sb strings.Builder
	gerror.WriteStackTrace(err, &sb)

	// 验证写入的结果
	output := sb.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "error_test.go")
	assert.Contains(t, output, "TestWriteStackTrace")
}

// TestChain 测试错误处理链
func TestChain(t *testing.T) {
	// 创建错误处理函数
	addUserID := func(err error) error {
		if err == nil {
			return nil
		}
		return gerror.WithContext(err, "用户ID", 42)
	}

	addRequestID := func(err error) error {
		if err == nil {
			return nil
		}
		return gerror.WithContext(err, "请求ID", "req-123")
	}

	// 链接错误处理函数
	chain := gerror.Chain(addUserID, addRequestID)

	// 创建一个初始错误
	err := gerror.New(gerror.CodeInvalidParam, "参数无效")

	// 应用链
	processed := chain(err)

	// 验证结果
	ctx := gerror.GetContext(processed)
	assert.Equal(t, 42, ctx["用户ID"])
	assert.Equal(t, "req-123", ctx["请求ID"])
}

// TestCause 测试获取原始错误
func TestCause(t *testing.T) {
	// 创建原始错误
	originalErr := errors.New("原始错误")

	// 包装错误
	wrappedErr := gerror.Wrap(originalErr, gerror.CodeInternal, "内部处理失败")

	// 类型断言为 gerror.Error 类型
	var gerr *gerror.Error
	assert.True(t, gerror.As(wrappedErr, &gerr))

	// 测试 Cause 方法
	cause := gerr.Cause()
	assert.Equal(t, originalErr, cause)

	// 测试多层包装的错误
	err1 := errors.New("错误1")
	err2 := gerror.Wrap(err1, gerror.CodeInternal, "错误2")
	err3 := gerror.Wrap(err2, gerror.CodeInternal, "错误3")

	cause = gerror.RootCause(err3)
	assert.Equal(t, err1, cause)

	// 测试对 nil 错误的处理
	assert.Nil(t, gerror.RootCause(nil))
}

// TestIsConnectionError 测试数据库连接错误识别
func TestIsConnectionError(t *testing.T) {
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
			name:     "gerror.CodeDBConnection",
			err:      gerror.New(gerror.CodeDBConnection, "数据库连接失败"),
			expected: true,
		},
		{
			name:     "connection消息",
			err:      errors.New("failed to connect to database"),
			expected: true,
		},
		{
			name:     "connect消息",
			err:      errors.New("could not connect to server"),
			expected: true,
		},
		{
			name:     "refused消息",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "其他错误",
			err:      errors.New("查询失败"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := gerror.IsConnectionError(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestEdgeCases 测试各种边缘情况
func TestEdgeCases(t *testing.T) {
	// 测试对 nil 错误的 GetCode
	assert.Equal(t, gerror.CodeUnknown, gerror.GetCode(nil))

	// 测试对 nil 错误的 GetMessage
	assert.Equal(t, "", gerror.GetMessage(nil))

	// 测试对 nil 错误的 GetContext
	assert.Nil(t, gerror.GetContext(nil))

	// 测试非 gerror.Error 类型的 WriteStackTrace
	var sb strings.Builder
	plainErr := errors.New("普通错误")
	gerror.WriteStackTrace(plainErr, &sb)
	assert.Contains(t, sb.String(), "错误不包含堆栈信息")

	// 测试 formatJSON 错误处理（这个很难直接测试，因为需要造成 json.Marshal 失败）

	// 测试错误的基础方法
	basicErr := gerror.New(gerror.CodeInternal, "测试错误")
	var gerr *gerror.Error
	assert.True(t, gerror.As(basicErr, &gerr))

	// 确保 Context 和 Unwrap 方法正确工作
	ctx := gerr.Context()
	assert.NotNil(t, ctx)
	assert.Empty(t, ctx)

	// 确保 WithContextMap 处理边缘情况
	nilMap := map[string]interface{}(nil)
	err := gerror.WithContextMap(basicErr, nilMap)
	assert.Equal(t, gerror.GetCode(basicErr), gerror.GetCode(err))

	// 空 map
	emptyMap := map[string]interface{}{}
	err = gerror.WithContextMap(basicErr, emptyMap)
	assert.Equal(t, gerror.GetCode(basicErr), gerror.GetCode(err))

	// 测试 WriteHTTPError 用一个非 JSON 格式的请求
	w := httptest.NewRecorder()
	gerror.WriteHTTPError(w, basicErr)
	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}
