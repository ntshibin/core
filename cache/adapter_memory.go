package cache

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// MemoryCacheConfig 内存缓存配置
type MemoryCacheConfig struct {
	// Policy 缓存策略：lru, fifo
	Policy string `yaml:"policy"`
}

// MemoryCache 内存存储实现
type MemoryCache struct {
	mutex           sync.RWMutex
	data            map[string]*memoryItem
	tags            map[string][]string
	maxSize         int
	cleanupInterval time.Duration
	stopCleanup     chan bool
	stats           *StatsCollector
	policy          Policy
	config          *MemoryCacheConfig
	listeners       []EventListener
}

// item 缓存项
type memoryItem struct {
	value      interface{}
	expiration *time.Time
	tags       []string
}

// NewMemoryCache 创建内存缓存实例
func NewMemoryCache(config *BaseConfig, cacheConfig *MemoryCacheConfig) *MemoryCache {
	cache := &MemoryCache{
		data:            make(map[string]*memoryItem),
		tags:            make(map[string][]string),
		config:          cacheConfig,
		stats:           NewStatsCollector(),
		policy:          NewLRUPolicy(),
		maxSize:         config.MaxSize,
		cleanupInterval: time.Duration(config.CleanupInterval) * time.Second,
		stopCleanup:     make(chan bool),
		listeners:       make([]EventListener, 0),
	}

	// 启动清理协程
	go cache.startCleanup()

	return cache
}

// Set 设置缓存
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查是否需要驱逐
	if len(c.data) >= c.maxSize {
		c.evictOne()
	}

	expiration := time.Now().Add(ttl)
	item := &memoryItem{
		value:      value,
		expiration: &expiration,
	}

	c.data[key] = item
	c.policy.Update(key, item)
	c.stats.IncrKeyCount()
	c.notifyListeners(EventTypeSet, key)

	return nil
}

// Get 获取缓存
func (c *MemoryCache) Get(ctx context.Context, key string, value interface{}) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		c.stats.IncrMisses()
		return ErrNotFound
	}

	if item.expiration != nil && time.Now().After(*item.expiration) {
		c.stats.IncrMisses()
		c.stats.IncrExpiredCount()
		return ErrNotFound
	}

	// 使用反射实现值的拷贝
	valuePtr := reflect.ValueOf(value)
	if valuePtr.Kind() != reflect.Ptr {
		return ErrInvalidValue
	}

	valueElem := valuePtr.Elem()
	cachedValue := reflect.ValueOf(item.value)
	if cachedValue.Kind() == reflect.Ptr {
		cachedValue = cachedValue.Elem()
	}

	if !cachedValue.Type().AssignableTo(valueElem.Type()) {
		return fmt.Errorf("cannot assign cached value of type %v to value of type %v", cachedValue.Type(), valueElem.Type())
	}

	valueElem.Set(cachedValue)
	c.stats.IncrHits()
	c.notifyListeners(EventTypeGet, key)

	return nil
}

// Delete 删除缓存
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, exists := c.data[key]; exists {
		// 删除标签关系
		for _, tag := range item.tags {
			if keys, ok := c.tags[tag]; ok {
				for i, k := range keys {
					if k == key {
						c.tags[tag] = append(keys[:i], keys[i+1:]...)
						break
					}
				}
			}
		}

		delete(c.data, key)
		c.stats.DecrKeyCount()
		c.notifyListeners(EventTypeDelete, key)
	}

	return nil
}

// Has 检查缓存是否存在
func (c *MemoryCache) Has(ctx context.Context, key string) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false, nil
	}

	if item.expiration != nil && time.Now().After(*item.expiration) {
		return false, nil
	}

	return true, nil
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*memoryItem)
	c.tags = make(map[string][]string)
	c.stats.Reset()
	c.notifyListeners(EventTypeClear, "")

	return nil
}

// GetStats 获取缓存统计信息
func (c *MemoryCache) GetStats(ctx context.Context) (*Stats, error) {
	stats := c.stats.GetStats()
	return &stats, nil
}

// HealthCheck 执行健康检查
func (c *MemoryCache) HealthCheck(ctx context.Context) (*Health, error) {
	stats := c.stats.GetStats()
	return &Health{
		Status: "healthy",
		Details: map[string]interface{}{
			"key_count":     stats.KeyCount,
			"hits":          stats.Hits,
			"misses":        stats.Misses,
			"evicted_count": stats.EvictedCount,
			"expired_count": stats.ExpiredCount,
		},
		Timestamp: time.Now(),
	}, nil
}

// MSet 批量设置缓存
func (c *MemoryCache) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, value := range items {
		// 检查是否需要驱逐
		if len(c.data) >= c.maxSize {
			c.evictOne()
		}

		expiration := time.Now().Add(ttl)
		item := &memoryItem{
			value:      value,
			expiration: &expiration,
		}

		c.data[key] = item
		c.policy.Update(key, item)
		c.stats.IncrKeyCount()
		c.notifyListeners(EventTypeSet, key)
	}

	return nil
}

// MGet 批量获取缓存
func (c *MemoryCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make(map[string]interface{})
	for _, key := range keys {
		item, exists := c.data[key]
		if !exists {
			c.stats.IncrMisses()
			continue
		}

		if item.expiration != nil && time.Now().After(*item.expiration) {
			c.stats.IncrMisses()
			c.stats.IncrExpiredCount()
			continue
		}

		result[key] = item.value
		c.stats.IncrHits()
		c.notifyListeners(EventTypeGet, key)
	}

	return result, nil
}

