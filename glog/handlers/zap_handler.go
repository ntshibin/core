// Package handlers provides a set of handlers for the glog logging system.
package handlers

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapHandler implements Handler interface using zap logger.
type ZapHandler struct {
	BaseHandler
	zapLogger *zap.Logger
	config    *ZapConfig
	mu        sync.RWMutex
}

// ZapConfig contains configuration for zap logger.
type ZapConfig struct {
	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `json:"development" yaml:"development" env:"LOG_ZAP_DEVELOPMENT" default:"false"`

	// Encoding sets the logger's encoding. Valid values are "json" and "console".
	Encoding string `json:"encoding" yaml:"encoding" env:"LOG_ZAP_ENCODING" default:"json"`

	// OutputPaths is a list of URLs or file paths to write logging output to.
	OutputPaths []string `json:"output_paths" yaml:"output_paths" env:"LOG_ZAP_OUTPUT_PATHS" default:"[\"stdout\"]"`

	// ErrorOutputPaths is a list of URLs to write internal logger errors to.
	ErrorOutputPaths []string `json:"error_output_paths" yaml:"error_output_paths" env:"LOG_ZAP_ERROR_OUTPUT_PATHS" default:"[\"stderr\"]"`

	// InitialFields is a collection of fields to add to the root logger.
	InitialFields map[string]interface{} `json:"initial_fields" yaml:"initial_fields"`

	// DisableCaller stops annotating logs with the calling function's file
	// name and line number.
	DisableCaller bool `json:"disable_caller" yaml:"disable_caller" env:"LOG_ZAP_DISABLE_CALLER" default:"false"`

	// DisableStacktrace disables automatic stacktrace capturing.
	DisableStacktrace bool `json:"disable_stacktrace" yaml:"disable_stacktrace" env:"LOG_ZAP_DISABLE_STACKTRACE" default:"false"`
}

// NewZapHandler creates a new ZapHandler with the given configuration.
func NewZapHandler(config *ZapConfig) (*ZapHandler, error) {
	if config == nil {
		config = &ZapConfig{
			Development:      false,
			Encoding:         "json",
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
	}

	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:       config.Development,
		Encoding:          config.Encoding,
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		OutputPaths:       config.OutputPaths,
		ErrorOutputPaths:  config.ErrorOutputPaths,
		InitialFields:     config.InitialFields,
		DisableCaller:     config.DisableCaller,
		DisableStacktrace: config.DisableStacktrace,
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %v", err)
	}

	return &ZapHandler{
		zapLogger: logger,
		config:    config,
	}, nil
}

// Handle processes the log entry using zap logger.
func (h *ZapHandler) Handle(logger *logrus.Logger, args ...interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Convert logrus level to zap level
	zapLevel := logrusLevelToZapLevel(logger.GetLevel())

	// Create zap fields from args
	fields := make([]zap.Field, 0)
	for i, arg := range args {
		fields = append(fields, zap.Any(fmt.Sprintf("arg%d", i), arg))
	}

	// Log with zap
	h.zapLogger.Log(zapLevel, fmt.Sprint(args...), fields...)

	// Pass to next handler if exists
	if h.Next != nil {
		h.Next.Handle(logger, args...)
	}
}

// Close closes the zap logger.
func (h *ZapHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.zapLogger != nil {
		return h.zapLogger.Sync()
	}
	return nil
}

// logrusLevelToZapLevel converts logrus level to zap level.
func logrusLevelToZapLevel(level logrus.Level) zapcore.Level {
	switch level {
	case logrus.TraceLevel, logrus.DebugLevel:
		return zapcore.DebugLevel
	case logrus.InfoLevel:
		return zapcore.InfoLevel
	case logrus.WarnLevel:
		return zapcore.WarnLevel
	case logrus.ErrorLevel:
		return zapcore.ErrorLevel
	case logrus.FatalLevel:
		return zapcore.FatalLevel
	case logrus.PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}
