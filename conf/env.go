package conf

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 环境变量正则表达式：匹配${VAR}或$VAR格式
var envVarRegex = regexp.MustCompile(`\${([^{}]+)}|\$([a-zA-Z0-9_]+)`)

// 环境变量默认值正则表达式：匹配${VAR:-default}格式
var envVarDefaultRegex = regexp.MustCompile(`\${([a-zA-Z0-9_]+):-([^}]*)}`)

// 定义常用的环境变量
const (
	EnvKeyRunMode = "RUN_MODE"
	EnvKeyConfDir = "CONF_DIR"
)

// 默认配置目录优先级：
// 1. 环境变量 CONF_DIR 指定的目录
// 2. 当前目录下的 etc 目录
// 3. 当前目录
var defaultConfDirs = []string{
	os.Getenv(EnvKeyConfDir),
	"etc",
	".",
}

// expandEnvVars 替换文本中的环境变量引用
func expandEnvVars(content string) string {
	// 先处理带默认值的环境变量
	content = envVarDefaultRegex.ReplaceAllStringFunc(content, func(match string) string {
		// 提取变量名和默认值
		parts := envVarDefaultRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		varName := parts[1]
		defaultValue := parts[2]

		// 获取环境变量，若不存在则使用默认值
		value := os.Getenv(varName)
		if value == "" {
			return defaultValue
		}
		return value
	})

	// 再处理普通环境变量
	content = envVarRegex.ReplaceAllStringFunc(content, func(match string) string {
		// 去除 ${ 和 } 或仅去除 $
		var varName string
		if strings.HasPrefix(match, "${") {
			varName = match[2 : len(match)-1]
		} else {
			varName = match[1:]
		}

		// 获取环境变量值，若不存在则返回原始文本
		value := os.Getenv(varName)
		if value == "" {
			return match // 保留原始引用
		}
		return value
	})

	return content
}

// GetRunMode 获取当前运行模式
// 支持：dev, test, prod 三种模式，默认为 dev
func GetRunMode() string {
	mode := os.Getenv(EnvKeyRunMode)
	if mode == "" {
		return "dev" // 默认为开发模式
	}
	return mode
}

// IsDevMode 是否为开发模式
func IsDevMode() bool {
	return GetRunMode() == "dev"
}

// IsTestMode 是否为测试模式
func IsTestMode() bool {
	return GetRunMode() == "test"
}

// IsProdMode 是否为生产模式
func IsProdMode() bool {
	return GetRunMode() == "prod"
}

// FindConfigFile 在默认目录下查找配置文件
func FindConfigFile(filename string) string {
	for _, dir := range defaultConfDirs {
		if dir == "" {
			continue
		}

		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return filename // 如果所有目录都未找到，返回原始文件名
}
