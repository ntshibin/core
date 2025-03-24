// Package handlers provides a set of handlers for the glog logging system.
// It implements the chain of responsibility pattern for log processing.
package handlers

import (
	"github.com/sirupsen/logrus"
)

// Handler defines the interface for log processing in the chain of responsibility pattern.
// Each handler can process logs and pass them to the next handler in the chain.
type Handler interface {
	// Handle processes the log entry with given logger and arguments
	Handle(logger *logrus.Logger, args ...interface{})
}

// BaseHandler provides common functionality for handlers.
// It implements the Handler interface and provides basic chaining capability.
type BaseHandler struct {
	Next Handler
}

// SetNext sets the next handler in the chain.
// It returns the next handler for easy chaining.
func (h *BaseHandler) SetNext(handler Handler) Handler {
	h.Next = handler
	return handler
}

// Handle provides default implementation for handlers.
// It simply passes control to the next handler in the chain if one exists.
func (h *BaseHandler) Handle(logger *logrus.Logger, args ...interface{}) {
	if h.Next != nil {
		h.Next.Handle(logger, args...)
	}
}
