package cache

import "time"

// BaseConfig 基础配置
type BaseConfig struct {
	// DefaultExpiration 默认过期时间
	DefaultExpiration time.Duration `yaml:"default_expiration"`
	// CleanupInterval 清理间隔时间
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
	// MaxSize 最大缓存条目数
	MaxSize int `yaml:"max_size"`
}

// Config 缓存配置
type Config struct {
	// Type 缓存类型：memory, redis, file
	Type string `yaml:"type"`
	// BaseConfig 基础配置
	BaseConfig BaseConfig `yaml:",inline"`
	// RedisConfig Redis配置
	RedisConfig RedisCacheConfig `yaml:"redis_config"`
	// FileConfig 文件缓存配置
	FileConfig FileCacheConfig `yaml:"file_config"`
	// MemoryConfig
	MemoryConfig MemoryCacheConfig `yaml:"memory_config"`
}
