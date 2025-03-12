package errorx

// GetErrorCode 从错误中获取错误码
// 如果错误是Error类型，返回其错误码
// 如果错误是其他类型，返回通用错误码
func GetErrorCode(err error) int {
	if err == nil {
		return HTTPCodeSuccess
	}

	// 尝试将错误转换为Error类型
	if e, ok := err.(*Error); ok {
		return e.Code()
	}

	// 非Error类型错误返回通用错误码
	return HTTPCodeError
}
