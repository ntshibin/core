package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCacheConfig Redis缓存配置
type RedisCacheConfig struct {
	// Redis连接地址
	Addr string
	// Redis密码
	Password string
	// Redis数据库
	DB int
}

// RedisCache Redis存储实现
type RedisCache struct {
	client    *redis.Client
	stats     *StatsCollector
	listeners []EventListener
	mutex     sync.RWMutex
	maxItems  int // 最大缓存项数量
}

// NewRedisCache 创建Redis缓存实例
func NewRedisCache(config *BaseConfig, cacheConfig *RedisCacheConfig) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     cacheConfig.Addr,
		Password: cacheConfig.Password,
		DB:       cacheConfig.DB,
	})

	return &RedisCache{
		client:    client,
		stats:     NewStatsCollector(),
		listeners: make([]EventListener, 0),
		maxItems:  config.MaxSize,
	}
}

// Set 设置缓存
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 检查当前缓存项数量
	keys, err := c.client.Keys(ctx, "*").Result()
	if err != nil {
		return err
	}

	// 如果超过最大数量，删除最早的缓存项
	if len(keys) >= c.maxItems {
		// 按时间排序获取所有键
		oldestKey := keys[0]
		for _, key := range keys {
			keyTTL, _ := c.client.TTL(ctx, key).Result()
			if keyTTL < c.client.TTL(ctx, oldestKey).Val() {
				oldestKey = key
			}
		}
		// 删除最早的缓存项
		c.client.Del(ctx, oldestKey)
	}
	// 序列化值
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 存储序列化后的数据
	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return err
	}

	c.stats.IncrKeyCount()
	c.notifyListeners(EventTypeSet, key)
	return nil
}

// Get 获取缓存
func (c *RedisCache) Get(ctx context.Context, key string, value interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get cache: %v", err)
	}

	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %v", err)
	}

	c.stats.IncrHits()
	c.notifyListeners(EventTypeGet, key)
	return nil
}

// Delete 删除缓存
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache: %v", err)
	}

	c.stats.DecrKeyCount()
	c.notifyListeners(EventTypeDelete, key)
	return nil
}

// Has 检查缓存是否存在
func (c *RedisCache) Has(ctx context.Context, key string) (bool, error) {
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache: %v", err)
	}

	return exists > 0, nil
}

// Clear 清空所有缓存
func (c *RedisCache) Clear(ctx context.Context) error {
	if err := c.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to clear cache: %v", err)
	}

	c.stats.Reset()
	c.notifyListeners(EventTypeClear, "")
	return nil
}

// MSet 批量设置缓存
func (c *RedisCache) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	pipe := c.client.Pipeline()
	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %v", err)
		}
		pipe.Set(ctx, key, data, ttl)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to set multiple caches: %v", err)
	}

	c.stats.IncrKeyCountBy(int64(len(items)))
	for key := range items {
		c.notifyListeners(EventTypeSet, key)
	}
	return nil
}

// MGet 批量获取缓存
func (c *RedisCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple caches: %v", err)
	}

	result := make(map[string]interface{})
	for i, value := range values {
		if value == nil {
			c.stats.IncrMisses()
			continue
		}

		var v interface{}
		if err := json.Unmarshal([]byte(value.(string)), &v); err != nil {
			return nil, fmt.Errorf("failed to unmarshal value: %v", err)
		}

		result[keys[i]] = v
		c.stats.IncrHits()
		c.notifyListeners(EventTypeGet, keys[i])
	}

	return result, nil
}

// MDelete 批量删除缓存
func (c *RedisCache) MDelete(ctx context.Context, keys []string) error {
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete multiple caches: %v", err)
	}

	c.stats.DecrKeyCountBy(int64(len(keys)))
	for _, key := range keys {
		c.notifyListeners(EventTypeDelete, key)
	}
	return nil
}

// SetWithTags 设置带标签的缓存
func (c *RedisCache) SetWithTags(ctx context.Context, key string, value interface{}, tags []string, ttl time.Duration) error {
	// 设置缓存值
	if err := c.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// 设置标签关系
	for _, tag := range tags {
		tagKey := fmt.Sprintf("tag:%s", tag)
		if err := c.client.SAdd(ctx, tagKey, key).Err(); err != nil {
			return fmt.Errorf("failed to set tag: %v", err)
		}
		if ttl > 0 {
			c.client.Expire(ctx, tagKey, ttl)
		}
	}

	return nil
}

// GetByTag 获取指定标签的所有缓存键
func (c *RedisCache) GetByTag(ctx context.Context, tag string) ([]string, error) {
	tagKey := fmt.Sprintf("tag:%s", tag)
	keys, err := c.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get tag keys: %v", err)
	}
	return keys, nil
}

// DeleteByTag 删除指定标签的所有缓存
func (c *RedisCache) DeleteByTag(ctx context.Context, tag string) error {
	tagKey := fmt.Sprintf("tag:%s", tag)
	keys, err := c.client.SMembers(ctx, tagKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get tag keys: %v", err)
	}

	if len(keys) > 0 {
		if err := c.MDelete(ctx, keys); err != nil {
			return err
		}
	}

	return c.client.Del(ctx, tagKey).Err()
}

// GetStats 获取缓存统计信息
func (c *RedisCache) GetStats(ctx context.Context) (*Stats, error) {
	stats := c.stats.GetStats()
	return &stats, nil
}

// HealthCheck 执行健康检查
func (c *RedisCache) HealthCheck(ctx context.Context) (*Health, error) {
	// 检查Redis连接
	if err := c.client.Ping(ctx).Err(); err != nil {
		return &Health{
			Status:    "unhealthy",
			Details:   map[string]interface{}{"error": err.Error()},
			Timestamp: time.Now(),
		}, nil
	}

	stats := c.stats.GetStats()
	return &Health{
		Status: "healthy",
		Details: map[string]interface{}{
			"key_count": stats.KeyCount,
			"hits":      stats.Hits,
			"misses":    stats.Misses,
		},
		Timestamp: time.Now(),
	}, nil
}

// AddEventListener 添加事件监听器
func (c *RedisCache) AddEventListener(listener EventListener) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.listeners = append(c.listeners, listener)
}

// RemoveEventListener 移除事件监听器
func (c *RedisCache) RemoveEventListener(listener EventListener) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i, l := range c.listeners {
		if l == listener {
			c.listeners = append(c.listeners[:i], c.listeners[i+1:]...)
			break
		}
	}
}

// notifyListeners 通知所有监听器
func (c *RedisCache) notifyListeners(eventType EventType, key string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	for _, listener := range c.listeners {
		listener.OnEvent(eventType, key)
	}
}

// ResetStats 重置统计信息
func (c *RedisCache) ResetStats(ctx context.Context) error {
	c.stats.Reset()
	return nil
}

// RedisLock Redis分布式锁实现
type RedisLock struct {
	client     *redis.Client
	key        string
	expiration time.Duration
}

// Lock 获取锁
func (l *RedisLock) Lock(ctx context.Context) error {
	ok, err := l.client.SetNX(ctx, l.key, time.Now().UnixNano(), l.expiration).Result()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}
	if !ok {
		return fmt.Errorf("lock already exists")
	}
	return nil
}

// Unlock 释放锁
func (l *RedisLock) Unlock(ctx context.Context) error {
	return l.client.Del(ctx, l.key).Err()
}

// Refresh 刷新锁的过期时间
func (l *RedisLock) Refresh(ctx context.Context) error {
	return l.client.Expire(ctx, l.key, l.expiration).Err()
}
