package gcache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ntshibin/core/gerror"
	"github.com/ntshibin/core/glog"
)

var (
	instance     Cache                                              // 全局缓存实例
	instanceOnce sync.Once                                          // 确保单例初始化只执行一次
	providers    = make(map[string]func(*Config) (Provider, error)) // 缓存提供者注册表
)

// 默认配置
var defaultConfig = &Config{
	Provider:    "memory",
	DefaultTTL:  time.Hour,
	Namespace:   "default",
	EnableDebug: false,
	MemoryConfig: &MemoryConfig{
		MaxSize:         10000,
		EvictionPolicy:  "LRU",
		CleanupInterval: time.Minute * 10,
	},
}

// 初始化时尝试使用 gconf 加载默认值
func init() {
	// 尝试使用 gconf 处理默认值
	if gcfg, ok := interface{}(nil).(interface{ LoadDefaultsFromStruct(interface{}) error }); ok {
		_ = gcfg.LoadDefaultsFromStruct(defaultConfig)
	}
}

// RegisterProvider 注册缓存提供者
// 参数:
//   - name: 提供者名称
//   - factory: 创建提供者实例的工厂函数
func RegisterProvider(name string, factory func(*Config) (Provider, error)) {
	if _, exists := providers[name]; exists {
		glog.Warnf("缓存提供者 %s 已经注册，将被覆盖", name)
	}
	providers[name] = factory
}

// GetCache 获取全局缓存实例
// 如果尚未初始化，则使用默认配置初始化
func GetCache() Cache {
	instanceOnce.Do(func() {
		var err error
		instance, err = NewCache(defaultConfig)
		if err != nil {
			glog.Errorf("使用默认配置初始化缓存失败: %v", err)
			// 回退到内存缓存
			memConfig := *defaultConfig
			memConfig.Provider = "memory"
			instance, err = NewCache(&memConfig)
			if err != nil {
				panic(fmt.Sprintf("初始化内存缓存失败: %v", err))
			}
		}
	})
	return instance
}

// NewCache 基于配置创建新的缓存实例
// 如果提供者不存在，返回错误
func NewCache(config *Config) (Cache, error) {
	if config == nil {
		return nil, gerror.New(CodeCacheConfigError, "缓存配置不能为空")
	}

	if config.Provider == "" {
		glog.Warn("未指定缓存提供者，使用默认内存缓存")
		config.Provider = "memory"
	}

	// 获取提供者工厂
	factory, exists := providers[config.Provider]
	if !exists {
		return nil, gerror.Newf(CodeCacheConfigError, "未知的缓存提供者: %s", config.Provider)
	}

	// 使用工厂创建提供者实例
	provider, err := factory(config)
	if err != nil {
		return nil, gerror.Wrapf(err, CodeCacheInitError, "初始化缓存提供者 %s 失败", config.Provider)
	}

	// 创建基础缓存实例
	cache := &baseCache{
		provider:    provider,
		config:      config,
		namespace:   config.Namespace,
		defaultTTL:  config.DefaultTTL,
		enableDebug: config.EnableDebug,
	}

	if config.EnableDebug {
		glog.Debugf("缓存初始化成功，提供者: %s, 命名空间: %s", config.Provider, config.Namespace)
	}

	return cache, nil
}

// Configure 使用新的配置重新配置全局缓存实例
// 这允许在运行时动态更改缓存配置
func Configure(config *Config) error {
	if config == nil {
		return gerror.New(CodeCacheConfigError, "配置不能为nil")
	}

	// 如果未设置提供者，使用默认提供者
	if config.Provider == "" {
		config.Provider = defaultConfig.Provider
	}

	// 如果未设置默认TTL，使用默认的TTL
	if config.DefaultTTL <= 0 {
		config.DefaultTTL = defaultConfig.DefaultTTL
	}

	// 如果未设置命名空间，使用默认命名空间
	if config.Namespace == "" {
		config.Namespace = defaultConfig.Namespace
	}

	// 创建新的缓存实例
	newCache, err := NewCache(config)
	if err != nil {
		return gerror.Wrapf(err, CodeCacheConfigError, "使用新配置创建缓存失败")
	}

	// 替换全局实例
	instance = newCache
	return nil
}

