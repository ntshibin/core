package cache

import (
	"sync"
)

var (
	instance ICache
	once     sync.Once
)

// GetInstance 获取缓存实例
func GetInstance() ICache {
	once.Do(func() {
		// 默认使用内存缓存
		config := BaseConfig{
			MaxSize:         10000,
			CleanupInterval: 10 * 60, // 10分钟
		}
		memoryConfig := MemoryCacheConfig{}
		instance = NewMemoryCache(&config, &memoryConfig)
	})
	return instance
}

// LoadConfig 加载配置并切换缓存实例
func LoadConfig(config *Config) error {
	var err error
	once = sync.Once{}
	once.Do(func() {
		switch config.Type {
		case "memory":
			instance = NewMemoryCache(&config.BaseConfig, &config.MemoryConfig)
		case "redis":
			instance = NewRedisCache(&config.BaseConfig, &config.RedisConfig)
		case "file":
			instance = NewFileCache(&config.BaseConfig, &config.FileConfig)
		default:
			err = ErrInvalidCacheType
		}
	})
	return err
}
