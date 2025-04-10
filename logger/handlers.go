package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

// BaseHandler 基础处理器
type BaseHandler struct {
	formatter Formatter
	level     LogLevel
}

// NewBaseHandler 创建基础处理器
func NewBaseHandler(formatter Formatter, level LogLevel) *BaseHandler {
	return &BaseHandler{
		formatter: formatter,
		level:     level,
	}
}

// Format 格式化日志事件
func (h *BaseHandler) Format(event LogEvent) ([]byte, error) {
	return h.formatter.Format(event)
}

// ShouldHandle 是否应该处理该事件
func (h *BaseHandler) ShouldHandle(event LogEvent) bool {
	return event.Level >= h.level
}

// Close 关闭处理器
func (h *BaseHandler) Close() error {
	return nil
}

// ConsoleHandler 控制台处理器
type ConsoleHandler struct {
	*BaseHandler
	writer io.Writer
}

// NewConsoleHandler 创建控制台处理器
func NewConsoleHandler(formatter Formatter, level LogLevel) *ConsoleHandler {
	return &ConsoleHandler{
		BaseHandler: NewBaseHandler(formatter, level),
		writer:      os.Stdout,
	}
}

// Handle 处理日志事件
func (h *ConsoleHandler) Handle(event LogEvent) error {
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

// FileHandler 文件处理器
type FileHandler struct {
	*BaseHandler
	writer io.WriteCloser
}

// NewFileHandler 创建文件处理器
func NewFileHandler(formatter Formatter, level LogLevel, filePath string) (*FileHandler, error) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %v", err)
		}
	}

	// 打开文件
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %v", err)
	}

	return &FileHandler{
		BaseHandler: NewBaseHandler(formatter, level),
		writer:      file,
	}, nil
}

// Handle 处理日志事件
func (h *FileHandler) Handle(event LogEvent) error {
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
func (h *FileHandler) Close() error {
	return h.writer.Close()
}

// FileRotateConfig 轮转配置
type FileRotateConfig struct {
	FilePath   string `yaml:"file_path" json:"file_path"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// DefaultFileRotateConfig 默认轮转配置
var DefaultFileRotateConfig = FileRotateConfig{
	FilePath:   "logs/app.log",
	MaxSize:    100,
	MaxBackups: 3,
	MaxAge:     28,
	Compress:   true,
}

// RotateFileHandler 轮转文件处理器
type RotateFileHandler struct {
	*BaseHandler
	writer *lumberjack.Logger
}

// NewRotateFileHandler 创建轮转文件处理器
func NewRotateFileHandler(formatter Formatter, level LogLevel, config FileRotateConfig) (*RotateFileHandler, error) {
	// 确保目录存在
	dir := filepath.Dir(config.FilePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %v", err)
		}
	}

	// 创建轮转写入器
	writer := &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	return &RotateFileHandler{
		BaseHandler: NewBaseHandler(formatter, level),
		writer:      writer,
	}, nil
}

// Handle 处理日志事件
func (h *RotateFileHandler) Handle(event LogEvent) error {
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
func (h *RotateFileHandler) Close() error {
	return h.writer.Close()
}

// MultiHandler 多处理器
type MultiHandler struct {
	handlers []Handler
	mu       sync.RWMutex
}

// NewMultiHandler 创建多处理器
func NewMultiHandler(handlers ...Handler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

// AddHandler 添加处理器
func (h *MultiHandler) AddHandler(handler Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers = append(h.handlers, handler)
}

// RemoveHandler 移除处理器
func (h *MultiHandler) RemoveHandler(handler Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, hdl := range h.handlers {
		if hdl == handler {
			h.handlers = append(h.handlers[:i], h.handlers[i+1:]...)
			break
		}
	}
}

// Handle 处理日志事件
func (h *MultiHandler) Handle(event LogEvent) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var lastErr error
	for _, handler := range h.handlers {
		if err := handler.Handle(event); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Close 关闭所有处理器
func (h *MultiHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var lastErr error
	for _, handler := range h.handlers {
		if err := handler.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// ShouldHandle 是否应该处理该事件
func (h *MultiHandler) ShouldHandle(event LogEvent) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, handler := range h.handlers {
		if handler.ShouldHandle(event) {
			return true
		}
	}
	return false
}

// Format 不应该被调用
func (h *MultiHandler) Format(event LogEvent) ([]byte, error) {
	return nil, fmt.Errorf("MultiHandler 不支持直接格式化")
}
