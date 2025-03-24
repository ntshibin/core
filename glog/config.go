// Package glog provides an enhanced logging system built on top of logrus
// with features like handler chains, multiple output targets, and structured logging.
package glog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ntshibin/core/glog/handlers"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogFormat 定义日志格式类型
type LogFormat string

const (
	// FormatJSON 表示JSON格式
	FormatJSON LogFormat = "json"
	// FormatText 表示文本格式
	FormatText LogFormat = "text"
)

// Config defines the configuration for the logger.
// It contains all settings that control the behavior of the logging system.
type Config struct {
	// Level specifies the minimum log level
	Level logrus.Level `json:"level" yaml:"level" env:"LOG_LEVEL" default:"info"`

	// Format specifies the log format (json or text)
	Format LogFormat `json:"format" yaml:"format" env:"LOG_FORMAT" default:"text"`

	// EnableConsole enables console output
	EnableConsole bool `json:"enable_console" yaml:"enable_console" env:"LOG_ENABLE_CONSOLE" default:"true"`

	// EnableFile enables file output
	EnableFile bool `json:"enable_file" yaml:"enable_file" env:"LOG_ENABLE_FILE" default:"false"`

	// ReportCaller enables the logging of caller information
	ReportCaller bool `json:"report_caller" yaml:"report_caller" env:"LOG_REPORT_CALLER" default:"false"`

	// TimestampFormat defines the format of timestamps in logs
	TimestampFormat string `json:"timestamp_format" yaml:"timestamp_format" env:"LOG_TIMESTAMP_FORMAT" default:"2006-01-02 15:04:05"`

	// DisableColors disables color output in console logging
	DisableColors bool `json:"disable_colors" yaml:"disable_colors" env:"LOG_DISABLE_COLORS" default:"false"`

	// FileConfig contains file-specific configurations
	FileConfig *FileConfig `json:"file_config" yaml:"file_config"`

	// CustomHandlers contains additional handlers to be added to the chain
	CustomHandlers []handlers.Handler `json:"-" yaml:"-"`

	// EnableAsync enables asynchronous logging
	EnableAsync bool `json:"enable_async" yaml:"enable_async" env:"LOG_ENABLE_ASYNC" default:"false"`

	// AsyncConfig contains configuration for asynchronous logging
	AsyncConfig *AsyncConfig `json:"async_config" yaml:"async_config"`

	// EnableZap enables zap logger
	EnableZap bool `json:"enable_zap" yaml:"enable_zap" env:"LOG_ENABLE_ZAP" default:"false"`

	// ZapConfig contains configuration for zap logger
	ZapConfig *handlers.ZapConfig `json:"zap_config" yaml:"zap_config"`
}

// FileConfig contains settings for file-based logging.
type FileConfig struct {
	// Filename specifies the log file path
	Filename string `json:"filename" yaml:"filename" env:"LOG_FILE_PATH" default:"app.log"`

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated
	MaxSize int `json:"max_size" yaml:"max_size" env:"LOG_FILE_MAX_SIZE" default:"100"`

	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `json:"max_backups" yaml:"max_backups" env:"LOG_FILE_MAX_BACKUPS" default:"10"`

	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `json:"max_age" yaml:"max_age" env:"LOG_FILE_MAX_AGE" default:"30"`

	// Compress determines if the rotated log files should be compressed
	Compress bool `json:"compress" yaml:"compress" env:"LOG_FILE_COMPRESS" default:"false"`
}

// AsyncConfig contains settings for asynchronous logging.
type AsyncConfig struct {
	// BufferSize is the maximum number of log entries that can be queued.
	// If the queue is full, new entries will be dropped with a warning.
	BufferSize int `json:"buffer_size" yaml:"buffer_size" env:"LOG_ASYNC_BUFFER_SIZE" default:"1000"`

	// BatchSize is the number of log entries to process in one batch.
	// Larger batches can improve throughput but might increase memory usage.
	BatchSize int `json:"batch_size" yaml:"batch_size" env:"LOG_ASYNC_BATCH_SIZE" default:"100"`

	// FlushInterval is the maximum duration to wait before flushing a batch.
	// Even if a batch is not full, it will be processed after this interval.
	FlushInterval time.Duration `json:"flush_interval" yaml:"flush_interval" env:"LOG_ASYNC_FLUSH_INTERVAL" default:"5s"`
}

// DefaultAsyncConfig returns default configuration for asynchronous logging.
func DefaultAsyncConfig() *AsyncConfig {
	return &AsyncConfig{
		BufferSize:    1000,        // Buffer up to 1000 log entries
		BatchSize:     100,         // Process in batches of 100
		FlushInterval: time.Second, // Flush at least once per second
	}
}

// DefaultFileConfig returns the default configuration for file-based logging.
func DefaultFileConfig() *FileConfig {
	return &FileConfig{
		Filename:   "app.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   false,
	}
}

