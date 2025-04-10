# 日志系统设计文档

## 1. 概述

这是一个基于设计模式的可扩展日志系统，支持多种输出目标、多种日志格式以及灵活的配置方式。该系统采用模块化设计，易于扩展和维护。

## 2. 设计原则

- **开闭原则**：系统易于扩展，无需修改已有代码
- **单一职责**：每个组件只负责单一功能
- **依赖倒置**：高层模块不依赖低层模块，二者都依赖抽象
- **接口隔离**：客户端不应依赖它不需要的接口
- **组合优于继承**：通过组合实现功能复用

## 3. 设计模式应用

### 3.1 策略模式

用于实现不同的日志处理器和格式化器策略。

- `Handler` 接口定义了日志处理策略
- `Formatter` 接口定义了日志格式化策略
- 客户端可以动态切换不同的处理器和格式化器

### 3.2 工厂模式

用于创建日志记录器及其组件。

- `LoggerFactory` 接口定义了创建日志记录器的方法
- `StandardLoggerFactory` 是标准实现

### 3.3 建造者模式

用于构建复杂的日志记录器配置。

- `BuilderInterface` 接口定义了建造者方法
- `StandardLoggerBuilder` 是标准实现
- 可以链式调用，逐步构建完整的日志记录器

### 3.4 单例模式

用于确保全局只有一个日志管理器。

- `LogManager` 是单例类
- `GetLogManager()` 确保返回唯一实例

### 3.5 装饰器模式

用于为日志记录器添加额外功能。

- 通过 `WithField` 和 `WithFields` 方法动态添加上下文
- 每次添加字段会返回新的记录器实例，不改变原实例

### 3.6 观察者模式

用于实现日志事件的分发。

- 日志事件通过多个处理器进行处理
- 每个处理器可以独立决定如何处理事件

## 4. 核心组件

### 4.1 接口定义

- `LoggerInterface`: 日志记录器接口
- `Handler`: 日志处理器接口
- `Formatter`: 日志格式化器接口
- `LoggerFactory`: 日志工厂接口
- `BuilderInterface`: 日志建造者接口

### 4.2 处理器实现

- `ConsoleHandler`: 控制台输出
- `FileHandler`: 文件输出
- `RotateFileHandler`: 轮转文件输出
- `MultiHandler`: 多目标输出
- `AsyncHandler`: 异步处理器
- `RemoteHandler`: 远程日志处理器
- `MemoryHandler`: 内存日志处理器
- `CustomHandler`: 自定义输出

### 4.3 格式化器实现

- `TextFormatter`: 文本格式
- `JSONFormatter`: JSON 格式
- `CustomFormatter`: 自定义格式

### 4.4 日志记录器实现

- `StandardLogger`: 标准日志记录器

### 4.5 工厂和建造者

- `StandardLoggerFactory`: 标准日志工厂
- `StandardLoggerBuilder`: 标准日志建造者

### 4.6 管理器

- `LogManager`: 日志管理器，管理多个日志记录器实例

## 5. 核心 API 概览

### 5.1 日志级别

```go
const (
    DebugLevel LogLevel = iota // 调试信息
    InfoLevel                  // 一般信息
    WarnLevel                  // 警告信息
    ErrorLevel                 // 错误信息
    FatalLevel                 // 致命错误
)
```

### 5.2 全局函数

```go
// 基本日志输出
Debug(msg string)
Info(msg string)
Warn(msg string)
Error(msg string)
Fatal(msg string) // 会导致程序退出

// 添加字段
WithField(key string, value interface{}) LoggerInterface
WithFields(fields map[string]interface{}) LoggerInterface

// 日志级别管理
SetLevel(level LogLevel)
GetLevel() LogLevel

// 日志同步和关闭
Sync() error
ShutdownSystem() error

// 初始化
Init(config LoggerConfig) error
InitSystem(config LoggerConfig) error
```

### 5.3 处理器 API

```go
// 添加各类处理器
AddConsoleHandler(level LogLevel)
AddFileHandler(filePath string, level LogLevel) error
AddRotateFileHandler(config FileRotateConfig, level LogLevel) error
AddAsyncHandler(handler Handler, queueSize int) error
AddRemoteHandler(config RemoteConfig, level LogLevel) error

// 创建内存日志记录器
CreateMemoryLogger(config MemoryConfig, level LogLevel) (*MemoryHandlerAPI, error)
```

