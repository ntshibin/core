package glog_test

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/ntshibin/core/glog"
	"github.com/ntshibin/core/glog/handlers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// 测试助手：创建自定义日志记录器
type testBuffer struct {
	bytes.Buffer
}

func (b *testBuffer) Close() error {
	return nil
}

// 自定义处理器实现
type testHandler struct {
	handlers.BaseHandler
	buffer *testBuffer
}

func (h *testHandler) Handle(logger *logrus.Logger, args ...interface{}) {
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			h.buffer.WriteString(msg)
		}
	}
	if h.Next != nil {
		h.Next.Handle(logger, args...)
	}
}

// 测试默认日志器
func TestDefaultLogger(t *testing.T) {
	// 初始化和配置是否执行
	logger := glog.GetLogger()
	assert.NotNil(t, logger)

	// 测试日志级别设置
	glog.SetLevel(logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, logger.GetLevel())

	// 测试日志级别设置回默认
	glog.SetLevel(logrus.InfoLevel)
}

// 测试日志级别过滤
func TestLogLevels(t *testing.T) {
	logger := glog.GetLogger()

	// 记录原日志级别，测试后恢复
	originalLevel := logger.GetLevel()
	defer logger.SetLevel(originalLevel)

	// 设置为Info级别，测试Debug级别日志不输出
	logger.SetLevel(logrus.InfoLevel)

	// 使用测试缓冲区而不是捕获标准输出
	buf := &testBuffer{}
	oldOutput := logger.Logger.Out
	logger.Logger.SetOutput(buf)
	defer logger.Logger.SetOutput(oldOutput)

	// 记录日志
	logger.Debug("此调试日志不应该输出")
	logger.Info("此信息日志应该输出")

	// 验证输出
	output := buf.String()
	assert.NotContains(t, output, "此调试日志不应该输出")
	assert.Contains(t, output, "此信息日志应该输出")
}

// 测试结构化日志
func TestStructuredLogging(t *testing.T) {
	logger := glog.GetLogger()

	// 使用测试缓冲区而不是捕获标准输出
	buf := &testBuffer{}
	oldOutput := logger.Logger.Out
	logger.Logger.SetOutput(buf)
	defer logger.Logger.SetOutput(oldOutput)

	// 记录结构化日志
	logger.WithField("用户", "张三").Info("用户登录")
	logger.WithFields(logrus.Fields{
		"IP地址": "192.168.1.1",
		"时间":   time.Now().Format(time.RFC3339),
	}).Info("访问记录")

	// 验证输出
	output := buf.String()
	assert.Contains(t, output, "用户登录")
	assert.Contains(t, output, "用户=\"张三\"")
	assert.Contains(t, output, "访问记录")
	assert.Contains(t, output, "IP地址=192.168.1.1")
}

// 测试自定义处理器
func TestCustomHandler(t *testing.T) {
	// 创建测试buffer
	buf := &testBuffer{}

	// 创建和配置日志记录器
	logger := &glog.Logger{
		Logger: logrus.New(),
	}
	logger.SetOutput(io.Discard) // 禁用默认输出

	// 自定义处理器 - 使用我们定义的测试处理器结构体
	customHandler := &testHandler{
		buffer: buf,
	}

	logger.AddHandler(customHandler)

	// 测试日志输出
	logger.Info("测试自定义处理器")

	assert.Contains(t, buf.String(), "测试自定义处理器")
}

// 测试静态方法
func TestStaticMethods(t *testing.T) {
	// 获取默认日志器并保存其输出
	logger := glog.GetLogger()
	oldOutput := logger.Logger.Out

	// 使用测试缓冲区
	buf := &testBuffer{}
	logger.Logger.SetOutput(buf)
	defer logger.Logger.SetOutput(oldOutput)

	// 使用静态方法记录日志
	glog.Info("静态信息日志")
	glog.WithField("模块", "测试").Error("静态错误日志")

	// 验证输出
	output := buf.String()
	assert.Contains(t, output, "静态信息日志")
	assert.Contains(t, output, "静态错误日志")
	assert.Contains(t, output, "模块=\"测试\"")
}

