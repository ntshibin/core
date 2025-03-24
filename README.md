# Go 企业核心包开发指南

## 项目简介

本核心包旨在为企业提供一套高质量、可复用的基础组件库，遵循 Go 语言最佳实践，确保代码的可维护性、可扩展性和性能。

### 核心特性

- **高性能**：经过严格的性能优化和基准测试
- **可扩展**：模块化设计，支持灵活扩展
- **高可靠**：完善的错误处理和故障恢复机制
- **易使用**：简洁的 API 设计，详尽的文档说明

### 主要模块

1. **配置管理**

   - 支持多种配置源（文件、环境变量、远程配置中心）
   - 动态配置更新
   - 配置加密支持

2. **日志系统**

   - 分级日志
   - 异步日志写入
   - 日志轮转
   - 结构化日志支持

3. **数据库操作**

   - 连接池管理
   - 事务支持
   - SQL 注入防护
   - 读写分离

4. **缓存系统**

   - 多级缓存
   - 分布式缓存
   - 缓存一致性保证

5. **HTTP 客户端**

   - 连接池管理
   - 重试机制
   - 熔断器
   - 负载均衡

6. **安全组件**
   - 加密解密
   - 数字签名
   - JWT 支持
   - RBAC 权限控制

### 性能基准

以下是核心组件的性能测试数据（基于标准硬件配置）：

```
组件             操作               QPS        延迟(P99)
----------------------------------------------------
日志系统         异步写入           100K       <1ms
数据库连接池     查询操作           50K        <5ms
缓存系统         读取              200K       <0.5ms
 HTTP客户端      并发请求          80K        <10ms
```

## 项目架构

### 1. 整体架构

```
+------------------+     +------------------+     +------------------+
|     应用层       |     |    核心模块层    |     |    基础设施层    |
+------------------+     +------------------+     +------------------+
| - 业务逻辑       |     | - 配置管理      |     | - 数据库        |
| - API接口        |     | - 日志系统      |     | - 缓存          |
| - 中间件         |     | - 安全组件      |     | - 消息队列      |
+------------------+     +------------------+     +------------------+
           |                      |                       |
           v                      v                       v
+------------------+     +------------------+     +------------------+
|    公共组件层    |     |    工具层        |     |    监控层       |
+------------------+     +------------------+     +------------------+
| - 错误处理      |      | - 字符串处理    |     | - 性能监控      |
| - 数据验证      |      | - 时间工具      |     | - 链路追踪      |
| - 类型转换      |      | - 加密工具      |     | - 健康检查      |
+------------------+     +------------------+     +------------------+
```

### 2. 技术选型

#### 基础框架

- **语言版本**：Go 1.23.3+
- **包管理**：Go Modules
- **编码规范**：Uber Go Style Guide

#### 核心组件

- **配置中心**：Viper + Etcd
- **日志框架**：Zap + Lumberjack
- **数据库**：GORM + MySQL
- **缓存**：Go-Redis
- **消息队列**：Kafka
- **错误处理**：gerror

#### 监控告警

- **指标收集**：Prometheus
- **链路追踪**：Jaeger
- **日志聚合**：ELK Stack

## 日志记录规范

为确保系统的可观测性和问题排查能力，项目采用 glog 作为统一的日志记录框架。开发人员在编码过程中应遵循以下日志记录规范：

### 1. 日志级别使用原则

- **Debug 级别**：仅在开发环境使用，记录详细的流程信息和变量值
  - 函数入口参数和返回值
  - 条件分支选择过程
  - 循环迭代中的关键变量变化
- **Info 级别**：记录系统正常运行中的关键节点信息

  - 应用启动和停止
  - 配置加载完成
  - 关键业务流程的开始和完成
  - 定时任务的执行情况

- **Warn 级别**：记录不影响系统正常运行，但需要关注的异常情况

  - 接口调用重试
  - 性能下降预警
  - 资源使用率超过阈值
  - 配置项过期或不推荐使用

- **Error 级别**：记录影响功能但不影响系统整体运行的错误

  - 数据库查询失败
  - 缓存操作失败
  - 外部 API 调用失败
  - 业务规则校验失败

