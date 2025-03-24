// Package health 提供了HTTP服务的健康检查功能
package health

import (
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/ntshibin/core/ghttp/internal/context"
)

// Status 健康状态
type Status string

const (
	// StatusUp 服务正常
	StatusUp Status = "UP"
	// StatusDown 服务不可用
	StatusDown Status = "DOWN"
	// StatusDegraded 服务降级
	StatusDegraded Status = "DEGRADED"
)

// CheckFunc 健康检查函数类型
type CheckFunc func() (Status, map[string]interface{})

// HealthCheck 健康检查组件
type HealthCheck struct {
	status       Status               // 整体健康状态
	name         string               // 服务名称
	version      string               // 服务版本
	startTime    time.Time            // 启动时间
	mutex        sync.RWMutex         // 读写锁
	dependencies map[string]CheckFunc // 依赖项检查
}

// NewHealthCheck 创建健康检查组件
func NewHealthCheck(name, version string) *HealthCheck {
	return &HealthCheck{
		status:       StatusUp,
		name:         name,
		version:      version,
		startTime:    time.Now(),
		dependencies: make(map[string]CheckFunc),
	}
}

// AddCheck 添加依赖项检查
func (hc *HealthCheck) AddCheck(name string, check CheckFunc) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.dependencies[name] = check
}

// RemoveCheck 移除依赖项检查
func (hc *HealthCheck) RemoveCheck(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	delete(hc.dependencies, name)
}

// SetStatus 设置服务健康状态
func (hc *HealthCheck) SetStatus(status Status) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.status = status
}

// Handler 返回健康检查HTTP处理函数
func (hc *HealthCheck) Handler() func(*context.Context) {
	return func(c *context.Context) {
		hc.mutex.RLock()
		defer hc.mutex.RUnlock()

		// 计算运行时间
		uptime := time.Since(hc.startTime)

		// 系统信息
		systemInfo := map[string]interface{}{
			"go_version":    runtime.Version(),
			"go_os":         runtime.GOOS,
			"go_arch":       runtime.GOARCH,
			"cpu_num":       runtime.NumCPU(),
			"goroutine_num": runtime.NumGoroutine(),
		}

		// 检查依赖项
		deps := make(map[string]map[string]interface{})
		overallStatus := hc.status

		for name, check := range hc.dependencies {
			status, details := check()
			deps[name] = map[string]interface{}{
				"status":  status,
				"details": details,
			}

			// 如果任何依赖项是DOWN状态，则整体状态是DOWN
			if status == StatusDown && overallStatus != StatusDown {
				overallStatus = StatusDown
			}

			// 如果任何依赖项是DEGRADED状态，且整体状态不是DOWN，则整体状态是DEGRADED
			if status == StatusDegraded && overallStatus != StatusDown {
				overallStatus = StatusDegraded
			}
		}

		// 健康检查响应
		response := map[string]interface{}{
			"status":       overallStatus,
			"name":         hc.name,
			"version":      hc.version,
			"timestamp":    time.Now().Format(time.RFC3339),
			"uptime":       uptime.String(),
			"uptime_sec":   uptime.Seconds(),
			"system":       systemInfo,
			"dependencies": deps,
		}

		// 设置HTTP状态码
		httpStatus := http.StatusOK
		if overallStatus == StatusDown {
			httpStatus = http.StatusServiceUnavailable
		} else if overallStatus == StatusDegraded {
			httpStatus = http.StatusOK // 降级状态也返回200，但包含降级信息
		}

		c.JSON(httpStatus, response)
	}
}

// LivenessHandler 返回存活检查HTTP处理函数 (轻量级)
func (hc *HealthCheck) LivenessHandler() func(*context.Context) {
	return func(c *context.Context) {
		hc.mutex.RLock()
		status := hc.status
		hc.mutex.RUnlock()

		if status == StatusDown {
			c.String(http.StatusServiceUnavailable, string(status))
			return
		}

		c.String(http.StatusOK, string(status))
	}
}

