// Package gerror provides enhanced error handling capabilities.
package gerror

import (
	"encoding/json"
	"net/http"
)

// HTTP状态码映射
var codeToHTTPStatus = map[Code]int{
	CodeUnknown:      http.StatusInternalServerError,
	CodeInternal:     http.StatusInternalServerError,
	CodeInvalidParam: http.StatusBadRequest,
	CodeUnauthorized: http.StatusUnauthorized,
	CodeForbidden:    http.StatusForbidden,
	CodeNotFound:     http.StatusNotFound,
	CodeTimeout:      http.StatusGatewayTimeout,
	CodeConflict:     http.StatusConflict,
	CodeExhausted:    http.StatusTooManyRequests,
}

// HTTPResponse 表示HTTP错误响应
type HTTPResponse struct {
	Code    int                    `json:"code"`              // 错误码
	Message string                 `json:"message"`           // 错误消息
	Details map[string]interface{} `json:"details,omitempty"` // 详细信息
}

// GetHTTPStatus 获取与错误相对应的HTTP状态码
func GetHTTPStatus(err error) int {
	code := GetCode(err)
	if status, ok := codeToHTTPStatus[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// GetHTTPResponse 生成适合HTTP响应的错误信息
func GetHTTPResponse(err error) HTTPResponse {
	code := GetCode(err)
	message := GetMessage(err)

	details := make(map[string]interface{})

	// 添加上下文信息到详细信息中
	if ctx := GetContext(err); ctx != nil {
		for k, v := range ctx {
			details[k] = v
		}
	}

	// 开发环境可以添加堆栈信息
	// 生产环境通常不应该返回堆栈信息给客户端

	return HTTPResponse{
		Code:    int(code),
		Message: message,
		Details: details,
	}
}

// WriteHTTPError 将错误写入HTTP响应
func WriteHTTPError(w http.ResponseWriter, err error) {
	status := GetHTTPStatus(err)
	resp := GetHTTPResponse(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// 序列化响应
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// 如果序列化失败，返回简单错误
		http.Error(w, `{"code":10001,"message":"内部错误"}`, http.StatusInternalServerError)
	}
}

// RegisterHTTPStatus 注册自定义错误码对应的HTTP状态码
func RegisterHTTPStatus(code Code, httpStatus int) {
	codeToHTTPStatus[code] = httpStatus
}
