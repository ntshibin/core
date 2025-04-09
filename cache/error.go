package cache

import "errors"

var (
	// ErrNotImplemented 未实现错误
	ErrNotImplemented = errors.New("not implemented")
	// ErrInvalidCacheType 无效的缓存类型
	ErrInvalidCacheType = errors.New("invalid cache type")
	// ErrNotFound 缓存未找到
	ErrNotFound = errors.New("cache not found")
	// ErrInvalidValue 无效的值
	ErrInvalidValue = errors.New("invalid value")
)
