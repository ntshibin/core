// Package handlers provides a set of handlers for the glog logging system.
// It implements the chain of responsibility pattern for log processing.
package handlers

import (
	"github.com/sirupsen/logrus"
)

// FormatterType defines the type of formatter to use
type FormatterType string

const (
	// JSONFormatter indicates JSON formatting should be used
	JSONFormatter FormatterType = "json"
	// TextFormatter indicates text formatting should be used
	TextFormatter FormatterType = "text"
)

// FormatterHandler handles log formatting.
// It can set the formatter to either JSON or text format.
type FormatterHandler struct {
	BaseHandler
	Format          FormatterType
	TimestampFormat string
	DisableColors   bool
}

// NewFormatterHandler creates a new formatter handler.
// It takes the formatter type, timestamp format, and color configuration.
func NewFormatterHandler(format FormatterType, timestampFormat string, disableColors bool) *FormatterHandler {
	return &FormatterHandler{
		Format:          format,
		TimestampFormat: timestampFormat,
		DisableColors:   disableColors,
	}
}

// Handle processes the log entry by setting the appropriate formatter.
// It configures the logger with either JSON or text formatter based on the handler's configuration.
// After processing, it calls the next handler in the chain if one exists.
func (h *FormatterHandler) Handle(logger *logrus.Logger, args ...interface{}) {
	// 设置格式化器
	if h.Format == JSONFormatter {
		formatter := &logrus.JSONFormatter{
			TimestampFormat: h.TimestampFormat,
		}
		logger.SetFormatter(formatter)
	} else {
		formatter := &logrus.TextFormatter{
			TimestampFormat: h.TimestampFormat,
			DisableColors:   h.DisableColors,
			FullTimestamp:   true,
		}
		logger.SetFormatter(formatter)
	}

	// 调用链中的下一个处理器
	if h.Next != nil {
		h.Next.Handle(logger, args...)
	}
}
