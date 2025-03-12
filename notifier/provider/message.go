package provider

import "errors"

type MessageType string

const (
	MessageTypeEmail MessageType = "email"
	MessageTypeSMS   MessageType = "sms"
)

type Message struct {
	// 邮件接收者
	To string `json:"to"`

	ProviderType MessageType `json:"provider_type"`

	Content MessageContent `json:"content"`
}

type MessageContent interface {
	Validate() error
}

type MessageEmail struct {
	// 邮件主题
	Subject string `json:"subject"`
	// 邮件内容
	Content string `json:"content"`
}

func (m *MessageEmail) Validate() error {
	if m.Subject == "" {
		return errors.New("subject is empty")
	}
	if m.Content == "" {
		return errors.New("content is empty")
	}
	return nil
}

type MessageSMS struct {
	// 短信签名
	SignName string `json:"sign_name"`
	// 短信模板Code
	TemplateCode string `json:"template_code"`
	// 短信模板参数
	TemplateParam map[string]interface{} `json:"template_param"`
}

func (m *MessageSMS) Validate() error {
	if m.SignName == "" {
		return errors.New("sign_name is empty")
	}
	if m.TemplateCode == "" {
		return errors.New("template_code is empty")
	}
	if len(m.TemplateParam) == 0 {
		return errors.New("template_param is empty")
	}
	return nil
}

type MessageRes struct {
	RequestID string `json:"request_id"`
}
