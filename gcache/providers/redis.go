package providers

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/ntshibin/core/gcache"
	"github.com/ntshibin/core/gerror"
	"github.com/ntshibin/core/glog"
	"github.com/redis/go-redis/v9"
)

// 定义Redis缓存相关常量
const (
	// 默认连接池大小
	DefaultRedisPoolSize = 10

	// 默认连接超时
	DefaultRedisConnTimeout = 5 * time.Second

	// 默认读取超时
	DefaultRedisReadTimeout = 3 * time.Second

	// 默认写入超时
	DefaultRedisWriteTimeout = 3 * time.Second

	// 默认最大重试次数
	DefaultRedisMaxRetries = 3
)

// RedisCache 实现基于Redis的缓存提供者
type RedisCache struct {
	client redis.UniversalClient // Redis客户端
	config *gcache.RedisConfig   // Redis配置
}

// NewRedisCache 创建一个新的Redis缓存
func NewRedisCache(config *gcache.Config) (gcache.Provider, error) {
	if config == nil || config.RedisConfig == nil {
		return nil, gerror.New(gcache.CodeCacheConfigError, "Redis缓存配置不能为空")
	}

	redisConfig := config.RedisConfig

	// 检查必要配置
	if len(redisConfig.Addresses) == 0 {
		return nil, gerror.New(gcache.CodeCacheConfigError, "Redis服务器地址不能为空")
	}

	// 使用默认值（如果未指定）
	if redisConfig.PoolSize <= 0 {
		redisConfig.PoolSize = DefaultRedisPoolSize
	}

	if redisConfig.ConnTimeout <= 0 {
		redisConfig.ConnTimeout = DefaultRedisConnTimeout
	}

	if redisConfig.ReadTimeout <= 0 {
		redisConfig.ReadTimeout = DefaultRedisReadTimeout
	}

	if redisConfig.WriteTimeout <= 0 {
		redisConfig.WriteTimeout = DefaultRedisWriteTimeout
	}

	if redisConfig.MaxRetries < 0 {
		redisConfig.MaxRetries = DefaultRedisMaxRetries
	}

	// 创建通用配置
	universalOptions := &redis.UniversalOptions{
		Addrs:        redisConfig.Addresses,
		Password:     redisConfig.Password,
		DB:           redisConfig.Database,
		PoolSize:     redisConfig.PoolSize,
		DialTimeout:  redisConfig.ConnTimeout,
		ReadTimeout:  redisConfig.ReadTimeout,
		WriteTimeout: redisConfig.WriteTimeout,
		MaxRetries:   redisConfig.MaxRetries,
	}

	// 处理集群/单机模式
	var client redis.UniversalClient
	if redisConfig.EnableCluster {
		glog.Info("使用Redis集群模式")
		client = redis.NewUniversalClient(universalOptions)
	} else if len(redisConfig.Addresses) > 1 {
		// 多个地址，但非集群 = Sentinel模式
		glog.Info("使用Redis Sentinel模式")
		universalOptions.MasterName = "mymaster" // 默认主名称
		client = redis.NewUniversalClient(universalOptions)
	} else {
		// 单节点模式
		glog.Info("使用Redis单节点模式")
		client = redis.NewClient(&redis.Options{
			Addr:         redisConfig.Addresses[0],
			Password:     redisConfig.Password,
			DB:           redisConfig.Database,
			PoolSize:     redisConfig.PoolSize,
			DialTimeout:  redisConfig.ConnTimeout,
			ReadTimeout:  redisConfig.ReadTimeout,
			WriteTimeout: redisConfig.WriteTimeout,
			MaxRetries:   redisConfig.MaxRetries,
		})
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.ConnTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, gerror.Wrapf(err, gcache.CodeCacheConnError, "连接Redis服务器失败: %s", strings.Join(redisConfig.Addresses, ","))
	}

	glog.Debug("Redis缓存连接成功")

	return &RedisCache{
		client: client,
		config: redisConfig,
	}, nil
}

