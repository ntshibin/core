package cache

import (
	"context"
	"time"
)

// ICache 缓存接口
type ICache interface {
	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Get 获取缓存
	Get(ctx context.Context, key string, value interface{}) error
	// Delete 删除缓存
	Delete(ctx context.Context, key string) error
	// Has 检查缓存是否存在
	Has(ctx context.Context, key string) (bool, error)
	// Clear 清空所有缓存
	Clear(ctx context.Context) error
	// GetStats 获取缓存统计信息
	GetStats(ctx context.Context) (*Stats, error)
	// HealthCheck 执行健康检查
	HealthCheck(ctx context.Context) (*Health, error)
	// MSet 批量设置缓存
	MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	// MGet 批量获取缓存
	MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
	// MDelete 批量删除缓存
	MDelete(ctx context.Context, keys []string) error
}

// Health 健康检查结果
type Health struct {
	Status    string                 `json:"status"`
	Details   map[string]interface{} `json:"details"`
	Timestamp time.Time              `json:"timestamp"`
}
