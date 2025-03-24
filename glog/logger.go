// Package glog provides an enhanced logging system built on top of logrus
// with features like handler chains, multiple output targets, and structured logging.
package glog

import (
	"fmt"
	"io"
	"sync"

	"github.com/ntshibin/core/glog/handlers"
	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger and provides additional functionality
// through a chain of responsibility pattern for log processing.
type Logger struct {
	*logrus.Logger
	chain        *handlers.Chain
	asyncHandler handlers.Handler // 用于存储异步处理器的引用，以便在关闭时调用Close()
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// GetLogger returns the global singleton Logger instance.
// It initializes the logger with default configuration on first call.
func GetLogger() *Logger {
	once.Do(func() {
		defaultLogger = &Logger{
			Logger: logrus.New(),
		}
		// Apply default configuration
		if err := ApplyDefaultConfig(defaultLogger); err != nil {
			defaultLogger.Error("Failed to apply default config:", err)
		}
	})

	return defaultLogger
}

// WithField returns a new entry with the specified field.
// This is used for structured logging to add contextual information to logs.
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithFields returns a new entry with the specified fields.
// This adds multiple key-value pairs as context to the log entry.
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithError returns a new entry with the error field set.
// This is a convenience method for logging errors with proper formatting.
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithContext returns a new entry with the context field set.
// This method is useful for adding application context to logs.
func (l *Logger) WithContext(ctx interface{}) *logrus.Entry {
	return l.Logger.WithField("context", ctx)
}

// WithTag adds a tag to the log entry.
// Tags help in filtering and categorizing logs.
func (l *Logger) WithTag(tag string) *logrus.Entry {
	return l.Logger.WithField("tag", tag)
}

// WithTags adds multiple tags to the log entry.
// This is useful for multi-dimensional categorization of logs.
func (l *Logger) WithTags(tags ...string) *logrus.Entry {
	return l.Logger.WithField("tags", tags)
}

// Info logs at the info level and processes through the handler chain.
// Info level is used for general operational information.
func (l *Logger) Info(args ...interface{}) {
	// 日志级别检查
	if l.Logger.IsLevelEnabled(logrus.InfoLevel) {
		// 使用处理器链处理日志
		if l.chain != nil {
			// 不再复制 Logger，而是创建一个新的具有相同配置的 Logger 实例
			tempLogger := logrus.New()
			tempLogger.SetLevel(l.Logger.GetLevel())
			tempLogger.SetFormatter(l.Logger.Formatter)
			tempLogger.SetReportCaller(l.Logger.ReportCaller)

			// 使用指针传递给 Process，tempLogger 已经是指针
			l.chain.Process(tempLogger, args...)
		}
		// 直接记录日志
		l.Logger.Info(args...)
	}
}

// Infof logs a formatted message at the info level.
// It uses fmt.Sprintf internally to format the message.
func (l *Logger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Info(msg)
}

// Error logs at the error level and processes through the handler chain.
// Error level indicates that something went wrong in the application.
func (l *Logger) Error(args ...interface{}) {
	// 日志级别检查
	if l.Logger.IsLevelEnabled(logrus.ErrorLevel) {
		// 使用处理器链处理日志
		if l.chain != nil {
			// 不再复制 Logger，而是创建一个新的具有相同配置的 Logger 实例
			tempLogger := logrus.New()
			tempLogger.SetLevel(l.Logger.GetLevel())
			tempLogger.SetFormatter(l.Logger.Formatter)
			tempLogger.SetReportCaller(l.Logger.ReportCaller)

			// 使用指针传递给 Process
			l.chain.Process(tempLogger, args...)
		}
		// 直接记录日志
		l.Logger.Error(args...)
	}
}

// Errorf logs a formatted message at the error level.
// It uses fmt.Sprintf internally to format the message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Error(msg)
}

// Debug logs at the debug level and processes through the handler chain.
// Debug level is typically used during development for detailed information.
func (l *Logger) Debug(args ...interface{}) {
	// 日志级别检查
	if l.Logger.IsLevelEnabled(logrus.DebugLevel) {
		// 使用处理器链处理日志
		if l.chain != nil {
			// 不再复制 Logger，而是创建一个新的具有相同配置的 Logger 实例
			tempLogger := logrus.New()
			tempLogger.SetLevel(l.Logger.GetLevel())
			tempLogger.SetFormatter(l.Logger.Formatter)
			tempLogger.SetReportCaller(l.Logger.ReportCaller)

			// 使用指针传递给 Process
			l.chain.Process(tempLogger, args...)
		}
		// 直接记录日志
		l.Logger.Debug(args...)
	}
}

// Debugf logs a formatted message at the debug level.
// It uses fmt.Sprintf internally to format the message.
func (l *Logger) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Debug(msg)
}

// Warn logs at the warn level and processes through the handler chain.
// Warn level indicates potentially harmful situations that require attention.
func (l *Logger) Warn(args ...interface{}) {
	// 日志级别检查
	if l.Logger.IsLevelEnabled(logrus.WarnLevel) {
		// 使用处理器链处理日志
		if l.chain != nil {
			// 不再复制 Logger，而是创建一个新的具有相同配置的 Logger 实例
			tempLogger := logrus.New()
			tempLogger.SetLevel(l.Logger.GetLevel())
			tempLogger.SetFormatter(l.Logger.Formatter)
			tempLogger.SetReportCaller(l.Logger.ReportCaller)

			// 使用指针传递给 Process
			l.chain.Process(tempLogger, args...)
		}
		// 直接记录日志
		l.Logger.Warn(args...)
	}
}

