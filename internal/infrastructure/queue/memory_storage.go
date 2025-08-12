package queue

import (
	"fmt"
	"sort"
	"sync"

	"print-service/internal/core/domain"
)

// MemoryJobStorage provides in-memory job storage implementation
type MemoryJobStorage struct {
	jobs  map[string]*domain.PrintJob
	mutex sync.RWMutex
}

// NewMemoryJobStorage creates a new in-memory job storage
func NewMemoryJobStorage() *MemoryJobStorage {
	return &MemoryJobStorage{
		jobs: make(map[string]*domain.PrintJob),
	}
}

// SaveJob saves a job to memory
func (m *MemoryJobStorage) SaveJob(job *domain.PrintJob) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}
	if job.ID == "" {
		return fmt.Errorf("job ID cannot be empty")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create a copy to avoid external modifications
	jobCopy := *job
	m.jobs[job.ID] = &jobCopy

	return nil
}

// GetJob retrieves a job by ID
func (m *MemoryJobStorage) GetJob(id string) (*domain.PrintJob, error) {
	if id == "" {
		return nil, fmt.Errorf("job ID cannot be empty")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	job, exists := m.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", id)
	}

	// Return a copy to avoid external modifications
	jobCopy := *job
	return &jobCopy, nil
}

// UpdateJob updates an existing job
func (m *MemoryJobStorage) UpdateJob(job *domain.PrintJob) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}
	if job.ID == "" {
		return fmt.Errorf("job ID cannot be empty")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.jobs[job.ID]; !exists {
		return fmt.Errorf("job not found: %s", job.ID)
	}

	// Create a copy to avoid external modifications
	jobCopy := *job
	m.jobs[job.ID] = &jobCopy

	return nil
}

// ListJobs lists jobs with pagination
func (m *MemoryJobStorage) ListJobs(offset, limit int) ([]*domain.PrintJob, error) {
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Convert map to slice for sorting
	jobs := make([]*domain.PrintJob, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobCopy := *job
		jobs = append(jobs, &jobCopy)
	}

	// Sort by creation time (newest first)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.After(jobs[j].CreatedAt)
	})

	// Apply pagination
	if offset >= len(jobs) {
		return []*domain.PrintJob{}, nil
	}

	end := offset + limit
	if end > len(jobs) {
		end = len(jobs)
	}

	return jobs[offset:end], nil
}

// DeleteJob deletes a job by ID
func (m *MemoryJobStorage) DeleteJob(id string) error {
	if id == "" {
		return fmt.Errorf("job ID cannot be empty")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.jobs[id]; !exists {
		return fmt.Errorf("job not found: %s", id)
	}

	delete(m.jobs, id)
	return nil
}

// GetJobCount returns the total number of jobs
func (m *MemoryJobStorage) GetJobCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.jobs)
}

// GetJobsByStatus returns jobs filtered by status
func (m *MemoryJobStorage) GetJobsByStatus(status domain.JobStatus) ([]*domain.PrintJob, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var jobs []*domain.PrintJob
	for _, job := range m.jobs {
		if job.Status == status {
			jobCopy := *job
			jobs = append(jobs, &jobCopy)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.After(jobs[j].CreatedAt)
	})

	return jobs, nil
}

// Clear removes all jobs (useful for testing)
func (m *MemoryJobStorage) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.jobs = make(map[string]*domain.PrintJob)
}
