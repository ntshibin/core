package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"
)

// FileCacheConfig 文件缓存配置
type FileCacheConfig struct {
	// Directory 缓存目录
	Directory string `yaml:"directory"`
}

// FileCache 文件存储实现
type FileCache struct {
	mutex           sync.RWMutex
	directory       string
	cleanupInterval time.Duration
	stopCleanup     chan bool
	stats           *StatsCollector
	tags            map[string][]string
	listeners       []EventListener
	data            map[string]*fileItem
}

// item 缓存项
type fileItem struct {
	Value      interface{} `json:"value"`
	Expiration *time.Time  `json:"expiration"`
	Tags       []string    `json:"tags"`
}

// NewFileCache 创建文件缓存实例
func NewFileCache(config *BaseConfig, cacheConfig *FileCacheConfig) *FileCache {
	cache := &FileCache{
		directory:       cacheConfig.Directory,
		cleanupInterval: time.Duration(config.CleanupInterval) * time.Second,
		stopCleanup:     make(chan bool),
		stats:           NewStatsCollector(),
		tags:            make(map[string][]string),
		listeners:       make([]EventListener, 0),
		data:            make(map[string]*fileItem),
	}

	// 确保目录存在
	if err := os.MkdirAll(cache.directory, 0755); err != nil {
		panic(fmt.Sprintf("failed to create cache directory: %v", err))
	}

	// 启动清理协程
	go cache.startCleanup()

	return cache
}

// Set 设置缓存
func (c *FileCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiration := time.Now().Add(ttl)
	item := &fileItem{
		Value:      value,
		Expiration: &expiration,
	}

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal cache item: %v", err)
	}

	filePath := filepath.Join(c.directory, key)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}

	c.stats.IncrKeyCount()
	c.notifyListeners(EventTypeSet, key)

	return nil
}

// Get 获取缓存
func (c *FileCache) Get(ctx context.Context, key string, value interface{}) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	filePath := filepath.Join(c.directory, key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.stats.IncrMisses()
			return ErrNotFound
		}
		return fmt.Errorf("failed to read cache file: %v", err)
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return fmt.Errorf("failed to unmarshal cache item: %v", err)
	}

	if item.Expiration != nil && time.Now().After(*item.Expiration) {
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
	cachedValue := reflect.ValueOf(item.Value)
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
func (c *FileCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	filePath := filepath.Join(c.directory, key)
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete cache file: %v", err)
	}

	// 删除标签关系
	if item, err := c.readItem(key); err == nil {
		for _, tag := range item.Tags {
			if keys, ok := c.tags[tag]; ok {
				for i, k := range keys {
					if k == key {
						c.tags[tag] = append(keys[:i], keys[i+1:]...)
						break
					}
				}
			}
		}
	}

	c.stats.DecrKeyCount()
	c.notifyListeners(EventTypeDelete, key)

	return nil
}

// Has 检查缓存是否存在
func (c *FileCache) Has(ctx context.Context, key string) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	filePath := filepath.Join(c.directory, key)
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check cache file: %v", err)
	}

	item, err := c.readItem(key)
	if err != nil {
		return false, err
	}

	if item.Expiration != nil && time.Now().After(*item.Expiration) {
		return false, nil
	}

	return true, nil
}

// Clear 清空所有缓存
func (c *FileCache) Clear(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := os.RemoveAll(c.directory); err != nil {
		return fmt.Errorf("failed to clear cache directory: %v", err)
	}

	if err := os.MkdirAll(c.directory, 0755); err != nil {
		return fmt.Errorf("failed to recreate cache directory: %v", err)
	}

	c.tags = make(map[string][]string)
	c.stats.Reset()
	c.notifyListeners(EventTypeClear, "")

	return nil
}

// GetStats 获取缓存统计信息
func (c *FileCache) GetStats(ctx context.Context) (*Stats, error) {
	stats := c.stats.GetStats()
	return &stats, nil
}

// HealthCheck 执行健康检查
func (c *FileCache) HealthCheck(ctx context.Context) (*Health, error) {
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
func (c *FileCache) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, value := range items {
		expiration := time.Now().Add(ttl)
		item := &fileItem{
			Value:      value,
			Expiration: &expiration,
		}

		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal cache item: %v", err)
		}

		filePath := filepath.Join(c.directory, key)
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write cache file: %v", err)
		}

		c.stats.IncrKeyCount()
		c.notifyListeners(EventTypeSet, key)
	}

	return nil
}

// MGet 批量获取缓存
func (c *FileCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make(map[string]interface{})
	for _, key := range keys {
		filePath := filepath.Join(c.directory, key)
		data, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				c.stats.IncrMisses()
				continue
			}
			return nil, fmt.Errorf("failed to read cache file: %v", err)
		}

		var item fileItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cache item: %v", err)
		}

		if item.Expiration != nil && time.Now().After(*item.Expiration) {
			c.stats.IncrMisses()
			c.stats.IncrExpiredCount()
			continue
		}

		result[key] = item.Value
		c.stats.IncrHits()
		c.notifyListeners(EventTypeGet, key)
	}

	return result, nil
}

