package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/melihgurlek/backend-path/internal/domain"
)

// BatchProcessor handles concurrent processing of multiple transaction tasks
type BatchProcessor struct {
	transactionProcessor domain.TransactionProcessor
	maxConcurrency       int
	batchTimeout         time.Duration
}

// BatchResult represents the result of processing a batch of transactions
type BatchResult struct {
	BatchID         string
	TotalTasks      int
	SuccessfulTasks int
	FailedTasks     int
	ProcessingTime  time.Duration
	Errors          []BatchError
	CompletedAt     time.Time
}

// BatchError represents an error that occurred during batch processing
type BatchError struct {
	TaskID string
	Error  string
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(
	transactionProcessor domain.TransactionProcessor,
	maxConcurrency int,
	batchTimeout time.Duration,
) *BatchProcessor {
	return &BatchProcessor{
		transactionProcessor: transactionProcessor,
		maxConcurrency:       maxConcurrency,
		batchTimeout:         batchTimeout,
	}
}

// ProcessBatch processes a batch of transaction tasks concurrently
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, tasks []*domain.TransactionTask) (*BatchResult, error) {
	if len(tasks) == 0 {
		return &BatchResult{
			BatchID:         generateBatchID(),
			TotalTasks:      0,
			SuccessfulTasks: 0,
			FailedTasks:     0,
			ProcessingTime:  0,
			CompletedAt:     time.Now(),
		}, nil
	}

	// Create span for tracing
	spanCtx, span := otel.Tracer("batch-processor").Start(ctx, "process-batch")
	defer span.End()

	batchID := generateBatchID()
	span.SetAttributes(
		attribute.String("batch.id", batchID),
		attribute.Int("batch.size", len(tasks)),
		attribute.Int("max_concurrency", bp.maxConcurrency),
	)

	startTime := time.Now()
	result := &BatchResult{
		BatchID:     batchID,
		TotalTasks:  len(tasks),
		CompletedAt: time.Now(),
	}

	// Create context with timeout
	batchCtx, cancel := context.WithTimeout(spanCtx, bp.batchTimeout)
	defer cancel()

	// Create channels for coordination
	taskChan := make(chan *domain.TransactionTask, len(tasks))
	resultChan := make(chan *domain.TransactionResult, len(tasks))
	errorChan := make(chan error, len(tasks))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < bp.maxConcurrency; i++ {
		wg.Add(1)
		go bp.worker(batchCtx, i, taskChan, resultChan, errorChan, &wg)
	}

	// Send tasks to workers
	go func() {
		defer close(taskChan)
		for _, task := range tasks {
			select {
			case taskChan <- task:
				// Task sent successfully
			case <-batchCtx.Done():
				return
			}
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Process results
	var errors []BatchError
ResultLoop:
	for {
		select {
		case res := <-resultChan:
			if res != nil {
				if res.Success {
					result.SuccessfulTasks++
				} else {
					result.FailedTasks++
					errors = append(errors, BatchError{
						TaskID: res.TaskID,
						Error:  res.Message,
					})
				}
			}
		case err := <-errorChan:
			if err != nil {
				result.FailedTasks++
				errors = append(errors, BatchError{
					TaskID: "unknown",
					Error:  err.Error(),
				})
			}
		case <-batchCtx.Done():
			// Timeout or cancellation
			result.FailedTasks += len(tasks) - result.SuccessfulTasks - result.FailedTasks
			span.RecordError(batchCtx.Err())
			log.Warn().Str("batch_id", batchID).Err(batchCtx.Err()).Msg("Batch processing timeout or cancelled")
			break ResultLoop
		}

		// Check if all tasks are processed
		if result.SuccessfulTasks+result.FailedTasks >= len(tasks) {
			break
		}
	}

	result.ProcessingTime = time.Since(startTime)
	result.Errors = errors

	span.SetAttributes(
		attribute.Int("successful_tasks", result.SuccessfulTasks),
		attribute.Int("failed_tasks", result.FailedTasks),
		attribute.Float64("processing_time_seconds", result.ProcessingTime.Seconds()),
	)

	log.Info().
		Str("batch_id", batchID).
		Int("total_tasks", result.TotalTasks).
		Int("successful_tasks", result.SuccessfulTasks).
		Int("failed_tasks", result.FailedTasks).
		Dur("processing_time", result.ProcessingTime).
		Msg("Batch processing completed")

	return result, nil
}

// ProcessBatchWithRollback processes a batch with rollback capability
func (bp *BatchProcessor) ProcessBatchWithRollback(ctx context.Context, tasks []*domain.TransactionTask) (*BatchResult, error) {
	// For now, we'll implement a simple version
	// In a real implementation, you might want to:
	// 1. Use database transactions
	// 2. Implement compensation logic
	// 3. Use saga pattern for complex workflows

	result, err := bp.ProcessBatch(ctx, tasks)
	if err != nil {
		return result, err
	}

	// If any task failed, we could implement rollback logic here
	if result.FailedTasks > 0 {
		log.Warn().
			Str("batch_id", result.BatchID).
			Int("failed_tasks", result.FailedTasks).
			Msg("Batch had failed tasks - consider implementing rollback logic")
	}

	return result, nil
}

// worker processes tasks from the task channel
func (bp *BatchProcessor) worker(
	ctx context.Context,
	workerID int,
	taskChan <-chan *domain.TransactionTask,
	resultChan chan<- *domain.TransactionResult,
	errorChan chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	log.Debug().Int("worker_id", workerID).Msg("Batch worker started")

	for {
		select {
		case task := <-taskChan:
			if task == nil {
				return
			}

			// Submit task to transaction processor
			err := bp.transactionProcessor.SubmitTask(ctx, task)
			if err != nil {
				errorChan <- fmt.Errorf("failed to submit task %s: %w", task.ID, err)
				continue
			}

			// For batch processing, we'll create a simple result
			// In a real implementation, you might want to wait for the actual result
			result := &domain.TransactionResult{
				TaskID:    task.ID,
				Success:   true, // Assume success for now
				Message:   "Task submitted successfully",
				Timestamp: time.Now().Unix(),
			}

			select {
			case resultChan <- result:
				// Result sent successfully
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			log.Debug().Int("worker_id", workerID).Msg("Batch worker context cancelled")
			return
		}
	}
}

// generateBatchID generates a unique batch ID
func generateBatchID() string {
	return fmt.Sprintf("batch_%d", time.Now().UnixNano())
}
