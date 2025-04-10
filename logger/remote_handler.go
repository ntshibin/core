package logger

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// RemoteDestination 远程日志目标类型
type RemoteDestination string

const (
	// HTTPDestination HTTP目标
	HTTPDestination RemoteDestination = "http"
	// TCPDestination TCP目标
	TCPDestination RemoteDestination = "tcp"
)

// RemoteConfig 远程日志配置
type RemoteConfig struct {
	// Destination 目标类型：http 或 tcp
	Destination RemoteDestination `yaml:"destination" json:"destination"`
	// Address 远程服务器地址，如 https://log-server.example.com/logs 或 log-server.example.com:9000
	Address string `yaml:"address" json:"address"`
	// Timeout 请求超时时间，单位毫秒
	Timeout int `yaml:"timeout" json:"timeout"`
	// BatchSize 批处理大小
	BatchSize int `yaml:"batch_size" json:"batch_size"`
	// RetryCount 重试次数
	RetryCount int `yaml:"retry_count" json:"retry_count"`
	// RetryInterval 重试间隔，单位毫秒
	RetryInterval int `yaml:"retry_interval" json:"retry_interval"`
	// Headers HTTP请求头
	Headers map[string]string `yaml:"headers" json:"headers"`
}

// DefaultRemoteConfig 默认远程配置
var DefaultRemoteConfig = RemoteConfig{
	Destination:   HTTPDestination,
	Address:       "http://localhost:8080/logs",
	Timeout:       3000,
	BatchSize:     10,
	RetryCount:    3,
	RetryInterval: 1000,
	Headers: map[string]string{
		"Content-Type": "application/json",
	},
}

// RemoteHandler 远程日志处理器
type RemoteHandler struct {
	*BaseHandler
	config     RemoteConfig
	buffer     []LogEvent
	client     *http.Client
	bufferLock sync.Mutex
	timer      *time.Timer
	closed     bool
}

// NewRemoteHandler 创建远程日志处理器
func NewRemoteHandler(formatter Formatter, level LogLevel, config RemoteConfig) (*RemoteHandler, error) {
	// 验证配置
	if config.Address == "" {
		return nil, fmt.Errorf("远程地址不能为空")
	}

	if config.Destination == HTTPDestination {
		_, err := url.Parse(config.Address)
		if err != nil {
			return nil, fmt.Errorf("无效的HTTP地址: %v", err)
		}
	}

	// 设置默认值
	if config.Timeout <= 0 {
		config.Timeout = 3000
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 10
	}
	if config.RetryCount < 0 {
		config.RetryCount = 3
	}
	if config.RetryInterval <= 0 {
		config.RetryInterval = 1000
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Millisecond,
	}

	h := &RemoteHandler{
		BaseHandler: NewBaseHandler(formatter, level),
		config:      config,
		buffer:      make([]LogEvent, 0, config.BatchSize),
		client:      client,
	}

	// 启动定时发送
	h.timer = time.AfterFunc(time.Second*5, h.sendBatch)

	return h, nil
}

// Handle 处理日志事件
func (h *RemoteHandler) Handle(event LogEvent) error {
	if !h.ShouldHandle(event) {
		return nil
	}

	h.bufferLock.Lock()
	defer h.bufferLock.Unlock()

	if h.closed {
		return fmt.Errorf("handler已关闭")
	}

	// 添加事件到缓冲区
	h.buffer = append(h.buffer, event)

	// 如果达到批处理大小，发送日志
	if len(h.buffer) >= h.config.BatchSize {
		go h.sendBatch()
	}

	return nil
}

// sendBatch 发送批量日志
func (h *RemoteHandler) sendBatch() {
	h.bufferLock.Lock()

	// 如果已关闭或缓冲区为空，不发送
	if h.closed || len(h.buffer) == 0 {
		h.bufferLock.Unlock()
		return
	}

	// 获取当前缓冲区数据并清空
	events := make([]LogEvent, len(h.buffer))
	copy(events, h.buffer)
	h.buffer = h.buffer[:0]

	// 重置定时器
	if !h.closed {
		h.timer.Reset(time.Second * 5)
	}

	h.bufferLock.Unlock()

	// 根据目标类型发送日志
	var err error
	switch h.config.Destination {
	case HTTPDestination:
		err = h.sendHTTP(events)
	case TCPDestination:
		err = h.sendTCP(events)
	default:
		err = fmt.Errorf("不支持的目标类型: %s", h.config.Destination)
	}

	if err != nil {
		// 简单打印错误，实际应用中可以有更复杂的错误处理
		fmt.Printf("发送远程日志失败: %v\n", err)
	}
}

// sendHTTP 通过HTTP发送日志
func (h *RemoteHandler) sendHTTP(events []LogEvent) error {
	// 格式化所有事件
	jsonData, err := h.formatBatch(events)
	if err != nil {
		return err
	}

	// 创建请求
	req, err := http.NewRequest("POST", h.config.Address, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// 添加请求头
	for key, value := range h.config.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求，带重试
	var resp *http.Response
	for i := 0; i <= h.config.RetryCount; i++ {
		resp, err = h.client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if i < h.config.RetryCount {
			time.Sleep(time.Duration(h.config.RetryInterval) * time.Millisecond)
		}
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// sendTCP 通过TCP发送日志
func (h *RemoteHandler) sendTCP(events []LogEvent) error {
	// 格式化所有事件
	jsonData, err := h.formatBatch(events)
	if err != nil {
		return err
	}

	// 解析地址
	addr := h.config.Address
	if addr == "" {
		return fmt.Errorf("TCP地址不能为空")
	}

	// 建立连接
	conn, err := net.DialTimeout("tcp", addr, time.Duration(h.config.Timeout)*time.Millisecond)
	if err != nil {
		return fmt.Errorf("TCP连接失败: %v", err)
	}
	defer conn.Close()

	// 设置写入超时
	err = conn.SetWriteDeadline(time.Now().Add(time.Duration(h.config.Timeout) * time.Millisecond))
	if err != nil {
		return fmt.Errorf("设置TCP写入超时失败: %v", err)
	}

	// 写入数据，确保完整发送
	bytesWritten, err := conn.Write(jsonData)
	if err != nil {
		return fmt.Errorf("TCP数据发送失败: %v", err)
	}

	if bytesWritten != len(jsonData) {
		return fmt.Errorf("TCP数据发送不完整: 已发送 %d, 总共 %d", bytesWritten, len(jsonData))
	}

	return nil
}

// formatBatch 批量格式化日志事件
func (h *RemoteHandler) formatBatch(events []LogEvent) ([]byte, error) {
	// 简单将所有JSON拼接成数组
	var buffer bytes.Buffer
	buffer.WriteString("[")

	for i, event := range events {
		data, err := h.Format(event)
		if err != nil {
			return nil, err
		}

		// 如果数据末尾有换行符，需要去除
		if len(data) > 0 && data[len(data)-1] == '\n' {
			data = data[:len(data)-1]
		}

		buffer.Write(data)

		if i < len(events)-1 {
			buffer.WriteString(",")
		}
	}

	buffer.WriteString("]")
	return buffer.Bytes(), nil
}

// Close 关闭处理器
func (h *RemoteHandler) Close() error {
	h.bufferLock.Lock()
	defer h.bufferLock.Unlock()

	if h.closed {
		return nil
	}

	h.closed = true
	h.timer.Stop()

	// 发送剩余的日志
	if len(h.buffer) > 0 {
		go h.sendBatch()
	}

	return nil
}
