package provider

import "context"

// NotificationSender 统一通知接口
type NotificationSender interface {
	Send(ctx context.Context, msg Message) (*MessageRes, error)
}
