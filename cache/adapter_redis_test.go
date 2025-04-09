package cache

import (
	"context"
	"testing"
	"time"
)

func checkRedisConnection() bool {
	config := &BaseConfig{
		MaxSize:         100,
		CleanupInterval: 60,
	}
	cacheConfig := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	cache := NewRedisCache(config, cacheConfig)
	ctx := context.Background()
	_, err := cache.HealthCheck(ctx)
	return err == nil
}

func TestRedisCache(t *testing.T) {
	if !checkRedisConnection() {
		t.Skip("Redis server is not available")
	}
	config := &BaseConfig{
		MaxSize:         100,
		CleanupInterval: 60,
	}
	cacheConfig := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	cache := NewRedisCache(config, cacheConfig)

	// 测试 Set 和 Get
	ctx := context.Background()
	key := "test_key"
	value := "test_value"
	if err := cache.Set(ctx, key, value, time.Minute); err != nil {
		t.Errorf("Set failed: %v", err)
	}

	var result string
	if err := cache.Get(ctx, key, &result); err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if result != value {
		t.Errorf("Expected %v, got %v", value, result)
	}

	// 测试 Has
	exists, err := cache.Has(ctx, key)
	if err != nil {
		t.Errorf("Has failed: %v", err)
	}
	if !exists {
		t.Error("Expected key to exist")
	}

	// 测试 Delete
	if err := cache.Delete(ctx, key); err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// 测试过期
	if err := cache.Set(ctx, key, value, time.Millisecond); err != nil {
		t.Errorf("Set failed: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := cache.Get(ctx, key, &result); err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}

	// 测试 Clear
	if err := cache.Set(ctx, key, value, time.Minute); err != nil {
		t.Errorf("Set failed: %v", err)
	}
	if err := cache.Clear(ctx); err != nil {
		t.Errorf("Clear failed: %v", err)
	}
	if err := cache.Get(ctx, key, &result); err != ErrNotFound {
		t.Errorf("Expected ErrNotFound after Clear, got %v", err)
	}
}

func TestRedisCacheWithTags(t *testing.T) {
	if !checkRedisConnection() {
		t.Skip("Redis server is not available")
	}
	config := &BaseConfig{
		MaxSize:         100,
		CleanupInterval: 60,
	}
	cacheConfig := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	cache := NewRedisCache(config, cacheConfig)

	ctx := context.Background()
	key := "test_key"
	value := "test_value"
	tags := []string{"tag1", "tag2"}

	// 测试 SetWithTags
	if err := cache.SetWithTags(ctx, key, value, tags, time.Minute); err != nil {
		t.Errorf("SetWithTags failed: %v", err)
	}

	// 测试 GetByTag
	keys, err := cache.GetByTag(ctx, "tag1")
	if err != nil {
		t.Errorf("GetByTag failed: %v", err)
	}
	if len(keys) != 1 || keys[0] != key {
		t.Errorf("Expected [%v], got %v", key, keys)
	}

	// 测试 DeleteByTag
	if err := cache.DeleteByTag(ctx, "tag1"); err != nil {
		t.Errorf("DeleteByTag failed: %v", err)
	}
	if err := cache.Get(ctx, key, &value); err != ErrNotFound {
		t.Errorf("Expected ErrNotFound after DeleteByTag, got %v", err)
	}
}

func TestRedisCacheStats(t *testing.T) {
	if !checkRedisConnection() {
		t.Skip("Redis server is not available")
	}
	config := &BaseConfig{
		MaxSize:         100,
		CleanupInterval: 60,
	}
	cacheConfig := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	cache := NewRedisCache(config, cacheConfig)

	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	// 测试统计信息
	if err := cache.Set(ctx, key, value, time.Minute); err != nil {
		t.Errorf("Set failed: %v", err)
	}

	stats, err := cache.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats failed: %v", err)
	}
	if stats.KeyCount != 1 {
		t.Errorf("Expected KeyCount 1, got %v", stats.KeyCount)
	}

	// 测试健康检查
	health, err := cache.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}
	if health.Status != "healthy" {
		t.Errorf("Expected status healthy, got %v", health.Status)
	}
}

func TestRedisCacheLock(t *testing.T) {
	if !checkRedisConnection() {
		t.Skip("Redis server is not available")
	}
	config := &BaseConfig{
		MaxSize:         100,
		CleanupInterval: 60,
	}
	cacheConfig := &RedisCacheConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	cache := NewRedisCache(config, cacheConfig)

	ctx := context.Background()
	key := "test_lock"
	lock := &RedisLock{
		client:     cache.client,
		key:        key,
		expiration: time.Second,
	}

	// 测试获取锁
	if err := lock.Lock(ctx); err != nil {
		t.Errorf("Lock failed: %v", err)
	}

	// 测试重复获取锁
	if err := lock.Lock(ctx); err == nil {
		t.Error("Expected error when locking twice")
	}

	// 测试刷新锁
	if err := lock.Refresh(ctx); err != nil {
		t.Errorf("Refresh failed: %v", err)
	}

	// 测试释放锁
	if err := lock.Unlock(ctx); err != nil {
		t.Errorf("Unlock failed: %v", err)
	}
}