// ReadinessHandler 返回就绪检查HTTP处理函数 (包含依赖项检查)
func (hc *HealthCheck) ReadinessHandler() func(*context.Context) {
	return func(c *context.Context) {
		hc.mutex.RLock()
		defer hc.mutex.RUnlock()

		status := hc.status
		deps := make(map[string]Status)

		// 检查所有依赖项
		for name, check := range hc.dependencies {
			depStatus, _ := check()
			deps[name] = depStatus

			// 如果任何依赖项是DOWN状态，则整体状态是DOWN
			if depStatus == StatusDown && status != StatusDown {
				status = StatusDown
			}

			// 如果任何依赖项是DEGRADED状态，且整体状态不是DOWN，则整体状态是DEGRADED
			if depStatus == StatusDegraded && status != StatusDown {
				status = StatusDegraded
			}
		}

		response := map[string]interface{}{
			"status":       status,
			"dependencies": deps,
		}

		httpStatus := http.StatusOK
		if status == StatusDown {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, response)
	}
}

// SimpleHandler 返回简单的健康检查HTTP处理函数 (只返回状态)
func (hc *HealthCheck) SimpleHandler() func(*context.Context) {
	return func(c *context.Context) {
		hc.mutex.RLock()
		status := hc.status
		hc.mutex.RUnlock()

		if status == StatusDown {
			c.String(http.StatusServiceUnavailable, string(status))
			return
		}

		c.String(http.StatusOK, string(status))
	}
}

// DBCheck 创建数据库健康检查
func DBCheck(ping func() error) CheckFunc {
	return func() (Status, map[string]interface{}) {
		start := time.Now()
		err := ping()
		latency := time.Since(start)

		details := map[string]interface{}{
			"latency_ms": latency.Milliseconds(),
		}

		if err != nil {
			details["error"] = err.Error()
			return StatusDown, details
		}

		return StatusUp, details
	}
}

// HTTPCheck 创建HTTP服务健康检查
func HTTPCheck(url string, timeout time.Duration) CheckFunc {
	client := &http.Client{
		Timeout: timeout,
	}

	return func() (Status, map[string]interface{}) {
		start := time.Now()
		resp, err := client.Get(url)
		latency := time.Since(start)

		details := map[string]interface{}{
			"url":        url,
			"latency_ms": latency.Milliseconds(),
		}

		if err != nil {
			details["error"] = err.Error()
			return StatusDown, details
		}
		defer resp.Body.Close()

		details["status_code"] = resp.StatusCode

		if resp.StatusCode >= 500 {
			return StatusDown, details
		} else if resp.StatusCode >= 400 {
			return StatusDegraded, details
		}

		return StatusUp, details
	}
}

// RedisCheck 创建Redis健康检查
func RedisCheck(ping func() error) CheckFunc {
	return func() (Status, map[string]interface{}) {
		start := time.Now()
		err := ping()
		latency := time.Since(start)

		details := map[string]interface{}{
			"latency_ms": latency.Milliseconds(),
		}

		if err != nil {
			details["error"] = err.Error()
			return StatusDown, details
		}

		return StatusUp, details
	}
}

// DiskSpaceCheck 创建磁盘空间健康检查
func DiskSpaceCheck(path string, warningThresholdPercent, criticalThresholdPercent float64) CheckFunc {
	return func() (Status, map[string]interface{}) {
		// 这里应该实现获取磁盘空间的逻辑
		// 由于需要依赖系统调用，这里只是示例结构

		// 假设我们获取了磁盘使用百分比
		usedPercent := 75.5 // 示例值

		details := map[string]interface{}{
			"path":               path,
			"used_percent":       usedPercent,
			"warning_threshold":  warningThresholdPercent,
			"critical_threshold": criticalThresholdPercent,
		}

		if usedPercent >= criticalThresholdPercent {
			return StatusDown, details
		} else if usedPercent >= warningThresholdPercent {
			return StatusDegraded, details
		}

		return StatusUp, details
	}
}

// MemoryCheck 创建内存使用健康检查
func MemoryCheck(warningThresholdPercent, criticalThresholdPercent float64) CheckFunc {
	return func() (Status, map[string]interface{}) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		details := map[string]interface{}{
			"alloc_mb":       m.Alloc / 1024 / 1024,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":         m.Sys / 1024 / 1024,
			"gc_cycles":      m.NumGC,
		}

		// 这里应该有更复杂的判断逻辑
		// 简单起见，返回UP状态
		return StatusUp, details
	}
}

// CustomCheck 创建自定义健康检查
func CustomCheck(check func() (bool, string, map[string]interface{})) CheckFunc {
	return func() (Status, map[string]interface{}) {
		isHealthy, message, details := check()

		if details == nil {
			details = make(map[string]interface{})
		}

		if message != "" {
			details["message"] = message
		}

		if isHealthy {
			return StatusUp, details
		}

		return StatusDown, details
	}
}
