package gconf_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ntshibin/core/gconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 基本配置结构体
type BasicConfig struct {
	Host    string `json:"host" env:"SERVER_HOST" default:"localhost"`
	Port    int    `json:"port" env:"SERVER_PORT" default:"8080"`
	LogPath string `json:"log_path" env:"LOG_PATH" default:"/var/log/app.log"`
	Debug   bool   `json:"debug" env:"APP_DEBUG" default:"false"`
}

// 嵌套配置结构体
type DatabaseConfig struct {
	Host     string `json:"host" env:"DB_HOST" default:"localhost"`
	Port     int    `json:"port" env:"DB_PORT" default:"3306"`
	Username string `json:"username" env:"DB_USER" default:"root"`
	Password string `json:"password" env:"DB_PASS" default:""`
	Database string `json:"database" env:"DB_NAME" default:"test"`
}

type RedisConfig struct {
	Host     string        `json:"host" env:"REDIS_HOST" default:"localhost"`
	Port     int           `json:"port" env:"REDIS_PORT" default:"6379"`
	Password string        `json:"password" env:"REDIS_PASS" default:""`
	DB       int           `json:"db" env:"REDIS_DB" default:"0"`
	Timeout  time.Duration `json:"timeout" env:"REDIS_TIMEOUT" default:"5s"`
}

