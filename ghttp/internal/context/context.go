// Package context 提供了HTTP请求的上下文处理
package context

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ntshibin/core/gerror"
)

// Context 是对 gin.Context 的封装
type Context struct {
	*gin.Context
}

// Result 表示API的返回结果
type Result struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 消息
	Data    interface{} `json:"data"`    // 数据
}

// Success 发送成功响应
func (c *Context) Success(data interface{}) {
	c.Context.JSON(http.StatusOK, Result{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// Fail 发送失败响应
func (c *Context) Fail(code int, message string) {
	c.Context.JSON(code, Result{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithCode 发送带错误码的失败响应
func (c *Context) ErrorWithCode(err error, httpStatus int) {
	var gErr *gerror.Error
	if gerror.As(err, &gErr) {
		c.Context.JSON(httpStatus, Result{
			Code:    int(gErr.Code()),
			Message: gErr.Message(),
			Data:    nil,
		})
	} else {
		c.Context.JSON(httpStatus, Result{
			Code:    httpStatus,
			Message: err.Error(),
			Data:    nil,
		})
	}
}

// Error 发送错误响应
func (c *Context) Error(err error) {
	c.ErrorWithCode(err, http.StatusInternalServerError)
}

// BadRequest 发送400错误响应
func (c *Context) BadRequest(err error) {
	c.ErrorWithCode(err, http.StatusBadRequest)
}

// Unauthorized 发送401错误响应
func (c *Context) Unauthorized(err error) {
	c.ErrorWithCode(err, http.StatusUnauthorized)
}

// Forbidden 发送403错误响应
func (c *Context) Forbidden(err error) {
	c.ErrorWithCode(err, http.StatusForbidden)
}

// NotFound 发送404错误响应
func (c *Context) NotFound(err error) {
	c.ErrorWithCode(err, http.StatusNotFound)
}
