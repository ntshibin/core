// Package gcache 提供一个灵活、高性能的缓存系统，支持多种存储后端。
// 它实现了策略模式，允许在内存、Redis和文件系统等不同的存储介质间无缝切换。
package gcache

import (
	"context"
	"time"

	"github.com/ntshibin/core/gerror"
)

// 预定义的缓存错误码
const (
	// 通用缓存错误（13000-13099）
	CodeCacheError      gerror.Code = 13000 // 一般缓存错误
	CodeCacheNotFound   gerror.Code = 13001 // 缓存项未找到
	CodeCacheFull       gerror.Code = 13002 // 缓存已满
	CodeCacheExpired    gerror.Code = 13003 // 缓存项已过期
	CodeCacheKeyInvalid gerror.Code = 13004 // 缓存键无效

	// 连接相关错误（13100-13199）
	CodeCacheConnError   gerror.Code = 13100 // 连接缓存后端失败
	CodeCacheConnTimeout gerror.Code = 13101 // 连接缓存后端超时
	CodeCacheConnClosed  gerror.Code = 13102 // 缓存连接已关闭
	CodeCacheUnavailable gerror.Code = 13103 // 缓存服务不可用
	CodeCacheAuthError   gerror.Code = 13104 // 缓存认证失败

	// 操作相关错误（13200-13299）
	CodeCacheSetError   gerror.Code = 13200 // 设置缓存值失败
	CodeCacheGetError   gerror.Code = 13201 // 获取缓存值失败
	CodeCacheDelError   gerror.Code = 13202 // 删除缓存值失败
	CodeCacheFlushError gerror.Code = 13203 // 清空缓存失败
	CodeCacheLockError  gerror.Code = 13204 // 缓存锁操作失败
	CodeCacheScanError  gerror.Code = 13205 // 扫描缓存键失败

	// 配置相关错误（13300-13399）
	CodeCacheConfigError gerror.Code = 13300 // 缓存配置错误
	CodeCacheInitError   gerror.Code = 13301 // 缓存初始化失败
)

// Provider 定义缓存提供者的通用接口
// 所有缓存实现都必须满足此接口
type Provider interface {
	// Get 从缓存获取值
	Get(ctx context.Context, key string) (interface{}, error)

	// GetMulti 从缓存获取多个值
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)

	// Set 设置缓存值
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete 删除缓存项
	Delete(ctx context.Context, key string) error

	// DeleteMulti 删除多个缓存项
	DeleteMulti(ctx context.Context, keys []string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Flush 清空所有缓存项
	Flush(ctx context.Context) error

	// GetTTL 获取缓存项的剩余生存时间
	GetTTL(ctx context.Context, key string) (time.Duration, error)

	// Close 关闭缓存提供者连接
	Close() error
}

// Cache 是缓存系统的主要接口
// 它包装了Provider接口，并提供了额外的功能
type Cache interface {
	Provider

	// WithNamespace 返回一个在指定命名空间下操作的Cache实例
	WithNamespace(namespace string) Cache

	// Name 返回缓存提供者的名称
	Name() string

	// Stats 返回缓存统计信息
	Stats() map[string]interface{}
}

// Item 表示缓存中的一个项
type Item struct {
	Key        string        // 缓存键
	Value      interface{}   // 缓存值
	TTL        time.Duration // 生存时间
	Expiration time.Time     // 过期时间点
}

// Config 缓存配置
type Config struct {
	// Provider 缓存提供者类型：memory, redis, file
	Provider string `json:"provider" yaml:"provider" env:"CACHE_PROVIDER" default:"memory"`

	// DefaultTTL 默认的缓存项生存时间
	DefaultTTL time.Duration `json:"default_ttl" yaml:"default_ttl" env:"CACHE_DEFAULT_TTL" default:"30m"`

	// Namespace 缓存命名空间，用于隔离不同的缓存键
	Namespace string `json:"namespace" yaml:"namespace" env:"CACHE_NAMESPACE" default:"app"`

	// EnableDebug 是否启用调试日志
	EnableDebug bool `json:"enable_debug" yaml:"enable_debug" env:"CACHE_ENABLE_DEBUG" default:"false"`

	// MemoryConfig 内存缓存的配置
	MemoryConfig *MemoryConfig `json:"memory_config" yaml:"memory_config"`

	// RedisConfig Redis缓存的配置
	RedisConfig *RedisConfig `json:"redis_config" yaml:"redis_config"`

	// FileConfig 文件缓存的配置
	FileConfig *FileConfig `json:"file_config" yaml:"file_config"`
}

