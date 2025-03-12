package errorx

// 响应状态码常量
const (
	HTTPCodeSuccess         = 0    // 成功
	HTTPCodeError           = 1000 // 通用错误
	HTTPCodeParamError      = 1001 // 参数错误
	HTTPCodeUnauthorized    = 1002 // 未授权
	HTTPCodeForbidden       = 1003 // 禁止访问
	HTTPCodeNotFound        = 1004 // 资源不存在
	HTTPCodeInternalError   = 1005 // 内部错误
	HTTPCodeValidationError = 1006 // 数据验证错误
	HTTPCodeDuplicateError  = 1007 // 数据重复错误
	HTTPCodeTimeoutError    = 1008 // 超时错误
	HTTPCodeThirdPartyError = 1009 // 第三方服务错误
)

// 预定义错误
var (
	// 通用错误
	ErrInternalServer = New(HTTPCodeInternalError, "内部服务错误")
	ErrParamInvalid   = New(HTTPCodeParamError, "参数无效")
	ErrUnauthorized   = New(HTTPCodeUnauthorized, "未授权访问")
	ErrForbidden      = New(HTTPCodeForbidden, "禁止访问")
	ErrNotFound       = New(HTTPCodeNotFound, "资源不存在")
	ErrTimeout        = New(HTTPCodeTimeoutError, "请求超时")

	// 业务错误
	ErrValidation        = New(HTTPCodeValidationError, "数据验证失败")
	ErrDuplicate         = New(HTTPCodeDuplicateError, "数据重复")
	ErrThirdParty        = New(HTTPCodeThirdPartyError, "第三方服务错误")
	ErrSendFailed        = New(HTTPCodeThirdPartyError, "发送通知失败")
	ErrInvalidConfig     = New(HTTPCodeParamError, "无效的驱动配置")
	ErrInvalidMessageType = New(HTTPCodeParamError, "无效的消息类型")
)
