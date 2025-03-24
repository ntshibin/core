package gconf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ntshibin/core/gerror"
	"github.com/ntshibin/core/glog"
	"gopkg.in/yaml.v3"
)

// 配置文件类型
const (
	// 配置文件格式
	FormatJSON = "json"
	FormatYAML = "yaml"
	FormatYML  = "yml"
	FormatENV  = "env"

	// 支持环境变量的标签名
	envTagName = "env"
	// 支持默认值的标签名
	defaultTagName = "default"
)

// 配置相关错误码
const (
	ErrLoadConfig   gerror.Code = 20000 // 加载配置错误
	ErrParseConfig  gerror.Code = 20001 // 解析配置错误
	ErrInvalidType  gerror.Code = 20002 // 配置类型错误
	ErrInvalidValue gerror.Code = 20003 // 配置值错误
)

// MustLoad 加载配置文件，如果失败则panic
func MustLoad(path string, v interface{}, loadEnv ...bool) {
	if err := Load(path, v, loadEnv...); err != nil {
		panic(err)
	}
}

// Load 加载指定路径的配置文件到指定的结构体
// 可通过 loadEnv 参数控制是否处理环境变量，默认处理
func Load(path string, v interface{}, loadEnv ...bool) error {
	glog.Debugf("从文件加载配置: %s", path)

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return gerror.Newf(ErrLoadConfig, "配置文件不存在: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return gerror.Wrapf(err, ErrLoadConfig, "读取配置文件失败: %s", path)
	}

	// 根据文件扩展名选择解析方法
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(content, v); err != nil {
			return gerror.Wrapf(err, ErrParseConfig, "解析JSON配置文件失败: %s", path)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, v); err != nil {
			return gerror.Wrapf(err, ErrParseConfig, "解析YAML配置文件失败: %s", path)
		}
	case ".env":
		if err := parseEnvFile(content, v); err != nil {
			return gerror.Wrapf(err, ErrParseConfig, "解析ENV配置文件失败: %s", path)
		}
	default:
		return gerror.Newf(ErrParseConfig, "不支持的配置文件类型: %s", ext)
	}

	// 默认处理默认值
	if err := processConfigWithOptions(v, false, true); err != nil {
		return err
	}

	// 判断是否需要处理环境变量
	processEnv := true
	if len(loadEnv) > 0 {
		processEnv = loadEnv[0]
	}

	if processEnv {
		glog.Debug("处理配置的环境变量覆盖")
		if err := processConfigWithOptions(v, true, false); err != nil {
			return err
		}
	}

	glog.Debug("配置加载完成")
	return nil
}

// parseEnvFile 解析.env文件内容到结构体
func parseEnvFile(content []byte, v interface{}) error {
	// 解析.env文件为key=value格式
	lines := strings.Split(string(content), "\n")
	envMap := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // 跳过空行和注释
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // 跳过格式不正确的行
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 去除值两侧的引号
		if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"' ||
			value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}

		envMap[key] = value
	}

	// 设置结构体字段的值
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return gerror.New(ErrInvalidType, "配置必须是结构体指针类型")
	}

	// 获取结构体类型和值
	val = val.Elem()
	typ := val.Type()

	// 遍历结构体字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}

		fieldType := typ.Field(i)
		envTag := fieldType.Tag.Get(envTagName)

		// 检查环境变量是否在文件中
		if envTag != "" {
			if envValue, ok := envMap[envTag]; ok {
				// 设置字段值
				if err := setFieldFromString(field, envValue); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// LoadFromStruct 直接从结构体加载配置，处理环境变量和默认值
func LoadFromStruct(v interface{}) error {
	glog.Debug("从结构体加载配置")
	return processConfig(v)
}

// MustLoadFromStruct 直接从结构体加载配置，如果失败则panic
func MustLoadFromStruct(v interface{}) {
	if err := LoadFromStruct(v); err != nil {
		panic(err)
	}
}

// LoadDefaultsFromStruct 从结构体加载配置，只处理默认值
func LoadDefaultsFromStruct(v interface{}) error {
	glog.Debug("从结构体加载默认值配置")
	return processConfigWithOptions(v, false, true)
}

// LoadEnvFromStruct 从结构体加载配置，只处理环境变量
func LoadEnvFromStruct(v interface{}) error {
	glog.Debug("从结构体加载环境变量配置")
	return processConfigWithOptions(v, true, false)
}

// processConfig 处理结构体中的环境变量和默认值标签
func processConfig(v interface{}) error {
	return processConfigWithOptions(v, true, true)
}

// processConfigWithOptions 处理结构体中的标签，可选择是否处理环境变量和默认值
func processConfigWithOptions(v interface{}, processEnv bool, processDefault bool) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return gerror.New(ErrInvalidType, "配置必须是指针类型")
	}

	// 获取指针指向的值
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return gerror.New(ErrInvalidType, "配置必须是结构体类型")
	}

	return processStructWithOptions(val, processEnv, processDefault)
}

// processStructWithOptions 递归处理结构体中的字段，可选择是否处理环境变量和默认值
func processStructWithOptions(val reflect.Value, processEnv bool, processDefault bool) error {
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 如果字段不可设置，跳过
		if !field.CanSet() {
			continue
		}

		// 处理嵌套结构体
		if field.Kind() == reflect.Struct {
			if err := processStructWithOptions(field, processEnv, processDefault); err != nil {
				return err
			}
			continue
		}

		// 处理指针类型的嵌套结构体
		if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			if err := processStructWithOptions(field.Elem(), processEnv, processDefault); err != nil {
				return err
			}
			continue
		}

		// 获取环境变量标签
		if processEnv {
			envKey, hasEnv := fieldType.Tag.Lookup(envTagName)
			if hasEnv {
				// 获取环境变量值
				envValue, exists := os.LookupEnv(envKey)
				if exists {
					glog.Debugf("从环境变量加载配置: %s=%s", envKey, envValue)
					// 设置字段值
					if err := setFieldFromString(field, envValue); err != nil {
						return gerror.Wrapf(err, ErrInvalidValue, "设置字段 %s 的环境变量值失败", fieldType.Name)
					}
					continue
				}
			}
		}

		// 如果没有环境变量值，检查默认值
		if processDefault {
			defaultValue, hasDefault := fieldType.Tag.Lookup(defaultTagName)
			if hasDefault && isFieldZeroValue(field) {
				glog.Debugf("使用默认值设置字段 %s=%s", fieldType.Name, defaultValue)
				// 设置默认值
				if err := setFieldFromString(field, defaultValue); err != nil {
					return gerror.Wrapf(err, ErrInvalidValue, "设置字段 %s 的默认值失败", fieldType.Name)
				}
			}
		}
	}

	return nil
}

// setFieldFromString 根据字符串值设置字段值
func setFieldFromString(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 特殊处理Duration类型
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
			return nil
		}

		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Slice:
		// 处理切片类型，假设用逗号分隔
		if field.Type().Elem().Kind() == reflect.String {
			values := strings.Split(value, ",")
			slice := reflect.MakeSlice(field.Type(), len(values), len(values))
			for i, v := range values {
				slice.Index(i).SetString(strings.TrimSpace(v))
			}
			field.Set(slice)
		} else {
			return fmt.Errorf("不支持的切片元素类型: %s", field.Type().Elem().Kind())
		}
	default:
		return fmt.Errorf("不支持的字段类型: %s", field.Kind())
	}

	return nil
}

// isFieldZeroValue 检查字段是否为零值
func isFieldZeroValue(field reflect.Value) bool {
	return reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface())
}
