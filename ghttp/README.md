# GHttp - Go HTTP 服务框架

GHttp 是对 Gin 框架的封装，提供了更简洁的 API，同时集成了 `gerror`、`glog` 和 `gconf` 等核心库，使得框架与业务逻辑更好地分离。

## 特性

- **轻量级封装**：对 Gin 框架的轻量级封装，保留其高性能特性
- **统一错误处理**：集成 `gerror` 包进行错误处理，支持错误码和错误信息
- **日志集成**：使用 `glog` 进行日志记录，提供丰富的日志功能
- **配置管理**：集成 `gconf` 包，支持从文件和环境变量加载配置
- **中间件支持**：内置多种常用中间件，如 CORS、认证、超时控制等
- **API 分组**：支持路由分组和嵌套分组
- **统一响应格式**：提供统一的 JSON 响应格式
- **优雅关闭**：支持服务器的优雅关闭

## 安装

```bash
go get github.com/ntshibin/core/ghttp
```

## 快速开始

### 基本用法

```go
package main

import (
    "github.com/ntshibin/core/ghttp"
    "github.com/ntshibin/core/glog"
)

func main() {
    // 加载默认配置
    config := ghttp.DefaultConfig()
    config.Port = 8080

    // 创建路由
    router := ghttp.New(config)

    // 注册路由
    router.GET("/", func(c *ghttp.Context) {
        c.Success(map[string]string{
            "message": "Hello, World!",
        })
    })

    // 启动服务
    if err := router.Run(); err != nil {
        glog.Fatalf("服务启动失败: %v", err)
    }
}
```

### 配置加载

```go
// 从配置文件加载
config, err := ghttp.LoadConfig("config.json")
if err != nil {
    glog.Fatalf("加载配置失败: %v", err)
}

// 或者从环境变量加载
config := ghttp.LoadConfigFromEnv()

// 或者使用默认配置
config := ghttp.DefaultConfig()
```

### 使用中间件

```go
// 全局中间件
router.Use(
    ghttp.Recovery(),
    ghttp.RequestID(),
    ghttp.CORS(ghttp.DefaultCORSConfig()),
    ghttp.Timeout(10 * time.Second),
)

// 组级别中间件
api := router.Group("/api")
api.Use(ghttp.Auth(myAuthFunction))
```

### API 分组

```go
// 创建 API 分组
api := router.Group("/api")
{
    // 用户相关 API
    api.GET("/users", GetUsers)
    api.POST("/users", CreateUser)
    api.GET("/users/:id", GetUser)
    api.PUT("/users/:id", UpdateUser)
    api.DELETE("/users/:id", DeleteUser)

    // 嵌套分组
    v2 := api.Group("/v2")
    {
        v2.GET("/users", GetUsersV2)
    }
}
```

### 错误处理

```go
func GetUser(c *ghttp.Context) {
    id := c.Param("id")

    user, err := findUser(id)
    if err != nil {
        // 使用 gerror 封装错误
        if isNotFoundError(err) {
            c.NotFound(gerror.New(gerror.CodeNotFound, "用户不存在"))
            return
        }

        c.Error(gerror.Wrap(err, gerror.CodeInternal, "获取用户失败"))
        return
    }

    c.Success(user)
}
```

### 认证中间件

```go
// 定义认证函数
func authenticate(c *ghttp.Context) error {
    token := c.GetHeader("Authorization")
    if token == "" {
        return gerror.New(gerror.CodeUnauthorized, "缺少认证令牌")
    }

    // 验证令牌...
    userID, err := verifyToken(token)
    if err != nil {
        return gerror.Wrap(err, gerror.CodeUnauthorized, "无效的认证令牌")
    }

    // 设置用户信息到上下文
    c.Set("userID", userID)
    return nil
}

// 使用认证中间件
authAPI := router.Group("/auth")
authAPI.Use(ghttp.Auth(authenticate))

// 认证后的 API
authAPI.GET("/profile", GetUserProfile)
```

## HTTP 配置选项

HTTP 服务器可以通过 `HTTPConfig` 结构体进行配置：

```go
type HTTPConfig struct {
    Port            int           `json:"port" env:"HTTP_PORT" default:"8080"`
    ReadTimeout     time.Duration `json:"read_timeout" env:"HTTP_READ_TIMEOUT" default:"5s"`
    WriteTimeout    time.Duration `json:"write_timeout" env:"HTTP_WRITE_TIMEOUT" default:"10s"`
    ShutdownTimeout time.Duration `json:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" default:"30s"`
    TrustedProxies  []string      `json:"trusted_proxies" env:"HTTP_TRUSTED_PROXIES" default:""`
    Mode            string        `json:"mode" env:"GIN_MODE" default:"release"`
}
```

## 中间件

GHttp 提供了多种常用中间件：

- **Recovery**：恢复中间件，捕获 panic 并返回 500 错误
- **RequestID**：请求 ID 中间件，为每个请求生成唯一 ID
- **CORS**：跨域中间件，支持 CORS 配置
- **Timeout**：超时中间件，设置请求超时时间
- **Auth**：认证中间件，验证请求是否包含有效的认证信息
- **RateLimit**：限流中间件，限制请求频率

## 响应格式

GHttp 提供了统一的 JSON 响应格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 响应数据
  }
}
```

