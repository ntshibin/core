// Package handlers provides a set of handlers for the glog logging system.
// It implements the chain of responsibility pattern for log processing.
package handlers

import (
	"github.com/sirupsen/logrus"
)

// Chain represents a chain of log handlers.
// It implements the chain of responsibility pattern, allowing multiple
// handlers to process log entries in sequence.
type Chain struct {
	head Handler
	tail Handler
}

// NewChain creates a new handler chain.
// The chain starts empty and handlers can be added to it.
func NewChain() *Chain {
	return &Chain{}
}

// Add appends a new handler to the chain.
// If the chain is empty, the handler becomes both the head and tail.
// Otherwise, the handler is added after the current tail.
func (c *Chain) Add(handler Handler) {
	if c.head == nil {
		c.head = handler
		c.tail = handler
		return
	}

	if base, ok := c.tail.(*BaseHandler); ok {
		base.SetNext(handler)
	}
	c.tail = handler
}

// Process runs the log entry through all handlers in the chain.
// It starts with the head handler, which then passes control to subsequent handlers.
// The logger parameter is passed to each handler for processing.
// Note: logger parameter must be a pointer to logrus.Logger, not a copied value.
func (c *Chain) Process(logger *logrus.Logger, args ...interface{}) {
	if c.head != nil {
		c.head.Handle(logger, args...)
	}
}
