// Package handlers provides a set of handlers for the glog logging system.
// It implements the chain of responsibility pattern for log processing.
package handlers

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// ConsoleHandler handles outputting logs to the console (stdout).
// It can be configured to use or disable colors in the output.
type ConsoleHandler struct {
	BaseHandler
	// 用于控制是否启用颜色
	DisableColors bool
	// 控制台输出目标，默认为os.Stdout
	Writer io.Writer
}

// NewConsoleHandler creates a new console handler.
// The disableColors parameter controls whether to use colored output.
func NewConsoleHandler(disableColors bool) *ConsoleHandler {
	return &ConsoleHandler{
		DisableColors: disableColors,
		Writer:        os.Stdout,
	}
}

// Handle processes the log entry for console output.
// It sets the output to stdout and configures color options.
// After processing, it calls the next handler in the chain if one exists.
func (h *ConsoleHandler) Handle(logger *logrus.Logger, args ...interface{}) {
	// 创建一个临时logger用于控制台输出
	tmpLogger := logrus.New()
	tmpLogger.SetOutput(os.Stdout)

	// 设置格式化器
	if formatter, ok := logger.Formatter.(*logrus.TextFormatter); ok {
		// 创建新的格式化器而不是复制，避免复制sync.Once
		newFormatter := &logrus.TextFormatter{
			DisableColors:             h.DisableColors,
			DisableTimestamp:          formatter.DisableTimestamp,
			ForceColors:               formatter.ForceColors,
			DisableSorting:            formatter.DisableSorting,
			DisableLevelTruncation:    formatter.DisableLevelTruncation,
			PadLevelText:              formatter.PadLevelText,
			QuoteEmptyFields:          formatter.QuoteEmptyFields,
			EnvironmentOverrideColors: formatter.EnvironmentOverrideColors,
			TimestampFormat:           formatter.TimestampFormat,
		}
		tmpLogger.SetFormatter(newFormatter)
	} else {
		// 使用原始格式化器
		tmpLogger.SetFormatter(logger.Formatter)
	}

	tmpLogger.SetLevel(logger.Level)

	// 根据日志级别记录到控制台
	switch {
	case logger.IsLevelEnabled(logrus.InfoLevel) && len(args) > 0:
		if msg, ok := args[0].(string); ok {
			tmpLogger.Info(msg)
		}
	case logger.IsLevelEnabled(logrus.WarnLevel) && len(args) > 0:
		if msg, ok := args[0].(string); ok {
			tmpLogger.Warn(msg)
		}
	case logger.IsLevelEnabled(logrus.ErrorLevel) && len(args) > 0:
		if msg, ok := args[0].(string); ok {
			tmpLogger.Error(msg)
		}
	case logger.IsLevelEnabled(logrus.DebugLevel) && len(args) > 0:
		if msg, ok := args[0].(string); ok {
			tmpLogger.Debug(msg)
		}
	}

	// 调用链中的下一个处理器
	if h.Next != nil {
		h.Next.Handle(logger, args...)
	}
}