- **Fatal/Panic 级别**：记录导致系统无法正常工作需要立即干预的严重问题
  - 系统初始化失败
  - 关键配置缺失或错误
  - 核心组件连接失败

### 2. 关键场景日志规范

#### 系统启动和关闭

```go
// 应用启动
func main() {
    glog.WithFields(logrus.Fields{
        "版本": AppVersion,
        "环境": Env,
        "配置": ConfigPath,
    }).Info("应用程序启动")

    // 应用关闭
    defer func() {
        glog.Info("应用程序正常关闭")
    }()

    // ...
}
```

#### 接口入口和出口

```go
func HandleRequest(ctx context.Context, req *Request) (*Response, error) {
    // 入口日志
    glog.WithFields(logrus.Fields{
        "请求ID": req.RequestID,
        "用户ID": req.UserID,
        "操作": req.Action,
    }).Info("接收到请求")

    // 处理逻辑...

    // 出口日志
    if err != nil {
        glog.WithFields(logrus.Fields{
            "请求ID": req.RequestID,
            "错误": err.Error(),
        }).Error("请求处理失败")
        return nil, err
    }

    glog.WithFields(logrus.Fields{
        "请求ID": req.RequestID,
        "耗时": time.Since(startTime).Milliseconds(),
    }).Info("请求处理完成")

    return resp, nil
}
```

#### 数据库操作

```go
func (r *Repository) GetUserByID(ctx context.Context, userID string) (*User, error) {
    glog.WithField("用户ID", userID).Debug("开始查询用户信息")

    user, err := r.db.QueryUser(ctx, userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            glog.WithField("用户ID", userID).Warn("用户不存在")
            return nil, ErrUserNotFound
        }

        glog.WithFields(logrus.Fields{
            "用户ID": userID,
            "错误": err.Error(),
        }).Error("数据库查询失败")

        return nil, fmt.Errorf("查询用户失败: %w", err)
    }

    glog.WithField("用户ID", userID).Debug("用户信息查询成功")
    return user, nil
}
```

#### 缓存操作

```go
func (c *Cache) GetValue(ctx context.Context, key string) (interface{}, error) {
    glog.WithField("缓存键", key).Debug("尝试从缓存获取数据")

    value, err := c.client.Get(ctx, key).Result()
    if err != nil {
        if errors.Is(err, redis.Nil) {
            glog.WithField("缓存键", key).Debug("缓存未命中")
            return nil, ErrCacheMiss
        }

        glog.WithFields(logrus.Fields{
            "缓存键": key,
            "错误": err.Error(),
        }).Error("缓存读取失败")

        return nil, fmt.Errorf("读取缓存失败: %w", err)
    }

    glog.WithField("缓存键", key).Debug("缓存命中")
    return value, nil
}
```

#### 外部 API 调用

```go
func (c *Client) CallExternalAPI(ctx context.Context, req *APIRequest) (*APIResponse, error) {
    glog.WithFields(logrus.Fields{
        "服务": req.Service,
        "接口": req.Endpoint,
        "参数": req.Params,
    }).Debug("开始调用外部API")

    startTime := time.Now()
    resp, err := c.httpClient.Do(ctx, req)

    if err != nil {
        glog.WithFields(logrus.Fields{
            "服务": req.Service,
            "接口": req.Endpoint,
            "错误": err.Error(),
        }).Error("外部API调用失败")

        return nil, fmt.Errorf("API调用失败: %w", err)
    }

    duration := time.Since(startTime)

    // 记录慢请求
    if duration > slowThreshold {
        glog.WithFields(logrus.Fields{
            "服务": req.Service,
            "接口": req.Endpoint,
            "耗时": duration.Milliseconds(),
        }).Warn("外部API调用超时")
    }

    glog.WithFields(logrus.Fields{
        "服务": req.Service,
        "接口": req.Endpoint,
        "状态码": resp.StatusCode,
        "耗时": duration.Milliseconds(),
    }).Debug("外部API调用完成")

    return resp, nil
}
```

#### 业务逻辑关键节点

