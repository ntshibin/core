// Package gerror provides enhanced error handling capabilities.
package gerror

import (
	"database/sql"
	"fmt"
	"strings"
)

// 数据库相关的错误码
const (
	CodeDBNotFound     Code = 11000 // 记录未找到
	CodeDBDuplicate    Code = 11001 // 重复记录
	CodeDBConstraint   Code = 11002 // 约束冲突
	CodeDBConnection   Code = 11003 // 连接错误
	CodeDBTransaction  Code = 11004 // 事务错误
	CodeDBQuery        Code = 11005 // 查询错误
	CodeDBExecution    Code = 11006 // 执行错误
	CodeDBTimeout      Code = 11007 // 数据库超时
	CodeDBUnavailable  Code = 11008 // 数据库不可用
	CodeDBUnauthorized Code = 11009 // 数据库未授权
)

func init() {
	// 注册数据库错误描述文本
	RegisterCodeText(CodeDBNotFound, "记录未找到")
	RegisterCodeText(CodeDBDuplicate, "记录已存在")
	RegisterCodeText(CodeDBConstraint, "违反数据库约束")
	RegisterCodeText(CodeDBConnection, "数据库连接错误")
	RegisterCodeText(CodeDBTransaction, "数据库事务错误")
	RegisterCodeText(CodeDBQuery, "数据库查询错误")
	RegisterCodeText(CodeDBExecution, "数据库执行错误")
	RegisterCodeText(CodeDBTimeout, "数据库操作超时")
	RegisterCodeText(CodeDBUnavailable, "数据库不可用")
	RegisterCodeText(CodeDBUnauthorized, "数据库访问未授权")
}

// IsNotFound 检查错误是否是"未找到记录"类型的错误
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是我们的错误类型
	if Is(err, New(CodeDBNotFound, "")) || Is(err, New(CodeNotFound, "")) {
		return true
	}

	// 检查标准库的sql.ErrNoRows
	if err == sql.ErrNoRows {
		return true
	}

	// 检查错误消息
	errMsg := err.Error()
	return strings.Contains(strings.ToLower(errMsg), "not found") ||
		strings.Contains(strings.ToLower(errMsg), "no rows") ||
		strings.Contains(strings.ToLower(errMsg), "does not exist")
}

// IsDuplicate 检查错误是否是"重复记录"类型的错误
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是我们的错误类型
	if Is(err, New(CodeDBDuplicate, "")) {
		return true
	}

	// 检查错误消息
	errMsg := err.Error()
	return strings.Contains(strings.ToLower(errMsg), "duplicate") ||
		strings.Contains(strings.ToLower(errMsg), "already exists") ||
		strings.Contains(strings.ToLower(errMsg), "unique violation")
}

// IsConstraintViolation 检查错误是否是"违反约束"类型的错误
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是我们的错误类型
	if Is(err, New(CodeDBConstraint, "")) {
		return true
	}

	// 检查错误消息
	errMsg := err.Error()
	return strings.Contains(strings.ToLower(errMsg), "constraint") ||
		strings.Contains(strings.ToLower(errMsg), "violation") ||
		strings.Contains(strings.ToLower(errMsg), "foreign key") ||
		strings.Contains(strings.ToLower(errMsg), "check constraint")
}

// IsConnectionError 检查错误是否是数据库连接错误
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是我们的错误类型
	if Is(err, New(CodeDBConnection, "")) {
		return true
	}

	// 检查错误消息
	errMsg := err.Error()
	return strings.Contains(strings.ToLower(errMsg), "connection") ||
		strings.Contains(strings.ToLower(errMsg), "connect") ||
		strings.Contains(strings.ToLower(errMsg), "network") ||
		strings.Contains(strings.ToLower(errMsg), "unreachable")
}

// WrapDBError 包装数据库错误，根据错误类型选择合适的错误码
func WrapDBError(err error, message string) error {
	if err == nil {
		return nil
	}

	var code Code

	switch {
	case err == sql.ErrNoRows || IsNotFound(err):
		code = CodeDBNotFound
	case IsDuplicate(err):
		code = CodeDBDuplicate
	case IsConstraintViolation(err):
		code = CodeDBConstraint
	case IsConnectionError(err):
		code = CodeDBConnection
	case strings.Contains(strings.ToLower(err.Error()), "timeout"):
		code = CodeDBTimeout
	case strings.Contains(strings.ToLower(err.Error()), "transaction"):
		code = CodeDBTransaction
	case strings.Contains(strings.ToLower(err.Error()), "query"):
		code = CodeDBQuery
	case strings.Contains(strings.ToLower(err.Error()), "execute"):
		code = CodeDBExecution
	default:
		code = CodeDBQuery // 默认为查询错误
	}

	return Wrap(err, code, message)
}

// NotFoundError 创建一个"记录未找到"错误
func NotFoundError(entity string, id interface{}) error {
	msg := "记录未找到"
	if entity != "" {
		if id != nil {
			msg = fmt.Sprintf("%s(ID:%v)不存在", entity, id)
		} else {
			msg = fmt.Sprintf("%s不存在", entity)
		}
	}
	return New(CodeDBNotFound, msg)
}

// DuplicateError 创建一个"重复记录"错误
func DuplicateError(entity string, field string, value interface{}) error {
	msg := "记录已存在"
	if entity != "" {
		if field != "" && value != nil {
			msg = fmt.Sprintf("%s中已存在%s为%v的记录", entity, field, value)
		} else {
			msg = fmt.Sprintf("%s记录已存在", entity)
		}
	}
	return New(CodeDBDuplicate, msg)
}
