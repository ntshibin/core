package conf

import (
	"os"
	"sync"
	"time"
)

// ConfigWatcher 配置监听器
type ConfigWatcher struct {
	file     string
	config   interface{}
	lastMod  time.Time
	callback func(interface{})
	stop     chan struct{}
	mu       sync.RWMutex
}

// NewConfigWatcher 创建配置监听器
func NewConfigWatcher(file string, config interface{}, callback func(interface{})) (*ConfigWatcher, error) {
	watcher := &ConfigWatcher{
		file:     file,
		config:   config,
		callback: callback,
		stop:     make(chan struct{}),
	}

	// 初始加载配置
	if err := LoadConfig(file, config); err != nil {
		return nil, err
	}

	// 获取文件最后修改时间
	info, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	watcher.lastMod = info.ModTime()

	// 启动监听
	go watcher.watch()

	return watcher, nil
}

// Stop 停止配置监听
func (w *ConfigWatcher) Stop() {
	close(w.stop)
}

// watch 监听配置文件变化
func (w *ConfigWatcher) watch() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			info, err := os.Stat(w.file)
			if err != nil {
				continue
			}

			w.mu.RLock()
			lastMod := w.lastMod
			w.mu.RUnlock()

			if info.ModTime().After(lastMod) {
				w.mu.Lock()
				w.lastMod = info.ModTime()
				if err := LoadConfig(w.file, w.config); err == nil && w.callback != nil {
					w.callback(w.config)
				}
				w.mu.Unlock()
			}
		}
	}
}
