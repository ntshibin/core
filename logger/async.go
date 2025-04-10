package logger

import (
	"sync"
)

// LogAsyncMode 定义异步模式枚举
type LogAsyncMode int

const (
	// SyncMode 同步模式 - 日志直接写入目标
	SyncMode LogAsyncMode = iota
	// AsyncMode 异步模式 - 日志通过队列异步写入
	AsyncMode
	// MixedMode 混合模式 - 根据配置有选择地使用异步
	MixedMode
)

// asyncConfig 全局异步配置
var asyncConfig struct {
	// 当前异步模式
	mode LogAsyncMode
	// 异步队列大小
	queueSize int
	// 异步处理器映射
	asyncHandlers map[Handler]*AsyncHandler
	// 原始处理器映射
	originalHandlers map[*AsyncHandler]Handler
	// 互斥锁
	mu sync.RWMutex
}

// 初始化异步配置
func init() {
	asyncConfig.mode = MixedMode // 默认为混合模式
	asyncConfig.queueSize = 1000 // 默认队列大小
	asyncConfig.asyncHandlers = make(map[Handler]*AsyncHandler)
	asyncConfig.originalHandlers = make(map[*AsyncHandler]Handler)
}

// SetAsyncMode 设置异步模式
func SetAsyncMode(mode LogAsyncMode) {
	asyncConfig.mu.Lock()
	defer asyncConfig.mu.Unlock()

	// 如果模式没有变化，不做处理
	if asyncConfig.mode == mode {
		return
	}

	// 保存新模式
	asyncConfig.mode = mode

	// 获取默认日志记录器
	logger, ok := GetDefaultLogger().(*StandardLogger)
	if !ok {
		return
	}

	// 应用新模式到所有处理器
	logger.mu.Lock()
	defer logger.mu.Unlock()

	// 根据新模式调整处理器
	var newHandlers []Handler
	for _, handler := range logger.handlers {
		newHandler := applyAsyncMode(handler, mode)
		newHandlers = append(newHandlers, newHandler)
	}

	// 替换处理器
	logger.handlers = newHandlers
}

// SetAsyncQueueSize 设置异步队列大小
func SetAsyncQueueSize(size int) {
	if size <= 0 {
		size = 1000
	}

	asyncConfig.mu.Lock()
	asyncConfig.queueSize = size
	asyncConfig.mu.Unlock()
}

// GetAsyncMode 获取当前异步模式
func GetAsyncMode() LogAsyncMode {
	asyncConfig.mu.RLock()
	defer asyncConfig.mu.RUnlock()
	return asyncConfig.mode
}

// applyAsyncMode 根据指定的异步模式处理处理器
func applyAsyncMode(handler Handler, mode LogAsyncMode) Handler {
	asyncConfig.mu.Lock()
	defer asyncConfig.mu.Unlock()

	// 如果是异步处理器，首先获取原始处理器
	if asyncHandler, ok := handler.(*AsyncHandler); ok {
		if originalHandler, exists := asyncConfig.originalHandlers[asyncHandler]; exists {
			handler = originalHandler
			delete(asyncConfig.asyncHandlers, originalHandler)
			delete(asyncConfig.originalHandlers, asyncHandler)
		}
	}

	// 根据模式决定是否需要包装为异步处理器
	switch mode {
	case AsyncMode:
		// 如果已经是异步处理器，直接返回
		if _, ok := handler.(*AsyncHandler); ok {
			return handler
		}

		// 创建新的异步处理器
		asyncHandler := NewAsyncHandler(handler, asyncConfig.queueSize)
		asyncConfig.asyncHandlers[handler] = asyncHandler
		asyncConfig.originalHandlers[asyncHandler] = handler
		return asyncHandler

	case SyncMode:
		// 同步模式，直接返回原始处理器
		return handler

	case MixedMode:
		// 混合模式，根据处理器类型决定
		// 对于文件、网络等IO密集型处理器，使用异步模式
		// 对于控制台等处理器，使用同步模式
		switch handler.(type) {
		case *FileHandler, *RotateFileHandler, *RemoteHandler:
			if _, ok := handler.(*AsyncHandler); ok {
				return handler
			}

			asyncHandler := NewAsyncHandler(handler, asyncConfig.queueSize)
			asyncConfig.asyncHandlers[handler] = asyncHandler
			asyncConfig.originalHandlers[asyncHandler] = handler
			return asyncHandler

		default:
			return handler
		}
	}

	return handler
}

// EnableAsyncMode 启用异步模式
func EnableAsyncMode() {
	SetAsyncMode(AsyncMode)
}

// EnableSyncMode 启用同步模式
func EnableSyncMode() {
	SetAsyncMode(SyncMode)
}

// EnableMixedMode 启用混合模式
func EnableMixedMode() {
	SetAsyncMode(MixedMode)
}

// IsAsyncMode 检查是否为异步模式
func IsAsyncMode() bool {
	return GetAsyncMode() == AsyncMode
}

// IsSyncMode 检查是否为同步模式
func IsSyncMode() bool {
	return GetAsyncMode() == SyncMode
}

// IsMixedMode 检查是否为混合模式
func IsMixedMode() bool {
	return GetAsyncMode() == MixedMode
}
