# Conf 包

Conf 是一个轻量级的 Go 配置管理包，支持多种配置格式和环境变量替换。

## 特性

- 支持多种配置格式：YAML、JSON、TOML
- 支持环境变量替换
- 支持多环境配置（开发、测试、生产）
- 支持配置热重载
- 支持配置验证
- 支持配置加密（TODO）

## 安装

```bash
go get github.com/ntshibin/core/conf
```

## 使用示例

### 基本使用

```go
package main

import (
	"fmt"
	"github.com/ntshibin/core/conf"
)

type Config struct {
	Server struct {
		Host string `yaml:"host" json:"host" toml:"host"`
		Port int    `yaml:"port" json:"port" toml:"port"`
	} `yaml:"server" json:"server" toml:"server"`
}

func main() {
	var config Config

	// 加载配置文件
	err := conf.LoadConfig("config.yaml", &config)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Server: %s:%d\n", config.Server.Host, config.Server.Port)
}
```

### 环境变量替换

配置文件支持环境变量替换，格式如下：

```yaml
server:
  host: ${HOST:-localhost}
  port: ${PORT:-8080}
```

支持两种格式：

- `${VAR}` 或 `$VAR`：使用环境变量值
- `${VAR:-default}`：如果环境变量不存在，使用默认值

### 多环境配置

支持根据环境加载不同的配置文件：

```go
// 加载 config_dev.yaml 或 config.yaml
err := conf.LoadConfigByEnv("config.yaml", &config)
```

环境由 `RUN_MODE` 环境变量决定：

- `dev`：开发环境
- `test`：测试环境
- `prod`：生产环境

### 配置热重载

支持配置文件热重载：

```go
type Config struct {
	Server struct {
		Host string
		Port int
	}
}

func onConfigChange(config interface{}) {
	fmt.Println("配置已更新")
}

func main() {
	var config Config

	// 创建配置监听器
	watcher, err := conf.NewConfigWatcher("config.yaml", &config, onConfigChange)
	if err != nil {
		panic(err)
	}
	defer watcher.Stop()

	// 程序继续运行...
}
```

### 配置验证

支持使用结构体标签进行配置验证：

```go
type Config struct {
	Server struct {
		Host string `validate:"required"`
		Port int    `validate:"required,min=1,max=65535"`
	}
}

func main() {
	var config Config

	// 加载配置
	err := conf.LoadConfig("config.yaml", &config)
	if err != nil {
		panic(err)
	}

	// 验证配置
	err = conf.Validate(&config)
	if err != nil {
		panic(err)
	}
}
```

支持的验证标签：

- `required`：字段必填
- `min`：最小值
- `max`：最大值
- `len`：长度
- `email`：邮箱格式
- `url`：URL 格式
- 更多标签请参考 [go-playground/validator](https://github.com/go-playground/validator)

## 配置文件查找

配置文件查找优先级：

1. 环境变量 `CONF_DIR` 指定的目录
2. 当前目录下的 `etc` 目录
3. 当前目录

## 注意事项

1. 配置结构体必须使用指针类型
2. 环境变量替换发生在配置文件解析之前
3. 配置热重载使用文件修改时间检测，可能存在延迟
4. 配置验证使用 go-playground/validator 包

## TODO

- [ ] 添加配置加密支持
- [ ] 添加配置变更通知机制
- [ ] 添加配置模板支持
- [ ] 添加配置合并功能
