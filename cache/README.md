# 缓存模块

这是一个通用的缓存模块，支持多种缓存实现，包括内存缓存、文件缓存和 Redis 缓存。该模块提供了统一的接口，可以方便地切换不同的缓存实现。

## 特性

- 支持多种缓存实现：
  - 内存缓存（MemoryCache）
  - 文件缓存（FileCache）
  - Redis 缓存（RedisCache）
- 统一的缓存接口
- 支持缓存项过期
- 支持标签管理
- 支持事件监听
- 支持分布式锁
- 提供统计信息和健康检查

## 快速开始

### 安装

```bash
go get github.com/ntshibin/project/core/cache
```

### 使用示例

```go
import "github.com/ntshibin/project/core/cache"

// 创建内存缓存
config := &cache.BaseConfig{
    MaxSize:         100,
    CleanupInterval: 60,
}
cacheConfig := &cache.MemoryCacheConfig{
    Policy: "lru",
}
memoryCache := cache.NewMemoryCache(config, cacheConfig)

// 创建文件缓存
fileConfig := &cache.FileCacheConfig{
    Directory: "/tmp/cache",
}
fileCache := cache.NewFileCache(config, fileConfig)

// 创建 Redis 缓存
redisConfig := &cache.RedisCacheConfig{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
}
redisCache := cache.NewRedisCache(config, redisConfig)

// 使用缓存
ctx := context.Background()
key := "test_key"
value := "test_value"

// 设置缓存
if err := memoryCache.Set(ctx, key, value, time.Minute); err != nil {
    log.Fatal(err)
}

// 获取缓存
var result string
if err := memoryCache.Get(ctx, key, &result); err != nil {
    log.Fatal(err)
}

// 使用标签
tags := []string{"tag1", "tag2"}
if err := memoryCache.SetWithTags(ctx, key, value, tags, time.Minute); err != nil {
    log.Fatal(err)
}

// 获取标签相关的键
keys, err := memoryCache.GetByTag(ctx, "tag1")
if err != nil {
    log.Fatal(err)
}

// 使用分布式锁
lock := &cache.MemoryLock{
    cache:      memoryCache,
    key:        "test_lock",
    expiration: time.Second,
}

if err := lock.Lock(ctx); err != nil {
    log.Fatal(err)
}
defer lock.Unlock(ctx)

// 获取统计信息
stats, err := memoryCache.GetStats(ctx)
if err != nil {
    log.Fatal(err)
}

// 健康检查
health, err := memoryCache.HealthCheck(ctx)
if err != nil {
    log.Fatal(err)
}
```

## 配置

### 基础配置

```go
type BaseConfig struct {
    // 最大缓存项数量
    MaxSize int
    // 清理间隔（秒）
    CleanupInterval int
}
```

### 内存缓存配置

```go
type MemoryCacheConfig struct {
    // 缓存策略：lru, fifo
    Policy string
}
```

### 文件缓存配置

```go
type FileCacheConfig struct {
    // 缓存目录
    Directory string
}
```

### Redis 缓存配置

```go
type RedisCacheConfig struct {
    // Redis 连接地址
    Addr string
    // Redis 密码
    Password string
    // Redis 数据库
    DB int
}
```

## 接口说明

### 缓存接口

```go
type ICache interface {
    // 设置缓存
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    // 获取缓存
    Get(ctx context.Context, key string, value interface{}) error
    // 删除缓存
    Delete(ctx context.Context, key string) error
    // 检查缓存是否存在
    Has(ctx context.Context, key string) (bool, error)
    // 清空所有缓存
    Clear(ctx context.Context) error
    // 获取统计信息
    GetStats(ctx context.Context) (*Stats, error)
    // 执行健康检查
    HealthCheck(ctx context.Context) (*Health, error)
    // 批量设置缓存
    MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
    // 批量获取缓存
    MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
    // 批量删除缓存
    MDelete(ctx context.Context, keys []string) error
    // 设置带标签的缓存
    SetWithTags(ctx context.Context, key string, value interface{}, tags []string, ttl time.Duration) error
    // 获取指定标签的所有缓存键
    GetByTag(ctx context.Context, tag string) ([]string, error)
    // 删除指定标签的所有缓存
    DeleteByTag(ctx context.Context, tag string) error
    // 添加事件监听器
    AddEventListener(listener EventListener)
    // 移除事件监听器
    RemoveEventListener(listener EventListener)
    // 重置统计信息
    ResetStats(ctx context.Context) error
}
```

### 分布式锁接口

```go
type ILock interface {
    // 获取锁
    Lock(ctx context.Context) error
    // 释放锁
    Unlock(ctx context.Context) error
    // 刷新锁的过期时间
    Refresh(ctx context.Context) error
}
```

## 事件类型

```go
const (
    // 设置缓存事件
    EventTypeSet EventType = iota
    // 获取缓存事件
    EventTypeGet
    // 删除缓存事件
    EventTypeDelete
    // 清空缓存事件
    EventTypeClear
)
```

## 统计信息

```go
type Stats struct {
    // 缓存键数量
    KeyCount int64
    // 命中次数
    Hits int64
    // 未命中次数
    Misses int64
    // 驱逐次数
    EvictedCount int64
    // 过期次数
    ExpiredCount int64
    // 最后更新时间
    LastUpdate time.Time
}
```

## 健康检查

```go
type Health struct {
    // 健康状态
    Status string
    // 详细信息
    Details map[string]interface{}
    // 时间戳
    Timestamp time.Time
}
```

## 测试

运行测试：

```bash
go test -v ./...
```

## 许可证

MIT
