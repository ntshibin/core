package logger

import (
	"context"
	"io"
	"sync"
)

// 全局生命周期管理
var (
	lifecycleOnce sync.Once
	shutdownFuncs []func() error
)

// Debug 输出Debug级别日志
func Debug(msg string) {
	GetDefaultLogger().Debug(msg)
}

// Info 输出Info级别日志
func Info(msg string) {
	GetDefaultLogger().Info(msg)
}

// Warn 输出Warn级别日志
func Warn(msg string) {
	GetDefaultLogger().Warn(msg)
}

// Error 输出Error级别日志
func Error(msg string) {
	GetDefaultLogger().Error(msg)
}

// Fatal 输出Fatal级别日志
func Fatal(msg string) {
	GetDefaultLogger().Fatal(msg)
}

// WithField 添加单个字段
func WithField(key string, value interface{}) LoggerInterface {
	return GetDefaultLogger().WithField(key, value)
}

// WithFields 添加多个字段
func WithFields(fields map[string]interface{}) LoggerInterface {
	return GetDefaultLogger().WithFields(fields)
}

// SetLevel 设置全局日志级别
func SetLevel(level LogLevel) {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		logger.SetLevel(level)
	}
}

// GetLevel 获取全局日志级别
func GetLevel() LogLevel {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		return logger.GetLevel()
	}
	return InfoLevel
}

// Sync 同步日志
func Sync() error {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		return logger.Sync()
	}
	return nil
}

// InitLifecycle 初始化生命周期管理
func InitLifecycle() {
	lifecycleOnce.Do(func() {
		// 注册关闭函数
		shutdownFuncs = append(shutdownFuncs, func() error {
			return GetLogManager().CloseAll()
		})

		// 处理信号
		go func() {
			// 等待信号或退出信号
			// 在实际应用中，这里应该捕获操作系统信号
			// 省略信号处理代码...
		}()
	})
}

// AddConsoleHandler 添加控制台处理器
func AddConsoleHandler(level LogLevel) {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		handler := NewConsoleHandler(NewTextFormatter(), level)
		logger.AddHandler(handler)
	}
}

// AddFileHandler 添加文件处理器
func AddFileHandler(filePath string, level LogLevel) error {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		handler, err := NewFileHandler(NewJSONFormatter(), level, filePath)
		if err != nil {
			return err
		}
		logger.AddHandler(handler)
	}
	return nil
}

// AddRotateFileHandler 添加轮转文件处理器
func AddRotateFileHandler(config FileRotateConfig, level LogLevel) error {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		handler, err := NewRotateFileHandler(NewJSONFormatter(), level, config)
		if err != nil {
			return err
		}
		logger.AddHandler(handler)
	}
	return nil
}

// AddWriter 添加自定义输出
func AddWriter(writer io.Writer, level LogLevel) {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		customHandler := &CustomHandler{
			BaseHandler: NewBaseHandler(NewTextFormatter(), level),
			writer:      writer,
		}
		logger.AddHandler(customHandler)
	}
}

// CustomHandler 自定义处理器
type CustomHandler struct {
	*BaseHandler
	writer io.Writer
}

// Handle 处理日志事件
func (h *CustomHandler) Handle(event LogEvent) error {
	if !h.ShouldHandle(event) {
		return nil
	}

	data, err := h.Format(event)
	if err != nil {
		return err
	}

	_, err = h.writer.Write(data)
	return err
}

// Close 关闭处理器
func (h *CustomHandler) Close() error {
	if closer, ok := h.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// AddAsyncHandler 添加异步处理器
func AddAsyncHandler(handler Handler, queueSize int) error {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		asyncHandler := NewAsyncHandler(handler, queueSize)
		logger.AddHandler(asyncHandler)
	}
	return nil
}

// AddRemoteHandler 添加远程日志处理器
func AddRemoteHandler(config RemoteConfig, level LogLevel) error {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		handler, err := NewRemoteHandler(NewJSONFormatter(), level, config)
		if err != nil {
			return err
		}
		logger.AddHandler(handler)
	}
	return nil
}

// CreateMemoryLogger 创建内存日志记录器和查询API
func CreateMemoryLogger(config MemoryConfig, level LogLevel) (*MemoryHandlerAPI, error) {
	handler := NewMemoryHandler(NewJSONFormatter(), level, config)

	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		logger.AddHandler(handler)
	}

	return NewMemoryHandlerAPI(handler), nil
}

// WithContext 将上下文添加到日志记录器
func WithContext(ctx context.Context) LoggerInterface {
	if logger, ok := GetDefaultLogger().(*StandardLogger); ok {
		return logger.WithContext(ctx)
	}
	return GetDefaultLogger()
}

// LoggerFromContext 从上下文获取日志记录器
func LoggerFromContext(ctx context.Context) LoggerInterface {
	if ctx == nil {
		return GetDefaultLogger()
	}

	// 获取上下文中的日志上下文
	logCtx := LogContextFromContext(ctx)
	if logCtx != nil {
		return GetDefaultLogger().WithContext(ctx)
	}

	return GetDefaultLogger()
}

// CreateTraceContext 创建带有追踪的上下文
func CreateTraceContext() context.Context {
	logCtx := NewContext().WithTrace(generateID())
	return WithLogContext(context.Background(), logCtx)
}

// StartLogSpan 开始一个日志跨度
func StartLogSpan(ctx context.Context, name string) (context.Context, LoggerInterface) {
	newCtx, _ := StartSpan(ctx, name)
	logger := GetDefaultLogger().(*StandardLogger).WithContext(newCtx)
	logger.Info("Started " + name)
	return newCtx, logger
}

// FinishLogSpan 结束一个日志跨度
func FinishLogSpan(ctx context.Context) {
	FinishSpan(ctx)
}

// SetAsyncLogging 设置是否使用异步日志
// enable=true 启用全异步模式，所有处理器都使用异步写入
// enable=false 启用同步模式，所有处理器都使用同步写入
func SetAsyncLogging(enable bool) {
	if enable {
		EnableAsyncMode()
	} else {
		EnableSyncMode()
	}
}

// SetMixedLogging 启用混合日志模式
// 在此模式下，文件和网络I/O使用异步，控制台使用同步
func SetMixedLogging() {
	EnableMixedMode()
}

// ConfigureAsync 配置异步日志
// mode: 异步模式 (0=同步, 1=异步, 2=混合)
// queueSize: 异步队列大小
func ConfigureAsync(mode int, queueSize int) {
	// 设置队列大小
	SetAsyncQueueSize(queueSize)

	// 设置模式
	switch mode {
	case 0:
		EnableSyncMode()
	case 1:
		EnableAsyncMode()
	case 2:
		EnableMixedMode()
	default:
		EnableMixedMode() // 默认使用混合模式
	}
}
