package providers

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ntshibin/core/gcache"
	"github.com/ntshibin/core/gerror"
	"github.com/ntshibin/core/glog"
)

// 定义文件缓存相关常量
const (
	// 默认缓存目录
	DefaultFileCacheDir = "/tmp/gcache"

	// 默认文件后缀
	DefaultFileSuffix = ".cache"

	// 默认文件权限
	DefaultFileMode = 0644

	// 默认垃圾回收间隔
	DefaultGCInterval = 15 * time.Minute

	// 文件内容格式版本
	FileFormatVersion = 1
)

// fileItem 表示缓存在文件中的项
type fileItem struct {
	Version    int         `json:"v"`     // 文件格式版本
	Key        string      `json:"k"`     // 缓存键
	Value      interface{} `json:"val"`   // 缓存值
	Expiration int64       `json:"exp"`   // 过期时间戳
	TTL        int64       `json:"ttl"`   // 生存时间（秒）
	CreateTime int64       `json:"ctime"` // 创建时间戳
}

// FileCache 实现基于文件的缓存提供者
type FileCache struct {
	dirPath  string        // 缓存文件存储目录
	suffix   string        // 缓存文件后缀
	fileMode os.FileMode   // 缓存文件权限，修改为os.FileMode
	mutex    sync.RWMutex  // 读写锁
	gcTicker *time.Ticker  // GC定时器
	stopChan chan struct{} // 停止信号
}

// NewFileCache 创建一个新的文件缓存
func NewFileCache(config *gcache.Config) (gcache.Provider, error) {
	// 使用默认配置（如果未提供）
	dirPath := DefaultFileCacheDir
	suffix := DefaultFileSuffix
	fileMode := os.FileMode(DefaultFileMode) // 修改为os.FileMode
	gcInterval := DefaultGCInterval

	if config != nil && config.FileConfig != nil {
		if config.FileConfig.DirPath != "" {
			dirPath = config.FileConfig.DirPath
		}

		if config.FileConfig.FileSuffix != "" {
			suffix = config.FileConfig.FileSuffix
		}

		if config.FileConfig.FileMode > 0 {
			fileMode = os.FileMode(config.FileConfig.FileMode) // 已经是正确的转换
		}

		if config.FileConfig.GcInterval > 0 {
			gcInterval = config.FileConfig.GcInterval
		}
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, gerror.Wrapf(err, gcache.CodeCacheInitError, "创建缓存目录失败: %s", dirPath)
	}

	fc := &FileCache{
		dirPath:  dirPath,
		suffix:   suffix,
		fileMode: fileMode,
		stopChan: make(chan struct{}),
	}

	// 启动GC定时器
	fc.gcTicker = time.NewTicker(gcInterval)
	go func() {
		for {
			select {
			case <-fc.gcTicker.C:
				if err := fc.gcExpired(); err != nil {
					glog.Errorf("文件缓存GC失败: %v", err)
				}
			case <-fc.stopChan:
				fc.gcTicker.Stop()
				return
			}
		}
	}()

	glog.Debugf("文件缓存初始化成功，目录: %s", dirPath)

	return fc, nil
}

// 将键转换为文件路径
func (c *FileCache) keyToFilePath(key string) string {
	// 使用MD5哈希或其他机制来生成文件名
	// 这里使用简单的URL编码作为示例
	encoded := strings.ReplaceAll(key, "/", "_")
	encoded = strings.ReplaceAll(encoded, ":", "_")
	encoded = strings.ReplaceAll(encoded, ".", "_")

	return filepath.Join(c.dirPath, encoded+c.suffix)
}

