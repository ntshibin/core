# gerror - 生产级 Go 错误处理库

gerror 是一个为 Go 语言设计的生产级错误处理库，旨在解决 Go 标准库中错误处理的局限性，并为企业级应用提供统一、强大且灵活的错误处理解决方案。

## 特性

- **结构化错误**：支持错误码、消息和上下文信息
- **错误链追踪**：包装错误并保留原始错误信息
- **堆栈跟踪**：自动捕获并记录错误产生的堆栈信息
- **HTTP 集成**：支持错误到 HTTP 状态码和响应的映射
- **数据库错误处理**：特定于数据库操作的错误识别和处理
- **微服务/RPC 支持**：适用于分布式系统的错误处理
- **错误分类**：提供错误类型检查功能
- **JSON 序列化**：支持以 JSON 格式序列化错误
- **可扩展性**：支持自定义错误码和处理器

## 安装

```bash
go get github.com/ntshibin/core/gerror
```

## 基本使用

### 创建错误

```go
import "github.com/ntshibin/core/gerror"

// 创建基本错误
err := gerror.New(gerror.CodeInvalidParam, "无效的用户ID")

// 创建格式化错误
userID := 101
err := gerror.Newf(gerror.CodeInvalidParam, "用户ID %d 不符合要求", userID)
```

### 包装错误

```go
// 包装标准错误
baseErr := errors.New("数据库连接失败")
err := gerror.Wrap(baseErr, gerror.CodeInternal, "用户数据读取失败")

// 格式化包装
err := gerror.Wrapf(baseErr, gerror.CodeInternal, "读取用户 %s 数据失败", username)
```

### 错误上下文

```go
// 添加单个上下文信息
err := gerror.New(gerror.CodeNotFound, "用户未找到")
err = gerror.WithContext(err, "用户ID", 12345)

// 添加多个上下文信息
err = gerror.WithContextMap(err, map[string]interface{}{
    "请求ID": "req-abc123",
    "时间戳": time.Now().Unix(),
})
```

### 错误检查

```go
// 使用预定义函数检查错误类型
if gerror.IsNotFound(err) {
    // 处理"未找到"类型的错误
}

if gerror.IsTimeout(err) {
    // 处理超时错误
}

// 使用错误码检查
if gerror.GetCode(err) == gerror.CodeInvalidParam {
    // 处理参数错误
}
```

### 错误格式化

```go
// 文本格式（默认）
errorText := gerror.FormatError(err, "text")
fmt.Println(errorText)

// JSON格式
errorJSON := gerror.FormatError(err, "json")
fmt.Println(errorJSON)
```

## Web 应用集成

gerror 提供了与 HTTP 服务的集成支持，便于在 Web 应用中使用：

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    user, err := getUserFromDB(r.Context(), userID)
    if err != nil {
        // 获取适当的HTTP状态码
        status := gerror.GetHTTPStatus(err)

        // 获取结构化的错误响应
        response := gerror.GetHTTPResponse(err)

        // 发送错误响应
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(response)
        return
    }

    // 处理成功情况...
}
```

更简便的方式：

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    user, err := getUserFromDB(r.Context(), userID)
    if err != nil {
        // 一步完成错误响应
        gerror.WriteHTTPError(w, err)
        return
    }

    // 处理成功情况...
}
```

## 数据库错误处理

针对数据库操作的特定错误处理：

```go
// 包装数据库错误，自动检测错误类型并使用合适的错误码
rows, err := db.Query("SELECT * FROM users WHERE id = ?", id)
if err != nil {
    return gerror.WrapDBError(err, "查询用户数据失败")
}

// 创建"记录未找到"错误
if rows.Next() == false {
    return gerror.NotFoundError("用户", id)
}

// 创建"重复记录"错误
return gerror.DuplicateError("用户", "邮箱", email)
```

## 微服务/RPC 错误处理

在微服务架构中处理错误：

```go
// 包装RPC调用错误
result, err := client.CallService(ctx, request)
if err != nil {
    return gerror.WrapRPCError(err, "UserService", "GetProfile", "获取用户资料失败")
}

// 特定类型的RPC错误
if serviceUnavailable {
    return gerror.New(gerror.CodeGRPCUnavailable, "用户服务当前不可用")
}

// 创建RPC超时错误
return gerror.RPCTimeoutError("UserService", "GetProfile", "3s")
```

