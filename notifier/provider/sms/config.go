package sms

import "errors"

type Config struct {
	AccessKeyID     string `yaml:"access_key_id" json:"access_key_id" mapstructure:"access_key_id"`             // 阿里云访问密钥ID
	AccessKeySecret string `yaml:"access_key_secret" json:"access_key_secret" mapstructure:"access_key_secret"` // 阿里云访问密钥密码           // 短信模板ID
	RegionID        string `yaml:"region_id" json:"region_id" mapstructure:"region_id"`                         // 地域ID
}

func (c *Config) Validate() error {
	if c.AccessKeyID == "" {
		return errors.New("ErrAccessKeyIDEmpty")
	}
	if c.AccessKeySecret == "" {
		return errors.New("ErrAccessKeySecretEmpty")
	}
	if c.RegionID == "" {
		return errors.New("ErrRegionIDEmpty")
	}
	return nil
}
