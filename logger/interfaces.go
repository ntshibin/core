package logger

import (
	"context"
)

// LogLevel 日志级别
type LogLevel int

// 日志级别常量
const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// LogEvent 日志事件
type LogEvent struct {
	Time    int64                  // 时间戳
	Level   LogLevel               // 日志级别
	Message string                 // 日志消息
	Fields  map[string]interface{} // 额外字段
	Caller  string                 // 调用者信息
	Context *LogContext            // 上下文信息
	Logger  string                 // 日志记录器名称
}

// LogContext 日志上下文
type LogContext struct {
	TraceID  string            // 追踪ID
	SpanID   string            // 跨度ID
	ParentID string            // 父跨度ID
	Tags     map[string]string // 上下文标签
}

// LoggerInterface 日志记录器接口
type LoggerInterface interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)

	WithField(key string, value interface{}) LoggerInterface
	WithFields(fields map[string]interface{}) LoggerInterface
	WithContext(ctx context.Context) LoggerInterface

	SetLevel(level LogLevel)
	GetLevel() LogLevel
	Sync() error
}

// Handler 日志处理器接口
type Handler interface {
	Handle(event LogEvent) error
	Format(event LogEvent) ([]byte, error)
	ShouldHandle(event LogEvent) bool
	Close() error
}

// Formatter 日志格式化器接口
type Formatter interface {
	Format(event LogEvent) ([]byte, error)
}

// LoggerFactory 日志工厂接口
type LoggerFactory interface {
	CreateLogger(name string) LoggerInterface
}

// BuilderInterface 日志建造者接口
type BuilderInterface interface {
	SetLevel(level LogLevel) BuilderInterface
	AddHandler(handler Handler) BuilderInterface
	Build() (LoggerInterface, error)
}
