package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// 上下文键类型
type contextKey string

// 上下文键
const (
	logContextKey contextKey = "logger_context"
)

// NewContext 创建新的日志上下文
func NewContext() *LogContext {
	return &LogContext{
		Tags: make(map[string]string),
	}
}

// LogContextFromContext 从context中获取日志上下文
func LogContextFromContext(ctx context.Context) *LogContext {
	if ctx == nil {
		return nil
	}
	if logCtx, ok := ctx.Value(logContextKey).(*LogContext); ok {
		return logCtx
	}
	return nil
}

// WithLogContext 将日志上下文添加到context中
func WithLogContext(ctx context.Context, logCtx *LogContext) context.Context {
	return context.WithValue(ctx, logContextKey, logCtx)
}

// WithTag 添加标签
func (c *LogContext) WithTag(key, value string) *LogContext {
	newCtx := &LogContext{
		TraceID:  c.TraceID,
		SpanID:   c.SpanID,
		ParentID: c.ParentID,
		Tags:     make(map[string]string),
	}

	// 复制现有标签
	for k, v := range c.Tags {
		newCtx.Tags[k] = v
	}

	// 添加新标签
	newCtx.Tags[key] = value

	return newCtx
}

// WithTrace 添加追踪ID
func (c *LogContext) WithTrace(traceID string) *LogContext {
	newCtx := &LogContext{
		TraceID:  traceID,
		SpanID:   c.SpanID,
		ParentID: c.ParentID,
		Tags:     make(map[string]string),
	}

	// 复制现有标签
	for k, v := range c.Tags {
		newCtx.Tags[k] = v
	}

	return newCtx
}

// WithSpan 添加跨度ID
func (c *LogContext) WithSpan(spanID string) *LogContext {
	newCtx := &LogContext{
		TraceID:  c.TraceID,
		SpanID:   spanID,
		ParentID: c.ParentID,
		Tags:     make(map[string]string),
	}

	// 复制现有标签
	for k, v := range c.Tags {
		newCtx.Tags[k] = v
	}

	return newCtx
}

// WithParent 添加父跨度ID
func (c *LogContext) WithParent(parentID string) *LogContext {
	newCtx := &LogContext{
		TraceID:  c.TraceID,
		SpanID:   c.SpanID,
		ParentID: parentID,
		Tags:     make(map[string]string),
	}

	// 复制现有标签
	for k, v := range c.Tags {
		newCtx.Tags[k] = v
	}

	return newCtx
}

// 生成唯一ID
func generateID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机数生成失败，使用时间戳作为备选
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// 全局跟踪上下文管理
var (
	traceInitOnce sync.Once
	traceInitFunc func()
)

// InitTraceContext 初始化追踪上下文
func InitTraceContext() {
	traceInitOnce.Do(func() {
		if traceInitFunc != nil {
			traceInitFunc()
		}
	})
}

// StartSpan 开始一个跟踪跨度
func StartSpan(ctx context.Context, name string) (context.Context, *LogContext) {
	var logCtx *LogContext

	// 从上下文中获取日志上下文
	parentLogCtx := LogContextFromContext(ctx)
	if parentLogCtx != nil {
		// 如果存在父上下文，继承追踪ID
		logCtx = &LogContext{
			TraceID:  parentLogCtx.TraceID,
			SpanID:   generateID(),
			ParentID: parentLogCtx.SpanID,
			Tags:     make(map[string]string),
		}
		logCtx.Tags["span_name"] = name
		logCtx.Tags["start_time"] = fmt.Sprintf("%d", time.Now().UnixNano())
	} else {
		// 如果不存在父上下文，创建新的追踪
		logCtx = &LogContext{
			TraceID:  generateID(),
			SpanID:   generateID(),
			ParentID: "",
			Tags:     make(map[string]string),
		}
		logCtx.Tags["span_name"] = name
		logCtx.Tags["start_time"] = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return WithLogContext(ctx, logCtx), logCtx
}

// FinishSpan 结束一个跟踪跨度
func FinishSpan(ctx context.Context) {
	logCtx := LogContextFromContext(ctx)
	if logCtx != nil {
		logCtx.Tags["end_time"] = fmt.Sprintf("%d", time.Now().UnixNano())
	}
}
