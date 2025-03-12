package convert

import (
	"fmt"
	"regexp"
)

// ValidateVariables 验证变量值是否符合规则
func ValidateVariables(msgMap map[string]interface{}, variables map[string]*VariableRule) (bool, error) {
	// 验证所有规则的Type字段是否存在
	for field, rule := range variables {
		if rule.Type == "" {
			return false, fmt.Errorf("字段 '%s' 的规则缺少必填参数type", field)
		}
	}

	// 验证所有必填字段是否存在
	for field, rule := range variables {
		if rule.Required {
			if _, exists := msgMap[field]; !exists {
				return false, fmt.Errorf("必填字段 '%s' 未提供", field)
			}
		}
	}

	// 验证msg中的字段
	for key, value := range msgMap {
		rule, exists := variables[key]
		if !exists {
			return false, fmt.Errorf("字段 '%s' 在模板变量中未定义", key)
		}

		// 类型验证
		switch rule.Type {
		case "string":
			str, ok := value.(string)
			if !ok {
				return false, fmt.Errorf("字段 '%s' 应为字符串类型", key)
			}
			// 字符串长度验证
			if rule.MinLen != nil && len(str) < *rule.MinLen {
				return false, fmt.Errorf("字段 '%s' 长度不能小于 %d", key, *rule.MinLen)
			}
			if rule.MaxLen != nil && len(str) > *rule.MaxLen {
				return false, fmt.Errorf("字段 '%s' 长度不能大于 %d", key, *rule.MaxLen)
			}
			// 正则表达式验证
			if rule.Pattern != nil {
				matched, err := regexp.MatchString(*rule.Pattern, str)
				if err != nil {
					return false, fmt.Errorf("字段 '%s' 正则表达式验证失败: %v", key, err)
				}
				if !matched {
					return false, fmt.Errorf("字段 '%s' 不匹配正则表达式 %s", key, *rule.Pattern)
				}
			}
		case "number":
			num, ok := value.(float64)
			if !ok {
				// 尝试将整数转换为float64
				if numInt, ok := value.(int); ok {
					num = float64(numInt)
				} else {
					return false, fmt.Errorf("字段 '%s' 应为数值类型", key)
				}
			}
			// 数值范围验证
			if rule.Min != nil && num < *rule.Min {
				return false, fmt.Errorf("字段 '%s' 不能小于 %v", key, *rule.Min)
			}
			if rule.Max != nil && num > *rule.Max {
				return false, fmt.Errorf("字段 '%s' 不能大于 %v", key, *rule.Max)
			}
		case "boolean":
			if _, ok := value.(bool); !ok {
				return false, fmt.Errorf("字段 '%s' 应为布尔类型", key)
			}
		default:
			return false, fmt.Errorf("字段 '%s' 的类型 '%s' 不支持", key, rule.Type)
		}

		// 枚举值验证
		if rule.Enum != nil {
			valid := false
			for _, enumVal := range rule.Enum {
				if value == enumVal {
					valid = true
					break
				}
			}
			if !valid {
				return false, fmt.Errorf("字段 '%s' 的值不在允许的枚举范围内", key)
			}
		}
	}
	return true, nil
}