// Get 从缓存获取值
func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	if key == "" {
		return nil, gcache.ErrCacheKeyInvalid
	}

	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, gcache.ErrCacheNotFound
		}
		return nil, gerror.Wrapf(err, gcache.CodeCacheGetError, "从Redis获取键失败: %s", key)
	}

	// 解析存储的值
	var value interface{}
	if err := json.Unmarshal([]byte(result), &value); err != nil {
		return nil, gerror.Wrapf(err, gcache.CodeCacheGetError, "解析Redis值失败: %s", key)
	}

	return value, nil
}

// GetMulti 从缓存获取多个值
func (c *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 使用管道批量获取
	pipe := c.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd, len(keys))

	for _, key := range keys {
		if key == "" {
			continue
		}
		cmds[key] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, gerror.Wrapf(err, gcache.CodeCacheGetError, "从Redis批量获取键失败")
	}

	// 处理结果
	result := make(map[string]interface{}, len(keys))
	for key, cmd := range cmds {
		val, err := cmd.Result()
		if err == nil {
			var value interface{}
			if err := json.Unmarshal([]byte(val), &value); err == nil {
				result[key] = value
			} else {
				glog.Warnf("解析Redis值失败: %s, 错误: %v", key, err)
			}
		}
	}

	return result, nil
}

// Set 设置缓存值
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return gcache.ErrCacheKeyInvalid
	}

	// 使用JSON编码值
	jsonData, err := json.Marshal(value)
	if err != nil {
		return gerror.Wrapf(err, gcache.CodeCacheSetError, "编码缓存值失败: %v", value)
	}

	if err := c.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return gerror.Wrapf(err, gcache.CodeCacheSetError, "设置Redis键失败: %s", key)
	}

	return nil
}

// Delete 删除缓存项
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return gcache.ErrCacheKeyInvalid
	}

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return gerror.Wrapf(err, gcache.CodeCacheDelError, "删除Redis键失败: %s", key)
	}

	return nil
}

// DeleteMulti 删除多个缓存项
func (c *RedisCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// 过滤空键
	validKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		if key != "" {
			validKeys = append(validKeys, key)
		}
	}

	if len(validKeys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, validKeys...).Err(); err != nil {
		return gerror.Wrapf(err, gcache.CodeCacheDelError, "批量删除Redis键失败")
	}

	return nil
}

// Exists 检查键是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, gcache.ErrCacheKeyInvalid
	}

	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, gerror.Wrapf(err, gcache.CodeCacheGetError, "检查Redis键存在性失败: %s", key)
	}

	return result > 0, nil
}

// Flush 清空所有缓存项
func (c *RedisCache) Flush(ctx context.Context) error {
	if err := c.client.FlushDB(ctx).Err(); err != nil {
		return gerror.Wrapf(err, gcache.CodeCacheFlushError, "清空Redis数据库失败")
	}

	glog.Warn("Redis缓存已清空 (FlushDB)")

	return nil
}

// GetTTL 获取缓存项的剩余生存时间
func (c *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if key == "" {
		return 0, gcache.ErrCacheKeyInvalid
	}

	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, gerror.Wrapf(err, gcache.CodeCacheGetError, "获取Redis键TTL失败: %s", key)
	}

	// Redis返回-2表示键不存在，-1表示键没有设置过期时间
	if ttl == -2 {
		return 0, gcache.ErrCacheNotFound
	} else if ttl == -1 {
		return -1, nil // -1表示永不过期
	}

	return ttl, nil
}

// Close 关闭缓存
func (c *RedisCache) Close() error {
	glog.Debug("关闭Redis缓存连接")
	return c.client.Close()
}

// 注册Redis缓存提供者
func init() {
	gcache.RegisterProvider("redis", func(config *gcache.Config) (gcache.Provider, error) {
		return NewRedisCache(config)
	})
}
