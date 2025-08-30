package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"mem_bank/pkg/logger"
)

// RedisQueue implements the Queue interface using Redis
type RedisQueue struct {
	client   *redis.Client
	logger   logger.Logger
	config   Config
	handlers map[string]JobHandler
	mu       sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewRedisQueue creates a new Redis-based queue
func NewRedisQueue(client *redis.Client, logger logger.Logger, config Config) *RedisQueue {
	// Set defaults
	if config.QueueName == "" {
		config.QueueName = "mem_bank_jobs"
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 30 * time.Second
	}
	if config.JobTimeout == 0 {
		config.JobTimeout = 5 * time.Minute
	}
	if config.ResultTTL == 0 {
		config.ResultTTL = 24 * time.Hour
	}
	if config.DefaultConcurrency == 0 {
		config.DefaultConcurrency = 5
	}
	if config.PollInterval == 0 {
		config.PollInterval = 1 * time.Second
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Hour
	}

	return &RedisQueue{
		client:   client,
		logger:   logger,
		config:   config,
		handlers: make(map[string]JobHandler),
		stopChan: make(chan struct{}),
	}
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, job *Job) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = q.config.MaxRetries
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshaling job: %w", err)
	}

	// Add job to priority queue (sorted set by priority and timestamp)
	score := float64(job.Priority)*1e9 + float64(job.CreatedAt.Unix())
	if err := q.client.ZAdd(ctx, q.getQueueKey(), redis.Z{
		Score:  score,
		Member: jobData,
	}).Err(); err != nil {
		return fmt.Errorf("enqueuing job: %w", err)
	}

	// Store job details separately for retrieval
	jobKey := q.getJobKey(job.ID)
	if err := q.client.Set(ctx, jobKey, jobData, q.config.ResultTTL).Err(); err != nil {
		return fmt.Errorf("storing job details: %w", err)
	}

	q.logger.WithFields(map[string]interface{}{
		"job_id":   job.ID,
		"job_type": job.Type,
		"priority": job.Priority,
	}).Info("Job enqueued")

	return nil
}

// EnqueueBatch adds multiple jobs to the queue
func (q *RedisQueue) EnqueueBatch(ctx context.Context, jobs []*Job) error {
	if len(jobs) == 0 {
		return nil
	}

	pipe := q.client.Pipeline()
	queueKey := q.getQueueKey()

	for _, job := range jobs {
		if job.ID == "" {
			job.ID = uuid.New().String()
		}
		if job.CreatedAt.IsZero() {
			job.CreatedAt = time.Now()
		}
		if job.MaxRetries == 0 {
			job.MaxRetries = q.config.MaxRetries
		}

		jobData, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("marshaling job %s: %w", job.ID, err)
		}

		score := float64(job.Priority)*1e9 + float64(job.CreatedAt.Unix())
		pipe.ZAdd(ctx, queueKey, redis.Z{
			Score:  score,
			Member: jobData,
		})

		jobKey := q.getJobKey(job.ID)
		pipe.Set(ctx, jobKey, jobData, q.config.ResultTTL)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("executing batch enqueue: %w", err)
	}

	q.logger.WithField("count", len(jobs)).Info("Jobs batch enqueued")
	return nil
}

// GetJob retrieves a job by ID
func (q *RedisQueue) GetJob(ctx context.Context, jobID string) (*Job, error) {
	jobKey := q.getJobKey(jobID)
	jobData, err := q.client.Get(ctx, jobKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("retrieving job: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("unmarshaling job: %w", err)
	}

	return &job, nil
}

// GetJobResult retrieves the result of a processed job
func (q *RedisQueue) GetJobResult(ctx context.Context, jobID string) (*JobResult, error) {
	resultKey := q.getResultKey(jobID)
	resultData, err := q.client.Get(ctx, resultKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job result not found: %s", jobID)
		}
		return nil, fmt.Errorf("retrieving job result: %w", err)
	}

	var result JobResult
	if err := json.Unmarshal([]byte(resultData), &result); err != nil {
		return nil, fmt.Errorf("unmarshaling job result: %w", err)
	}

	return &result, nil
}

// StartConsuming starts consuming jobs from the queue
func (q *RedisQueue) StartConsuming(ctx context.Context, concurrency int) error {
	if concurrency <= 0 {
		concurrency = q.config.DefaultConcurrency
	}

	q.logger.WithField("concurrency", concurrency).Info("Starting job consumer")

	// Start worker goroutines
	for i := 0; i < concurrency; i++ {
		q.wg.Add(1)
		go q.worker(ctx, i)
	}

	// Start cleanup goroutine
	q.wg.Add(1)
	go q.cleanup(ctx)

	return nil
}

// RegisterHandler registers a job handler for a specific job type
func (q *RedisQueue) RegisterHandler(jobType string, handler JobHandler) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.handlers[jobType] = handler
	q.logger.WithFields(map[string]interface{}{
		"job_type": jobType,
		"handler":  handler.Name(),
	}).Info("Job handler registered")
}

// StopConsuming stops consuming jobs
func (q *RedisQueue) StopConsuming() error {
	q.logger.Info("Stopping job consumer")
	close(q.stopChan)
	q.wg.Wait()
	return nil
}

// Close closes the queue
func (q *RedisQueue) Close() error {
	return q.StopConsuming()
}

// worker processes jobs from the queue
func (q *RedisQueue) worker(ctx context.Context, workerID int) {
	defer q.wg.Done()

	logger := q.logger.WithField("worker_id", workerID)
	logger.Info("Worker started")

	ticker := time.NewTicker(q.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker stopped due to context cancellation")
			return
		case <-q.stopChan:
			logger.Info("Worker stopped")
			return
		case <-ticker.C:
			q.processNextJob(ctx, logger)
		}
	}
}

