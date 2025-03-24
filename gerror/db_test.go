package gerror_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/ntshibin/core/gerror"
	"github.com/stretchr/testify/assert"
)

func TestIsNotFound(t *testing.T) {
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
			name:     "sql.ErrNoRows",
			err:      sql.ErrNoRows,
			expected: true,
		},
		{
			name:     "gerror.CodeDBNotFound",
			err:      gerror.New(gerror.CodeDBNotFound, "记录未找到"),
			expected: true,
		},
		{
			name:     "gerror.CodeNotFound",
			err:      gerror.New(gerror.CodeNotFound, "资源不存在"),
			expected: true,
		},
		{
			name:     "not found消息",
			err:      errors.New("record not found"),
			expected: true,
		},
		{
			name:     "no rows消息",
			err:      errors.New("no rows in result set"),
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
			result := gerror.IsNotFound(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsDuplicate(t *testing.T) {
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
			name:     "gerror.CodeDBDuplicate",
			err:      gerror.New(gerror.CodeDBDuplicate, "记录已存在"),
			expected: true,
		},
		{
			name:     "duplicate消息",
			err:      errors.New("duplicate key value violates unique constraint"),
			expected: true,
		},
		{
			name:     "already exists消息",
			err:      errors.New("key already exists"),
			expected: true,
		},
		{
			name:     "unique violation消息",
			err:      errors.New("unique violation"),
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
			result := gerror.IsDuplicate(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsConstraintViolation(t *testing.T) {
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
			name:     "gerror.CodeDBConstraint",
			err:      gerror.New(gerror.CodeDBConstraint, "约束冲突"),
			expected: true,
		},
		{
			name:     "constraint消息",
			err:      errors.New("violates foreign key constraint"),
			expected: true,
		},
		{
			name:     "violation消息",
			err:      errors.New("integrity constraint violation"),
			expected: true,
		},
		{
			name:     "foreign key消息",
			err:      errors.New("foreign key violation"),
			expected: true,
		},
		{
			name:     "check constraint消息",
			err:      errors.New("check constraint failed"),
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
			result := gerror.IsConstraintViolation(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWrapDBError(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		message      string
		expectedCode gerror.Code
	}{
		{
			name:         "空错误",
			err:          nil,
			message:      "测试消息",
			expectedCode: 0, // 应该返回nil
		},
		{
			name:         "sql.ErrNoRows",
			err:          sql.ErrNoRows,
			message:      "查询用户失败",
			expectedCode: gerror.CodeDBNotFound,
		},
		{
			name:         "重复记录错误",
			err:          errors.New("duplicate key value"),
			message:      "创建用户失败",
			expectedCode: gerror.CodeDBDuplicate,
		},
		{
			name:         "约束错误",
			err:          errors.New("violates foreign key constraint"),
			message:      "创建订单失败",
			expectedCode: gerror.CodeDBConstraint,
		},
		{
			name:         "连接错误",
			err:          errors.New("connection refused"),
			message:      "连接数据库失败",
			expectedCode: gerror.CodeDBConnection,
		},
		{
			name:         "超时错误",
			err:          errors.New("query timeout"),
			message:      "查询超时",
			expectedCode: gerror.CodeDBTimeout,
		},
		{
			name:         "事务错误",
			err:          errors.New("transaction aborted"),
			message:      "事务失败",
			expectedCode: gerror.CodeDBTransaction,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrappedErr := gerror.WrapDBError(tc.err, tc.message)

			if tc.err == nil {
				assert.Nil(t, wrappedErr)
				return
			}

			assert.NotNil(t, wrappedErr)
			assert.Equal(t, tc.expectedCode, gerror.GetCode(wrappedErr))
			assert.Equal(t, tc.message, gerror.GetMessage(wrappedErr))
			assert.Contains(t, wrappedErr.Error(), tc.err.Error())
		})
	}
}

func TestNotFoundError(t *testing.T) {
	// 测试不带ID
	err1 := gerror.NotFoundError("用户", nil)
	assert.Equal(t, gerror.CodeDBNotFound, gerror.GetCode(err1))
	assert.Contains(t, gerror.GetMessage(err1), "用户不存在")

	// 测试带ID
	err2 := gerror.NotFoundError("订单", 12345)
	assert.Equal(t, gerror.CodeDBNotFound, gerror.GetCode(err2))
	assert.Contains(t, gerror.GetMessage(err2), "订单(ID:12345)不存在")

	// 测试空实体
	err3 := gerror.NotFoundError("", nil)
	assert.Equal(t, gerror.CodeDBNotFound, gerror.GetCode(err3))
	assert.Contains(t, gerror.GetMessage(err3), "记录未找到")
}

func TestDuplicateError(t *testing.T) {
	// 测试不带字段和值
	err1 := gerror.DuplicateError("用户", "", nil)
	assert.Equal(t, gerror.CodeDBDuplicate, gerror.GetCode(err1))
	assert.Contains(t, gerror.GetMessage(err1), "用户记录已存在")

	// 测试带字段和值
	err2 := gerror.DuplicateError("用户", "邮箱", "test@example.com")
	assert.Equal(t, gerror.CodeDBDuplicate, gerror.GetCode(err2))
	assert.Contains(t, gerror.GetMessage(err2), "用户中已存在邮箱为test@example.com的记录")

	// 测试空实体
	err3 := gerror.DuplicateError("", "", nil)
	assert.Equal(t, gerror.CodeDBDuplicate, gerror.GetCode(err3))
	assert.Contains(t, gerror.GetMessage(err3), "记录已存在")
}