// ConfigureFromFile 从配置文件加载缓存配置
// 支持 JSON, YAML 等格式，由文件扩展名决定
func ConfigureFromFile(filePath string) error {
	// 尝试导入 gconf
	gcfgLoader, ok := interface{}(nil).(interface {
		Load(string, interface{}, ...bool) error
	})
	if !ok {
		return gerror.New(CodeCacheConfigError, "gconf 包不可用，无法从文件加载配置")
	}

	var config Config
	if err := gcfgLoader.Load(filePath, &config); err != nil {
		return gerror.Wrapf(err, CodeCacheConfigError, "从文件加载缓存配置失败: %s", filePath)
	}

	return Configure(&config)
}

// baseCache 是对Provider接口的基本实现
type baseCache struct {
	provider    Provider      // 缓存提供者
	config      *Config       // 缓存配置
	namespace   string        // 缓存命名空间
	defaultTTL  time.Duration // 默认TTL
	enableDebug bool          // 是否启用调试
}

// 键名处理，添加命名空间前缀
func (c *baseCache) makeKey(key string) string {
	if key == "" {
		return ""
	}
	if c.namespace == "" {
		return key
	}
	return c.namespace + ":" + key
}

// WithNamespace 返回使用指定命名空间的新缓存实例
func (c *baseCache) WithNamespace(namespace string) Cache {
	if namespace == c.namespace {
		return c
	}

	// 创建新实例，共享底层Provider
	return &baseCache{
		provider:    c.provider,
		config:      c.config,
		namespace:   namespace,
		defaultTTL:  c.defaultTTL,
		enableDebug: c.enableDebug,
	}
}

// Name 返回缓存提供者名称
func (c *baseCache) Name() string {
	return c.config.Provider
}

// Stats 返回缓存统计信息
func (c *baseCache) Stats() map[string]interface{} {
	// 基本统计信息
	stats := map[string]interface{}{
		"provider":  c.config.Provider,
		"namespace": c.namespace,
	}

	// TODO: 从提供者获取更多统计信息

	return stats
}

// Get 从缓存获取值
func (c *baseCache) Get(ctx context.Context, key string) (interface{}, error) {
	if key == "" {
		return nil, ErrCacheKeyInvalid
	}

	namespacedKey := c.makeKey(key)
	if c.enableDebug {
		glog.Debugf("缓存Get操作：键=%s", namespacedKey)
	}

	value, err := c.provider.Get(ctx, namespacedKey)
	if err != nil {
		if gerror.GetCode(err) == CodeCacheNotFound {
			if c.enableDebug {
				glog.Debugf("缓存未命中：键=%s", namespacedKey)
			}
			return nil, ErrCacheNotFound
		}
		return nil, gerror.Wrapf(err, CodeCacheGetError, "获取缓存失败：键=%s", key)
	}

	if c.enableDebug {
		glog.Debugf("缓存命中：键=%s", namespacedKey)
	}

	return value, nil
}

// GetMulti 从缓存获取多个值
func (c *baseCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 添加命名空间前缀
	namespacedKeys := make([]string, len(keys))
	for i, key := range keys {
		if key == "" {
			return nil, gerror.Wrapf(ErrCacheKeyInvalid, CodeCacheKeyInvalid, "键列表中包含空键，索引=%d", i)
		}
		namespacedKeys[i] = c.makeKey(key)
	}

	if c.enableDebug {
		glog.Debugf("缓存GetMulti操作：键数量=%d", len(namespacedKeys))
	}

	// 调用提供者的方法
	result, err := c.provider.GetMulti(ctx, namespacedKeys)
	if err != nil {
		return nil, gerror.Wrapf(err, CodeCacheGetError, "批量获取缓存失败：键数量=%d", len(keys))
	}

	// 还原没有命名空间的键
	finalResult := make(map[string]interface{}, len(result))
	prefixLen := len(c.namespace)

	for nk, value := range result {
		// 删除命名空间前缀
		var originalKey string
		if c.namespace != "" {
			originalKey = nk[prefixLen+1:] // +1 for the ':'
		} else {
			originalKey = nk
		}
		finalResult[originalKey] = value
	}

	if c.enableDebug {
		glog.Debugf("缓存GetMulti结果：找到=%d, 请求=%d", len(finalResult), len(keys))
	}

	return finalResult, nil
}