// gcExpired 清理过期的缓存文件
func (c *FileCache) gcExpired() error {
	glog.Debug("文件缓存开始GC")

	// 遍历缓存目录
	count := 0
	total := 0

	err := filepath.Walk(c.dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录和非缓存文件
		if info.IsDir() || !strings.HasSuffix(info.Name(), c.suffix) {
			return nil
		}

		total++

		// 读取文件内容
		data, err := os.ReadFile(path)
		if err != nil {
			glog.Warnf("读取缓存文件失败: %s, %v", path, err)
			// 删除损坏的文件
			os.Remove(path)
			count++
			return nil
		}

		// 解析缓存项
		var item fileItem
		if err := json.Unmarshal(data, &item); err != nil {
			glog.Warnf("解析缓存文件失败: %s, %v", path, err)
			// 删除损坏的文件
			os.Remove(path)
			count++
			return nil
		}

		// 检查是否过期
		if item.Expiration > 0 && item.Expiration < time.Now().Unix() {
			if err := os.Remove(path); err != nil {
				glog.Warnf("删除过期缓存文件失败: %s, %v", path, err)
				return nil
			}
			count++
		}

		return nil
	})

	if err != nil {
		return gerror.Wrapf(err, gcache.CodeCacheError, "文件缓存GC失败")
	}

	glog.Debugf("文件缓存GC完成: 删除 %d/%d 个文件", count, total)

	return nil
}

// Get 从缓存获取值
func (c *FileCache) Get(ctx context.Context, key string) (interface{}, error) {
	if key == "" {
		return nil, gcache.ErrCacheKeyInvalid
	}

	// 获取文件路径
	filePath := c.keyToFilePath(key)

	// 检查文件是否存在
	c.mutex.RLock()
	data, err := os.ReadFile(filePath)
	c.mutex.RUnlock()

	if err != nil {
		if os.IsNotExist(err) {
			return nil, gcache.ErrCacheNotFound
		}
		return nil, gerror.Wrapf(err, gcache.CodeCacheGetError, "读取缓存文件失败: %s", key)
	}

	// 解析缓存项
	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, gerror.Wrapf(err, gcache.CodeCacheGetError, "解析缓存文件失败: %s", key)
	}

	// 检查是否过期
	if item.Expiration > 0 && item.Expiration < time.Now().Unix() {
		// 在读锁下检测到过期，需要切换到写锁来删除文件
		c.mutex.Lock()
		// 重新检查一次过期状态（有可能在获取写锁期间被其他协程修改）
		if item.Expiration > 0 && item.Expiration < time.Now().Unix() {
			os.Remove(filePath) // 删除过期文件
			c.mutex.Unlock()
			glog.Debugf("文件缓存：删除过期键: %s, 过期时间: %d", key, item.Expiration)
			return nil, gcache.ErrCacheNotFound
		}
		c.mutex.Unlock()
	}

	return item.Value, nil
}

// GetMulti 从缓存获取多个值
func (c *FileCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{}, len(keys))

	for _, key := range keys {
		if key == "" {
			continue
		}

		value, err := c.Get(ctx, key)
		if err == nil {
			result[key] = value
		}
	}

	return result, nil
}

// Set 设置缓存值
func (c *FileCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return gcache.ErrCacheKeyInvalid
	}

	// 计算过期时间
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).Unix()
	}

	// 创建缓存项
	item := fileItem{
		Version:    FileFormatVersion,
		Key:        key,
		Value:      value,
		Expiration: expiration,
		TTL:        int64(ttl.Seconds()),
		CreateTime: time.Now().Unix(),
	}

	// 序列化为JSON
	data, err := json.Marshal(item)
	if err != nil {
		return gerror.Wrapf(err, gcache.CodeCacheSetError, "序列化缓存项失败: %s", key)
	}

	// 获取文件路径
	filePath := c.keyToFilePath(key)

	// 写入文件
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := os.WriteFile(filePath, data, c.fileMode); err != nil {
		return gerror.Wrapf(err, gcache.CodeCacheSetError, "写入缓存文件失败: %s", key)
	}

	return nil
}

// Delete 删除缓存项
func (c *FileCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return gcache.ErrCacheKeyInvalid
	}

	// 获取文件路径
	filePath := c.keyToFilePath(key)

	// 删除文件
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return gerror.Wrapf(err, gcache.CodeCacheDelError, "删除缓存文件失败: %s", key)
	}

	return nil
}

