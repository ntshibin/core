// Package ratelimit 提供了HTTP请求限流功能
package ratelimit

import (
	"sync"
	"time"
)

// Limiter 限流接口
type Limiter interface {
	// Allow 判断是否允许请求通过
	Allow(key string) bool
	// Reset 重置限流记录
	Reset(key string)
}

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	rate       float64      // 令牌产生速率 (每秒)
	capacity   float64      // 桶容量
	tokens     sync.Map     // 当前令牌数量 key -> tokens
	lastAccess sync.Map     // 上次访问时间 key -> time.Time
	mutex      sync.RWMutex // 读写锁
}

// NewTokenBucket 创建令牌桶限流器
// rate: 令牌产生速率 (每秒)
// capacity: 桶容量
func NewTokenBucket(rate, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:     rate,
		capacity: capacity,
	}
}

// Allow 判断是否允许请求通过
func (tb *TokenBucket) Allow(key string) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()

	// 获取上次访问时间
	var lastTime time.Time
	if val, ok := tb.lastAccess.Load(key); ok {
		lastTime = val.(time.Time)
	}

	// 获取当前令牌数量
	var currentTokens float64
	if val, ok := tb.tokens.Load(key); ok {
		currentTokens = val.(float64)
	} else {
		// 首次访问，桶是满的
		currentTokens = tb.capacity
	}

	// 根据时间差计算新增令牌
	elapsed := now.Sub(lastTime).Seconds()
	newTokens := elapsed * tb.rate

	// 更新令牌数量，不超过桶容量
	if currentTokens+newTokens > tb.capacity {
		currentTokens = tb.capacity
	} else {
		currentTokens += newTokens
	}

	// 判断是否有足够的令牌
	if currentTokens < 1.0 {
		return false
	}

	// 消耗一个令牌
	currentTokens--

	// 更新状态
	tb.tokens.Store(key, currentTokens)
	tb.lastAccess.Store(key, now)

	return true
}

// Reset 重置指定key的限流记录
func (tb *TokenBucket) Reset(key string) {
	tb.tokens.Delete(key)
	tb.lastAccess.Delete(key)
}

// FixedWindow 固定时间窗口限流器
type FixedWindow struct {
	limit     int           // 时间窗口内的请求上限
	window    time.Duration // 时间窗口大小
	counters  sync.Map      // 当前计数 key -> count
	startTime sync.Map      // 窗口开始时间 key -> time.Time
	mutex     sync.RWMutex  // 读写锁
}

// NewFixedWindow 创建固定时间窗口限流器
// limit: 时间窗口内的请求上限
// window: 时间窗口大小
func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:  limit,
		window: window,
	}
}

// Allow 判断是否允许请求通过
func (fw *FixedWindow) Allow(key string) bool {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	now := time.Now()

	// 获取窗口开始时间
	var start time.Time
	if val, ok := fw.startTime.Load(key); ok {
		start = val.(time.Time)
	} else {
		// 首次访问，设置窗口开始时间
		start = now
		fw.startTime.Store(key, start)
	}

	// 检查是否需要重置窗口
	if now.Sub(start) > fw.window {
		// 重置窗口
		fw.startTime.Store(key, now)
		fw.counters.Store(key, 1)
		return true
	}

	// 获取当前计数
	var count int
	if val, ok := fw.counters.Load(key); ok {
		count = val.(int)
	}

	// 判断是否超过限制
	if count >= fw.limit {
		return false
	}

	// 增加计数
	fw.counters.Store(key, count+1)

	return true
}

// Reset 重置指定key的限流记录
func (fw *FixedWindow) Reset(key string) {
	fw.counters.Delete(key)
	fw.startTime.Delete(key)
}

// SlidingWindow 滑动时间窗口限流器
type SlidingWindow struct {
	limit      int           // 时间窗口内的请求上限
	window     time.Duration // 时间窗口大小
	timestamps sync.Map      // 请求时间戳列表 key -> []time.Time
	mutex      sync.RWMutex  // 读写锁
}

// NewSlidingWindow 创建滑动时间窗口限流器
// limit: 时间窗口内的请求上限
// window: 时间窗口大小
func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		limit:  limit,
		window: window,
	}
}

// Allow 判断是否允许请求通过
func (sw *SlidingWindow) Allow(key string) bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.window)

	// 获取请求时间戳列表
	var timestamps []time.Time
	if val, ok := sw.timestamps.Load(key); ok {
		timestamps = val.([]time.Time)
	}

	// 移除窗口外的时间戳
	validIdx := 0
	for ; validIdx < len(timestamps); validIdx++ {
		if timestamps[validIdx].After(windowStart) {
			break
		}
	}

	if validIdx > 0 {
		timestamps = timestamps[validIdx:]
	}

	// 判断是否超过限制
	if len(timestamps) >= sw.limit {
		// 更新时间戳列表
		sw.timestamps.Store(key, timestamps)
		return false
	}

	// 添加新的时间戳
	timestamps = append(timestamps, now)
	sw.timestamps.Store(key, timestamps)

	return true
}

// Reset 重置指定key的限流记录
func (sw *SlidingWindow) Reset(key string) {
	sw.timestamps.Delete(key)
}
