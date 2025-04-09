// Package conf 提供配置文件加载功能，支持从YAML加载配置并支持环境变量替换
package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// LoadConfig 从文件加载配置到结构体
// config 应该是指向结构体的指针
func LoadConfig(file string, config interface{}) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// 替换环境变量
	expandedContent := expandEnvVars(string(content))

	// 根据文件扩展名选择解析方式
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal([]byte(expandedContent), config)
	case ".json":
		err = json.Unmarshal([]byte(expandedContent), config)
	case ".toml":
		err = toml.Unmarshal([]byte(expandedContent), config)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return err
	}

	return nil
}

// MustLoad 从文件加载配置，若失败则panic
func MustLoad(file string, config interface{}) {
	if err := LoadConfig(file, config); err != nil {
		panic(err)
	}
}

// LoadConfigByEnv 根据当前环境加载配置文件
// 会查找以下配置文件（假设基础文件名为config.yaml）：
// 1. config_[env].yaml (如: config_dev.yaml)
// 2. config.yaml (作为默认)
// 环境变量 RUN_MODE 决定当前环境，如未设置则为dev
func LoadConfigByEnv(baseFilename string, config interface{}) error {
	// 获取不带扩展名的文件名和扩展名
	ext := filepath.Ext(baseFilename)
	base := baseFilename[:len(baseFilename)-len(ext)]

	// 构建特定环境的文件名
	envMode := GetRunMode()
	envFilename := base + "_" + envMode + ext

	// 查找环境特定的配置文件
	envFile := FindConfigFile(envFilename)
	if _, err := os.Stat(envFile); err == nil {
		// 找到环境特定的配置文件
		return LoadConfig(envFile, config)
	}

	// 尝试加载基础配置文件
	baseFile := FindConfigFile(baseFilename)
	return LoadConfig(baseFile, config)
}

// MustLoadByEnv 根据当前环境加载配置，若失败则panic
func MustLoadByEnv(baseFilename string, config interface{}) {
	if err := LoadConfigByEnv(baseFilename, config); err != nil {
		panic(err)
	}
}