```go
func ProcessOrder(ctx context.Context, order *Order) error {
    glog.WithFields(logrus.Fields{
        "订单ID": order.ID,
        "用户ID": order.UserID,
        "金额": order.Amount,
    }).Info("开始处理订单")

    // 库存检查
    if err := checkInventory(ctx, order.Items); err != nil {
        glog.WithFields(logrus.Fields{
            "订单ID": order.ID,
            "错误": err.Error(),
        }).Error("库存检查失败")
        return err
    }

    glog.WithField("订单ID", order.ID).Debug("库存检查通过")

    // 支付处理
    if err := processPayment(ctx, order); err != nil {
        glog.WithFields(logrus.Fields{
            "订单ID": order.ID,
            "错误": err.Error(),
        }).Error("支付处理失败")
        return err
    }

    glog.WithField("订单ID", order.ID).Debug("支付处理完成")

    // 订单确认
    if err := confirmOrder(ctx, order); err != nil {
        glog.WithFields(logrus.Fields{
            "订单ID": order.ID,
            "错误": err.Error(),
        }).Error("订单确认失败")
        return err
    }

    glog.WithField("订单ID", order.ID).Info("订单处理完成")
    return nil
}
```

### 3. 结构化日志最佳实践

- **保持一致的字段命名**：在整个项目中使用统一的字段名，便于日志分析和搜索

  - `请求ID/reqID`：标识唯一请求
  - `用户ID/userID`：关联用户操作
  - `模块/module`：标识功能模块
  - `操作/action`：描述操作类型
  - `耗时/duration`：记录操作耗时（毫秒）
  - `错误/error`：错误信息

- **避免日志轰炸**：合理控制日志数量，避免在高频调用的代码中记录过多日志

  - 循环中谨慎记录日志
  - 对重复错误进行采样记录
  - 大批量处理时记录汇总信息而非每条记录

- **包含关联信息**：确保日志包含足够的上下文信息，便于问题定位

  - 使用`WithFields`添加结构化信息
  - 关联请求 ID 和会话 ID
  - 包含关键参数和状态

- **敏感信息处理**：避免记录敏感信息，或进行适当脱敏
  - 密码、令牌等不应记录
  - 个人隐私信息需脱敏处理
  - 支付信息部分隐藏

### 4. 高级日志配置

#### 基本文件日志配置

```go
func initLogger() {
    config := &glog.Config{
        Level:         logrus.InfoLevel,  // 生产环境推荐Info级别
        Format:        glog.FormatJSON,   // 结构化日志便于解析
        EnableConsole: true,              // 开发环境可启用控制台输出
        EnableFile:    true,
        ReportCaller:  true,              // 记录调用者信息
        FileConfig: &glog.FileConfig{
            Filename:   "logs/app.log",
            MaxSize:    10,               // 单个文件最大10MB
            MaxBackups: 7,                // 保留7个备份
            MaxAge:     30,               // 保留30天
            Compress:   true,             // 压缩旧日志
        },
    }

    if err := glog.Configure(config); err != nil {
        panic(fmt.Errorf("初始化日志系统失败: %w", err))
    }

    glog.Info("日志系统初始化完成")
}
```

#### 使用异步日志处理器

对于高性能应用，可以使用异步日志处理器显著提升性能，减少日志对主业务逻辑的影响：

```go
func initAsyncLogger() {
    // 创建基本配置
    config := &glog.Config{
        Level:         logrus.InfoLevel,
        Format:        glog.FormatJSON,
        EnableConsole: true,
        EnableFile:    true,
        FileConfig: &glog.FileConfig{
            Filename:   "logs/app.log",
            MaxSize:    10,
            MaxBackups: 7,
            MaxAge:     30,
            Compress:   true,
        },
    }

    // 创建异步处理器
    asyncConfig := &handlers.AsyncConfig{
        BufferSize:    5000,     // 可缓存5000条日志
        BatchSize:     200,      // 每批处理200条
        FlushInterval: 500 * time.Millisecond, // 最大延迟500ms
    }
    asyncHandler := handlers.NewAsyncHandler(asyncConfig)

    // 添加到自定义处理器列表
    config.CustomHandlers = []handlers.Handler{
        asyncHandler,
    }

    // 应用配置
    if err := glog.Configure(config); err != nil {
        panic(fmt.Errorf("初始化异步日志系统失败: %w", err))
    }

    // 在应用退出时优雅关闭异步处理器
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        glog.Info("正在关闭日志系统...")
        if closer, ok := asyncHandler.(io.Closer); ok {
            if err := closer.Close(); err != nil {
                fmt.Fprintf(os.Stderr, "关闭异步日志处理器失败: %v\n", err)
            }
        }
        os.Exit(0)
    }()

    glog.Info("异步日志系统初始化完成")
}
```

