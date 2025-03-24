// Package handlers provides a set of handlers for the glog logging system.
// It implements the chain of responsibility pattern for log processing.
package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogEntry represents a log entry in the async queue.
// It contains all necessary information to process a log entry asynchronously.
type LogEntry struct {
	// Logger is the logger instance that generated this entry
	Logger *logrus.Logger
	// Args are the original arguments passed to the log function
	Args []interface{}
	// Level represents the logging level for this entry
	Level logrus.Level
}

// AsyncHandler implements asynchronous logging functionality.
// It buffers log entries in a channel and processes them in a separate goroutine,
// allowing the main application to continue execution without waiting for log I/O.
// This handler is particularly useful for high-throughput applications where logging
// should not block the main execution path.
type AsyncHandler struct {
	BaseHandler
	queue         chan LogEntry // Buffered channel for log entries
	batchSize     int           // Number of entries to process in one batch
	flushInterval time.Duration // Maximum duration to wait before processing
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	// Flag to track if Close has been called
	closed bool
	// Mutex to protect closed flag
	mu sync.RWMutex
}

// AsyncConfig defines the configuration for AsyncHandler.
// It controls the behavior of the async logging process.
type AsyncConfig struct {
	// BufferSize is the size of the channel buffer.
	// Larger values allow more logs to be queued before blocking or dropping entries.
	BufferSize int

	// BatchSize is the number of log entries to process in one batch.
	// Larger batches may improve throughput but increase latency.
	BatchSize int

	// FlushInterval is the maximum time to wait before processing a batch.
	// Even if the batch is not full, it will be processed after this duration.
	FlushInterval time.Duration
}

// DefaultAsyncConfig returns the default configuration for AsyncHandler.
// These defaults provide a reasonable balance between performance and resource usage.
func DefaultAsyncConfig() *AsyncConfig {
	return &AsyncConfig{
		BufferSize:    1000,        // Buffer up to 1000 log entries
		BatchSize:     100,         // Process in batches of 100
		FlushInterval: time.Second, // Flush at least once per second
	}
}

// NewAsyncHandler creates a new AsyncHandler with the specified configuration.
// If nil is passed for config, default configuration is used.
// The handler starts a background goroutine for processing logs.
func NewAsyncHandler(config *AsyncConfig) *AsyncHandler {
	if config == nil {
		config = DefaultAsyncConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	h := &AsyncHandler{
		queue:         make(chan LogEntry, config.BufferSize),
		batchSize:     config.BatchSize,
		flushInterval: config.FlushInterval,
		ctx:           ctx,
		cancel:        cancel,
		closed:        false,
	}

	// Start the background worker
	h.wg.Add(1)
	go h.processLogs()

	return h
}

// Handle processes the log entry asynchronously.
// It queues the log entry for async processing and then calls the next handler.
// If the queue is full, the log entry is dropped with a warning.
func (h *AsyncHandler) Handle(logger *logrus.Logger, args ...interface{}) {
	// Check if handler is closed
	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		// If closed, just pass to next handler
		if h.Next != nil {
			h.Next.Handle(logger, args...)
		}
		return
	}
	h.mu.RUnlock()

	// Determine the log level - default to Info if not determinable
	level := logrus.InfoLevel
	if len(args) > 0 {
		// Try to extract level from the logger's current state
		level = logger.GetLevel()
	}

	// Try to queue the log entry
	select {
	case h.queue <- LogEntry{Logger: logger, Args: args, Level: level}:
		// Log entry queued successfully
	default:
		// Queue is full, log a warning and drop the entry
		logrus.Warnf("Async log queue is full, dropping log entry. Consider increasing buffer size.")
	}

	// Call the next handler in the chain
	if h.Next != nil {
		h.Next.Handle(logger, args...)
	}
}

// processLogs processes log entries in batches.
// This is a background goroutine that runs until the handler is closed.
// It processes entries in batches for efficiency, or when the flush interval expires.
func (h *AsyncHandler) processLogs() {
	defer h.wg.Done()

	batch := make([]LogEntry, 0, h.batchSize)
	ticker := time.NewTicker(h.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			// Process remaining entries before shutting down
			h.processBatch(batch)
			return

		case entry, ok := <-h.queue:
			if !ok {
				// Channel closed, process remaining entries and exit
				h.processBatch(batch)
				return
			}

			batch = append(batch, entry)
			if len(batch) >= h.batchSize {
				h.processBatch(batch)
				batch = make([]LogEntry, 0, h.batchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				h.processBatch(batch)
				batch = make([]LogEntry, 0, h.batchSize)
			}
		}
	}
}

// processBatch processes a batch of log entries.
// It respects the log level of each entry and logs accordingly.
func (h *AsyncHandler) processBatch(batch []LogEntry) {
	if len(batch) == 0 {
		return
	}

	for _, entry := range batch {
		// Log at the appropriate level
		switch entry.Level {
		case logrus.DebugLevel:
			entry.Logger.Debug(entry.Args...)
		case logrus.InfoLevel:
			entry.Logger.Info(entry.Args...)
		case logrus.WarnLevel:
			entry.Logger.Warn(entry.Args...)
		case logrus.ErrorLevel:
			entry.Logger.Error(entry.Args...)
		case logrus.FatalLevel:
			// Be careful with Fatal as it calls os.Exit(1)
			// Consider downgrading to Error in async context
			entry.Logger.Error(append([]interface{}{"[FATAL]"}, entry.Args...))
		case logrus.PanicLevel:
			// Be careful with Panic as it calls panic()
			// Consider downgrading to Error in async context
			entry.Logger.Error(append([]interface{}{"[PANIC]"}, entry.Args...))
		default:
			entry.Logger.Info(entry.Args...)
		}
	}
}

// Close gracefully shuts down the AsyncHandler.
// It waits for all queued log entries to be processed before returning.
// This method should be called when the application is shutting down.
func (h *AsyncHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return nil // Already closed
	}

	// Mark as closed
	h.closed = true

	// Signal the worker to stop
	h.cancel()

	// Wait for all goroutines to finish with a timeout
	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		// Graceful shutdown completed
	case <-time.After(5 * time.Second):
		// Timeout - some logs might be lost
		logrus.Warn("AsyncHandler Close timed out after 5s, some logs might be lost")
	}

	// Close the queue
	close(h.queue)

	return nil
}