type AppConfig struct {
	Basic    BasicConfig    `json:"basic"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	LogLevel string         `json:"log_level" env:"LOG_LEVEL" default:"info"`
	Tags     []string       `json:"tags" env:"APP_TAGS" default:"api,service,backend"`
}

// 服务器配置
type ServerConfig struct {
	Host string `json:"host" env:"SERVER_HOST" default:"localhost"`
	Port int    `json:"port" env:"SERVER_PORT" default:"8080"`
}

// 测试用配置
type TestConfig struct {
	Server   ServerConfig `json:"server"`
	Database struct {
		MaxConn    int           `json:"max_conn" env:"DB_MAX_CONN" default:"10"`
		MaxTimeout time.Duration `json:"max_timeout" env:"DB_MAX_TIMEOUT" default:"30s"`
	} `json:"database"`
}

func TestLoadFromStruct(t *testing.T) {
	// 清除可能影响测试的环境变量
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("REDIS_TIMEOUT")
	os.Unsetenv("APP_TAGS")

	// 创建一个新的配置实例
	config := &AppConfig{}

	// 从结构体加载配置（使用默认值）
	err := gconf.LoadFromStruct(config)
	require.NoError(t, err)

	// 验证默认值是否正确设置
	assert.Equal(t, "localhost", config.Basic.Host)
	assert.Equal(t, 8080, config.Basic.Port)
	assert.Equal(t, "/var/log/app.log", config.Basic.LogPath)
	assert.Equal(t, false, config.Basic.Debug)

	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 3306, config.Database.Port)
	assert.Equal(t, "root", config.Database.Username)
	assert.Equal(t, "", config.Database.Password)
	assert.Equal(t, "test", config.Database.Database)

	assert.Equal(t, "localhost", config.Redis.Host)
	assert.Equal(t, 6379, config.Redis.Port)
	assert.Equal(t, "", config.Redis.Password)
	assert.Equal(t, 0, config.Redis.DB)
	assert.Equal(t, 5*time.Second, config.Redis.Timeout)

	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, []string{"api", "service", "backend"}, config.Tags)
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// 设置环境变量
	os.Setenv("SERVER_HOST", "api.example.com")
	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("REDIS_TIMEOUT", "10s")
	os.Setenv("APP_TAGS", "web,frontend,mobile")

	defer func() {
		// 测试结束后清理环境变量
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("REDIS_TIMEOUT")
		os.Unsetenv("APP_TAGS")
	}()

	// 创建一个新的配置实例
	config := &AppConfig{}

	// 从结构体加载配置（使用环境变量覆盖默认值）
	err := gconf.LoadFromStruct(config)
	require.NoError(t, err)

	// 验证环境变量是否正确覆盖默认值
	assert.Equal(t, "api.example.com", config.Basic.Host)
	assert.Equal(t, 9000, config.Basic.Port)
	assert.Equal(t, "db.example.com", config.Database.Host)
	assert.Equal(t, 10*time.Second, config.Redis.Timeout)
	assert.Equal(t, []string{"web", "frontend", "mobile"}, config.Tags)

	// 验证未设置环境变量的字段仍然使用默认值
	assert.Equal(t, "/var/log/app.log", config.Basic.LogPath)
	assert.Equal(t, 3306, config.Database.Port)
}

func TestLoadFromJSONFile(t *testing.T) {
	// 创建临时JSON配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	configContent := `{
		"basic": {
			"host": "config.example.com",
			"port": 7000
		},
		"database": {
			"host": "db-config.example.com",
			"port": 5432,
			"username": "admin",
			"database": "production"
		},
		"log_level": "debug"
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 创建一个新的配置实例
	config := &AppConfig{}

	// 从JSON文件加载配置
	err = gconf.Load(configPath, config)
	require.NoError(t, err)

	// 环境变量不应该影响测试，所以这里应该看到的是文件中的原始值
	assert.Equal(t, "config.example.com", config.Basic.Host)       // 文件中的值
	assert.Equal(t, 7000, config.Basic.Port)                       // 文件中的值
	assert.Equal(t, "db-config.example.com", config.Database.Host) // 文件中的值
	assert.Equal(t, 5432, config.Database.Port)                    // 文件中的值
	assert.Equal(t, "admin", config.Database.Username)             // 文件中的值
	assert.Equal(t, "production", config.Database.Database)        // 文件中的值
	assert.Equal(t, "debug", config.LogLevel)                      // 文件中的值

	// 验证文件中未设置的字段是否使用默认值
	assert.Equal(t, "localhost", config.Redis.Host)                     // 默认值
	assert.Equal(t, 6379, config.Redis.Port)                            // 默认值
	assert.Equal(t, 5*time.Second, config.Redis.Timeout)                // 默认值
	assert.Equal(t, []string{"api", "service", "backend"}, config.Tags) // 默认值

	// 注释掉 LoadWithEnv 相关测试，因为该方法已不存在
	/*
		// 现在测试环境变量的覆盖
		// 设置一些环境变量，测试优先级
		os.Setenv("SERVER_HOST", "env.example.com")
		os.Setenv("DB_PORT", "3307")

		defer func() {
			os.Unsetenv("SERVER_HOST")
			os.Unsetenv("DB_PORT")
		}()

		// 使用 LoadWithEnv 加载配置并处理环境变量
		config2 := &AppConfig{}
		err = gconf.LoadWithEnv(configPath, config2)
		require.NoError(t, err)

		// 验证环境变量是否覆盖了文件中的值
		assert.Equal(t, "env.example.com", config2.Basic.Host) // 环境变量覆盖
		assert.Equal(t, 3307, config2.Database.Port)           // 环境变量覆盖
	*/
}

// 测试处理无效配置值
func TestInvalidValues(t *testing.T) {
	os.Setenv("SERVER_PORT", "invalid_port")
	defer os.Unsetenv("SERVER_PORT")

	config := &BasicConfig{}
	err := gconf.LoadFromStruct(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "设置字段 Port 的环境变量值失败")
}

// 测试 MustLoad 函数
func TestMustLoad(t *testing.T) {
	// 创建一个不存在的配置文件路径
	nonExistentPath := "/non/existent/config.json"

	// MustLoad 应该会 panic
	assert.Panics(t, func() {
		config := &AppConfig{}
		gconf.MustLoad(nonExistentPath, config)
	})
}

// 测试 MustLoadFromStruct 函数
func TestMustLoadFromStruct(t *testing.T) {
	os.Setenv("SERVER_PORT", "invalid_port")
	defer os.Unsetenv("SERVER_PORT")

	// MustLoadFromStruct 应该会 panic
	assert.Panics(t, func() {
		config := &BasicConfig{}
		gconf.MustLoadFromStruct(config)
	})
}

// 空配置结构体测试
type EmptyConfig struct {
	// 无字段
}

// 测试处理空配置结构体
func TestEmptyConfig(t *testing.T) {
	config := &EmptyConfig{}
	err := gconf.LoadFromStruct(config)
	assert.NoError(t, err)
}

// 测试基本配置
func TestBasicConfig(t *testing.T) {
	// 定义初始配置
	config := TestConfig{
		Server: ServerConfig{
			Host: "localhost", // 这个不会被覆盖，因为没有对应的环境变量
			Port: 8080,        // 这个会被环境变量覆盖
		},
		Database: struct {
			MaxConn    int           `json:"max_conn" env:"DB_MAX_CONN" default:"10"`
			MaxTimeout time.Duration `json:"max_timeout" env:"DB_MAX_TIMEOUT" default:"30s"`
		}{
			MaxConn:    10,
			MaxTimeout: 30 * time.Second,
		},
	}

	// 设置环境变量
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_MAX_CONN", "20")

	// 从结构体加载配置，处理环境变量和默认值
	err := gconf.LoadFromStruct(&config)
	assert.NoError(t, err)

	// 验证环境变量覆盖
	assert.Equal(t, "localhost", config.Server.Host)
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, 20, config.Database.MaxConn)
	assert.Equal(t, 30*time.Second, config.Database.MaxTimeout)

	// 清理环境变量
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DB_MAX_CONN")
}
