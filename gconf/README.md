# GConf - 灵活的 Go 配置管理库

GConf 是一个简单而强大的 Go 配置管理库，支持从配置文件和环境变量加载配置，并支持默认值设置。

## 特性

- **多种配置来源**：支持从配置文件、环境变量加载配置
- **默认值支持**：为配置字段设置默认值
- **灵活的优先级机制**：环境变量 > 配置文件 > 默认值
- **精细控制处理流程**：可以选择性地处理环境变量或默认值
- **类型转换**：自动将字符串配置值转换为正确的类型
- **嵌套结构体支持**：支持复杂的嵌套配置结构
- **切片支持**：支持字符串切片类型的配置项
- **错误处理**: 使用 gerror 提供清晰的错误信息和代码

## 安装

```bash
go get github.com/ntshibin/core/gconf
```

## 基本使用

### 定义配置结构体

```go
type Config struct {
    Host     string        `json:"host" env:"SERVER_HOST" default:"localhost"`
    Port     int           `json:"port" env:"SERVER_PORT" default:"8080"`
    Debug    bool          `json:"debug" env:"APP_DEBUG" default:"false"`
    LogLevel string        `json:"log_level" env:"LOG_LEVEL" default:"info"`
    Tags     []string      `json:"tags" env:"APP_TAGS" default:"api,service"`
    Timeout  time.Duration `json:"timeout" env:"APP_TIMEOUT" default:"5s"`
}
```

### 从结构体加载配置

只应用默认值和环境变量，不加载配置文件：

```go
import (
    "fmt"
    "github.com/ntshibin/core/gconf"
)

func main() {
    // 创建配置实例
    config := &Config{}

    // 加载配置（处理环境变量和默认值）
    if err := gconf.LoadFromStruct(config); err != nil {
        panic(err)
    }

    fmt.Printf("配置: %+v\n", config)
}
```

### 从配置文件加载

从 JSON 文件加载配置并处理环境变量（默认行为）：

```go
import (
    "fmt"
    "github.com/ntshibin/core/gconf"
)

func main() {
    config := &Config{}

    // 从JSON文件加载配置（默认会应用默认值和处理环境变量）
    if err := gconf.Load("config.json", config); err != nil {
        panic(err)
    }

    fmt.Printf("配置: %+v\n", config)
}
```

从配置文件加载但不处理环境变量：

```go
// 加载配置文件但不处理环境变量
if err := gconf.Load("config.json", config, false); err != nil {
    panic(err)
}
```

### 使用 Must 函数

```go
import (
    "github.com/ntshibin/core/gconf"
)

func main() {
    config := &Config{}

    // 如果加载失败，将会 panic
    gconf.MustLoad("config.json", config)

    // 或者只处理结构体
    gconf.MustLoadFromStruct(config)
}
```

## 精细控制配置加载流程

GConf 提供了几个函数来精细控制配置加载流程：

```go
// 只处理默认值，不处理环境变量
gconf.LoadDefaultsFromStruct(config)

// 只处理环境变量，不处理默认值
gconf.LoadEnvFromStruct(config)

// 加载配置文件，默认处理环境变量和默认值
gconf.Load("config.json", config)

// 加载配置文件，但禁用环境变量处理
gconf.Load("config.json", config, false)
```

## 嵌套配置示例

GConf 支持各种嵌套结构体配置：

```go
type DatabaseConfig struct {
    Host     string `json:"host" env:"DB_HOST" default:"localhost"`
    Port     int    `json:"port" env:"DB_PORT" default:"3306"`
    Username string `json:"username" env:"DB_USER" default:"root"`
    Password string `json:"password" env:"DB_PASS" default:""`
}

type AppConfig struct {
    Server   Config         `json:"server"`
    Database DatabaseConfig `json:"database"`
    LogPath  string         `json:"log_path" env:"LOG_PATH" default:"/var/log/app.log"`
}
```

## 标签说明

- **json**: 指定 JSON 配置文件中的字段名
- **env**: 指定环境变量名
- **default**: 指定默认值

## 优先级

配置值的优先级从高到低：

1. 环境变量
2. 配置文件中的值
3. 默认值

## 类型支持

GConf 支持多种数据类型的自动转换：

- 字符串: `string`
- 整数: `int`, `int8`, `int16`, `int32`, `int64`
- 无符号整数: `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- 浮点数: `float32`, `float64`
- 布尔值: `bool`
- 字符串切片: `[]string`
- 时间间隔: `time.Duration`

## 错误处理

GConf 使用 gerror 包提供清晰的错误信息，错误码包括：

- `ErrLoadConfig` (20000): 加载配置错误
- `ErrParseConfig` (20001): 解析配置错误
- `ErrInvalidType` (20002): 配置类型错误
- `ErrInvalidValue` (20003): 配置值错误

## 示例：完整配置加载流程

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"

    "github.com/ntshibin/core/gconf"
)

type ServerConfig struct {
    Host    string        `json:"host" env:"SERVER_HOST" default:"localhost"`
    Port    int           `json:"port" env:"SERVER_PORT" default:"8080"`
    Timeout time.Duration `json:"timeout" env:"SERVER_TIMEOUT" default:"30s"`
}

type DatabaseConfig struct {
    Host     string `json:"host" env:"DB_HOST" default:"localhost"`
    Port     int    `json:"port" env:"DB_PORT" default:"3306"`
    User     string `json:"user" env:"DB_USER" default:"root"`
    Password string `json:"password" env:"DB_PASSWORD" default:""`
    Database string `json:"database" env:"DB_NAME" default:"test"`
}

type Config struct {
    Server   ServerConfig   `json:"server"`
    Database DatabaseConfig `json:"database"`
    Debug    bool           `json:"debug" env:"APP_DEBUG" default:"false"`
    LogLevel string         `json:"log_level" env:"LOG_LEVEL" default:"info"`
    Tags     []string       `json:"tags" env:"APP_TAGS" default:"api,service,backend"`
}

func main() {
    // 设置环境变量进行测试
    os.Setenv("SERVER_PORT", "9000")
    os.Setenv("DB_USER", "admin")
    os.Setenv("APP_DEBUG", "true")

    // 创建配置实例
    config := &Config{}

    // 方法1: 从结构体加载配置（应用默认值和环境变量）
    if err := gconf.LoadFromStruct(config); err != nil {
        log.Fatalf("加载配置失败: %v", err)
    }
    fmt.Printf("方法1 - 从结构体加载: %+v\n", config)

    // 方法2: 从配置文件加载
    config = &Config{} // 重置配置
    if err := gconf.Load("config.json", config); err != nil {
        // 如果配置文件不存在，使用默认值
        fmt.Println("配置文件不存在，使用默认配置和环境变量")
        if err := gconf.LoadFromStruct(config); err != nil {
            log.Fatalf("加载默认配置失败: %v", err)
        }
    }
    fmt.Printf("方法2 - 从文件加载: %+v\n", config)

    // 清理环境变量
    os.Unsetenv("SERVER_PORT")
    os.Unsetenv("DB_USER")
    os.Unsetenv("APP_DEBUG")
}
```

## 注意事项

- 所有需要被设置的字段必须是可导出的（首字母大写）
- 默认值和环境变量值都是字符串，将被转换为字段的实际类型
- 字符串切片使用逗号分隔的值表示（如：`"item1,item2,item3"`）
- 当前支持 JSON 格式的配置文件，可以扩展支持其他格式如 YAML 和 TOML
- 对于时间间隔（`time.Duration`）使用如 "5s", "1m30s" 的格式