// MDelete 批量删除缓存
func (c *FileCache) MDelete(ctx context.Context, keys []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, key := range keys {
		filePath := filepath.Join(c.directory, key)
		if err := os.Remove(filePath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to delete cache file: %v", err)
		}

		// 删除标签关系
		if item, err := c.readItem(key); err == nil {
			for _, tag := range item.Tags {
				if keys, ok := c.tags[tag]; ok {
					for i, k := range keys {
						if k == key {
							c.tags[tag] = append(keys[:i], keys[i+1:]...)
							break
						}
					}
				}
			}
		}

		c.stats.DecrKeyCount()
		c.notifyListeners(EventTypeDelete, key)
	}

	return nil
}

// SetWithTags 设置带标签的缓存
func (c *FileCache) SetWithTags(ctx context.Context, key string, value interface{}, tags []string, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiration := time.Now().Add(ttl)
	item := &fileItem{
		Value:      value,
		Expiration: &expiration,
		Tags:       tags,
	}

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal cache item: %v", err)
	}

	filePath := filepath.Join(c.directory, key)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}

	// 更新标签关系
	for _, tag := range tags {
		c.tags[tag] = append(c.tags[tag], key)
	}

	c.stats.IncrKeyCount()
	c.notifyListeners(EventTypeSet, key)

	return nil
}

// GetByTag 获取指定标签的所有缓存键
func (c *FileCache) GetByTag(ctx context.Context, tag string) ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if keys, ok := c.tags[tag]; ok {
		return keys, nil
	}
	return nil, nil
}

// DeleteByTag 删除指定标签的所有缓存
func (c *FileCache) DeleteByTag(ctx context.Context, tag string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if keys, ok := c.tags[tag]; ok {
		for _, key := range keys {
			filePath := filepath.Join(c.directory, key)
			if err := os.Remove(filePath); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return fmt.Errorf("failed to delete cache file: %v", err)
			}

			// 删除标签关系
			if item, err := c.readItem(key); err == nil {
				for _, t := range item.Tags {
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
			}

			c.stats.DecrKeyCount()
			c.notifyListeners(EventTypeDelete, key)
		}
		delete(c.tags, tag)
	}

	return nil
}

// AddEventListener 添加事件监听器
func (c *FileCache) AddEventListener(listener EventListener) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.listeners = append(c.listeners, listener)
}

// RemoveEventListener 移除事件监听器
func (c *FileCache) RemoveEventListener(listener EventListener) {
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
func (c *FileCache) notifyListeners(eventType EventType, key string) {
	for _, listener := range c.listeners {
		listener.OnEvent(eventType, key)
	}
}

// ResetStats 重置统计信息
func (c *FileCache) ResetStats(ctx context.Context) error {
	c.stats.Reset()
	return nil
}

// FileLock 文件分布式锁实现
type FileLock struct {
	cache      *FileCache
	key        string
	expiration time.Duration
	value      int64
}

// Lock 获取锁
func (l *FileLock) Lock(ctx context.Context) error {
	l.cache.mutex.Lock()
	defer l.cache.mutex.Unlock()

	// 检查锁是否已存在
	if item, exists := l.cache.data[l.key]; exists {
		if item.Expiration != nil && time.Now().After(*item.Expiration) {
			delete(l.cache.data, l.key)
		} else {
			return fmt.Errorf("lock already exists")
		}
	}

	// 设置锁
	expiration := time.Now().Add(l.expiration)
	l.value = time.Now().UnixNano()
	l.cache.data[l.key] = &fileItem{
		Value:      l.value,
		Expiration: &expiration,
	}

	return nil
}

// Unlock 释放锁
func (l *FileLock) Unlock(ctx context.Context) error {
	l.cache.mutex.Lock()
	defer l.cache.mutex.Unlock()

	if item, exists := l.cache.data[l.key]; exists {
		if item.Value == l.value {
			delete(l.cache.data, l.key)
			return nil
		}
		return fmt.Errorf("lock value mismatch")
	}

	return fmt.Errorf("lock not found")
}

// Refresh 刷新锁的过期时间
func (l *FileLock) Refresh(ctx context.Context) error {
	l.cache.mutex.Lock()
	defer l.cache.mutex.Unlock()

	if item, exists := l.cache.data[l.key]; exists {
		if item.Value == l.value {
			expiration := time.Now().Add(l.expiration)
			item.Expiration = &expiration
			return nil
		}
		return fmt.Errorf("lock value mismatch")
	}

	return fmt.Errorf("lock not found")
}

// readItem 读取缓存项
func (c *FileCache) readItem(key string) (*fileItem, error) {
	filePath := filepath.Join(c.directory, key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

// startCleanup 启动清理协程
func (c *FileCache) startCleanup() {
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
func (c *FileCache) deleteExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	entries, err := os.ReadDir(c.directory)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		key := entry.Name()
		item, err := c.readItem(key)
		if err != nil {
			continue
		}

		if item.Expiration != nil && now.After(*item.Expiration) {
			// 删除标签关系
			for _, tag := range item.Tags {
				if keys, ok := c.tags[tag]; ok {
					for i, k := range keys {
						if k == key {
							c.tags[tag] = append(keys[:i], keys[i+1:]...)
							break
						}
					}
				}
			}

			os.Remove(filepath.Join(c.directory, key))
			c.stats.DecrKeyCount()
			c.stats.IncrExpiredCount()
			c.notifyListeners(EventTypeDelete, key)
		}
	}
}
