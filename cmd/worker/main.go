package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"print-service/internal/core/services"
	"print-service/internal/pkg/config"
	"print-service/internal/pkg/pool"
	"print-service/internal/infrastructure/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	log := logger.NewStructuredLogger(&cfg.Logger)
	defer log.Sync()

	// Create worker pool
	workerPool := pool.NewWorkerPool(cfg.Worker.PoolSize, log)

	// Initialize queue service
	queueService, err := services.NewQueueService(cfg.Queue, log)
	if err != nil {
		log.Error("Failed to initialize queue service", "error", err)
		os.Exit(1)
	}

	// Initialize print service
	printService, err := services.NewPrintService(cfg.Print, log)
	if err != nil {
		log.Error("Failed to initialize print service", "error", err)
		os.Exit(1)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker pool
	workerPool.Start(ctx, func(job interface{}) error {
		return printService.ProcessJob(ctx, job)
	})

	// Start queue consumer
	go func() {
		log.Info("Starting queue consumer")
		if err := queueService.StartConsumer(ctx, workerPool.Submit); err != nil {
			log.Error("Queue consumer failed", "error", err)
			cancel()
		}
	}()

	log.Info("Worker started", "pool_size", cfg.Worker.PoolSize)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")

	// Cancel context to stop all workers
	cancel()

	// Give workers time to finish current jobs
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	workerPool.Stop(shutdownCtx)
	queueService.Stop(shutdownCtx)

	log.Info("Worker exited")
}