### 5.4 上下文相关 API

```go
// 上下文操作
WithContext(ctx context.Context) LoggerInterface
LoggerFromContext(ctx context.Context) LoggerInterface
CreateTraceContext() context.Context
StartLogSpan(ctx context.Context, name string) (context.Context, LoggerInterface)
FinishLogSpan(ctx context.Context)
```

## 6. 使用示例

### 6.1 基本使用

```go
package main

import "github.com/ntshibin/core/logger"

func main() {
    // 使用默认配置初始化
    logger.Init(logger.DefaultLoggerConfig)

    // 输出各级别日志
    logger.Debug("这是一条Debug日志")
    logger.Info("这是一条Info日志")
    logger.Warn("这是一条Warn日志")
    logger.Error("这是一条Error日志")
    // logger.Fatal("这是一条Fatal日志") // 会导致程序终止

    // 程序退出前同步日志
    logger.Sync()
}
```

### 6.2 完整配置示例

```go
package main

import (
    "time"
    "github.com/ntshibin/core/logger"
)

func main() {
    // 创建完整配置
    config := logger.LoggerConfig{
        Name:          "app-logger",
        Level:         "debug",
        Encoding:      "json",
        EnableConsole: true,
        EnableFile:    false,
        EnableRotate:  true,
        RotateConfig: logger.FileRotateConfig{
            FilePath:   "logs/app.log",
            MaxSize:    100,
            MaxBackups: 10,
            MaxAge:     30,
            Compress:   true,
        },
        EnableAsync:    true,
        AsyncQueueSize: 1000,
        EnableMemory:   true,
        MemoryConfig: logger.MemoryConfig{
            Capacity:        1000,
            ExpireTime:      time.Hour * 24,
            CleanupInterval: time.Minute * 10,
        },
        EnableTrace: true,
        CallerSkip:  false,
    }

    // 初始化日志系统
    err := logger.InitSystem(config)
    if err != nil {
        panic(err)
    }
    defer logger.ShutdownSystem()

    // 基本日志记录
    logger.Info("系统已初始化")
}
```

### 6.3 添加上下文字段

```go
// 添加单个字段
logger.WithField("user_id", "12345").Info("用户登录")

// 添加多个字段
fields := map[string]interface{}{
    "user_id":   "12345",
    "ip":        "192.168.1.1",
    "action":    "login",
    "timestamp": time.Now().Unix(),
}
logger.WithFields(fields).Info("用户操作记录")
```

### 6.4 使用调用链跟踪

```go
package main

import (
    "context"
    "github.com/ntshibin/core/logger"
)

func main() {
    // 初始化带跟踪功能的日志
    config := logger.LoggerConfig{
        Name:         "app",
        Level:        "debug",
        Encoding:     "json",
        EnableTrace:  true,
        EnableConsole: true,
    }
    logger.InitSystem(config)
    defer logger.ShutdownSystem()

    // 创建跟踪上下文
    ctx := logger.CreateTraceContext()

    // 处理请求
    handleRequest(ctx)
}

func handleRequest(ctx context.Context) {
    // 开始一个新的跨度
    ctx, log := logger.StartLogSpan(ctx, "handleRequest")
    defer logger.FinishLogSpan(ctx)

    log.WithField("path", "/api/users").Info("收到请求")

    // 处理业务逻辑
    validateRequest(ctx)
    processData(ctx)

    log.Info("请求处理完成")
}

func validateRequest(ctx context.Context) {
    ctx, log := logger.StartLogSpan(ctx, "validateRequest")
    defer logger.FinishLogSpan(ctx)

    log.Info("验证请求")
}

func processData(ctx context.Context) {
    ctx, log := logger.StartLogSpan(ctx, "processData")
    defer logger.FinishLogSpan(ctx)

    log.Info("处理数据")
}
```

### 6.5 使用 YAML 配置文件

