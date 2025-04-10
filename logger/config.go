package logger

// LoggerConfig 日志配置结构体
type LoggerConfig struct {
	// 日志记录器名称
	Name string `yaml:"name" json:"name"`
	// 日志级别: debug, info, warn, error, fatal
	Level string `yaml:"level" json:"level"`
	// 输出格式: json, text
	Encoding string `yaml:"encoding" json:"encoding"`
	// 是否跳过调用者信息
	CallerSkip bool `yaml:"caller_skip" json:"caller_skip"`

	// 控制台输出配置
	EnableConsole bool `yaml:"enable_console" json:"enable_console"`

	// 文件输出配置
	EnableFile bool   `yaml:"enable_file" json:"enable_file"`
	FilePath   string `yaml:"file_path" json:"file_path"`

	// 日志轮转配置
	EnableRotate bool             `yaml:"enable_rotate" json:"enable_rotate"`
	Rotate       FileRotateConfig `yaml:"rotate" json:"rotate"`

	// 异步日志配置
	EnableAsync    bool `yaml:"enable_async" json:"enable_async"`
	AsyncQueueSize int  `yaml:"async_queue_size" json:"async_queue_size"`

	// 远程日志配置
	EnableRemote bool         `yaml:"enable_remote" json:"enable_remote"`
	Remote       RemoteConfig `yaml:"remote" json:"remote"`

	// 内存日志配置
	EnableMemory bool         `yaml:"enable_memory" json:"enable_memory"`
	Memory       MemoryConfig `yaml:"memory" json:"memory"`

	// 调用链跟踪配置
	EnableTrace bool `yaml:"enable_trace" json:"enable_trace"`
}

// DefaultLoggerConfig 默认日志配置
var DefaultLoggerConfig = LoggerConfig{
	Name:           "default",
	Level:          "info",
	Encoding:       "json",
	CallerSkip:     false,
	EnableConsole:  true,
	EnableFile:     false,
	FilePath:       "logs/app.log",
	EnableRotate:   false,
	Rotate:         DefaultFileRotateConfig,
	EnableAsync:    false,
	AsyncQueueSize: 1000,
	EnableRemote:   false,
	Remote:         DefaultRemoteConfig,
	EnableMemory:   false,
	Memory:         DefaultMemoryConfig,
	EnableTrace:    false,
}

// LoadConfig 初始化日志系统
func LoadConfig(config LoggerConfig) error {
	// 解析日志级别
	level, err := ParseLevel(config.Level)
	if err != nil {
		return err
	}

	// 创建默认处理器
	var handlers []Handler

	// 添加控制台处理器
	if config.EnableConsole {
		var formatter Formatter
		if config.Encoding == "json" {
			formatter = NewJSONFormatter()
		} else {
			formatter = NewTextFormatter()
		}
		handlers = append(handlers, NewConsoleHandler(formatter, level))
	}

	// 添加文件处理器
	if config.EnableFile && !config.EnableRotate {
		handler, err := NewFileHandler(NewJSONFormatter(), level, config.FilePath)
		if err != nil {
			return err
		}
		handlers = append(handlers, handler)
	}

	// 添加轮转文件处理器
	if config.EnableRotate {
		handler, err := NewRotateFileHandler(NewJSONFormatter(), level, config.Rotate)
		if err != nil {
			return err
		}
		handlers = append(handlers, handler)
	}

	// 添加远程日志处理器
	if config.EnableRemote {
		handler, err := NewRemoteHandler(NewJSONFormatter(), level, config.Remote)
		if err != nil {
			return err
		}
		handlers = append(handlers, handler)
	}

	// 添加内存日志处理器
	if config.EnableMemory {
		handler := NewMemoryHandler(NewJSONFormatter(), level, config.Memory)
		handlers = append(handlers, handler)
	}

	// 根据异步配置处理处理器
	if config.EnableAsync {
		// 启用全局异步模式
		SetAsyncMode(MixedMode)
		// 设置异步配置
		SetAsyncQueueSize(config.AsyncQueueSize)
	}

	// 创建日志记录器
	logger := NewStandardLogger(config.Name, level, handlers...)

	// 替换默认日志记录器
	manager := GetLogManager()
	manager.mu.Lock()
	manager.loggers["default"] = logger
	manager.mu.Unlock()

	return nil
}

// InitWithFileLog 初始化日志系统并启用文件日志
// 这是一个便捷方法，用于快速启用文件日志功能
func InitWithFileLog(level string, filePath string) error {
	config := DefaultLoggerConfig
	config.Level = level
	config.EnableFile = true
	config.FilePath = filePath

	return LoadConfig(config)
}

// InitWithRotateLog 初始化日志系统并启用轮转文件日志
// 这是一个便捷方法，用于快速启用轮转文件日志功能
func InitWithRotateLog(level string, rotateConfig FileRotateConfig) error {
	config := DefaultLoggerConfig
	config.Level = level
	config.EnableRotate = true
	config.Rotate = rotateConfig

	return LoadConfig(config)
}

// InitWithAsyncMode 初始化日志系统并设置异步模式
// 这是一个便捷方法，用于快速设置异步模式和队列大小
func InitWithAsyncMode(mode LogAsyncMode, queueSize int) error {
	// 设置异步配置
	SetAsyncQueueSize(queueSize)
	SetAsyncMode(mode)
	return nil
}