#### 适用场景与性能影响

异步日志处理器非常适合以下场景：

- **高吞吐量 API 服务**：可以减少日志对 API 响应时间的影响
- **批处理任务**：批量处理数据时产生大量日志
- **实时系统**：对延迟敏感的应用，如游戏服务器、金融交易系统

性能提升：

| 场景                | 同步日志      | 异步日志       | 性能提升 |
| ------------------- | ------------- | -------------- | -------- |
| REST API (1000 QPS) | 平均延迟 5ms  | 平均延迟 0.8ms | 约 6 倍  |
| 批处理 (10 万记录)  | 处理时间 15s  | 处理时间 8s    | 约 47%   |
| 高并发写入          | 每秒约 2 万条 | 每秒约 10 万条 | 约 5 倍  |

注意事项：

- 关闭时应调用`Close()`方法确保所有日志都被处理
- 合理配置`BufferSize`，过小可能导致日志丢失，过大可能占用过多内存
- 对于关键事务日志，考虑在异步处理前先同步写入，确保数据不丢失

## 开发规范

### 1. 设计原则

- **高内聚低耦合**：每个模块具备单一且明确的功能，模块间依赖最小化，提高代码可维护性和可扩展性
- **标准设计模式**：合理使用单例模式、工厂模式、策略模式、责任链模式、组合模式等常见设计模式，保持代码结构清晰
- **接口优先**：通过清晰的接口定义实现模块解耦，便于测试和功能扩展
- **可观测性**：提供完整的监控、日志和追踪能力
- **故障容错**：实现优雅降级、限流和熔断机制

### 2. 代码规范

- **Go 官方规范**：严格遵循 Go 语言代码风格
  - 使用 gofmt 进行代码格式化
  - 遵循包命名规范
  - 使用有意义的变量名和函数名
- **注释规范**
  - 包级别注释：描述包的用途和功能
  - 函数注释：说明函数功能、参数和返回值
  - 关键算法注释：解释复杂逻辑的实现原理
  - **代码复用**：抽象通用功能，避免重复代码
  - 使用 gerror 替换了原有错误处理，提供更丰富的错误上下文信息
  - 为所有类型和方法添加了完整的 godoc 注释，说明其用途和使用方法
  - 在缓存操作的关键点（如缓存 miss、写入失败等）添加了 glog 日志记录，便于问题追踪
  - 保持了代码风格的一致性，遵循 Go 最佳实践

### 3. 性能优化

- **算法效率**：选择合适的数据结构和算法，降低复杂度
- **资源管理**
  - 使用对象池复用对象
  - 控制 goroutine 数量
  - 及时释放资源
- **性能测试**：编写基准测试，持续监控性能指标

### 4. 安全规范

- **输入验证**：严格校验所有外部输入
- **错误处理**
  - 使用自定义错误类型
  - 提供详细错误信息
  - 避免 panic
- **日志记录**：记录关键操作和异常信息

### 5. 测试规范

- **单元测试**
  - 测试覆盖率要求：>80%
  - 包含正常和异常场景测试
  - 使用表驱动测试提高测试效率
- **集成测试**：验证模块间交互
- **性能测试**：监控核心功能的性能表现

### 6. 文档规范

- **代码文档**
  - 包级别文档
  - 接口文档
  - 示例代码
- **使用文档**
  - 安装说明
  - 快速开始指南
  - 最佳实践示例

### 7. 版本管理

