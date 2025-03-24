package gcache_test

import (
	"context"
	"testing"
	"time"

	"github.com/ntshibin/core/gcache"
	_ "github.com/ntshibin/core/gcache/providers" // 引入缓存提供者
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试Redis缓存提供者，但这需要真实的Redis服务器才能运行
// 这个测试默认是被跳过的，除非在开发环境中有Redis可用
func TestRedisCache(t *testing.T) {
	// 跳过这个测试，因为它需要真实的Redis连接
	t.Skip("这个测试需要真实的Redis服务器，默认跳过")

	// 如果想运行这个测试，请提供可用的Redis服务器地址
	redisAddr := "localhost:6379"

	// 创建Redis缓存配置
	config := &gcache.Config{
		Provider:    "redis",
		DefaultTTL:  time.Minute,
		Namespace:   "test",
		EnableDebug: true,
		RedisConfig: &gcache.RedisConfig{
			Addresses:    []string{redisAddr},
			Password:     "",
			Database:     0,
			PoolSize:     5,
			ConnTimeout:  time.Second * 2,
			ReadTimeout:  time.Second * 2,
			WriteTimeout: time.Second * 2,
			MaxRetries:   3,
		},
	}

	// 测试Redis连接
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis服务器不可用: %v", err)
		return
	}

	// 创建Redis缓存客户端
	cache, err := gcache.NewCache(config)
	require.NoError(t, err, "创建Redis缓存客户端失败")
	defer cache.Close()

	// 清空测试数据库
	defer func() {
		client.FlushDB(ctx)
	}()

	// 测试基本的设置和获取
	t.Run("基本设置和获取", func(t *testing.T) {
		// 设置值
		err := cache.Set(ctx, "testKey", "testValue", 0)
		require.NoError(t, err, "设置值失败")

		// 从缓存获取
		value, err := cache.Get(ctx, "testKey")
		require.NoError(t, err, "获取值失败")
		assert.Equal(t, "testValue", value)
	})

	// 测试过期
	t.Run("过期", func(t *testing.T) {
		// 设置带过期时间的值
		err := cache.Set(ctx, "expireKey", "willExpire", time.Second)
		require.NoError(t, err, "设置带过期时间的值失败")

		// 验证可以获取
		value, err := cache.Get(ctx, "expireKey")
		require.NoError(t, err, "获取值失败")
		assert.Equal(t, "willExpire", value)

		// 等待过期
		time.Sleep(time.Second * 2)

		// 验证无法获取
		_, err = cache.Get(ctx, "expireKey")
		assert.Equal(t, gcache.ErrCacheNotFound, err, "过期后应该返回未找到错误")
	})

	// 测试批量操作
	t.Run("批量操作", func(t *testing.T) {
		// 批量设置多个值
		for i := 1; i <= 3; i++ {
			key := "multi" + string('0'+byte(i))
			val := "value" + string('0'+byte(i))
			err := cache.Set(ctx, key, val, 0)
			require.NoError(t, err, "批量设置值失败")
		}

		// 批量获取
		values, err := cache.GetMulti(ctx, []string{"multi1", "multi2", "multi3"})
		require.NoError(t, err, "批量获取值失败")
		assert.Equal(t, 3, len(values), "应该获取到3个值")
		assert.Equal(t, "value2", values["multi2"])

		// 批量删除
		err = cache.DeleteMulti(ctx, []string{"multi1", "multi3"})
		require.NoError(t, err, "批量删除值失败")

		// 验证删除结果
		exists, err := cache.Exists(ctx, "multi1")
		require.NoError(t, err, "检查键存在性失败")
		assert.False(t, exists, "multi1应该已被删除")

		exists, err = cache.Exists(ctx, "multi2")
		require.NoError(t, err, "检查键存在性失败")
		assert.True(t, exists, "multi2不应被删除")
	})

	// 测试命名空间
	t.Run("命名空间", func(t *testing.T) {
		// 使用原始命名空间
		err := cache.Set(ctx, "nsKey", "defaultNS", 0)
		require.NoError(t, err, "设置默认命名空间的值失败")

		// 使用其他命名空间
		otherCache := cache.WithNamespace("other")
		err = otherCache.Set(ctx, "nsKey", "otherNS", 0)
		require.NoError(t, err, "设置其他命名空间的值失败")

		// 验证两个命名空间互不干扰
		val1, err := cache.Get(ctx, "nsKey")
		require.NoError(t, err, "获取默认命名空间的值失败")
		assert.Equal(t, "defaultNS", val1)

		val2, err := otherCache.Get(ctx, "nsKey")
		require.NoError(t, err, "获取其他命名空间的值失败")
		assert.Equal(t, "otherNS", val2)
	})

	// 测试TTL
	t.Run("TTL", func(t *testing.T) {
		// 设置带过期时间的值
		err := cache.Set(ctx, "ttlKey", "haveTTL", time.Second*10)
		require.NoError(t, err, "设置带过期时间的值失败")

		// 获取TTL
		ttl, err := cache.GetTTL(ctx, "ttlKey")
		require.NoError(t, err, "获取TTL失败")
		assert.True(t, ttl > 0 && ttl <= time.Second*10, "TTL应该在合理范围内")
	})
}
