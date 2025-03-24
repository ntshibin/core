package providers

import (
	"context"
	"sync"
	"time"

	"github.com/ntshibin/core/gcache"
	"github.com/ntshibin/core/glog"
)

// 定义内存缓存相关的常量
const (
	// 内存缓存默认大小
	DefaultMemoryCacheSize = 10000

	// 内存缓存默认淘汰策略
	DefaultEvictionPolicy = "LRU"

	// 内存缓存默认清理间隔
	DefaultCleanupInterval = 10 * time.Minute

	// 支持的淘汰策略
	EvictionLRU  = "LRU"  // 最近最少使用
	EvictionLFU  = "LFU"  // 最不经常使用
	EvictionFIFO = "FIFO" // 先进先出
)

// memoryItem 表示内存缓存中的一个项
type memoryItem struct {
	key         string        // 缓存键
	value       interface{}   // 缓存值
	expiration  time.Time     // 过期时间
	ttl         time.Duration // 生存时间
	accessTime  time.Time     // 最后访问时间（用于LRU）
	accessCount int64         // 访问计数（用于LFU）
	createTime  time.Time     // 创建时间（用于FIFO）
}

// 检查项是否已过期
func (i *memoryItem) isExpired() bool {
	if i.expiration.IsZero() {
		return false
	}
	return time.Now().After(i.expiration)
}

// MemoryCache 实现基于内存的缓存提供者
type MemoryCache struct {
	items    map[string]*memoryItem // 缓存项
	mutex    sync.RWMutex           // 读写锁
	config   *gcache.MemoryConfig   // 配置
	janitor  *time.Ticker           // 清理定时器
	stopChan chan struct{}          // 停止信号
	stats    *memoryCacheStats      // 统计信息
}

// memoryCacheStats 记录内存缓存的统计信息
type memoryCacheStats struct {
	hits             int64        // 缓存命中次数
	misses           int64        // 缓存未命中次数
	sets             int64        // 设置操作次数
	deletes          int64        // 删除操作次数
	evictions        int64        // 淘汰次数
	expirations      int64        // 过期次数
	currentItemCount int64        // 当前缓存项数量
	lastCleanup      time.Time    // 最后清理时间
	startTime        time.Time    // 启动时间
	mutex            sync.RWMutex // 统计数据锁
}

// 创建新的内存缓存
func NewMemoryCache(config *gcache.Config) (gcache.Provider, error) {
	// 使用默认内存配置（如果未提供）
	memConfig := &gcache.MemoryConfig{
		MaxSize:         DefaultMemoryCacheSize,
		EvictionPolicy:  DefaultEvictionPolicy,
		CleanupInterval: DefaultCleanupInterval,
	}

	if config != nil && config.MemoryConfig != nil {
		// 使用用户配置，但确保配置合理
		if config.MemoryConfig.MaxSize > 0 {
			memConfig.MaxSize = config.MemoryConfig.MaxSize
		}

		if config.MemoryConfig.CleanupInterval > 0 {
			memConfig.CleanupInterval = config.MemoryConfig.CleanupInterval
		}

		if config.MemoryConfig.EvictionPolicy != "" {
			memConfig.EvictionPolicy = config.MemoryConfig.EvictionPolicy
		}
	}

	// 验证淘汰策略
	switch memConfig.EvictionPolicy {
	case EvictionLRU, EvictionLFU, EvictionFIFO:
		// 支持的策略
	default:
		glog.Warnf("不支持的淘汰策略 %s，使用默认策略 LRU", memConfig.EvictionPolicy)
		memConfig.EvictionPolicy = EvictionLRU
	}

	stats := &memoryCacheStats{
		startTime:   time.Now(),
		lastCleanup: time.Now(),
	}

	cache := &MemoryCache{
		items:    make(map[string]*memoryItem),
		config:   memConfig,
		stopChan: make(chan struct{}),
		stats:    stats,
	}

	// 启动定期清理过期项的协程
	cache.startJanitor()

	return cache, nil
}

// startJanitor 启动后台清理过期项的协程
func (c *MemoryCache) startJanitor() {
	c.janitor = time.NewTicker(c.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-c.janitor.C:
				c.deleteExpired()
			case <-c.stopChan:
				c.janitor.Stop()
				return
			}
		}
	}()
}