// MDelete 批量删除缓存
func (c *MemoryCache) MDelete(ctx context.Context, keys []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, key := range keys {
		if item, exists := c.data[key]; exists {
			// 删除标签关系
			for _, tag := range item.tags {
				if keys, ok := c.tags[tag]; ok {
					for i, k := range keys {
						if k == key {
							c.tags[tag] = append(keys[:i], keys[i+1:]...)
							break
						}
					}
				}
			}

			delete(c.data, key)
			c.stats.DecrKeyCount()
			c.notifyListeners(EventTypeDelete, key)
		}
	}

	return nil
}

// SetWithTags 设置带标签的缓存
func (c *MemoryCache) SetWithTags(ctx context.Context, key string, value interface{}, tags []string, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查是否需要驱逐
	if len(c.data) >= c.maxSize {
		c.evictOne()
	}

	expiration := time.Now().Add(ttl)
	item := &memoryItem{
		value:      value,
		expiration: &expiration,
		tags:       tags,
	}

	// 更新标签关系
	for _, tag := range tags {
		c.tags[tag] = append(c.tags[tag], key)
	}

	c.data[key] = item
	c.policy.Update(key, item)
	c.stats.IncrKeyCount()
	c.notifyListeners(EventTypeSet, key)

	return nil
}

// GetByTag 获取指定标签的所有缓存键
func (c *MemoryCache) GetByTag(ctx context.Context, tag string) ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if keys, ok := c.tags[tag]; ok {
		return keys, nil
	}
	return nil, nil
}

// DeleteByTag 删除指定标签的所有缓存
func (c *MemoryCache) DeleteByTag(ctx context.Context, tag string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if keys, ok := c.tags[tag]; ok {
		for _, key := range keys {
			if item, exists := c.data[key]; exists {
				// 删除标签关系
				for _, t := range item.tags {
					if t != tag {
						if ks, ok := c.tags[t]; ok {
							for i, k := range ks {
								if k == key {
									c.tags[t] = append(ks[:i], ks[i+1:]...)
									break
								}
							}
						}
					}
				}

				delete(c.data, key)
				c.stats.DecrKeyCount()
				c.notifyListeners(EventTypeDelete, key)
			}
		}
		delete(c.tags, tag)
	}

	return nil
}

// AddEventListener 添加事件监听器
func (c *MemoryCache) AddEventListener(listener EventListener) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.listeners = append(c.listeners, listener)
}

// RemoveEventListener 移除事件监听器
func (c *MemoryCache) RemoveEventListener(listener EventListener) {
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
func (c *MemoryCache) notifyListeners(eventType EventType, key string) {
	for _, listener := range c.listeners {
		listener.OnEvent(eventType, key)
	}
}

// ResetStats 重置统计信息
func (c *MemoryCache) ResetStats(ctx context.Context) error {
	c.stats.Reset()
	return nil
}

// MemoryLock 内存分布式锁实现
type MemoryLock struct {
	cache      *MemoryCache
	key        string
	expiration time.Duration
	value      int64
}

// Lock 获取锁
func (l *MemoryLock) Lock(ctx context.Context) error {
	l.cache.mutex.Lock()
	defer l.cache.mutex.Unlock()

	if item, exists := l.cache.data[l.key]; exists {
		if item.expiration != nil && time.Now().After(*item.expiration) {
			delete(l.cache.data, l.key)
		} else {
			return fmt.Errorf("lock already exists")
		}
	}

	expiration := time.Now().Add(l.expiration)
	l.value = time.Now().UnixNano()
	l.cache.data[l.key] = &memoryItem{
		value:      l.value,
		expiration: &expiration,
	}

	return nil
}

// Unlock 释放锁
func (l *MemoryLock) Unlock(ctx context.Context) error {
	l.cache.mutex.Lock()
	defer l.cache.mutex.Unlock()

	if item, exists := l.cache.data[l.key]; exists {
		if item.value == l.value {
			delete(l.cache.data, l.key)
			return nil
		}
	}

	return fmt.Errorf("lock not found or value mismatch")
}

// Refresh 刷新锁的过期时间
func (l *MemoryLock) Refresh(ctx context.Context) error {
	l.cache.mutex.Lock()
	defer l.cache.mutex.Unlock()

	if item, exists := l.cache.data[l.key]; exists {
		if item.value == l.value {
			expiration := time.Now().Add(l.expiration)
			item.expiration = &expiration
			return nil
		}
	}

	return fmt.Errorf("lock not found or value mismatch")
}

// evictOne 根据策略驱逐一个缓存项
func (c *MemoryCache) evictOne() {
	if c.policy == nil {
		c.policy = NewLRUPolicy()
	}
	key := c.policy.Evict(c.data)
	if key != "" {
		if item, exists := c.data[key]; exists {
			// 删除标签关系
			for _, tag := range item.tags {
				if keys, ok := c.tags[tag]; ok {
					for i, k := range keys {
						if k == key {
							c.tags[tag] = append(keys[:i], keys[i+1:]...)
							break
						}
					}
				}
			}

			delete(c.data, key)
			c.stats.DecrKeyCount()
			c.stats.IncrEvictedCount()
			c.notifyListeners(EventTypeDelete, key)
		}
	}
}

// startCleanup 启动清理协程
func (c *MemoryCache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// deleteExpired 删除过期的缓存项
func (c *MemoryCache) deleteExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.data {
		if item.expiration != nil && now.After(*item.expiration) {
			// 删除标签关系
			for _, tag := range item.tags {
				if keys, ok := c.tags[tag]; ok {
					for i, k := range keys {
						if k == key {
							c.tags[tag] = append(keys[:i], keys[i+1:]...)
							break
						}
					}
				}
			}

			delete(c.data, key)
			c.stats.DecrKeyCount()
			c.stats.IncrExpiredCount()
			c.notifyListeners(EventTypeDelete, key)
		}
	}
}
