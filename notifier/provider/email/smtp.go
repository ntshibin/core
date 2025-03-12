package email

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/ntshibin/core/errorx"
	"github.com/ntshibin/core/notifier/provider"
	"gopkg.in/gomail.v2"
)

// SMTPEmail SMTP 邮件实现
type SMTPEmail struct {
	conf   *Config
	auth   smtp.Auth
	dialer *gomail.Dialer
}

// NewSMTPEmail 创建 SMTP 邮件实例
func NewSMTPEmail(conf *Config) provider.NotificationSender {
	// 创建SMTP认证
	auth := smtp.PlainAuth("", conf.Username, conf.Password, conf.Host)

	// 创建邮件发送器
	dialer := gomail.NewDialer(conf.Host, conf.Port, conf.Username, conf.Password)
	dialer.SSL = conf.UseSSL // 根据配置决定是否使用SSL

	return &SMTPEmail{
		conf:   conf,
		auth:   auth,
		dialer: dialer,
	}
}

// Send 发送邮件
func (s *SMTPEmail) Send(ctx context.Context, msg provider.Message) (res *provider.MessageRes, err error) {
	notice, ok := msg.Content.(*provider.MessageEmail)
	if !ok {
		return nil, errors.New("invalid message content type")
	}
	// 创建邮件消息
	m := gomail.NewMessage()
	m.SetHeader("From", s.conf.From)
	m.SetHeader("To", msg.To)
	m.SetHeader("Subject", notice.Subject)
	m.SetBody("text/plain", notice.Content)

	// 发送邮件
	if err := s.dialer.DialAndSend(m); err != nil {
		return nil, fmt.Errorf("%w: %v", errorx.ErrSendFailed, err)
	}
	requestID := fmt.Sprintf("%s_%s", strings.ReplaceAll(msg.To, "@", "_"), ctx.Value("request_id"))
	res = &provider.MessageRes{
		RequestID: requestID,
	}
	fmt.Printf("📧 [SMTP] 发送邮件至 %s，主题: %s，内容: %s\n", msg.To, notice.Subject, notice.Content)
	return
}
