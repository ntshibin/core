package cache

// EventType 事件类型
type EventType int

const (
	// EventTypeSet 设置缓存事件
	EventTypeSet EventType = iota
	// EventTypeGet 获取缓存事件
	EventTypeGet
	// EventTypeDelete 删除缓存事件
	EventTypeDelete
	// EventTypeClear 清空缓存事件
	EventTypeClear
)

// EventListener 事件监听器接口
type EventListener interface {
	// OnEvent 处理缓存事件
	OnEvent(eventType EventType, key string)
}
