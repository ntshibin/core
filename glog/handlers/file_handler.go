// Package handlers provides a set of handlers for the glog logging system.
// It implements the chain of responsibility pattern for log processing.
package handlers

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// FileHandler implements file output processing.
// It writes log entries to a file with rotation capabilities using lumberjack.
type FileHandler struct {
	BaseHandler
	FilePath   string // Path to the log file
	MaxSize    int    // Maximum size in megabytes before rotation
	MaxBackups int    // Maximum number of old log files to retain
	MaxAge     int    // Maximum days to retain old log files
	Compress   bool   // Whether to compress rotated files
	writer     io.Writer
}

// NewFileHandler creates a new file handler.
// It initializes the log directory and configures log rotation based on the provided parameters.
// Panics if it cannot create the log directory.
func NewFileHandler(filePath string, maxSize, maxBackups, maxAge int, compress bool) *FileHandler {
	// 确保日志目录存在
	logDir := filepath.Dir(filePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	// 配置日志轮转
	rotateLogger := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}

	return &FileHandler{
		FilePath:   filePath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
		writer:     rotateLogger,
	}
}

// Handle processes the log entry for file output.
// It creates a temporary logger that writes to the file and logs the message
// based on the original logger's level.
// After processing, it calls the next handler in the chain if one exists.
func (h *FileHandler) Handle(logger *logrus.Logger, args ...interface{}) {
	// 如果writer未初始化，则初始化
	if h.writer == nil {
		rotateLogger := &lumberjack.Logger{
			Filename:   h.FilePath,
			MaxSize:    h.MaxSize,
			MaxBackups: h.MaxBackups,
			MaxAge:     h.MaxAge,
			Compress:   h.Compress,
		}
		h.writer = rotateLogger
	}

	// 创建一个临时logger用于写入文件
	tmpLogger := logrus.New()
	tmpLogger.SetFormatter(logger.Formatter)
	tmpLogger.SetLevel(logger.Level)
	tmpLogger.SetReportCaller(logger.ReportCaller)
	tmpLogger.SetOutput(h.writer)

	// 直接写入日志到文件
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