// DeleteMulti 删除多个缓存项
func (c *FileCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	var lastErr error

	for _, key := range keys {
		if key == "" {
			continue
		}

		if err := c.Delete(ctx, key); err != nil {
			lastErr = err
			glog.Warnf("删除缓存文件失败: %s, %v", key, err)
		}
	}

	return lastErr
}

// Exists 检查键是否存在
func (c *FileCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, gcache.ErrCacheKeyInvalid
	}

	// 获取文件路径
	filePath := c.keyToFilePath(key)

	// 检查文件是否存在
	c.mutex.RLock()
	data, err := os.ReadFile(filePath)
	c.mutex.RUnlock()

	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, gerror.Wrapf(err, gcache.CodeCacheGetError, "读取缓存文件失败: %s", key)
	}

	// 解析缓存项
	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return false, gerror.Wrapf(err, gcache.CodeCacheGetError, "解析缓存文件失败: %s", key)
	}

	// 检查是否过期
	if item.Expiration > 0 && item.Expiration < time.Now().Unix() {
		// 在读锁下检测到过期，需要切换到写锁来删除文件
		c.mutex.Lock()
		// 重新检查一次过期状态（有可能在获取写锁期间被其他协程修改）
		if item.Expiration > 0 && item.Expiration < time.Now().Unix() {
			os.Remove(filePath) // 删除过期文件
			c.mutex.Unlock()
			glog.Debugf("文件缓存：检查键存在性时删除过期键: %s, 过期时间: %d", key, item.Expiration)
			return false, nil
		}
		c.mutex.Unlock()
	}

	return true, nil
}

// Flush 清空所有缓存项
func (c *FileCache) Flush(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 遍历缓存目录删除所有文件
	count := 0

	err := filepath.Walk(c.dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录和非缓存文件
		if info.IsDir() || !strings.HasSuffix(info.Name(), c.suffix) {
			return nil
		}

		// 删除文件
		if err := os.Remove(path); err != nil {
			glog.Warnf("删除缓存文件失败: %s, %v", path, err)
			return nil
		}

		count++
		return nil
	})

	if err != nil {
		return gerror.Wrapf(err, gcache.CodeCacheFlushError, "清空文件缓存失败")
	}

	glog.Debugf("文件缓存已清空: 删除 %d 个文件", count)

	return nil
}

// GetTTL 获取缓存项的剩余生存时间
func (c *FileCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if key == "" {
		return 0, gcache.ErrCacheKeyInvalid
	}

	// 获取文件路径
	filePath := c.keyToFilePath(key)

	// 检查文件是否存在
	c.mutex.RLock()
	data, err := os.ReadFile(filePath)
	c.mutex.RUnlock()

	if err != nil {
		if os.IsNotExist(err) {
			return 0, gcache.ErrCacheNotFound
		}
		return 0, gerror.Wrapf(err, gcache.CodeCacheGetError, "读取缓存文件失败: %s", key)
	}

	// 解析缓存项
	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return 0, gerror.Wrapf(err, gcache.CodeCacheGetError, "解析缓存文件失败: %s", key)
	}

	// 检查是否过期
	if item.Expiration > 0 {
		now := time.Now().Unix()
		if item.Expiration < now {
			// 在读锁下检测到过期，需要切换到写锁来删除文件
			c.mutex.Lock()
			// 重新检查一次过期状态（有可能在获取写锁期间被其他协程修改）
			if item.Expiration < time.Now().Unix() {
				os.Remove(filePath)
				c.mutex.Unlock()
				glog.Debugf("文件缓存：获取TTL时删除过期键: %s, 过期时间: %d", key, item.Expiration)
				return 0, gcache.ErrCacheNotFound
			}
			c.mutex.Unlock()
		}

		// 计算剩余时间
		remaining := time.Duration(item.Expiration-now) * time.Second
		return remaining, nil
	}

	// 没有设置过期时间
	return -1, nil
}

// Close 关闭缓存
func (c *FileCache) Close() error {
	close(c.stopChan)
	glog.Debug("文件缓存已关闭")
	return nil
}

// 注册文件缓存提供者
func init() {
	gcache.RegisterProvider("file", func(config *gcache.Config) (gcache.Provider, error) {
		return NewFileCache(config)
	})
}