// processNextJob processes the next available job
func (q *RedisQueue) processNextJob(ctx context.Context, workerLogger logger.Logger) {
	queueKey := q.getQueueKey()

	// Pop the highest priority job (highest score)
	redisResult, err := q.client.ZPopMax(ctx, queueKey, 1).Result()
	if err != nil {
		if err != redis.Nil {
			workerLogger.WithError(err).Error("Failed to pop job from queue")
		}
		return
	}

	if len(redisResult) == 0 {
		return // No jobs available
	}

	jobData := redisResult[0].Member.(string)
	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		workerLogger.WithError(err).Error("Failed to unmarshal job")
		return
	}

	jobLogger := workerLogger.WithFields(map[string]interface{}{
		"job_id":   job.ID,
		"job_type": job.Type,
		"retries":  job.Retries,
	})

	jobLogger.Info("Processing job")
	start := time.Now()

	// Get handler for job type
	q.mu.RLock()
	handler, exists := q.handlers[job.Type]
	q.mu.RUnlock()

	if !exists {
		q.handleJobFailure(ctx, &job, fmt.Errorf("no handler for job type: %s", job.Type), jobLogger)
		return
	}

	// Create job timeout context
	jobCtx, cancel := context.WithTimeout(ctx, q.config.JobTimeout)
	defer cancel()

	// Process job
	result, err := handler.Handle(jobCtx, &job)
	duration := time.Since(start)

	if err != nil {
		jobLogger.WithError(err).WithField("duration", duration).Error("Job processing failed")
		q.handleJobFailure(ctx, &job, err, jobLogger)
		return
	}

	// Job completed successfully
	result.JobID = job.ID
	result.Status = JobStatusCompleted
	result.Duration = duration
	result.CreatedAt = time.Now()

	q.storeJobResult(ctx, result, jobLogger)
	jobLogger.WithField("duration", duration).Info("Job completed successfully")
}

// handleJobFailure handles job processing failures with retry logic
func (q *RedisQueue) handleJobFailure(ctx context.Context, job *Job, jobErr error, jobLogger logger.Logger) {
	job.Retries++
	job.Error = jobErr.Error()

	if job.Retries < job.MaxRetries {
		// Retry the job with exponential backoff
		delay := time.Duration(job.Retries) * q.config.RetryDelay

		jobLogger.WithFields(map[string]interface{}{
			"retry_in": delay,
			"retries":  job.Retries,
		}).Warn("Job failed, will retry")

		// Re-enqueue job with delay using proper context handling
		q.wg.Add(1)
		go func() {
			defer q.wg.Done()

			timer := time.NewTimer(delay)
			defer timer.Stop()

			select {
			case <-timer.C:
				if err := q.Enqueue(ctx, job); err != nil {
					jobLogger.WithError(err).Error("Failed to re-enqueue job for retry")
				}
			case <-ctx.Done():
				jobLogger.Info("Context cancelled, skipping job retry")
				return
			case <-q.stopChan:
				jobLogger.Info("Queue stopped, skipping job retry")
				return
			}
		}()
	} else {
		// Maximum retries exceeded
		now := time.Now()
		job.FailedAt = &now

		result := &JobResult{
			JobID:     job.ID,
			Status:    JobStatusFailed,
			Error:     jobErr.Error(),
			Duration:  0,
			CreatedAt: now,
		}

		q.storeJobResult(ctx, result, jobLogger)
		jobLogger.Error("Job failed permanently after max retries")
	}
}

// storeJobResult stores the job result in Redis
func (q *RedisQueue) storeJobResult(ctx context.Context, result *JobResult, jobLogger logger.Logger) {
	resultKey := q.getResultKey(result.JobID)
	resultData, err := json.Marshal(result)
	if err != nil {
		jobLogger.WithError(err).Error("Failed to marshal job result")
		return
	}

	if err := q.client.Set(ctx, resultKey, resultData, q.config.ResultTTL).Err(); err != nil {
		jobLogger.WithError(err).Error("Failed to store job result")
	}
}

// cleanup performs periodic cleanup of old jobs and results
func (q *RedisQueue) cleanup(ctx context.Context) {
	defer q.wg.Done()

	ticker := time.NewTicker(q.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-q.stopChan:
			return
		case <-ticker.C:
			q.performCleanup(ctx)
		}
	}
}

// performCleanup removes old completed job results
func (q *RedisQueue) performCleanup(ctx context.Context) {
	// This is a simplified cleanup - in production you might want more sophisticated logic
	pattern := q.getResultKey("*")
	keys, err := q.client.Keys(ctx, pattern).Result()
	if err != nil {
		q.logger.WithError(err).Error("Failed to get result keys for cleanup")
		return
	}

	cleaned := 0
	for _, key := range keys {
		ttl := q.client.TTL(ctx, key).Val()
		if ttl < 0 { // Key has no expiration
			if err := q.client.Expire(ctx, key, q.config.ResultTTL).Err(); err != nil {
				q.logger.WithError(err).WithField("key", key).Warn("Failed to set TTL for result key")
			} else {
				cleaned++
			}
		}
	}

	if cleaned > 0 {
		q.logger.WithField("count", cleaned).Info("Cleaned up job results")
	}
}

// Redis key helpers
func (q *RedisQueue) getQueueKey() string {
	return fmt.Sprintf("%s:queue", q.config.QueueName)
}

func (q *RedisQueue) getJobKey(jobID string) string {
	return fmt.Sprintf("%s:job:%s", q.config.QueueName, jobID)
}

func (q *RedisQueue) getResultKey(jobID string) string {
	return fmt.Sprintf("%s:result:%s", q.config.QueueName, jobID)
}