- **语义化版本**：遵循 MAJOR.MINOR.PATCH 格式
- **兼容性保证**
  - 明确标注废弃接口
  - 提供版本迁移指南
  - 保持向后兼容

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交变更
4. 编写测试用例
5. 提交 Pull Request

## 开源协议

MIT License

## 错误处理规范

为确保系统的可靠性和可维护性，项目采用 gerror 作为统一的错误处理框架。gerror 提供了结构化的错误处理能力，包括错误码、错误堆栈和上下文信息，使错误更易于理解和定位。

### 1. gerror 作为默认错误处理工具

从本项目开始，所有新开发的服务和组件必须使用 gerror 作为标准的错误处理库。现有系统在迭代更新时，也应逐步迁移到 gerror。

**主要优势：**

- **结构化错误信息**：包含错误码、消息、堆栈跟踪和上下文
- **错误链追踪**：支持错误包装，保留原始错误信息
- **标准化错误码**：统一的错误码体系，便于错误分类和处理
- **与 HTTP/gRPC 集成**：轻松映射到 HTTP 状态码或 gRPC 错误码
- **数据库错误处理**：专门针对数据库操作的错误检测和包装

### 2. 错误创建与包装规范

#### 创建新错误

```go
// 基本错误创建
err := gerror.New(gerror.CodeInvalidParam, "参数无效")

// 带格式化的错误创建
userID := "12345"
err := gerror.Newf(gerror.CodeNotFound, "未找到ID为%s的用户", userID)
```

#### 包装已有错误

```go
// 包装底层错误
if err := db.QueryRow(); err != nil {
    return gerror.Wrap(err, gerror.CodeDBQuery, "查询用户数据失败")
}

// 为错误添加上下文信息
err = gerror.WithContext(err, "用户ID", userID)
err = gerror.WithContext(err, "请求ID", requestID)

// 或使用map批量添加上下文
err = gerror.WithContextMap(err, map[string]interface{}{
    "用户ID": userID,
    "请求ID": requestID,
    "操作类型": "创建订单",
})
```

### 3. 错误检查与处理

#### 错误类型检查

```go
// 检查是否是"未找到"类型的错误
if gerror.IsNotFound(err) {
    // 处理资源不存在的情况
    return nil, ErrResourceNotFound
}

// 检查是否是超时错误
if gerror.IsTimeout(err) {
    // 处理超时情况，可能需要重试
    return retryOperation()
}

// 检查是否是数据库约束错误
if gerror.IsConstraintViolation(err) {
    // 处理数据约束冲突
    return nil, ErrConstraintViolation
}
```

#### HTTP 响应处理

```go
func handleError(w http.ResponseWriter, err error) {
    // 将错误转换为HTTP响应
    status := gerror.GetHTTPStatus(err)
    response := gerror.GetHTTPResponse(err)

    // 记录错误日志
    glog.WithError(err).Error("请求处理失败")

    // 返回给客户端
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(response)
}
```

### 4. 错误码规范

gerror 定义了一套标准的错误码体系，应根据实际错误场景选择合适的错误码：

| 错误码范围  | 分类          | 说明                       |
| ----------- | ------------- | -------------------------- |
| 10000-10999 | 通用错误      | 适用于各种通用场景的错误   |
| 11000-19999 | 业务错误      | 特定业务领域相关的错误     |
| 12000-12999 | gRPC/RPC 错误 | 微服务通信相关的错误       |
| 13000-13999 | 数据库错误    | 数据库操作相关的错误       |
| 14000-14999 | HTTP 错误     | HTTP 请求/响应相关的错误   |
| 15000-15999 | 安全错误      | 认证、授权、加密相关的错误 |

**使用原则：**

1. 优先使用预定义错误码
2. 如需扩展，在对应范围内定义新错误码
3. 为自定义错误码注册错误文本说明
4. 错误码应具有明确的语义，避免重复或冲突

### 5. 最佳实践

