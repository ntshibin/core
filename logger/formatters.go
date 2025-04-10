package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// JSONFormatter JSON格式化器
type JSONFormatter struct{}

// NewJSONFormatter 创建JSON格式化器
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format 格式化日志事件为JSON
func (f *JSONFormatter) Format(event LogEvent) ([]byte, error) {
	data := make(map[string]interface{})

	// 添加基本字段
	data["timestamp"] = time.Unix(0, event.Time).Format(time.RFC3339Nano)
	data["level"] = levelToString(event.Level)
	data["message"] = event.Message
	data["pid"] = os.Getpid() // 添加进程ID

	// 添加调用者信息
	if event.Caller != "" {
		data["caller"] = event.Caller
	}

	// 添加上下文信息
	if event.Context != nil {
		if event.Context.TraceID != "" {
			data["trace_id"] = event.Context.TraceID
		}
		if event.Context.SpanID != "" {
			data["span_id"] = event.Context.SpanID
		}
		if event.Context.ParentID != "" {
			data["parent_id"] = event.Context.ParentID
		}
		if len(event.Context.Tags) > 0 {
			data["tags"] = event.Context.Tags
		}
	}

	// 添加自定义字段
	for k, v := range event.Fields {
		// 避免覆盖基本字段
		if _, exists := data[k]; !exists {
			data[k] = v
		}
	}

	// 转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 添加换行符
	return append(jsonData, '\n'), nil
}

// TextFormatter 文本格式化器
type TextFormatter struct{}

// NewTextFormatter 创建文本格式化器
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format 格式化日志事件为文本
func (f *TextFormatter) Format(event LogEvent) ([]byte, error) {
	// 基本信息
	timeStr := time.Unix(0, event.Time).Format("2006-01-02 15:04:05.000")
	levelStr := levelToString(event.Level)
	pid := os.Getpid() // 获取进程ID

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[%s] [%s] [pid:%d] %s", timeStr, levelStr, pid, event.Message))

	// 添加调用者信息
	if event.Caller != "" {
		builder.WriteString(fmt.Sprintf(" [%s]", event.Caller))
	}

	// 添加上下文信息
	if event.Context != nil {
		if event.Context.TraceID != "" {
			builder.WriteString(fmt.Sprintf(" trace=%s", event.Context.TraceID))
		}
		if event.Context.SpanID != "" {
			builder.WriteString(fmt.Sprintf(" span=%s", event.Context.SpanID))
		}
	}

	// 添加自定义字段
	if len(event.Fields) > 0 {
		builder.WriteString(" | ")
		first := true
		for k, v := range event.Fields {
			if !first {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
	}

	builder.WriteString("\n")
	return []byte(builder.String()), nil
}

// levelToString 将日志级别转换为字符串
func levelToString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel 解析日志级别字符串
func ParseLevel(level string) (LogLevel, error) {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return InfoLevel, fmt.Errorf("无效的日志级别: %s", level)
	}
}