// MemoryConfig 内存缓存配置
type MemoryConfig struct {
	// MaxSize 最大缓存项数量，0表示无限制
	MaxSize int `json:"max_size" yaml:"max_size" env:"CACHE_MEMORY_MAX_SIZE" default:"1000"`

	// EvictionPolicy 淘汰策略：LRU, LFU, FIFO
	EvictionPolicy string `json:"eviction_policy" yaml:"eviction_policy" env:"CACHE_MEMORY_EVICTION_POLICY" default:"LRU"`

	// CleanupInterval 过期项清理间隔
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval" env:"CACHE_MEMORY_CLEANUP_INTERVAL" default:"5m"`
}

// RedisConfig Redis缓存配置
type RedisConfig struct {
	// Addresses Redis服务器地址列表
	Addresses []string `json:"addresses" yaml:"addresses" env:"CACHE_REDIS_ADDRESSES"`

	// Password Redis认证密码
	Password string `json:"password" yaml:"password" env:"CACHE_REDIS_PASSWORD"`

	// Database Redis数据库索引
	Database int `json:"database" yaml:"database" env:"CACHE_REDIS_DATABASE" default:"0"`

	// PoolSize 连接池大小
	PoolSize int `json:"pool_size" yaml:"pool_size" env:"CACHE_REDIS_POOL_SIZE" default:"10"`

	// ConnTimeout 连接超时时间
	ConnTimeout time.Duration `json:"conn_timeout" yaml:"conn_timeout" env:"CACHE_REDIS_CONN_TIMEOUT" default:"5s"`

	// ReadTimeout 读取超时时间
	ReadTimeout time.Duration `json:"read_timeout" yaml:"read_timeout" env:"CACHE_REDIS_READ_TIMEOUT" default:"3s"`

	// WriteTimeout 写入超时时间
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout" env:"CACHE_REDIS_WRITE_TIMEOUT" default:"3s"`

	// MaxRetries 最大重试次数
	MaxRetries int `json:"max_retries" yaml:"max_retries" env:"CACHE_REDIS_MAX_RETRIES" default:"3"`

	// EnableTLS 是否启用TLS连接
	EnableTLS bool `json:"enable_tls" yaml:"enable_tls" env:"CACHE_REDIS_ENABLE_TLS" default:"false"`

	// EnableCluster 是否启用集群模式
	EnableCluster bool `json:"enable_cluster" yaml:"enable_cluster" env:"CACHE_REDIS_ENABLE_CLUSTER" default:"false"`
}

// FileConfig 文件缓存配置
type FileConfig struct {
	// DirPath 缓存文件存储目录
	DirPath string `json:"dir_path" yaml:"dir_path" env:"CACHE_FILE_DIR_PATH" default:"./cache"`

	// FileSuffix 缓存文件后缀
	FileSuffix string `json:"file_suffix" yaml:"file_suffix" env:"CACHE_FILE_SUFFIX" default:".cache"`

	// FileMode 缓存文件权限
	FileMode uint32 `json:"file_mode" yaml:"file_mode" env:"CACHE_FILE_MODE" default:"0644"`

	// GcInterval 垃圾回收间隔
	GcInterval time.Duration `json:"gc_interval" yaml:"gc_interval" env:"CACHE_FILE_GC_INTERVAL" default:"10m"`
}

// ErrCacheNotFound 缓存项未找到错误
var ErrCacheNotFound = gerror.New(CodeCacheNotFound, "缓存项未找到")

// ErrCacheKeyInvalid 缓存键无效错误
var ErrCacheKeyInvalid = gerror.New(CodeCacheKeyInvalid, "缓存键无效")
