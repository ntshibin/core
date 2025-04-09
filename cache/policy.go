package cache

// Policy 缓存策略接口
type Policy interface {
	// Update 更新缓存项
	Update(key string, item *memoryItem)
	// Evict 驱逐一个缓存项
	Evict(data map[string]*memoryItem) string
}

// LRUPolicy LRU策略实现
type LRUPolicy struct {
	keys []string
}

// NewLRUPolicy 创建LRU策略
func NewLRUPolicy() *LRUPolicy {
	return &LRUPolicy{
		keys: make([]string, 0),
	}
}

// Update 更新缓存项
func (p *LRUPolicy) Update(key string, item *memoryItem) {
	// 移除旧的位置
	for i, k := range p.keys {
		if k == key {
			p.keys = append(p.keys[:i], p.keys[i+1:]...)
			break
		}
	}
	// 添加到末尾
	p.keys = append(p.keys, key)
}

// Evict 驱逐一个缓存项
func (p *LRUPolicy) Evict(data map[string]*memoryItem) string {
	if len(p.keys) == 0 {
		return ""
	}
	// 移除最旧的项
	key := p.keys[0]
	p.keys = p.keys[1:]
	return key
}

// FIFOPolicy FIFO策略实现
type FIFOPolicy struct {
	keys []string
}

// NewFIFOPolicy 创建FIFO策略
func NewFIFOPolicy() *FIFOPolicy {
	return &FIFOPolicy{
		keys: make([]string, 0),
	}
}

// Update 更新缓存项
func (p *FIFOPolicy) Update(key string, item *memoryItem) {
	// 检查是否已存在
	for _, k := range p.keys {
		if k == key {
			return
		}
	}
	// 添加到末尾
	p.keys = append(p.keys, key)
}

// Evict 驱逐一个缓存项
func (p *FIFOPolicy) Evict(data map[string]*memoryItem) string {
	if len(p.keys) == 0 {
		return ""
	}
	// 移除最旧的项
	key := p.keys[0]
	p.keys = p.keys[1:]
	return key
}
