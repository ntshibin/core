package conf

import "github.com/go-playground/validator/v10"

var validate = validator.New()

// Validate 验证配置结构体是否符合校验规则
// 支持使用结构体tag进行校验，如 `validate:"required"`
func Validate(config interface{}) error {
	return validate.Struct(config)
}

// MustValidate 验证配置，若不符合则panic
func MustValidate(config interface{}) {
	if err := Validate(config); err != nil {
		panic(err)
	}
}