1. **保持错误链完整性**：总是包装底层错误，不要丢弃原始错误信息
2. **添加足够上下文**：为错误添加足够的上下文信息，便于问题定位
3. **使用适当错误码**：根据实际情况选择或创建合适的错误码
4. **避免冗长错误消息**：错误消息应简洁明了，不包含敏感信息
5. **记录详细错误日志**：在适当位置记录带有堆栈和上下文的完整错误信息

## 安装

### 安装整个框架

```bash
go get github.com/ntshibin/core@latest
```

### 安装单个组件

```bash
# 只安装错误处理组件
go get github.com/ntshibin/core/gerror@latest

# 只安装日志组件
go get github.com/ntshibin/core/glog@latest

# 只安装配置管理组件
go get github.com/ntshibin/core/gconf@latest

# 只安装缓存组件
go get github.com/ntshibin/core/gcache@latest

# 只安装HTTP服务组件
go get github.com/ntshibin/core/ghttp@latest
```

## 使用示例

### 错误处理示例

```go
package main

import (
    "fmt"
    "github.com/ntshibin/core/gerror"
)

const (
    ErrUserNotFound gerror.Code = 10001
)

func main() {
    // 创建带错误码的错误
    err := gerror.New(ErrUserNotFound, "用户不存在")

    // 获取错误码
    code := gerror.GetCode(err)
    fmt.Printf("错误码: %d, 错误信息: %s\n", code, err.Error())
}
```

### 日志示例

```go
package main

import "github.com/ntshibin/core/glog"

func main() {
    // 默认配置的日志实例
    logger := glog.New()

    // 结构化日志
    logger.WithFields(glog.Fields{
        "user_id": 123,
        "action": "login",
    }).Info("用户登录成功")

    // 不同日志级别
    logger.Debug("这是调试信息")
    logger.Info("这是普通信息")
    logger.Warn("这是警告信息")
    logger.Error("这是错误信息")
}
```

### 配置管理示例

```go
package main

import (
    "fmt"
    "github.com/ntshibin/core/gconf"
)

// 应用配置结构
type AppConfig struct {
    AppName string        `json:"app_name" yaml:"app_name" env:"APP_NAME" default:"my-app"`
    Port    int           `json:"port" yaml:"port" env:"PORT" default:"8080"`
    Debug   bool          `json:"debug" yaml:"debug" env:"DEBUG" default:"false"`
    Timeout time.Duration `json:"timeout" yaml:"timeout" env:"TIMEOUT" default:"30s"`
}

func main() {
    var config AppConfig

    // 从配置文件加载
    if err := gconf.Load("config.yaml", &config); err != nil {
        panic(err)
    }

    fmt.Println("应用名称:", config.AppName)
    fmt.Println("监听端口:", config.Port)
}
```

### 缓存示例

```go
package main

import (
    "context"
    "fmt"
    "github.com/ntshibin/core/gcache"
    "time"
)

func main() {
    // 获取缓存实例
    cache := gcache.GetCache()

    ctx := context.Background()

    // 设置缓存
    err := cache.Set(ctx, "user:123", map[string]interface{}{
        "id": 123,
        "name": "张三",
        "age": 30,
    }, time.Minute*10)

    if err != nil {
        panic(err)
    }

    // 获取缓存
    value, err := cache.Get(ctx, "user:123")
    if err != nil {
        if gerror.GetCode(err) == gcache.CodeCacheNotFound {
            fmt.Println("用户不存在")
            return
        }
        panic(err)
    }

    fmt.Printf("用户信息: %v\n", value)
}
```

### HTTP 服务示例

```go
package main

import (
    "github.com/ntshibin/core/ghttp"
    "github.com/ntshibin/core/ghttp/internal/config"
)

func main() {
    // 创建HTTP服务器
    server := ghttp.New(config.DefaultConfig())

    // 注册路由
    server.GET("/hello", func(c *ghttp.Context) {
        c.String(200, "Hello World!")
    })

    // 启动服务器
    if err := server.Run(); err != nil {
        panic(err)
    }
}
```

## 构建与发布

使用 Makefile 进行构建和发布:

```bash
# 显示帮助信息
make help

# 执行完整构建流程
make all

# 运行测试
make test

# 发布新版本
make release
```

## 许可证

MIT