```yaml
# logger_config.yaml
log:
  name: "app-logger"
  level: "debug"
  encoding: "json"
  caller_skip: true

  enable_console: true

  enable_file: false
  file_path: "logs/app.log"

  enable_rotate: true
  rotate:
    file_path: "logs/app.log"
    max_size: 100
    max_backups: 10
    max_age: 30
    compress: true

  enable_async: true
  async_queue_size: 1000

  enable_memory: true
  memory:
    capacity: 1000
    expire_time: 86400
    cleanup_interval: 600

  enable_trace: true
```

```go
package main

import (
    "github.com/ntshibin/core/conf"
    "github.com/ntshibin/core/logger"
)

func main() {
    // 从配置文件加载配置
    var config struct {
        Log logger.LoggerConfig `yaml:"log"`
    }
    conf.LoadConfig("configs/logger_config.yaml", &config)

    // 初始化日志系统
    logger.InitSystem(config.Log)
    defer logger.ShutdownSystem()

    logger.Info("系统已使用配置文件初始化")
}
```

### 6.6 内存日志查询和监听

```go
// 创建内存日志API
memConfig := logger.MemoryConfig{
    Capacity:        1000,
    ExpireTime:      time.Hour * 24,
    CleanupInterval: time.Minute * 5,
}
memAPI, _ := logger.CreateMemoryLogger(memConfig, logger.InfoLevel)

// 记录一些日志
for i := 0; i < 10; i++ {
    logger.WithField("index", i).Info("测试日志")
}

// 查询最新的5条日志
latestLogs := memAPI.GetLatest(5)
for _, log := range latestLogs {
    fmt.Printf("[%s] %s\n", log.Event.Level, log.Event.Message)
}

// 实时监听日志
logChan := memAPI.SubscribeToLogs(10)
go func() {
    for event := range logChan {
        fmt.Printf("实时日志: [%s] %s\n", event.Level, event.Message)
    }
}()

// 使用一段时间后取消订阅
time.Sleep(time.Second * 10)
memAPI.UnsubscribeFromLogs(logChan)
```

## 7. 高级功能

### 7.1 异步日志

异步日志处理器可以将日志记录操作与应用程序主线程分离，避免日志记录阻塞主业务流程。

```go
// 创建文件处理器并包装为异步处理器
fileHandler, _ := logger.NewFileHandler(logger.NewJSONFormatter(), logger.InfoLevel, "logs/app.log")
logger.AddAsyncHandler(fileHandler, 1000)

// 记录日志 - 不会阻塞
logger.Info("异步日志示例")

// 应用程序退出前确保日志写入
logger.Sync()
```

### 7.2 远程日志

```go
// 创建HTTP远程日志配置
httpConfig := logger.RemoteConfig{
    Destination:   logger.HTTPDestination,
    Address:       "https://log-server.example.com/logs",
    Timeout:       3000, // 毫秒
    BatchSize:     100,
    RetryCount:    3,
    RetryInterval: 1000, // 毫秒
    Headers: map[string]string{
        "Content-Type":  "application/json",
        "Authorization": "Bearer token123",
    },
}

// 添加远程日志处理器
logger.AddRemoteHandler(httpConfig, logger.WarnLevel)
```

## 8. 性能考虑

### 8.1 异步日志最佳实践

- 对于高性能要求的应用，使用异步日志处理器
- 设置合适的队列大小，避免内存占用过大
- 程序退出前务必调用`Sync()`确保所有日志被处理

### 8.2 内存日志注意事项

- 合理设置容量，避免占用过多内存
- 设置适当的过期时间，自动清理旧日志
- 高并发场景谨慎使用监听器，可能造成性能瓶颈

### 8.3 远程日志配置建议

- 设置批处理大小，减少网络请求次数
- 配置合理的重试策略，提高可靠性
- 考虑使用异步处理器包装远程处理器，进一步降低延迟

### 8.4 结构化上下文优化

- 仅在关键路径上使用追踪和跨度功能
- 避免在循环中创建大量跨度
- 合理使用字段，避免记录过多无用信息

## 9. 最佳实践