// DefaultConfig returns the default configuration for the logger.
func DefaultConfig() *Config {
	cfg := &Config{
		Level:           logrus.InfoLevel,
		Format:          FormatText,
		EnableConsole:   true,
		EnableFile:      false,
		ReportCaller:    false,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
		EnableAsync:     false,
		FileConfig:      DefaultFileConfig(),
		AsyncConfig:     DefaultAsyncConfig(),
	}

	// 尝试使用 gconf 处理默认值
	if gcfg, ok := interface{}(nil).(interface{ LoadDefaultsFromStruct(interface{}) error }); ok {
		_ = gcfg.LoadDefaultsFromStruct(cfg)
	}

	return cfg
}

// ApplyConfig applies the configuration to the logger.
// It sets up outputs, formatters, and handlers according to the provided configuration.
// Returns an error if any setup operation fails.
func ApplyConfig(logger *Logger, config *Config) error {
	var outputs []io.Writer

	// 添加控制台输出
	if config.EnableConsole {
		outputs = append(outputs, os.Stdout)
	}

	// 添加文件输出
	if config.EnableFile && config.FileConfig != nil {
		// 确保日志目录存在
		logDir := filepath.Dir(config.FileConfig.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// 如果FileConfig未设置默认值，则使用合理的默认值
		fileConfig := config.FileConfig
		if fileConfig.MaxSize <= 0 {
			fileConfig.MaxSize = 10 // 默认10MB
		}
		if fileConfig.MaxBackups <= 0 {
			fileConfig.MaxBackups = 5 // 默认保留5个备份
		}
		if fileConfig.MaxAge <= 0 {
			fileConfig.MaxAge = 7 // 默认保留7天
		}

		// 创建lumberjack日志轮转器
		fileWriter := &lumberjack.Logger{
			Filename:   fileConfig.Filename,
			MaxSize:    fileConfig.MaxSize,
			MaxBackups: fileConfig.MaxBackups,
			MaxAge:     fileConfig.MaxAge,
			Compress:   fileConfig.Compress,
		}

		outputs = append(outputs, fileWriter)
	}

	// 设置输出
	var output io.Writer
	if len(outputs) == 0 {
		// 如果没有启用任何输出，默认使用标准输出
		output = os.Stdout
	} else if len(outputs) == 1 {
		output = outputs[0]
	} else {
		output = io.MultiWriter(outputs...)
	}
	logger.SetOutput(output)

	// 初始化处理器链
	chain := handlers.NewChain()

	// 添加zap处理器
	if config.EnableZap {
		zapHandler, err := handlers.NewZapHandler(config.ZapConfig)
		if err != nil {
			return fmt.Errorf("failed to create zap handler: %v", err)
		}
		chain.Add(zapHandler)
		logger.zapHandler = zapHandler
	}

	// 设置默认格式化器
	if config.Format == FormatJSON {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: config.TimestampFormat,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: config.TimestampFormat,
			DisableColors:   config.DisableColors,
			FullTimestamp:   true,
		})
	}

	// Set caller reporting
	logger.SetReportCaller(config.ReportCaller)

	// Add level handler to control log level
	chain.Add(handlers.NewLevelHandler(config.Level))

	// Add formatter handler based on format
	var formatterType handlers.FormatterType
	if config.Format == FormatJSON {
		formatterType = handlers.JSONFormatter
	} else {
		formatterType = handlers.TextFormatter
	}

	// Add formatter handler
	chain.Add(handlers.NewFormatterHandler(
		formatterType,
		config.TimestampFormat,
		config.DisableColors,
	))

	// Add async handler if enabled
	var asyncHandler *handlers.AsyncHandler
	if config.EnableAsync {
		asyncConfig := config.AsyncConfig
		if asyncConfig == nil {
			asyncConfig = DefaultAsyncConfig()
		}

		// 创建并添加异步处理器
		asyncHandler = handlers.NewAsyncHandler(&handlers.AsyncConfig{
			BufferSize:    asyncConfig.BufferSize,
			BatchSize:     asyncConfig.BatchSize,
			FlushInterval: asyncConfig.FlushInterval,
		})
		chain.Add(asyncHandler)

		// 保存异步处理器以便能在应用关闭时正确关闭
		logger.asyncHandler = asyncHandler
	}

	// Add console handler if enabled
	if config.EnableConsole {
		chain.Add(handlers.NewConsoleHandler(config.DisableColors))
	}

	// Add file handler if enabled
	if config.EnableFile && config.FileConfig != nil {
		fileHandler := handlers.NewFileHandler(
			config.FileConfig.Filename,
			config.FileConfig.MaxSize,
			config.FileConfig.MaxBackups,
			config.FileConfig.MaxAge,
			config.FileConfig.Compress,
		)
		chain.Add(fileHandler)
	}

	// Add custom handlers
	for _, handler := range config.CustomHandlers {
		chain.Add(handler)
	}

	// Store the chain in the logger
	logger.chain = chain

	return nil
}

// ApplyDefaultConfig applies the default configuration to the logger.
// This is a convenience function to quickly set up a logger with reasonable defaults.
func ApplyDefaultConfig(logger *Logger) error {
	return ApplyConfig(logger, DefaultConfig())
}
