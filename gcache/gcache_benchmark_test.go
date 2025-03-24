package gcache_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ntshibin/core/gcache"
	_ "github.com/ntshibin/core/gcache/providers" // 初始化所有缓存提供者
)

// BenchmarkMemoryCache 测试内存缓存性能
func BenchmarkMemoryCache(b *testing.B) {
	ctx := context.Background()

	config := &gcache.Config{
		Provider:    "memory",
		DefaultTTL:  time.Hour,
		Namespace:   "bench",
		EnableDebug: false, // 关闭调试以提高性能
		MemoryConfig: &gcache.MemoryConfig{
			MaxSize:         100000,
			EvictionPolicy:  "LRU",
			CleanupInterval: time.Minute,
		},
	}

	cache, err := gcache.NewCache(config)
	if err != nil {
		b.Fatalf("创建缓存实例失败: %v", err)
	}
	defer cache.Close()

	// 预热缓存
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key:%d", i)
		value := fmt.Sprintf("value:%d", i)
		if err := cache.Set(ctx, key, value, 0); err != nil {
			b.Fatalf("预热缓存失败: %v", err)
		}
	}

	// 测试Get性能
	b.Run("Get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key:%d", i%1000)
			_, _ = cache.Get(ctx, key)
		}
	})

	// 测试Set性能
	b.Run("Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key:%d", i)
			value := strconv.Itoa(i)
			_ = cache.Set(ctx, key, value, 0)
		}
	})

	// 测试Exists性能
	b.Run("Exists", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key:%d", i%1000)
			_, _ = cache.Exists(ctx, key)
		}
	})

	// 测试Delete性能
	b.Run("Delete", func(b *testing.B) {
		// 先设置一些值
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-key:%d", i)
			_ = cache.Set(ctx, key, "to-be-deleted", 0)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-key:%d", i)
			_ = cache.Delete(ctx, key)
		}
	})

	// 测试GetMulti性能
	b.Run("GetMulti", func(b *testing.B) {
		batchSize := 100

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			keys := make([]string, batchSize)
			for j := 0; j < batchSize; j++ {
				keys[j] = fmt.Sprintf("key:%d", (i*batchSize+j)%1000)
			}
			_, _ = cache.GetMulti(ctx, keys)
		}
	})
}

// BenchmarkFileCache 测试文件缓存性能
func BenchmarkFileCache(b *testing.B) {
	ctx := context.Background()

	config := &gcache.Config{
		Provider:    "file",
		DefaultTTL:  time.Hour,
		Namespace:   "bench",
		EnableDebug: false, // 关闭调试以提高性能
		FileConfig: &gcache.FileConfig{
			DirPath:    "/tmp/gcache_bench",
			FileSuffix: ".cache",
			FileMode:   0644,
			GcInterval: time.Minute,
		},
	}

	cache, err := gcache.NewCache(config)
	if err != nil {
		b.Fatalf("创建缓存实例失败: %v", err)
	}
	defer cache.Close()

	// 预热缓存（文件缓存只预热少量项以避免文件系统开销）
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key:%d", i)
		value := fmt.Sprintf("value:%d", i)
		if err := cache.Set(ctx, key, value, 0); err != nil {
			b.Fatalf("预热缓存失败: %v", err)
		}
	}

	// 测试Get性能
	b.Run("Get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key:%d", i%100)
			_, _ = cache.Get(ctx, key)
		}
	})

	// 测试Set性能
	b.Run("Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key-set:%d", i%100) // 限制数量以避免创建太多文件
			value := strconv.Itoa(i)
			_ = cache.Set(ctx, key, value, 0)
		}
	})

	// 测试Exists性能
	b.Run("Exists", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key:%d", i%100)
			_, _ = cache.Exists(ctx, key)
		}
	})

	// 测试Delete性能（文件缓存删除操作较慢，限制数量）
	b.Run("Delete", func(b *testing.B) {
		maxTests := 100
		if b.N > maxTests {
			b.N = maxTests
		}

		// 先设置一些值
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-key:%d", i)
			_ = cache.Set(ctx, key, "to-be-deleted", 0)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-key:%d", i)
			_ = cache.Delete(ctx, key)
		}
	})
}

// 基准测试缓存命名空间性能
func BenchmarkNamespace(b *testing.B) {
	ctx := context.Background()

	// 创建基础缓存
	baseCache, err := gcache.NewCache(&gcache.Config{
		Provider:    "memory",
		Namespace:   "base",
		EnableDebug: false,
	})
	if err != nil {
		b.Fatalf("创建基础缓存失败: %v", err)
	}
	defer baseCache.Close()

	// 预热
	for i := 0; i < 100; i++ {
		_ = baseCache.Set(ctx, fmt.Sprintf("key:%d", i), i, 0)
	}

	b.Run("WithNamespace", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ns := fmt.Sprintf("ns%d", i%10)
			_ = baseCache.WithNamespace(ns)
		}
	})

	b.Run("NamespaceIsolation", func(b *testing.B) {
		// 创建多个命名空间
		namespaces := make([]gcache.Cache, 10)
		for i := 0; i < 10; i++ {
			namespaces[i] = baseCache.WithNamespace(fmt.Sprintf("ns%d", i))
			// 每个命名空间设置相同的键但值不同
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("isokey:%d", j)
				_ = namespaces[i].Set(ctx, key, fmt.Sprintf("ns%d-value%d", i, j), 0)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			nsIndex := i % 10
			keyIndex := i % 10
			key := fmt.Sprintf("isokey:%d", keyIndex)
			_, _ = namespaces[nsIndex].Get(ctx, key)
		}
	})
}

// 基准测试默认全局缓存实例
func BenchmarkGlobalCache(b *testing.B) {
	ctx := context.Background()
	cache := gcache.GetCache()

	// 预热
	for i := 0; i < 100; i++ {
		_ = cache.Set(ctx, fmt.Sprintf("global:%d", i), i, 0)
	}

	b.Run("GlobalGet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("global:%d", i%100)
			_, _ = cache.Get(ctx, key)
		}
	})

	b.Run("GlobalSet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("global:%d", i%100)
			_ = cache.Set(ctx, key, i, 0)
		}
	})
}