// deleteExpired 删除所有过期的项
func (c *MemoryCache) deleteExpired() {
	now := time.Now()
	expiredKeys := make([]string, 0)

	// 查找所有过期的项
	c.mutex.RLock()
	for k, v := range c.items {
		if v.isExpired() {
			expiredKeys = append(expiredKeys, k)
		}
	}
	c.mutex.RUnlock()

	// 删除过期项
	if len(expiredKeys) > 0 {
		c.mutex.Lock()
		for _, k := range expiredKeys {
			if item, found := c.items[k]; found && item.isExpired() {
				delete(c.items, k)
				c.stats.mutex.Lock()
				c.stats.expirations++
				c.stats.currentItemCount--
				c.stats.mutex.Unlock()
			}
		}
		c.mutex.Unlock()

		if len(expiredKeys) > 0 {
			glog.Debugf("内存缓存：已清理 %d 个过期项", len(expiredKeys))
		}
	}

	// 更新最后清理时间
	c.stats.mutex.Lock()
	c.stats.lastCleanup = now
	c.stats.mutex.Unlock()
}

// Get 从缓存获取值
func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	if key == "" {
		return nil, gcache.ErrCacheKeyInvalid
	}

	c.mutex.RLock()
	item, found := c.items[key]
	c.mutex.RUnlock()

	// 缓存未命中
	if !found {
		c.stats.mutex.Lock()
		c.stats.misses++
		c.stats.mutex.Unlock()
		return nil, gcache.ErrCacheNotFound
	}

	// 检查是否过期
	if item.isExpired() {
		c.mutex.Lock()
		delete(c.items, key) // 删除过期项
		c.mutex.Unlock()

		c.stats.mutex.Lock()
		c.stats.expirations++
		c.stats.misses++
		c.stats.currentItemCount--
		c.stats.mutex.Unlock()

		return nil, gcache.ErrCacheNotFound
	}

	// 更新访问统计
	c.mutex.Lock()
	item.accessTime = time.Now()
	item.accessCount++
	c.mutex.Unlock()

	c.stats.mutex.Lock()
	c.stats.hits++
	c.stats.mutex.Unlock()

	return item.value, nil
}

// GetMulti 从缓存获取多个值
func (c *MemoryCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{}, len(keys))

	for _, key := range keys {
		if key == "" {
			continue
		}

		value, err := c.Get(ctx, key)
		if err == nil {
			result[key] = value
		}
	}

	return result, nil
}

// Set 设置缓存值
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return gcache.ErrCacheKeyInvalid
	}

	// 如果ttl <= 0，则使用永不过期
	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	item := &memoryItem{
		key:        key,
		value:      value,
		expiration: expiration,
		ttl:        ttl,
		accessTime: time.Now(),
		createTime: time.Now(),
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查是否需要淘汰
	if c.config.MaxSize > 0 && len(c.items) >= c.config.MaxSize {
		// 当前缓存已满，需要淘汰一个项
		if _, exists := c.items[key]; !exists {
			c.evictOne()
		}
	}

	// 检查键是否已存在，获取当前计数
	_, exists := c.items[key]
	c.items[key] = item

	c.stats.mutex.Lock()
	c.stats.sets++
	if !exists {
		c.stats.currentItemCount++
	}
	c.stats.mutex.Unlock()

	return nil
}

// Delete 删除缓存项
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return gcache.ErrCacheKeyInvalid
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, found := c.items[key]; found {
		delete(c.items, key)

		c.stats.mutex.Lock()
		c.stats.deletes++
		c.stats.currentItemCount--
		c.stats.mutex.Unlock()
	}

	return nil
}

// DeleteMulti 删除多个缓存项
func (c *MemoryCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := 0
	for _, key := range keys {
		if key == "" {
			continue
		}

		if _, found := c.items[key]; found {
			delete(c.items, key)
			count++
		}
	}

	if count > 0 {
		c.stats.mutex.Lock()
		c.stats.deletes += int64(count)
		c.stats.currentItemCount -= int64(count)
		c.stats.mutex.Unlock()
	}

	return nil
}

// Exists 检查键是否存在
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, gcache.ErrCacheKeyInvalid
	}

	c.mutex.RLock()
	item, found := c.items[key]
	c.mutex.RUnlock()

	if !found {
		return false, nil
	}

	// 检查是否过期
	if item.isExpired() {
		c.mutex.Lock()
		delete(c.items, key)
		c.mutex.Unlock()

		c.stats.mutex.Lock()
		c.stats.expirations++
		c.stats.currentItemCount--
		c.stats.mutex.Unlock()

		return false, nil
	}

	return true, nil
}

