// Package handlers provides a set of handlers for the glog logging system.
// It implements the chain of responsibility pattern for log processing.
package handlers

import (
	"github.com/sirupsen/logrus"
)

// LevelHandler implements level-based filtering of log entries.
// It ensures that only log entries meeting the minimum level requirement are processed.
type LevelHandler struct {
	BaseHandler
	Level logrus.Level
}

// NewLevelHandler creates a new level handler with the specified minimum level.
// Log entries below this level will not be processed.
func NewLevelHandler(level logrus.Level) *LevelHandler {
	return &LevelHandler{
		Level: level,
	}
}

// Handle processes the log entry based on the configured level.
// It only passes control to the next handler if the log level is at or above the minimum level.
func (h *LevelHandler) Handle(logger *logrus.Logger, args ...interface{}) {
	// Set the level
	logger.SetLevel(h.Level)

	// Call the next handler
	if h.Next != nil {
		h.Next.Handle(logger, args...)
	}
}
