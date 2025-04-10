package logger

import (
	"fmt"
	"sync"
)

// AsyncHandler 异步处理器
type AsyncHandler struct {
	handler   Handler
	queue     chan LogEvent
	wg        sync.WaitGroup
	closeOnce sync.Once
	closed    bool
	mu        sync.RWMutex
}

// NewAsyncHandler 创建异步处理器
func NewAsyncHandler(handler Handler, queueSize int) *AsyncHandler {
	if queueSize <= 0 {
		queueSize = 1000
	}

	h := &AsyncHandler{
		handler: handler,
		queue:   make(chan LogEvent, queueSize),
	}

	// 启动工作协程
	h.wg.Add(1)
	go h.worker()

	return h
}

// worker 处理日志事件工作协程
func (h *AsyncHandler) worker() {
	defer h.wg.Done()

	for event := range h.queue {
		_ = h.handler.Handle(event)
	}
}

// Handle 处理日志事件
func (h *AsyncHandler) Handle(event LogEvent) error {
	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		return fmt.Errorf("handler已关闭")
	}
	h.mu.RUnlock()

	// 非阻塞发送，避免队列满导致应用程序阻塞
	select {
	case h.queue <- event:
		return nil
	default:
		return fmt.Errorf("队列已满，丢弃事件")
	}
}

// Format 格式化日志事件
func (h *AsyncHandler) Format(event LogEvent) ([]byte, error) {
	return h.handler.Format(event)
}

// ShouldHandle 是否应该处理该事件
func (h *AsyncHandler) ShouldHandle(event LogEvent) bool {
	return h.handler.ShouldHandle(event)
}

// Close 关闭处理器
func (h *AsyncHandler) Close() error {
	var err error
	h.closeOnce.Do(func() {
		h.mu.Lock()
		h.closed = true
		close(h.queue)
		h.mu.Unlock()

		// 等待所有事件处理完成
		h.wg.Wait()

		// 关闭内部处理器
		err = h.handler.Close()
	})
	return err
}

// Sync 同步等待所有事件处理完成
func (h *AsyncHandler) Sync() error {
	// 创建一个同步通道
	done := make(chan struct{})

	// 发送一个特殊事件，用于标记队列末尾
	syncEvent := LogEvent{
		Message: "_SYNC_", // 特殊消息，只用于同步
	}

	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		return fmt.Errorf("handler已关闭")
	}
	h.mu.RUnlock()

	// 创建处理同步事件的协程
	go func() {

		// 监听队列中的特殊事件
		for event := range h.queue {
			if event.Message == "_SYNC_" {
				close(done)
				return
			}
			_ = h.handler.Handle(event)
		}
	}()

	// 发送同步事件
	select {
	case h.queue <- syncEvent:
		// 等待同步完成
		<-done
		return nil
	default:
		return fmt.Errorf("队列已满，无法同步")
	}
}