## 错误码参考

gerror 预定义了多组错误码：

### 通用错误码 (10000-10999)

| 错误码 | 常量名           | 说明       |
| ------ | ---------------- | ---------- |
| 10000  | CodeUnknown      | 未知错误   |
| 10001  | CodeInternal     | 内部错误   |
| 10002  | CodeInvalidParam | 参数错误   |
| 10003  | CodeUnauthorized | 未授权     |
| 10004  | CodeForbidden    | 禁止访问   |
| 10005  | CodeNotFound     | 资源不存在 |
| 10006  | CodeTimeout      | 超时错误   |
| 10007  | CodeConflict     | 冲突错误   |
| 10008  | CodeExhausted    | 资源耗尽   |

### 数据库错误码 (11000-11999)

| 错误码 | 常量名             | 说明         |
| ------ | ------------------ | ------------ |
| 11000  | CodeDBNotFound     | 记录未找到   |
| 11001  | CodeDBDuplicate    | 重复记录     |
| 11002  | CodeDBConstraint   | 约束冲突     |
| 11003  | CodeDBConnection   | 连接错误     |
| 11004  | CodeDBTransaction  | 事务错误     |
| 11005  | CodeDBQuery        | 查询错误     |
| 11006  | CodeDBExecution    | 执行错误     |
| 11007  | CodeDBTimeout      | 数据库超时   |
| 11008  | CodeDBUnavailable  | 数据库不可用 |
| 11009  | CodeDBUnauthorized | 数据库未授权 |

### gRPC/微服务错误码 (12000-12999)

| 错误码 | 常量名                   | 说明       |
| ------ | ------------------------ | ---------- |
| 12000  | CodeGRPCCanceled         | 请求被取消 |
| 12001  | CodeGRPCUnknown          | 未知错误   |
| 12002  | CodeGRPCInvalidArgument  | 无效参数   |
| 12003  | CodeGRPCDeadlineExceeded | 请求超时   |
| 12004  | CodeGRPCNotFound         | 资源未找到 |
| 12005  | CodeGRPCAlreadyExists    | 资源已存在 |
| 12006  | CodeGRPCPermissionDenied | 权限拒绝   |
| ...    | ...                      | ...        |

## 自定义错误码

可以根据业务需求注册自定义错误码：

```go
// 定义自定义错误码
const (
    CodeUserInactive   gerror.Code = 20001
    CodeUserSuspended  gerror.Code = 20002
    CodeUserDeleted    gerror.Code = 20003
)

func init() {
    // 注册错误码文本
    gerror.RegisterCodeText(CodeUserInactive, "用户未激活")
    gerror.RegisterCodeText(CodeUserSuspended, "用户已被暂停")
    gerror.RegisterCodeText(CodeUserDeleted, "用户已被删除")

    // 可选：注册HTTP状态码映射
    gerror.RegisterHTTPStatus(CodeUserInactive, http.StatusForbidden)
    gerror.RegisterHTTPStatus(CodeUserSuspended, http.StatusForbidden)
    gerror.RegisterHTTPStatus(CodeUserDeleted, http.StatusGone)
}
```

## 最佳实践

1. **始终包装错误**：始终使用 Wrap 或 WithContext 保留原始错误信息和上下文
2. **统一错误码**：在项目中统一定义和使用错误码，确保一致性
3. **有意义的错误消息**：提供具体、有帮助的错误消息，避免模糊描述
4. **分层错误处理**：在不同层（DAO、Service、API）适当包装和添加上下文
5. **避免泄露敏感信息**：在向客户端返回错误时过滤敏感信息
6. **记录详细错误**：在服务端记录包含堆栈和上下文的完整错误信息
7. **合理使用 HTTP 状态码**：确保 HTTP 状态码与错误类型匹配

## 示例

gerror 包含两个完整的示例，展示了库的各种功能：

- `examples/basic_usage.go`: 展示基本功能和用法
- `examples/web_example.go`: 展示在 Web 应用中的集成

运行示例：

```bash
go run gerror/examples/basic_usage.go
go run gerror/examples/web_example.go
```

## 许可证

MIT License
