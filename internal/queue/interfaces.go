package queue

import (
	"context"
	"time"
)

// Job represents a unit of work to be processed asynchronously
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    int                    `json:"priority"` // Higher numbers = higher priority
	MaxRetries  int                    `json:"max_retries"`
	Retries     int                    `json:"retries"`
	CreatedAt   time.Time              `json:"created_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
)

// JobResult represents the result of job processing
type JobResult struct {
	JobID     string                 `json:"job_id"`
	Status    JobStatus              `json:"status"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	CreatedAt time.Time              `json:"created_at"`
}

// Producer defines the interface for job producers
type Producer interface {
	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, job *Job) error

	// EnqueueBatch adds multiple jobs to the queue
	EnqueueBatch(ctx context.Context, jobs []*Job) error

	// GetJob retrieves a job by ID
	GetJob(ctx context.Context, jobID string) (*Job, error)

	// GetJobResult retrieves the result of a processed job
	GetJobResult(ctx context.Context, jobID string) (*JobResult, error)

	// Close closes the producer
	Close() error
}

// Consumer defines the interface for job consumers
type Consumer interface {
	// StartConsuming starts consuming jobs from the queue
	StartConsuming(ctx context.Context, concurrency int) error

	// RegisterHandler registers a job handler for a specific job type
	RegisterHandler(jobType string, handler JobHandler)

	// StopConsuming stops consuming jobs
	StopConsuming() error

	// Close closes the consumer
	Close() error
}

// JobHandler defines the interface for job processing handlers
type JobHandler interface {
	// Handle processes a job and returns the result
	Handle(ctx context.Context, job *Job) (*JobResult, error)

	// Name returns the handler name
	Name() string

	// JobType returns the job type this handler processes
	JobType() string
}

// Queue combines producer and consumer interfaces
type Queue interface {
	Producer
	Consumer
}

// Stats represents queue statistics
type Stats struct {
	PendingJobs    int64 `json:"pending_jobs"`
	ProcessingJobs int64 `json:"processing_jobs"`
	CompletedJobs  int64 `json:"completed_jobs"`
	FailedJobs     int64 `json:"failed_jobs"`
	TotalJobs      int64 `json:"total_jobs"`
}

// Monitor defines the interface for queue monitoring
type Monitor interface {
	// GetStats returns queue statistics
	GetStats(ctx context.Context) (*Stats, error)

	// GetFailedJobs returns a list of failed jobs
	GetFailedJobs(ctx context.Context, limit, offset int) ([]*Job, error)

	// RetryFailedJob retries a failed job
	RetryFailedJob(ctx context.Context, jobID string) error

	// PurgeCompletedJobs removes completed jobs older than the specified duration
	PurgeCompletedJobs(ctx context.Context, olderThan time.Duration) (int64, error)
}

// Config holds queue configuration
type Config struct {
	// Redis connection settings
	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`

	// Queue settings
	QueueName       string        `mapstructure:"queue_name"`
	MaxRetries      int           `mapstructure:"max_retries"`
	RetryDelay      time.Duration `mapstructure:"retry_delay"`
	JobTimeout      time.Duration `mapstructure:"job_timeout"`
	ResultTTL       time.Duration `mapstructure:"result_ttl"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`

	// Consumer settings
	DefaultConcurrency int           `mapstructure:"default_concurrency"`
	PollInterval       time.Duration `mapstructure:"poll_interval"`

	// Monitoring settings
	StatsEnabled        bool          `mapstructure:"stats_enabled"`
	StatsUpdateInterval time.Duration `mapstructure:"stats_update_interval"`
}