// 测试日志格式
func TestLogFormatting(t *testing.T) {
	// 获取默认日志器和保存其配置
	logger := glog.GetLogger()
	oldOutput := logger.Logger.Out
	oldFormatter := logger.Logger.Formatter

	// 使用标准输出进行目视检查
	logger.Logger.SetOutput(os.Stdout)

	// 测试结束时恢复原始设置
	defer func() {
		logger.Logger.SetOutput(oldOutput)
		logger.Logger.SetFormatter(oldFormatter)
	}()

	// 配置日志格式为JSON
	config := &glog.Config{
		Format: glog.FormatJSON,
	}

	// 应用配置
	err := glog.Configure(config)
	assert.NoError(t, err)

	// 记录测试日志以供人工检查
	t.Log("以下日志应以JSON格式输出:")
	glog.Info("JSON格式日志测试")

	// 检查格式化器类型
	_, isJSONFormatter := logger.Logger.Formatter.(*logrus.JSONFormatter)
	assert.True(t, isJSONFormatter, "日志格式化器应为JSONFormatter类型")
}

// 测试异步处理器
func TestAsyncHandler(t *testing.T) {
	// 保存原来的配置
	oldConfig := glog.DefaultConfig()
	defer glog.Configure(oldConfig)

	// 配置日志格式和异步处理
	config := &glog.Config{
		EnableAsync: true,
		AsyncConfig: &glog.AsyncConfig{
			BufferSize:    100,
			BatchSize:     10,
			FlushInterval: 100 * time.Millisecond,
		},
	}

	// 应用配置
	err := glog.Configure(config)
	assert.NoError(t, err)

	// 发送多条日志
	for i := 0; i < 20; i++ {
		glog.Info("异步测试日志")
	}

	// 优雅关闭，确保所有日志都被处理
	err = glog.Close()
	assert.NoError(t, err)
}

// 测试自定义配置
func TestCustomConfig(t *testing.T) {
	// 创建独立的测试日志记录器
	logger := &glog.Logger{
		Logger: logrus.New(),
	}

	// 记录原始日志级别
	originalLevel := logger.Logger.GetLevel()
	t.Logf("原始日志级别: %v", originalLevel)

	// 使用标准输出以便目视检查
	logger.Logger.SetOutput(os.Stdout)

	// 创建配置
	config := &glog.Config{
		Format:        glog.FormatText,
		DisableColors: true,
	}

	// 应用配置
	err := glog.ApplyConfig(logger, config)
	assert.NoError(t, err)

	// 直接设置日志级别
	logger.Logger.SetLevel(logrus.WarnLevel)

	// 验证级别设置
	assert.Equal(t, logrus.WarnLevel, logger.Logger.GetLevel(),
		"日志级别应该被设置为 WarnLevel")

	// 记录测试日志
	t.Log("以下日志只应显示警告和错误级别:")
	logger.Debug("调试日志-不应该输出")
	logger.Info("信息日志-不应该输出")
	logger.Warn("警告日志-应该输出")
	logger.Error("错误日志-应该输出")

	// 检查格式化器类型
	_, isTextFormatter := logger.Logger.Formatter.(*logrus.TextFormatter)
	assert.True(t, isTextFormatter, "日志格式化器应为TextFormatter类型")
}

// 计数处理器用于测试链式调用
type countingHandler struct {
	handlers.BaseHandler
	Count *int
}

// 实现 Handle 方法，每次调用时增加计数
func (h *countingHandler) Handle(logger *logrus.Logger, args ...interface{}) {
	*h.Count++
	if h.Next != nil {
		h.Next.Handle(logger, args...)
	}
}

// 测试链式日志
func TestChainLogging(t *testing.T) {
	// 创建计数器
	var count1, count2 int

	// 创建独立的测试日志记录器
	logger := &glog.Logger{
		Logger: logrus.New(),
	}

	// 禁用默认输出
	logger.SetOutput(io.Discard)

	// 创建两个计数处理器
	handler1 := &countingHandler{Count: &count1}
	handler2 := &countingHandler{Count: &count2}

	// 设置链式关系
	handler1.Next = handler2

	// 添加处理器到日志记录器
	logger.AddHandler(handler1)

	// 发送测试日志
	logger.Info("测试链式处理器1")
	logger.Error("测试链式处理器2")

	// 验证两个处理器都被调用了正确的次数
	assert.Equal(t, 2, count1, "第一个处理器应被调用2次")
	assert.Equal(t, 2, count2, "第二个处理器应被调用2次")
}