// Flush 清空所有缓存项
func (c *MemoryCache) Flush(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	itemCount := len(c.items)
	c.items = make(map[string]*memoryItem)

	c.stats.mutex.Lock()
	c.stats.currentItemCount = 0
	c.stats.mutex.Unlock()

	glog.Debugf("内存缓存：已清空 %d 个缓存项", itemCount)

	return nil
}

// GetTTL 获取缓存项的剩余生存时间
func (c *MemoryCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if key == "" {
		return 0, gcache.ErrCacheKeyInvalid
	}

	c.mutex.RLock()
	item, found := c.items[key]
	c.mutex.RUnlock()

	if !found {
		return 0, gcache.ErrCacheNotFound
	}

	// 检查是否过期
	if item.isExpired() {
		c.mutex.Lock()
		delete(c.items, key)
		c.mutex.Unlock()

		c.stats.mutex.Lock()
		c.stats.expirations++
		c.stats.currentItemCount--
		c.stats.mutex.Unlock()

		return 0, gcache.ErrCacheNotFound
	}

	// 如果没有设置过期时间
	if item.expiration.IsZero() {
		return -1, nil // -1 表示永不过期
	}

	// 计算剩余时间
	remaining := time.Until(item.expiration)
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

// Close 关闭缓存
func (c *MemoryCache) Close() error {
	// 停止清理协程
	close(c.stopChan)

	// 清空缓存
	c.mutex.Lock()
	c.items = make(map[string]*memoryItem)
	c.mutex.Unlock()

	glog.Debug("内存缓存已关闭")

	return nil
}

// GetStats 获取缓存统计信息
func (c *MemoryCache) GetStats() map[string]interface{} {
	c.stats.mutex.RLock()
	defer c.stats.mutex.RUnlock()

	uptime := time.Since(c.stats.startTime)

	return map[string]interface{}{
		"hits":              c.stats.hits,
		"misses":            c.stats.misses,
		"sets":              c.stats.sets,
		"deletes":           c.stats.deletes,
		"evictions":         c.stats.evictions,
		"expirations":       c.stats.expirations,
		"current_items":     c.stats.currentItemCount,
		"max_items":         c.config.MaxSize,
		"eviction_policy":   c.config.EvictionPolicy,
		"last_cleanup_time": c.stats.lastCleanup,
		"uptime_seconds":    int64(uptime.Seconds()),
		"hit_rate":          c.calculateHitRate(),
	}
}

// calculateHitRate 计算缓存命中率
func (c *MemoryCache) calculateHitRate() float64 {
	total := c.stats.hits + c.stats.misses
	if total == 0 {
		return 0
	}
	return float64(c.stats.hits) / float64(total) * 100
}

// evictOne 根据淘汰策略删除一个缓存项
func (c *MemoryCache) evictOne() {
	if len(c.items) == 0 {
		return
	}

	var keyToEvict string

	switch c.config.EvictionPolicy {
	case EvictionLRU:
		// 淘汰最近最少使用的项
		var oldest time.Time
		for k, v := range c.items {
			if oldest.IsZero() || v.accessTime.Before(oldest) {
				oldest = v.accessTime
				keyToEvict = k
			}
		}

	case EvictionLFU:
		// 淘汰最不经常使用的项
		var leastCount int64 = -1
		for k, v := range c.items {
			if leastCount == -1 || v.accessCount < leastCount {
				leastCount = v.accessCount
				keyToEvict = k
			}
		}

	case EvictionFIFO:
		// 淘汰最先创建的项
		var oldest time.Time
		for k, v := range c.items {
			if oldest.IsZero() || v.createTime.Before(oldest) {
				oldest = v.createTime
				keyToEvict = k
			}
		}

	default:
		// 默认使用LRU
		var oldest time.Time
		for k, v := range c.items {
			if oldest.IsZero() || v.accessTime.Before(oldest) {
				oldest = v.accessTime
				keyToEvict = k
			}
		}
	}

	if keyToEvict != "" {
		delete(c.items, keyToEvict)

		c.stats.mutex.Lock()
		c.stats.evictions++
		c.stats.currentItemCount--
		c.stats.mutex.Unlock()

		glog.Debugf("内存缓存：已淘汰键 %s（策略：%s）", keyToEvict, c.config.EvictionPolicy)
	}
}

// 初始化：注册内存缓存提供者
func init() {
	gcache.RegisterProvider("memory", func(config *gcache.Config) (gcache.Provider, error) {
		return NewMemoryCache(config)
	})
}
