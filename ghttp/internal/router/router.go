// Package router 提供了HTTP路由管理
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ntshibin/core/ghttp/internal/config"
	"github.com/ntshibin/core/ghttp/internal/context"
)

// Router 是对 gin.Engine 的封装
type Router struct {
	Engine *gin.Engine
	Server *http.Server
	Config config.HTTPConfig
}

// HandlerFunc 定义了处理HTTP请求的函数类型
type HandlerFunc func(*context.Context)

// RouterGroup 是对 gin.RouterGroup 的封装
type RouterGroup struct {
	group *gin.RouterGroup
}

// Group 创建一个新的路由组
func (r *Router) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	ginHandlers := convertToGinHandlers(handlers)
	return &RouterGroup{group: r.Engine.Group(relativePath, ginHandlers...)}
}

// Use 添加全局中间件
func (r *Router) Use(handlers ...HandlerFunc) {
	r.Engine.Use(convertToGinHandlers(handlers)...)
}

// Static 添加静态文件服务
func (r *Router) Static(relativePath, root string) {
	r.Engine.Static(relativePath, root)
}

// StaticFile 添加单个静态文件
func (r *Router) StaticFile(relativePath, filepath string) {
	r.Engine.StaticFile(relativePath, filepath)
}

// StaticFS 添加自定义的文件系统
func (r *Router) StaticFS(relativePath string, fs http.FileSystem) {
	r.Engine.StaticFS(relativePath, fs)
}

// handleHTTPMethod 处理HTTP请求方法
func (r *Router) handleHTTPMethod(httpMethod, relativePath string, handlers ...HandlerFunc) {
	r.Engine.Handle(httpMethod, relativePath, convertToGinHandlers(handlers)...)
}

// GET 处理GET请求
func (r *Router) GET(relativePath string, handlers ...HandlerFunc) {
	r.handleHTTPMethod(http.MethodGet, relativePath, handlers...)
}

// POST 处理POST请求
func (r *Router) POST(relativePath string, handlers ...HandlerFunc) {
	r.handleHTTPMethod(http.MethodPost, relativePath, handlers...)
}

// PUT 处理PUT请求
func (r *Router) PUT(relativePath string, handlers ...HandlerFunc) {
	r.handleHTTPMethod(http.MethodPut, relativePath, handlers...)
}

// DELETE 处理DELETE请求
func (r *Router) DELETE(relativePath string, handlers ...HandlerFunc) {
	r.handleHTTPMethod(http.MethodDelete, relativePath, handlers...)
}

// PATCH 处理PATCH请求
func (r *Router) PATCH(relativePath string, handlers ...HandlerFunc) {
	r.handleHTTPMethod(http.MethodPatch, relativePath, handlers...)
}

// HEAD 处理HEAD请求
func (r *Router) HEAD(relativePath string, handlers ...HandlerFunc) {
	r.handleHTTPMethod(http.MethodHead, relativePath, handlers...)
}

// OPTIONS 处理OPTIONS请求
func (r *Router) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	r.handleHTTPMethod(http.MethodOptions, relativePath, handlers...)
}

// Group 在当前组中创建一个新的路由组
func (g *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	ginHandlers := convertToGinHandlers(handlers)
	return &RouterGroup{group: g.group.Group(relativePath, ginHandlers...)}
}

// Use 为当前组添加中间件
func (g *RouterGroup) Use(handlers ...HandlerFunc) {
	g.group.Use(convertToGinHandlers(handlers)...)
}

// handleHTTPMethod 处理HTTP请求方法
func (g *RouterGroup) handleHTTPMethod(httpMethod, relativePath string, handlers ...HandlerFunc) {
	g.group.Handle(httpMethod, relativePath, convertToGinHandlers(handlers)...)
}

// GET 处理GET请求
func (g *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) {
	g.handleHTTPMethod(http.MethodGet, relativePath, handlers...)
}

// POST 处理POST请求
func (g *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) {
	g.handleHTTPMethod(http.MethodPost, relativePath, handlers...)
}

// PUT 处理PUT请求
func (g *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) {
	g.handleHTTPMethod(http.MethodPut, relativePath, handlers...)
}

// DELETE 处理DELETE请求
func (g *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) {
	g.handleHTTPMethod(http.MethodDelete, relativePath, handlers...)
}

// PATCH 处理PATCH请求
func (g *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) {
	g.handleHTTPMethod(http.MethodPatch, relativePath, handlers...)
}

// HEAD 处理HEAD请求
func (g *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) {
	g.handleHTTPMethod(http.MethodHead, relativePath, handlers...)
}

// OPTIONS 处理OPTIONS请求
func (g *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	g.handleHTTPMethod(http.MethodOptions, relativePath, handlers...)
}

// 工具函数，将自定义的 HandlerFunc 转换为 gin.HandlerFunc
func convertToGinHandlers(handlers []HandlerFunc) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = func(c *gin.Context) {
			handler(&context.Context{Context: c})
		}
	}
	return ginHandlers
}