错误响应：

```json
{
  "code": 10004, // 错误码
  "message": "资源不存在", // 错误消息
  "data": null
}
```

## 完整示例

```go
package main

import (
    "time"

    "github.com/ntshibin/core/gerror"
    "github.com/ntshibin/core/gconf"
    "github.com/ntshibin/core/ghttp"
    "github.com/ntshibin/core/glog"
)

// AppConfig 应用配置
type AppConfig struct {
    HTTP ghttp.HTTPConfig `json:"http"`
    Log  glog.LogConfig   `json:"log"`
}

func main() {
    // 加载配置
    var config AppConfig
    err := gconf.Load("config.json", &config)
    if err != nil {
        panic(err)
    }

    // 初始化日志
    glog.Init(config.Log)

    // 创建路由
    router := ghttp.New(config.HTTP)

    // 使用中间件
    router.Use(
        ghttp.Recovery(),
        ghttp.RequestID(),
        ghttp.CORS(ghttp.DefaultCORSConfig()),
    )

    // 注册路由
    setupRoutes(router)

    // 启动服务
    glog.Infof("服务启动于端口: %d", config.HTTP.Port)
    if err := router.Run(); err != nil {
        glog.Fatalf("服务启动失败: %v", err)
    }
}

func setupRoutes(router *ghttp.Router) {
    // 首页
    router.GET("/", func(c *ghttp.Context) {
        c.Success(map[string]string{
            "message": "API 服务已启动",
            "version": "1.0.0",
        })
    })

    // API 分组
    api := router.Group("/api")
    {
        // 用户 API
        api.GET("/users", getUsers)
        api.POST("/users", createUser)
        api.GET("/users/:id", getUser)
        api.PUT("/users/:id", updateUser)
        api.DELETE("/users/:id", deleteUser)
    }

    // 认证 API
    authAPI := router.Group("/auth")
    authAPI.POST("/login", login)

    // 受保护的 API
    protectedAPI := router.Group("/protected")
    protectedAPI.Use(ghttp.Auth(authenticate))
    {
        protectedAPI.GET("/profile", getProfile)
    }
}

// 认证函数
func authenticate(c *ghttp.Context) error {
    token := c.GetHeader("Authorization")
    if token == "" {
        return gerror.New(gerror.CodeUnauthorized, "缺少认证令牌")
    }

    // 验证令牌...
    if token != "valid-token" {
        return gerror.New(gerror.CodeUnauthorized, "无效的认证令牌")
    }

    c.Set("userID", "123456")
    return nil
}

// API 处理函数...
func getUsers(c *ghttp.Context) {
    // 获取用户列表...
    c.Success([]map[string]interface{}{
        {"id": "1", "name": "用户1"},
        {"id": "2", "name": "用户2"},
    })
}

func getUser(c *ghttp.Context) {
    id := c.Param("id")

    // 模拟查找用户
    if id == "999" {
        c.NotFound(gerror.New(gerror.CodeNotFound, "用户不存在"))
        return
    }

    c.Success(map[string]interface{}{
        "id":   id,
        "name": "测试用户",
    })
}

func createUser(c *ghttp.Context) {
    var user struct {
        Name  string `json:"name" binding:"required"`
        Email string `json:"email" binding:"required,email"`
    }

    if err := c.ShouldBindJSON(&user); err != nil {
        c.BadRequest(gerror.Wrap(err, gerror.CodeInvalidParam, "无效的用户数据"))
        return
    }

    // 创建用户...
    c.Success(map[string]interface{}{
        "id":      "123",
        "name":    user.Name,
        "email":   user.Email,
        "created": time.Now(),
    })
}

func updateUser(c *ghttp.Context) {
    // 更新用户...
}

func deleteUser(c *ghttp.Context) {
    // 删除用户...
}

func login(c *ghttp.Context) {
    var loginData struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&loginData); err != nil {
        c.BadRequest(gerror.Wrap(err, gerror.CodeInvalidParam, "无效的登录数据"))
        return
    }

    // 验证用户名和密码...
    if loginData.Username != "admin" || loginData.Password != "password" {
        c.Unauthorized(gerror.New(gerror.CodeUnauthorized, "无效的用户名或密码"))
        return
    }

    c.Success(map[string]string{
        "token": "valid-token",
    })
}

func getProfile(c *ghttp.Context) {
    userID := c.GetString("userID")

    c.Success(map[string]interface{}{
        "id":   userID,
        "name": "测试用户",
        "role": "admin",
    })
}
```

## 错误处理

GHttp 集成了 `gerror` 包进行错误处理，支持以下常见的错误响应方法：

- `c.Error(err)`: 返回 500 错误
- `c.BadRequest(err)`: 返回 400 错误
- `c.Unauthorized(err)`: 返回 401 错误
- `c.Forbidden(err)`: 返回 403 错误
- `c.NotFound(err)`: 返回 404 错误

错误响应会自动识别 `gerror.Error` 类型，提取其中的错误码和错误信息。
