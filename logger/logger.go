package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// GetDefaultLogger 获取默认日志记录器
func GetDefaultLogger() LoggerInterface {
	return GetLogManager().loggers["default"]
}

// StandardLogger 标准日志记录器
type StandardLogger struct {
	name       string
	level      LogLevel
	handlers   []Handler
	fields     map[string]interface{}
	context    *LogContext
	mu         sync.RWMutex
	callerSkip int
}

// NewStandardLogger 创建标准日志记录器
func NewStandardLogger(name string, level LogLevel, handlers ...Handler) *StandardLogger {
	return &StandardLogger{
		name:       name,
		level:      level,
		handlers:   handlers,
		fields:     make(map[string]interface{}),
		context:    nil,
		callerSkip: 2,
	}
}

// SetLevel 设置日志级别
func (l *StandardLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 获取日志级别
func (l *StandardLogger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// AddHandler 添加处理器
func (l *StandardLogger) AddHandler(handler Handler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers = append(l.handlers, handler)
}

// RemoveHandler 移除处理器
func (l *StandardLogger) RemoveHandler(handler Handler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, h := range l.handlers {
		if h == handler {
			l.handlers = append(l.handlers[:i], l.handlers[i+1:]...)
			break
		}
	}
}

// Debug 输出Debug级别日志
func (l *StandardLogger) Debug(msg string) {
	l.log(DebugLevel, msg)
}

// Info 输出Info级别日志
func (l *StandardLogger) Info(msg string) {
	l.log(InfoLevel, msg)
}

// Warn 输出Warn级别日志
func (l *StandardLogger) Warn(msg string) {
	l.log(WarnLevel, msg)
}

// Error 输出Error级别日志
func (l *StandardLogger) Error(msg string) {
	l.log(ErrorLevel, msg)
}

// Fatal 输出Fatal级别日志
func (l *StandardLogger) Fatal(msg string) {
	l.log(FatalLevel, msg)
	os.Exit(1)
}

// log 处理日志记录
func (l *StandardLogger) log(level LogLevel, msg string) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	l.mu.RUnlock()

	// 创建日志事件
	event := LogEvent{
		Time:    time.Now().UnixNano(),
		Level:   level,
		Message: msg,
		Fields:  make(map[string]interface{}),
		Context: l.context,
		Logger:  l.name,
	}

	// 复制字段
	l.mu.RLock()
	for k, v := range l.fields {
		event.Fields[k] = v
	}
	l.mu.RUnlock()

	// 添加调用者信息
	if caller := l.getCaller(); caller != "" {
		event.Caller = caller
	}

	// 发送给所有处理器
	for _, handler := range l.handlers {
		_ = handler.Handle(event)
	}
}

// getCaller 获取调用者信息
func (l *StandardLogger) getCaller() string {
	_, file, line, ok := runtime.Caller(l.callerSkip)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// WithField 添加单个字段
func (l *StandardLogger) WithField(key string, value interface{}) LoggerInterface {
	return l.WithFields(map[string]interface{}{key: value})
}

// WithFields 添加多个字段
func (l *StandardLogger) WithFields(fields map[string]interface{}) LoggerInterface {
	newLogger := &StandardLogger{
		name:       l.name,
		level:      l.level,
		handlers:   l.handlers,
		fields:     make(map[string]interface{}),
		context:    l.context,
		callerSkip: l.callerSkip,
	}

	// 复制现有字段
	l.mu.RLock()
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	l.mu.RUnlock()

	// 添加新字段
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithContext 添加上下文
func (l *StandardLogger) WithContext(ctx context.Context) LoggerInterface {
	// 从上下文中获取日志上下文
	logCtx := LogContextFromContext(ctx)
	if logCtx == nil {
		logCtx = &LogContext{}
	}

	newLogger := &StandardLogger{
		name:       l.name,
		level:      l.level,
		handlers:   l.handlers,
		fields:     make(map[string]interface{}),
		context:    logCtx,
		callerSkip: l.callerSkip,
	}

	// 复制现有字段
	l.mu.RLock()
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	l.mu.RUnlock()

	return newLogger
}

// Sync 同步所有处理器
func (l *StandardLogger) Sync() error {
	var lastErr error
	for _, handler := range l.handlers {
		if h, ok := handler.(*AsyncHandler); ok {
			if err := h.Sync(); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

// Close 关闭所有处理器
func (l *StandardLogger) Close() error {
	var lastErr error
	for _, handler := range l.handlers {
		if err := handler.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// LogManager 日志管理器
type LogManager struct {
	loggers map[string]LoggerInterface
	factory LoggerFactory
	mu      sync.RWMutex
}

var (
	defaultManager *LogManager
	managerOnce    sync.Once
)

// GetLogManager 获取默认日志管理器
func GetLogManager() *LogManager {
	managerOnce.Do(func() {
		defaultManager = &LogManager{
			loggers: make(map[string]LoggerInterface),
			factory: &StandardLoggerFactory{
				defaultLevel:     InfoLevel,
				defaultFormatter: NewJSONFormatter(),
			},
		}

		// 创建默认日志记录器
		defaultManager.loggers["default"] = defaultManager.factory.CreateLogger("default")
	})
	return defaultManager
}

// CloseAll 关闭所有日志记录器
func (m *LogManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for _, logger := range m.loggers {
		if l, ok := logger.(*StandardLogger); ok {
			if err := l.Close(); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

// StandardLoggerFactory 标准日志工厂
type StandardLoggerFactory struct {
	defaultLevel     LogLevel
	defaultFormatter Formatter
	defaultHandlers  []Handler
	mu               sync.RWMutex
}

// NewStandardLoggerFactory 创建标准日志工厂
func NewStandardLoggerFactory() *StandardLoggerFactory {
	return &StandardLoggerFactory{
		defaultLevel:     InfoLevel,
		defaultFormatter: NewJSONFormatter(),
	}
}

// CreateLogger 创建日志记录器
func (f *StandardLoggerFactory) CreateLogger(name string) LoggerInterface {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var handlers []Handler
	if len(f.defaultHandlers) > 0 {
		handlers = f.defaultHandlers
	} else {
		handlers = []Handler{
			NewConsoleHandler(f.defaultFormatter, f.defaultLevel),
		}
	}

	return NewStandardLogger(name, f.defaultLevel, handlers...)
}
