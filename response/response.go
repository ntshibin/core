package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ntshibin/core/errorx"
)

// Response 通用响应结构
// Code: 响应状态码，0表示成功，非0表示错误
// Message: 响应消息，成功时为"success"，错误时为具体错误信息
// Data: 响应数据，可以是任意类型
type Response struct {
	Code    int         `json:"code"`               // 状态码
	Message string      `json:"message"`            // 响应消息
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	TraceID string      `json:"trace_id,omitempty"` // 请求追踪ID
}

// Success 返回成功响应
// ctx: gin上下文
// data: 响应数据
func Success(ctx *gin.Context, data interface{}) {
	traceID := ctx.GetString("trace_id")
	ctx.JSON(http.StatusOK, &Response{
		Code:    errorx.HTTPCodeSuccess,
		Message: "success",
		Data:    data,
		TraceID: traceID,
	})
}

// Error 返回错误响应
// ctx: gin上下文
// err: 错误信息
func Error(ctx *gin.Context, err error) {
	traceID := ctx.GetString("trace_id")
	ctx.JSON(http.StatusOK, &Response{
		Code:    errorx.GetErrorCode(err),
		Message: err.Error(),
		TraceID: traceID,
	})
}
