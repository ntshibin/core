package email

import "errors"

type Config struct {
	Host     string `yaml:"host" json:"host" mapstructure:"host"`             // 邮件服务器地址
	Port     int    `yaml:"port" json:"port" mapstructure:"port"`             // 邮件服务器端口
	Username string `yaml:"username" json:"username" mapstructure:"username"` // 邮件服务器用户名
	Password string `yaml:"password" json:"password" mapstructure:"password"` // 邮件服务器密码
	From     string `yaml:"from" json:"from" mapstructure:"from"`             // 发件人邮箱地址
	UseSSL   bool   `yaml:"use_ssl" json:"use_ssl" mapstructure:"use_ssl"`    // 是否使用SSL
	UseTLS   bool   `yaml:"use_tls" json:"use_tls" mapstructure:"use_tls"`    // 是否使用TLS
}

func (c *Config) Validate() (err error) {
	if c.Host == "" {
		return errors.New("邮件服务器地址不能为空")
	}

	if c.Port == 0 {
		return errors.New("邮件服务器端口不能为空")
	}

	if c.Username == "" {
		return errors.New("邮件服务器用户名不能为空")
	}
	if c.Password == "" {
		return errors.New("邮件服务器密码不能为空")
	}
	if c.From == "" {
		return errors.New("发件人邮箱地址不能为空")
	}
	return
}
