package gcache_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ntshibin/core/gcache"
	_ "github.com/ntshibin/core/gcache/providers" // 初始化所有缓存提供者
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryCache 测试内存缓存提供者
func TestMemoryCache(t *testing.T) {
	config := &gcache.Config{
		Provider:    "memory",
		DefaultTTL:  time.Minute,
		Namespace:   "test",
		EnableDebug: true,
		MemoryConfig: &gcache.MemoryConfig{
			MaxSize:         100,
			EvictionPolicy:  "LRU",
			CleanupInterval: time.Second * 5,
		},
	}

	testCacheProvider(t, config)
}

// TestFileCache 测试文件缓存提供者
func TestFileCache(t *testing.T) {
	// 创建临时目录
	tmpDir := "/tmp/gcache_test_" + time.Now().Format("20060102150405")
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err, "创建临时目录失败")
	defer os.RemoveAll(tmpDir) // 测试结束后清理

	config := &gcache.Config{
		Provider:    "file",
		DefaultTTL:  time.Minute,
		Namespace:   "test",
		EnableDebug: true,
		FileConfig: &gcache.FileConfig{
			DirPath:    tmpDir, // 使用绝对路径
			FileSuffix: ".cache",
			FileMode:   0644,
			GcInterval: time.Second * 5,
		},
	}

	testCacheProvider(t, config)
}

// testCacheProvider 对指定的缓存提供者执行通用测试
func testCacheProvider(t *testing.T, config *gcache.Config) {
	ctx := context.Background()

	// 创建缓存实例
	cache, err := gcache.NewCache(config)
	require.NoError(t, err, "创建缓存实例失败")
	require.NotNil(t, cache, "缓存实例不应为空")

	// 测试完成后清理
	defer func() {
		err := cache.Flush(ctx)
		require.NoError(t, err, "清空缓存失败")

		err = cache.Close()
		require.NoError(t, err, "关闭缓存失败")
	}()

	// 测试基本的设置和获取
	t.Run("基本设置和获取", func(t *testing.T) {
		err := cache.Set(ctx, "testKey", "testValue", 0) // 使用默认TTL
		require.NoError(t, err, "设置缓存值失败")

		value, err := cache.Get(ctx, "testKey")
		require.NoError(t, err, "获取缓存值失败")
		assert.Equal(t, "testValue", value)
	})

	// 测试不存在的键
	t.Run("获取不存在键", func(t *testing.T) {
		_, err := cache.Get(ctx, "不存在的键")
		assert.Error(t, err)
		assert.Equal(t, gcache.ErrCacheNotFound, err)
	})

	// 测试存在性检查
	t.Run("检查键存在性", func(t *testing.T) {
		err := cache.Set(ctx, "existKey", 123, 0)
		require.NoError(t, err, "设置缓存值失败")

		exists, err := cache.Exists(ctx, "existKey")
		require.NoError(t, err, "检查键存在性失败")
		assert.True(t, exists, "键应该存在")

		exists, err = cache.Exists(ctx, "不存在的键")
		require.NoError(t, err, "检查键存在性失败")
		assert.False(t, exists, "键不应该存在")
	})

	// 测试删除
	t.Run("删除缓存项", func(t *testing.T) {
		err := cache.Set(ctx, "delKey", "要删除的值", 0)
		require.NoError(t, err, "设置缓存值失败")

		err = cache.Delete(ctx, "delKey")
		require.NoError(t, err, "删除缓存项失败")

		exists, err := cache.Exists(ctx, "delKey")
		require.NoError(t, err, "检查键存在性失败")
		assert.False(t, exists, "键应该已被删除")
	})

	// 测试过期
	t.Run("过期检查", func(t *testing.T) {
		// 使用1秒的过期时间
		expiryKey := "expireKey_" + time.Now().Format("150405")
		err := cache.Set(ctx, expiryKey, "过期值", time.Second)
		require.NoError(t, err, "设置缓存值失败")

		// 确认当前键存在
		value, err := cache.Get(ctx, expiryKey)
		require.NoError(t, err, "获取缓存值失败")
		assert.Equal(t, "过期值", value, "应该能获取到未过期的值")

		// 等待过期 - 等待时间长一些，确保文件缓存也能检测到过期
		time.Sleep(time.Second * 2)

		// 检查键是否已过期
		_, err = cache.Get(ctx, expiryKey)
		if config.Provider == "file" {
			t.Logf("文件缓存提示：检查过期键 %s", expiryKey)
		}
		assert.Error(t, err, "过期后应该返回错误")
		assert.Equal(t, gcache.ErrCacheNotFound, err, "过期后应该返回未找到错误")
	})

	// 测试TTL
	t.Run("获取TTL", func(t *testing.T) {
		err := cache.Set(ctx, "ttlKey", "TTL测试", time.Second*5)
		require.NoError(t, err, "设置缓存值失败")

		ttl, err := cache.GetTTL(ctx, "ttlKey")
		require.NoError(t, err, "获取TTL失败")
		assert.True(t, ttl > 0 && ttl <= time.Second*5, "TTL应该在合理范围内")
	})

	// 测试批量操作
	t.Run("批量操作", func(t *testing.T) {
		// 批量设置（通过循环）
		for i := 1; i <= 5; i++ {
			key := fmt.Sprintf("batch%d", i)
			val := fmt.Sprintf("value%d", i)
			err := cache.Set(ctx, key, val, 0)
			require.NoError(t, err, "批量设置失败")
		}

		// 批量获取
		keys := []string{"batch1", "batch2", "batch3", "batch4", "batch5"}
		values, err := cache.GetMulti(ctx, keys)
		require.NoError(t, err, "批量获取失败")
		assert.Equal(t, 5, len(values), "应该获取5个值")
		assert.Equal(t, "value3", values["batch3"])

		// 批量删除
		err = cache.DeleteMulti(ctx, []string{"batch1", "batch3", "batch5"})
		require.NoError(t, err, "批量删除失败")

		// 验证删除结果
		exists, err := cache.Exists(ctx, "batch1")
		require.NoError(t, err)
		assert.False(t, exists, "batch1应该已被删除")

		exists, err = cache.Exists(ctx, "batch2")
		require.NoError(t, err)
		assert.True(t, exists, "batch2不应被删除")
	})

	// 测试命名空间
	t.Run("命名空间", func(t *testing.T) {
		// 使用原始缓存
		err := cache.Set(ctx, "nsKey", "默认命名空间", 0)
		require.NoError(t, err, "设置缓存值失败")

		// 使用其他命名空间
		otherCache := cache.WithNamespace("other")
		err = otherCache.Set(ctx, "nsKey", "其他命名空间", 0)
		require.NoError(t, err, "设置缓存值失败")

		// 验证两个命名空间互不干扰
		val1, err := cache.Get(ctx, "nsKey")
		require.NoError(t, err)
		assert.Equal(t, "默认命名空间", val1)

		val2, err := otherCache.Get(ctx, "nsKey")
		require.NoError(t, err)
		assert.Equal(t, "其他命名空间", val2)
	})

	// 测试缓存统计信息
	t.Run("缓存统计", func(t *testing.T) {
		stats := cache.Stats()
		assert.NotNil(t, stats, "缓存统计不应为空")
		assert.Equal(t, config.Provider, stats["provider"], "提供者名称应匹配")
	})
}