// Set 设置缓存值
func (c *baseCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return ErrCacheKeyInvalid
	}

	// 使用默认TTL（如果未指定）
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	namespacedKey := c.makeKey(key)
	if c.enableDebug {
		glog.Debugf("缓存Set操作：键=%s, TTL=%v", namespacedKey, ttl)
	}

	err := c.provider.Set(ctx, namespacedKey, value, ttl)
	if err != nil {
		return gerror.Wrapf(err, CodeCacheSetError, "设置缓存失败：键=%s", key)
	}

	return nil
}

// Delete 删除缓存项
func (c *baseCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrCacheKeyInvalid
	}

	namespacedKey := c.makeKey(key)
	if c.enableDebug {
		glog.Debugf("缓存Delete操作：键=%s", namespacedKey)
	}

	err := c.provider.Delete(ctx, namespacedKey)
	if err != nil {
		return gerror.Wrapf(err, CodeCacheDelError, "删除缓存失败：键=%s", key)
	}

	return nil
}

// DeleteMulti 删除多个缓存项
func (c *baseCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// 添加命名空间前缀
	namespacedKeys := make([]string, len(keys))
	for i, key := range keys {
		if key == "" {
			return gerror.Wrapf(ErrCacheKeyInvalid, CodeCacheKeyInvalid, "键列表中包含空键，索引=%d", i)
		}
		namespacedKeys[i] = c.makeKey(key)
	}

	if c.enableDebug {
		glog.Debugf("缓存DeleteMulti操作：键数量=%d", len(namespacedKeys))
	}

	err := c.provider.DeleteMulti(ctx, namespacedKeys)
	if err != nil {
		return gerror.Wrapf(err, CodeCacheDelError, "批量删除缓存失败：键数量=%d", len(keys))
	}

	return nil
}

// Exists 检查键是否存在
func (c *baseCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrCacheKeyInvalid
	}

	namespacedKey := c.makeKey(key)
	if c.enableDebug {
		glog.Debugf("缓存Exists操作：键=%s", namespacedKey)
	}

	exists, err := c.provider.Exists(ctx, namespacedKey)
	if err != nil {
		return false, gerror.Wrapf(err, CodeCacheGetError, "检查缓存键存在性失败：键=%s", key)
	}

	return exists, nil
}

// Flush 清空所有缓存项
func (c *baseCache) Flush(ctx context.Context) error {
	if c.enableDebug {
		glog.Debugf("缓存Flush操作：命名空间=%s", c.namespace)
	}

	// TODO: 如果提供者支持，只清空当前命名空间

	err := c.provider.Flush(ctx)
	if err != nil {
		return gerror.Wrapf(err, CodeCacheFlushError, "清空缓存失败")
	}

	return nil
}

// GetTTL 获取缓存项的剩余生存时间
func (c *baseCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if key == "" {
		return 0, ErrCacheKeyInvalid
	}

	namespacedKey := c.makeKey(key)
	if c.enableDebug {
		glog.Debugf("缓存GetTTL操作：键=%s", namespacedKey)
	}

	ttl, err := c.provider.GetTTL(ctx, namespacedKey)
	if err != nil {
		return 0, gerror.Wrapf(err, CodeCacheGetError, "获取缓存TTL失败：键=%s", key)
	}

	return ttl, nil
}

// Close 关闭缓存
func (c *baseCache) Close() error {
	if c.enableDebug {
		glog.Debug("关闭缓存连接")
	}

	return c.provider.Close()
}