// Warnf logs a formatted message at the warn level.
// It uses fmt.Sprintf internally to format the message.
func (l *Logger) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Warn(msg)
}

// Fatal logs at the fatal level and processes through the handler chain.
// After logging, the program will exit with status code 1.
func (l *Logger) Fatal(args ...interface{}) {
	// 日志级别检查
	if l.Logger.IsLevelEnabled(logrus.FatalLevel) {
		// 使用处理器链处理日志
		if l.chain != nil {
			// 不再复制 Logger，而是创建一个新的具有相同配置的 Logger 实例
			tempLogger := logrus.New()
			tempLogger.SetLevel(l.Logger.GetLevel())
			tempLogger.SetFormatter(l.Logger.Formatter)
			tempLogger.SetReportCaller(l.Logger.ReportCaller)

			// 使用指针传递给 Process
			l.chain.Process(tempLogger, args...)
		}
		// 直接记录日志
		l.Logger.Fatal(args...)
	}
}

// Fatalf logs a formatted message at the fatal level.
// It uses fmt.Sprintf internally to format the message.
// After logging, the program will exit with status code 1.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Fatal(msg)
}

// Panic logs at the panic level and processes through the handler chain.
// After logging, the program will panic with the message.
func (l *Logger) Panic(args ...interface{}) {
	// 日志级别检查
	if l.Logger.IsLevelEnabled(logrus.PanicLevel) {
		// 使用处理器链处理日志
		if l.chain != nil {
			// 不再复制 Logger，而是创建一个新的具有相同配置的 Logger 实例
			tempLogger := logrus.New()
			tempLogger.SetLevel(l.Logger.GetLevel())
			tempLogger.SetFormatter(l.Logger.Formatter)
			tempLogger.SetReportCaller(l.Logger.ReportCaller)

			// 使用指针传递给 Process
			l.chain.Process(tempLogger, args...)
		}
		// 直接记录日志
		l.Logger.Panic(args...)
	}
}

// Panicf logs a formatted message at the panic level.
// It uses fmt.Sprintf internally to format the message.
// After logging, the program will panic with the message.
func (l *Logger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Panic(msg)
}

// AddHandler adds a new handler to the chain.
// This allows for extending the logger with custom processing logic.
func (l *Logger) AddHandler(handler handlers.Handler) {
	if l.chain == nil {
		l.chain = handlers.NewChain()
	}
	l.chain.Add(handler)
}

// Close gracefully shuts down the logger and its handlers.
// It ensures that all pending log entries are processed before returning.
func (l *Logger) Close() error {
	if l.asyncHandler != nil {
		if closer, ok := l.asyncHandler.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}

// ConfigureFromFile loads logger configuration from a file.
// Currently not implemented, reserved for future use.
func (l *Logger) ConfigureFromFile(filename string) error {
	// 实现从文件加载配置的逻辑
	// TODO: 根据需要实现配置文件的加载和解析
	return fmt.Errorf("not implemented yet")
}

// Global convenience functions that use the default logger

// Info logs at info level using the default logger.
// This is a convenience function for quick logging without managing a logger instance.
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof logs a formatted message at info level using the default logger.
// This is a convenience function that formats the message before logging.
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Error logs at error level using the default logger.
// This is a convenience function for logging errors without managing a logger instance.
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf logs a formatted message at error level using the default logger.
// This is a convenience function that formats the error message before logging.
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Debug logs at debug level using the default logger.
// This is a convenience function for debug logging without managing a logger instance.
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf logs a formatted message at debug level using the default logger.
// This is a convenience function that formats the debug message before logging.
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Warn logs at warn level using the default logger.
// This is a convenience function for warning logs without managing a logger instance.
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf logs a formatted message at warn level using the default logger.
// This is a convenience function that formats the warning message before logging.
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Fatal logs at fatal level using the default logger.
// This is a convenience function for fatal errors without managing a logger instance.
// After logging, the program will exit with status code 1.
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf logs a formatted message at fatal level using the default logger.
// This is a convenience function that formats the fatal message before logging.
// After logging, the program will exit with status code 1.
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Panic logs at panic level using the default logger.
// This is a convenience function for panic-level errors without managing a logger instance.
// After logging, the program will panic with the message.
func Panic(args ...interface{}) {
	GetLogger().Panic(args...)
}

// Panicf logs a formatted message at panic level using the default logger.
// This is a convenience function that formats the panic message before logging.
// After logging, the program will panic with the message.
func Panicf(format string, args ...interface{}) {
	GetLogger().Panicf(format, args...)
}

// WithField returns a new entry with the specified field using the default logger.
// This is a convenience function for structured logging with the default logger.
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithFields returns a new entry with the specified fields using the default logger.
// This is a convenience function for adding multiple fields to the default logger entry.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithError returns a new entry with the error field set using the default logger.
// This is a convenience function for logging errors with the default logger.
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}

// SetLevel sets the logging level for the default logger.
// This controls which severity of logs will be output.
func SetLevel(level logrus.Level) {
	GetLogger().SetLevel(level)
}

// Configure applies the given config to the default logger.
// This is the main way to customize the default logger's behavior.
func Configure(config *Config) error {
	return ApplyConfig(GetLogger(), config)
}

// Close gracefully shuts down the default logger.
// This ensures that all pending log entries are processed before returning.
func Close() error {
	return GetLogger().Close()
}