// TestGlobalCache 测试全局缓存实例
func TestGlobalCache(t *testing.T) {
	ctx := context.Background()

	// 使用全局缓存
	cache := gcache.GetCache()
	require.NotNil(t, cache, "全局缓存实例不应为空")

	// 简单的设置和获取测试
	err := cache.Set(ctx, "globalKey", "globalValue", 0)
	require.NoError(t, err, "设置全局缓存值失败")

	value, err := cache.Get(ctx, "globalKey")
	require.NoError(t, err, "获取全局缓存值失败")
	assert.Equal(t, "globalValue", value)

	// 清理测试数据
	err = cache.Delete(ctx, "globalKey")
	require.NoError(t, err, "删除全局缓存值失败")
}

// TestConfigure 测试重新配置全局缓存
func TestConfigure(t *testing.T) {
	ctx := context.Background()

	// 使用新配置初始化全局缓存
	config := &gcache.Config{
		Provider:    "memory",
		DefaultTTL:  time.Second * 30,
		Namespace:   "reconfigured",
		EnableDebug: true,
	}

	err := gcache.Configure(config)
	require.NoError(t, err, "重新配置全局缓存失败")

	// 获取全局缓存
	cache := gcache.GetCache()

	// 测试新配置
	err = cache.Set(ctx, "configKey", "configTest", 0)
	require.NoError(t, err, "设置缓存值失败")

	value, err := cache.Get(ctx, "configKey")
	require.NoError(t, err, "获取缓存值失败")
	assert.Equal(t, "configTest", value)

	// 验证命名空间
	stats := cache.Stats()
	assert.Equal(t, "reconfigured", stats["namespace"])

	// 清理
	err = cache.Delete(ctx, "configKey")
	require.NoError(t, err, "删除缓存值失败")
}

// TestInvalidCacheKey 测试无效缓存键
func TestInvalidCacheKey(t *testing.T) {
	ctx := context.Background()

	// 创建缓存实例
	cache, err := gcache.NewCache(&gcache.Config{
		Provider: "memory",
	})
	require.NoError(t, err, "创建缓存实例失败")

	// 测试空键
	_, err = cache.Get(ctx, "")
	assert.Equal(t, gcache.ErrCacheKeyInvalid, err)

	err = cache.Set(ctx, "", "value", 0)
	assert.Equal(t, gcache.ErrCacheKeyInvalid, err)

	err = cache.Delete(ctx, "")
	assert.Equal(t, gcache.ErrCacheKeyInvalid, err)

	_, err = cache.Exists(ctx, "")
	assert.Equal(t, gcache.ErrCacheKeyInvalid, err)

	_, err = cache.GetTTL(ctx, "")
	assert.Equal(t, gcache.ErrCacheKeyInvalid, err)
}

// TestRegisterProvider 测试注册自定义缓存提供者
func TestRegisterProvider(t *testing.T) {
	// 注册一个简单的内存缓存提供者
	gcache.RegisterProvider("custom", func(config *gcache.Config) (gcache.Provider, error) {
		return gcache.NewCache(&gcache.Config{Provider: "memory"})
	})

	// 创建缓存实例
	cache, err := gcache.NewCache(&gcache.Config{
		Provider: "custom",
	})

	require.NoError(t, err, "创建自定义缓存实例失败")
	require.NotNil(t, cache, "缓存实例不应为空")

	// 验证能正常运行
	ctx := context.Background()
	err = cache.Set(ctx, "customKey", "customValue", 0)
	require.NoError(t, err)

	value, err := cache.Get(ctx, "customKey")
	require.NoError(t, err)
	assert.Equal(t, "customValue", value)

	// 清理
	err = cache.Delete(ctx, "customKey")
	require.NoError(t, err)

	err = cache.Close()
	require.NoError(t, err)
}
