package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// Stats 缓存统计信息
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

// StatsCollector 统计信息收集器
type StatsCollector struct {
	stats Stats
	mutex sync.RWMutex
}

// NewStatsCollector 创建统计信息收集器
func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		stats: Stats{
			LastUpdate: time.Now(),
		},
	}
}

// GetStats 获取统计信息
func (s *StatsCollector) GetStats() Stats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.stats
}

// IncrKeyCount 增加键数量
func (s *StatsCollector) IncrKeyCount() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.KeyCount++
	s.stats.LastUpdate = time.Now()
}

// DecrKeyCount 减少键数量
func (s *StatsCollector) DecrKeyCount() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.stats.KeyCount > 0 {
		s.stats.KeyCount--
	}
	s.stats.LastUpdate = time.Now()
}

// IncrKeyCountBy 增加指定数量的键
func (s *StatsCollector) IncrKeyCountBy(count int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.KeyCount += count
	s.stats.LastUpdate = time.Now()
}

// DecrKeyCountBy 减少指定数量的键
func (s *StatsCollector) DecrKeyCountBy(count int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.stats.KeyCount > count {
		s.stats.KeyCount -= count
	} else {
		s.stats.KeyCount = 0
	}
	s.stats.LastUpdate = time.Now()
}

// IncrHits 增加命中次数
func (s *StatsCollector) IncrHits() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.Hits++
	s.stats.LastUpdate = time.Now()
}

// IncrMisses 增加未命中次数
func (s *StatsCollector) IncrMisses() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.Misses++
	s.stats.LastUpdate = time.Now()
}

// IncrEvictedCount 增加驱逐次数
func (s *StatsCollector) IncrEvictedCount() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.EvictedCount++
	s.stats.LastUpdate = time.Now()
}

// IncrExpiredCount 增加过期次数
func (s *StatsCollector) IncrExpiredCount() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.ExpiredCount++
	s.stats.LastUpdate = time.Now()
}

// Reset 重置统计信息
func (s *StatsCollector) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats = Stats{
		LastUpdate: time.Now(),
	}
}

// HitRatio 获取命中率
func (s *StatsCollector) HitRatio() float64 {
	hits := atomic.LoadInt64(&s.stats.Hits)
	total := hits + atomic.LoadInt64(&s.stats.Misses)
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total)
}
