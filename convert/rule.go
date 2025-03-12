package convert

// VariableRule 变量验证规则
type VariableRule struct {
	Type     string        `json:"type" binding:"required"` // string, number, boolean
	Required bool          `json:"required" default:"false"`
	MinLen   *int          `json:"min_len,omitempty"` // 字符串最小长度
	MaxLen   *int          `json:"max_len,omitempty"` // 字符串最大长度
	Min      *float64      `json:"min,omitempty"`     // 数值最小值
	Max      *float64      `json:"max,omitempty"`     // 数值最大值
	Pattern  *string       `json:"pattern,omitempty"` // 正则表达式
	Enum     []interface{} `json:"enum,omitempty"`    // 枚举值
}
