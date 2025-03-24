# GCache - 高性能灵活的 Go 缓存库

GCache 是一个灵活、高性能的 Go 缓存系统，支持多种存储后端。它实现了策略模式，允许在内存、Redis 和文件系统等不同的存储介质间无缝切换，同时提供了一致的 API 接口。

## 特性

- **多种缓存提供者**：支持内存缓存、文件缓存和 Redis 缓存
- **一致的 API**：无论使用哪种缓存提供者，API 保持一致
- **命名空间隔离**：支持通过命名空间隔离不同的缓存数据
- **丰富的配置选项**：可针对不同提供者配置相应的特性
- **过期机制**：支持设置缓存项 TTL (Time-To-Live)
- **垃圾回收**：自动清理过期的缓存项
- **批量操作**：支持批量获取和删除缓存项
- **统计信息**：提供缓存使用统计

## 架构设计

GCache 采用接口分离和策略模式设计，主要组件包括：

1. **Provider 接口**：定义缓存提供者必须实现的基本方法
2. **Cache 接口**：扩展 Provider 接口，提供更多高级功能
3. **baseCache**：Cache 接口的基本实现，包含通用逻辑
4. **MemoryCache**：基于内存的缓存实现，支持多种淘汰策略
5. **FileCache**：基于文件系统的持久化缓存
6. **RedisCache**：基于 Redis 的分布式缓存

## 安装

```bash
go get github.com/ntshibin/core/gcache
```

## 基本使用

### 创建缓存实例

```go
import (
    "context"
    "time"

    "github.com/ntshibin/core/gcache"
    _ "github.com/ntshibin/core/gcache/providers" // 初始化所有缓存提供者
)

// 创建内存缓存
memConfig := &gcache.Config{
    Provider:   "memory",
    DefaultTTL: time.Hour,
    Namespace:  "users",
    MemoryConfig: &gcache.MemoryConfig{
        MaxSize:         10000,
        EvictionPolicy:  "LRU",
        CleanupInterval: time.Minute * 10,
    },
}
memCache, err := gcache.NewCache(memConfig)
if err != nil {
    // 处理错误
}
defer memCache.Close()

// 创建文件缓存
fileConfig := &gcache.Config{
    Provider:   "file",
    DefaultTTL: time.Hour,
    Namespace:  "products",
    FileConfig: &gcache.FileConfig{
        DirPath:    "/tmp/gcache",
        FileSuffix: ".cache",
        FileMode:   0644,
        GcInterval: time.Minute * 15,
    },
}
fileCache, err := gcache.NewCache(fileConfig)
if err != nil {
    // 处理错误
}
defer fileCache.Close()

// 创建Redis缓存
redisConfig := &gcache.Config{
    Provider:   "redis",
    DefaultTTL: time.Hour,
    Namespace:  "sessions",
    RedisConfig: &gcache.RedisConfig{
        Addresses:    []string{"localhost:6379"},
        Password:     "",
        Database:     0,
        PoolSize:     10,
        ConnTimeout:  time.Second * 5,
        ReadTimeout:  time.Second * 3,
        WriteTimeout: time.Second * 3,
    },
}
redisCache, err := gcache.NewCache(redisConfig)
if err != nil {
    // 处理错误
}
defer redisCache.Close()
```

### 使用全局缓存

```go
// 获取默认的全局缓存实例
cache := gcache.GetCache()

// 使用自定义配置初始化全局缓存
config := &gcache.Config{
    Provider:   "memory",
    DefaultTTL: time.Minute * 30,
    Namespace:  "global",
}
err := gcache.Configure(config)
if err != nil {
    // 处理错误
}
```

### 基本操作

```go
ctx := context.Background()

// 设置缓存项
err := cache.Set(ctx, "user:1001", userObj, time.Hour)
if err != nil {
    // 处理错误
}

// 获取缓存项
value, err := cache.Get(ctx, "user:1001")
if err != nil {
    if err == gcache.ErrCacheNotFound {
        // 缓存未命中
    } else {
        // 处理其他错误
    }
}

// 检查键是否存在
exists, err := cache.Exists(ctx, "user:1001")
if err != nil {
    // 处理错误
}
if exists {
    // 键存在
}

// 删除缓存项
err = cache.Delete(ctx, "user:1001")
if err != nil {
    // 处理错误
}

// 获取TTL
ttl, err := cache.GetTTL(ctx, "user:1001")
if err != nil {
    // 处理错误
}
```

### 批量操作

```go
// 批量获取
keys := []string{"user:1001", "user:1002", "user:1003"}
values, err := cache.GetMulti(ctx, keys)
if err != nil {
    // 处理错误
}

// 批量删除
err = cache.DeleteMulti(ctx, keys)
if err != nil {
    // 处理错误
}
```

### 使用命名空间

```go
// 创建特定命名空间的缓存
userCache := cache.WithNamespace("users")
productCache := cache.WithNamespace("products")

// 在不同命名空间下操作相同的键
err := userCache.Set(ctx, "1001", userObj, 0)
err = productCache.Set(ctx, "1001", productObj, 0)

// 获取不同命名空间下的值
userValue, _ := userCache.Get(ctx, "1001")   // 获取用户对象
prodValue, _ := productCache.Get(ctx, "1001") // 获取产品对象
```

### 清空缓存

```go
// 清空所有缓存项
err := cache.Flush(ctx)
if err != nil {
    // 处理错误
}
```

## 自定义缓存提供者

您可以通过实现 `gcache.Provider` 接口并注册提供者来创建自定义缓存：

```go
func init() {
    gcache.RegisterProvider("custom", func(config *gcache.Config) (gcache.Provider, error) {
        // 创建并返回自定义缓存提供者的实例
        return NewCustomCache(config)
    })
}
```

## 性能测试

GCache 提供了高性能的缓存实现，以下是不同缓存提供者的基准测试结果：

### 内存缓存

- **Get 操作**: ~462 ns/op，每秒约 2.4 百万次请求
- **GetMulti 操作**: ~57 μs/op，每秒约 20k 次批量请求

### 全局缓存

- **Get 操作**: ~1.06 μs/op，每秒约 1 百万次请求
- **Set 操作**: ~1.29 μs/op，每秒约 944k 次请求

### 命名空间操作

- **创建命名空间**: ~156 ns/op
- **命名空间隔离查询**: ~380 ns/op

文件缓存和 Redis 缓存的性能将取决于底层存储介质的性能特性。

## 错误处理

GCache 使用 `gerror` 包提供了详细的错误信息，主要错误类型包括：

- `ErrCacheNotFound`: 缓存项未找到
- `ErrCacheKeyInvalid`: 缓存键无效
- 其他错误: 连接错误、操作错误等

## 线程安全

GCache 所有的缓存提供者都是线程安全的，可以在并发环境中安全使用。

## 许可证

本项目采用 MIT 许可证，详情请参阅 LICENSE 文件。
