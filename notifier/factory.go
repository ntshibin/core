package notifier

import (
	"fmt"
	"sync"

	"github.com/ntshibin/core/errorx"
	"github.com/ntshibin/core/notifier/provider"
)

// ProviderFactory 工厂类
type ProviderFactory struct {
	providers map[string]provider.NotificationSender
	mu        sync.RWMutex
}

var factoryInstance *ProviderFactory
var once sync.Once

// GetFactory 获取工厂实例
func GetFactory() *ProviderFactory {
	once.Do(func() {
		factoryInstance = &ProviderFactory{
			providers: make(map[string]provider.NotificationSender),
		}
	})
	return factoryInstance
}

// RegisterProvider 注册通知提供商
func (f *ProviderFactory) RegisterProvider(driver, providerName string, sender provider.NotificationSender) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := fmt.Sprintf("%s:%s", driver, providerName)
	f.providers[key] = sender
}

// GetProvider 获取通知提供商
func (f *ProviderFactory) GetProvider(driver, providerName string) (provider.NotificationSender, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", driver, providerName)
	if sender, exists := f.providers[key]; exists {
		return sender, nil
	}
	return nil, errorx.New(errorx.HTTPCodeNotFound, fmt.Sprintf("未找到通知提供商: %s", key))
}
