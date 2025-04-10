package logger

import (
	"container/ring"
	"strings"
	"sync"
	"time"
)

// MemoryConfig 内存日志配置
type MemoryConfig struct {
	// Capacity 容量，即保留的最大日志条数
	Capacity int `yaml:"capacity" json:"capacity"`
	// ExpireTime 过期时间，超过此时间的日志会被自动清理
	ExpireTime time.Duration `yaml:"expire_time" json:"expire_time"`
	// CleanupInterval 清理间隔
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// DefaultMemoryConfig 默认内存配置
var DefaultMemoryConfig = MemoryConfig{
	Capacity:        1000,
	ExpireTime:      time.Hour * 24,
	CleanupInterval: time.Minute * 10,
}

// LogEntry 日志条目
type LogEntry struct {
	Event         LogEvent  // 日志事件
	FormattedData []byte    // 格式化后的数据
	Time          time.Time // 记录时间
}

// MemoryHandler 内存日志处理器
type MemoryHandler struct {
	*BaseHandler
	buffer    *ring.Ring      // 环形缓冲区
	entries   []LogEntry      // 日志条目列表
	config    MemoryConfig    // 配置
	mu        sync.RWMutex    // 互斥锁
	ticker    *time.Ticker    // 清理定时器
	done      chan struct{}   // 关闭通道
	listeners []chan LogEvent // 监听器列表，用于实时接收日志事件
	listMu    sync.RWMutex    // 监听器互斥锁
}

// NewMemoryHandler 创建内存日志处理器
func NewMemoryHandler(formatter Formatter, level LogLevel, config MemoryConfig) *MemoryHandler {
	// 设置默认值
	if config.Capacity <= 0 {
		config.Capacity = 1000
	}
	if config.ExpireTime <= 0 {
		config.ExpireTime = time.Hour * 24
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = time.Minute * 10
	}

	h := &MemoryHandler{
		BaseHandler: NewBaseHandler(formatter, level),
		buffer:      ring.New(config.Capacity),
		entries:     make([]LogEntry, 0, config.Capacity),
		config:      config,
		done:        make(chan struct{}),
		listeners:   make([]chan LogEvent, 0),
	}

	// 启动清理定时器
	h.ticker = time.NewTicker(config.CleanupInterval)
	go h.scheduleCleanup()

	return h
}

// scheduleCleanup 定时清理过期日志
func (h *MemoryHandler) scheduleCleanup() {
	for {
		select {
		case <-h.ticker.C:
			h.cleanup()
		case <-h.done:
			return
		}
	}
}

// cleanup 清理过期日志
func (h *MemoryHandler) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.entries) == 0 {
		return
	}

	now := time.Now()
	var validEntries []LogEntry

	for _, entry := range h.entries {
		if now.Sub(entry.Time) < h.config.ExpireTime {
			validEntries = append(validEntries, entry)
		}
	}

	h.entries = validEntries
}

// Handle 处理日志事件
func (h *MemoryHandler) Handle(event LogEvent) error {
	if !h.ShouldHandle(event) {
		return nil
	}

	data, err := h.Format(event)
	if err != nil {
		return err
	}

	// 创建日志条目
	entry := LogEntry{
		Event:         event,
		FormattedData: data,
		Time:          time.Now(),
	}

	// 添加到环形缓冲区和列表
	h.mu.Lock()
	h.buffer.Value = entry
	h.buffer = h.buffer.Next()
	h.entries = append(h.entries, entry)
	h.mu.Unlock()

	// 通知所有监听器
	h.notifyListeners(event)

	return nil
}

// notifyListeners 通知所有监听器
func (h *MemoryHandler) notifyListeners(event LogEvent) {
	h.listMu.RLock()
	defer h.listMu.RUnlock()

	for _, listener := range h.listeners {
		select {
		case listener <- event:
			// 成功发送
		default:
			// 通道已满，丢弃事件
		}
	}
}

// Close 关闭处理器
func (h *MemoryHandler) Close() error {
	h.ticker.Stop()
	close(h.done)

	// 关闭所有监听器
	h.listMu.Lock()
	for _, listener := range h.listeners {
		close(listener)
	}
	h.listeners = nil
	h.listMu.Unlock()

	return nil
}

// MemoryHandlerAPI 内存日志处理器API
type MemoryHandlerAPI struct {
	handler *MemoryHandler
}

// NewMemoryHandlerAPI 创建内存日志处理器API
func NewMemoryHandlerAPI(handler *MemoryHandler) *MemoryHandlerAPI {
	return &MemoryHandlerAPI{handler: handler}
}

// GetLatest 获取最新的n条日志
func (api *MemoryHandlerAPI) GetLatest(n int) []LogEntry {
	api.handler.mu.RLock()
	defer api.handler.mu.RUnlock()

	if len(api.handler.entries) == 0 {
		return []LogEntry{}
	}

	if n <= 0 || n > len(api.handler.entries) {
		n = len(api.handler.entries)
	}

	start := len(api.handler.entries) - n
	return api.handler.entries[start:]
}

// GetByLevel 根据级别获取日志
func (api *MemoryHandlerAPI) GetByLevel(level LogLevel, n int) []LogEntry {
	api.handler.mu.RLock()
	defer api.handler.mu.RUnlock()

	var result []LogEntry
	for _, entry := range api.handler.entries {
		if entry.Event.Level == level {
			result = append(result, entry)
		}
	}

	if n > 0 && n < len(result) {
		start := len(result) - n
		return result[start:]
	}

	return result
}

// GetByTimeRange 根据时间范围获取日志
func (api *MemoryHandlerAPI) GetByTimeRange(start, end time.Time) []LogEntry {
	api.handler.mu.RLock()
	defer api.handler.mu.RUnlock()

	var result []LogEntry
	for _, entry := range api.handler.entries {
		if (entry.Time.After(start) || entry.Time.Equal(start)) &&
			(entry.Time.Before(end) || entry.Time.Equal(end)) {
			result = append(result, entry)
		}
	}

	return result
}

// GetContaining 获取包含特定文本的日志
func (api *MemoryHandlerAPI) GetContaining(text string, n int) []LogEntry {
	api.handler.mu.RLock()
	defer api.handler.mu.RUnlock()

	var result []LogEntry
	for _, entry := range api.handler.entries {
		if strings.Contains(entry.Event.Message, text) {
			result = append(result, entry)
		}
	}

	if n > 0 && n < len(result) {
		start := len(result) - n
		return result[start:]
	}

	return result
}

// SubscribeToLogs 订阅日志事件
func (api *MemoryHandlerAPI) SubscribeToLogs(bufferSize int) chan LogEvent {
	if bufferSize <= 0 {
		bufferSize = 10
	}

	listener := make(chan LogEvent, bufferSize)

	api.handler.listMu.Lock()
	api.handler.listeners = append(api.handler.listeners, listener)
	api.handler.listMu.Unlock()

	return listener
}

// UnsubscribeFromLogs 取消订阅日志事件
func (api *MemoryHandlerAPI) UnsubscribeFromLogs(listener chan LogEvent) {
	api.handler.listMu.Lock()
	defer api.handler.listMu.Unlock()

	for i, l := range api.handler.listeners {
		if l == listener {
			api.handler.listeners = append(api.handler.listeners[:i], api.handler.listeners[i+1:]...)
			close(listener)
			break
		}
	}
}
