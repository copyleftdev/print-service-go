package middleware

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PerformanceMonitor tracks system performance metrics
type PerformanceMonitor struct {
	activeJobs     int64
	totalJobs      int64
	successfulJobs int64
	failedJobs     int64
	maxConcurrent  int64
	startTime      time.Time
	mutex          sync.RWMutex
	logger         *zap.Logger
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger *zap.Logger, maxConcurrent int64) *PerformanceMonitor {
	return &PerformanceMonitor{
		maxConcurrent: maxConcurrent,
		startTime:     time.Now(),
		logger:        logger,
	}
}

// PerformanceMiddleware provides performance monitoring and rate limiting
func (pm *PerformanceMonitor) PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if we're at concurrent job limit (based on rigor test findings)
		currentActive := atomic.LoadInt64(&pm.activeJobs)
		if currentActive >= pm.maxConcurrent {
			pm.logger.Warn("Concurrent job limit reached",
				zap.Int64("active_jobs", currentActive),
				zap.Int64("max_concurrent", pm.maxConcurrent))

			c.JSON(429, gin.H{
				"error":          "System at capacity",
				"message":        "Too many concurrent requests. Please try again later.",
				"active_jobs":    currentActive,
				"max_concurrent": pm.maxConcurrent,
			})
			c.Abort()
			return
		}

		// Track job start
		atomic.AddInt64(&pm.activeJobs, 1)
		atomic.AddInt64(&pm.totalJobs, 1)
		startTime := time.Now()

		// Process request
		c.Next()

		// Track job completion
		atomic.AddInt64(&pm.activeJobs, -1)
		duration := time.Since(startTime)

		// Track success/failure based on status code
		if c.Writer.Status() < 400 {
			atomic.AddInt64(&pm.successfulJobs, 1)
		} else {
			atomic.AddInt64(&pm.failedJobs, 1)
		}

		// Log performance metrics
		pm.logger.Info("Request completed",
			zap.Duration("duration", duration),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("active_jobs", atomic.LoadInt64(&pm.activeJobs)),
			zap.String("path", c.Request.URL.Path))
	}
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	total := atomic.LoadInt64(&pm.totalJobs)
	successful := atomic.LoadInt64(&pm.successfulJobs)
	failed := atomic.LoadInt64(&pm.failedJobs)
	active := atomic.LoadInt64(&pm.activeJobs)

	var successRate float64
	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
	}

	uptime := time.Since(pm.startTime)

	return map[string]interface{}{
		"active_jobs":     active,
		"total_jobs":      total,
		"successful_jobs": successful,
		"failed_jobs":     failed,
		"success_rate":    successRate,
		"max_concurrent":  pm.maxConcurrent,
		"capacity_usage":  float64(active) / float64(pm.maxConcurrent) * 100,
		"uptime_seconds":  uptime.Seconds(),
		"jobs_per_minute": float64(total) / uptime.Minutes(),
	}
}

// IsAtCapacity checks if system is at or near capacity
func (pm *PerformanceMonitor) IsAtCapacity() bool {
	active := atomic.LoadInt64(&pm.activeJobs)
	return active >= pm.maxConcurrent
}

// GetCapacityUsage returns current capacity usage percentage
func (pm *PerformanceMonitor) GetCapacityUsage() float64 {
	active := atomic.LoadInt64(&pm.activeJobs)
	return float64(active) / float64(pm.maxConcurrent) * 100
}