1. **使用上下文**: 总是使用 `WithField` 和 `WithFields` 添加上下文信息
2. **适当级别**: 选择合适的日志级别，避免日志过多或过少
3. **结构化日志**: 优先使用结构化日志（JSON 格式），便于分析和处理
4. **合理配置**: 生产环境建议使用文件轮转，避免日志文件过大
5. **日志同步**: 程序退出前调用 `Sync()` 确保日志写入
6. **错误处理**: 正确处理日志相关的错误，避免循环记录
7. **命名规范**: 为不同模块创建不同的日志记录器，便于定位问题

## 10. 故障排除

### 10.1 常见问题

1. **日志没有输出到文件**

   - 检查文件路径是否正确
   - 确认应用有写入权限
   - 检查是否正确配置了文件处理器

2. **日志级别过滤不生效**

   - 确认正确设置了日志级别
   - 检查处理器级别是否与期望一致

3. **异步日志丢失**

   - 确保程序退出前调用 `Sync()` 和 `ShutdownSystem()`
   - 检查队列大小是否足够

4. **远程日志发送失败**
   - 检查网络连接
   - 验证服务器地址和端口
   - 检查认证凭据是否正确

### 10.2 调试技巧

- 临时将日志级别设置为 `DebugLevel` 以获取更多信息
- 使用控制台处理器辅助调试
- 检查日志文件权限和磁盘空间

## 快速启用文件日志

如果您的日志没有写入文件，您可以使用以下方式快速启用文件日志功能：

### 方式一：使用便捷方法

```go
package main

import (
    "github.com/ntshibin/core/logger"
)

func main() {
    // 初始化日志系统，启用文件日志
    err := logger.InitWithFileLog("info", "logs/app.log")
    if err != nil {
        panic(err)
    }

    // 使用日志
    logger.Info("启动HTTP服务器")
    logger.Error("发生错误")
}
```

### 方式二：使用配置对象

```go
package main

import (
    "github.com/ntshibin/core/logger"
)

func main() {
    // 创建配置
    config := logger.DefaultLoggerConfig
    config.EnableFile = true        // 启用文件日志
    config.FilePath = "logs/app.log" // 设置文件路径

    // 初始化日志系统
    err := logger.Init(config)
    if err != nil {
        panic(err)
    }

    // 使用日志
    logger.Info("启动HTTP服务器")
    logger.Error("发生错误")
}
```

### 方式三：使用轮转文件日志

对于生产环境，建议使用轮转文件日志，避免日志文件过大：

```go
package main

import (
    "github.com/ntshibin/core/logger"
)

func main() {
    // 创建轮转配置
    rotateConfig := logger.DefaultFileRotateConfig
    rotateConfig.FilePath = "logs/app.log"
    rotateConfig.MaxSize = 100     // 100MB
    rotateConfig.MaxBackups = 10   // 保留10个备份

    // 初始化日志系统，启用轮转文件日志
    err := logger.InitWithRotateLog("info", rotateConfig)
    if err != nil {
        panic(err)
    }

    // 使用日志
    logger.Info("启动HTTP服务器")
    logger.Error("发生错误")
}
```

## 异步/同步日志模式

logger 支持三种日志模式：

1. **同步模式**：所有日志直接写入目标，适合调试和简单应用
2. **异步模式**：所有日志通过队列异步写入，提高性能，适合高负载应用
3. **混合模式**：I/O 密集型处理器（文件、网络）使用异步，其他使用同步，平衡性能和实时性

### 通过配置文件设置

```yaml
log:
  # 启用异步模式
  enable_async: true
  # 异步队列大小
  async_queue_size: 1000
```

### 通过代码设置

```go
// 设置全异步模式
logger.SetAsyncLogging(true)

// 设置同步模式
logger.SetAsyncLogging(false)

// 设置混合模式
logger.SetMixedLogging()

// 自定义配置
logger.ConfigureAsync(1, 2000) // 异步模式，队列大小2000
```

### 异步模式说明

- **同步模式 (0)**：所有日志同步写入，日志不会丢失，但可能影响应用性能
- **异步模式 (1)**：所有日志异步写入，提高性能，但队列满时可能丢失日志
- **混合模式 (2)**：优先考虑控制台实时性，同时提高文件/网络日志性能

### 程序退出前同步日志

在应用程序结束前，建议调用 Sync()来确保所有异步日志都已写入：

```go
func main() {
    // 应用逻辑...

    // 程序结束前同步日志
    logger.Sync()
}
```
