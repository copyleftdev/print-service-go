package services

import (
	"context"
	"fmt"
	"sync"

	"print-service/internal/infrastructure/logger"
	"print-service/internal/pkg/config"
)

// QueueService provides job queue functionality
type QueueService struct {
	jobs    chan interface{}
	config  config.QueueConfig
	logger  logger.Logger
	running bool
	mutex   sync.RWMutex
}

// NewQueueService creates a new queue service
func NewQueueService(cfg config.QueueConfig, logger logger.Logger) (*QueueService, error) {
	return &QueueService{
		jobs:   make(chan interface{}, 100), // Buffer for jobs
		config: cfg,
		logger: logger.With("service", "queue"),
	}, nil
}

// Enqueue adds a job to the queue
func (qs *QueueService) Enqueue(job interface{}) error {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	if !qs.running {
		return fmt.Errorf("queue service is not running")
	}

	select {
	case qs.jobs <- job:
		qs.logger.Debug("Job enqueued successfully")
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

// StartConsumer starts consuming jobs from the queue
func (qs *QueueService) StartConsumer(ctx context.Context, handler func(interface{}) error) error {
	qs.mutex.Lock()
	qs.running = true
	qs.mutex.Unlock()

	qs.logger.Info("Starting queue consumer")

	for {
		select {
		case job := <-qs.jobs:
			if err := handler(job); err != nil {
				qs.logger.Error("Job processing failed", "error", err)
			}
		case <-ctx.Done():
			qs.logger.Info("Queue consumer stopped")
			return ctx.Err()
		}
	}
}

// Stop stops the queue service
func (qs *QueueService) Stop(ctx context.Context) {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	qs.running = false
	close(qs.jobs)
	qs.logger.Info("Queue service stopped")
}
