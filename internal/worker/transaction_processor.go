package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/melihgurlek/backend-path/internal/domain"
	"github.com/melihgurlek/backend-path/pkg/metrics"
)

// TransactionProcessorImpl implements domain.TransactionProcessor
type TransactionProcessorImpl struct {
	transactionService domain.TransactionService
	balanceService     domain.BalanceService

	// Worker pool configuration
	numWorkers int
	queueSize  int

	// Channels for task processing
	taskQueue   chan *domain.TransactionTask
	resultQueue chan *domain.TransactionResult
	stopChan    chan struct{}

	// Worker management
	workers  []*worker
	workerWg sync.WaitGroup

	// Atomic counters for statistics
	totalProcessed  int64
	successfulTasks int64
	failedTasks     int64
	activeWorkers   int32

	// Processing time tracking
	processTimes     []time.Duration
	processTimeMutex sync.RWMutex

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// worker represents a single worker in the pool
type worker struct {
	id        int
	processor *TransactionProcessorImpl
	ctx       context.Context
}

// NewTransactionProcessor creates a new transaction processor
func NewTransactionProcessor(
	transactionService domain.TransactionService,
	balanceService domain.BalanceService,
	numWorkers int,
	queueSize int,
) *TransactionProcessorImpl {
	ctx, cancel := context.WithCancel(context.Background())

	return &TransactionProcessorImpl{
		transactionService: transactionService,
		balanceService:     balanceService,
		numWorkers:         numWorkers,
		queueSize:          queueSize,
		taskQueue:          make(chan *domain.TransactionTask, queueSize),
		resultQueue:        make(chan *domain.TransactionResult, queueSize),
		stopChan:           make(chan struct{}),
		workers:            make([]*worker, 0, numWorkers),
		ctx:                ctx,
		cancel:             cancel,
	}
}

// Start starts the worker pool
func (p *TransactionProcessorImpl) Start(ctx context.Context) error {
	log.Info().Int("workers", p.numWorkers).Int("queue_size", p.queueSize).Msg("Starting transaction processor")

	// Start workers
	for i := 0; i < p.numWorkers; i++ {
		w := &worker{
			id:        i,
			processor: p,
			ctx:       ctx,
		}
		p.workers = append(p.workers, w)

		p.workerWg.Add(1)
		go w.start()
	}

	// Start result processor
	go p.processResults()

	log.Info().Msg("Transaction processor started successfully")
	return nil
}

// Stop gracefully stops the worker pool
func (p *TransactionProcessorImpl) Stop(ctx context.Context) error {
	log.Info().Msg("Stopping transaction processor")

	// Signal all workers to stop
	close(p.stopChan)
	p.cancel()

	// Wait for all workers to finish
	p.workerWg.Wait()

	// Close channels
	close(p.taskQueue)
	close(p.resultQueue)

	log.Info().Msg("Transaction processor stopped successfully")
	return nil
}

// SubmitTask submits a transaction task to the processing queue
func (p *TransactionProcessorImpl) SubmitTask(ctx context.Context, task *domain.TransactionTask) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	if task.ID == "" {
		return errors.New("task ID cannot be empty")
	}

	if task.Amount <= 0 {
		return errors.New("task amount must be positive")
	}

	// Create span for tracing
	_, span := otel.Tracer("transaction-processor").Start(ctx, "submit-task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task.id", task.ID),
		attribute.String("task.type", task.Type),
		attribute.Int("task.user_id", task.UserID),
		attribute.Float64("task.amount", task.Amount),
		attribute.Int("task.priority", task.Priority),
	)

	// Try to submit task to queue with timeout
	select {
	case p.taskQueue <- task:
		log.Debug().Str("task_id", task.ID).Msg("Task submitted to queue")
		metrics.TransactionQueueSize.Set(float64(len(p.taskQueue)))
		return nil
	case <-time.After(5 * time.Second):
		span.RecordError(errors.New("queue timeout"))
		return errors.New("queue is full, task submission timeout")
	case <-ctx.Done():
		span.RecordError(ctx.Err())
		return ctx.Err()
	}
}

// GetStats returns current processing statistics
func (p *TransactionProcessorImpl) GetStats() *domain.ProcessingStats {
	p.processTimeMutex.RLock()
	defer p.processTimeMutex.RUnlock()

	var avgProcessTime float64
	if len(p.processTimes) > 0 {
		var total time.Duration
		for _, pt := range p.processTimes {
			total += pt
		}
		avgProcessTime = float64(total) / float64(len(p.processTimes))
	}

	return &domain.ProcessingStats{
		TotalProcessed:     atomic.LoadInt64(&p.totalProcessed),
		SuccessfulTasks:    atomic.LoadInt64(&p.successfulTasks),
		FailedTasks:        atomic.LoadInt64(&p.failedTasks),
		QueueSize:          len(p.taskQueue),
		ActiveWorkers:      int(atomic.LoadInt32(&p.activeWorkers)),
		AverageProcessTime: avgProcessTime,
	}
}

// start starts a worker goroutine
func (w *worker) start() {
	defer w.processor.workerWg.Done()

	log.Debug().Int("worker_id", w.id).Msg("Worker started")

	for {
		select {
		case task := <-w.processor.taskQueue:
			if task == nil {
				return
			}
			w.processTask(task)
		case <-w.processor.stopChan:
			log.Debug().Int("worker_id", w.id).Msg("Worker stopping")
			return
		case <-w.ctx.Done():
			log.Debug().Int("worker_id", w.id).Msg("Worker context cancelled")
			return
		}
	}
}

// processTask processes a single transaction task
func (w *worker) processTask(task *domain.TransactionTask) {
	startTime := time.Now()
	atomic.AddInt32(&w.processor.activeWorkers, 1)
	defer atomic.AddInt32(&w.processor.activeWorkers, -1)

	// Create span for tracing
	_, span := otel.Tracer("transaction-processor").Start(context.Background(), "process-task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task.id", task.ID),
		attribute.String("task.type", task.Type),
		attribute.Int("task.user_id", task.UserID),
		attribute.Float64("task.amount", task.Amount),
		attribute.Int("worker.id", w.id),
	)

	result := &domain.TransactionResult{
		TaskID:    task.ID,
		Timestamp: time.Now().Unix(),
	}

	// Process the task based on type
	var err error
	switch task.Type {
	case "credit":
		err = w.processor.transactionService.Credit(task.UserID, task.Amount)
	case "debit":
		err = w.processor.transactionService.Debit(task.UserID, task.Amount)
	case "transfer":
		if task.ToUserID == nil {
			err = errors.New("transfer requires to_user_id")
		} else {
			err = w.processor.transactionService.Transfer(task.UserID, *task.ToUserID, task.Amount)
		}
	default:
		err = fmt.Errorf("unknown transaction type: %s", task.Type)
	}

	// Record result
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = err.Error()
		atomic.AddInt64(&w.processor.failedTasks, 1)
		span.RecordError(err)
		log.Error().Err(err).Str("task_id", task.ID).Int("worker_id", w.id).Msg("Task processing failed")
	} else {
		result.Success = true
		result.Message = "Task processed successfully"
		atomic.AddInt64(&w.processor.successfulTasks, 1)
		log.Debug().Str("task_id", task.ID).Int("worker_id", w.id).Msg("Task processed successfully")
	}

	atomic.AddInt64(&w.processor.totalProcessed, 1)

	// Record processing time
	processTime := time.Since(startTime)
	w.processor.processTimeMutex.Lock()
	w.processor.processTimes = append(w.processor.processTimes, processTime)
	// Keep only last 1000 processing times to avoid memory growth
	if len(w.processor.processTimes) > 1000 {
		w.processor.processTimes = w.processor.processTimes[1:]
	}
	w.processor.processTimeMutex.Unlock()

	span.SetAttributes(attribute.Float64("process_time_seconds", processTime.Seconds()))

	// Send result to result queue
	select {
	case w.processor.resultQueue <- result:
		// Result sent successfully
	default:
		log.Warn().Str("task_id", task.ID).Msg("Result queue full, dropping result")
	}

	// Update metrics
	metrics.TransactionProcessingDuration.WithLabelValues(task.Type).Observe(processTime.Seconds())
	if result.Success {
		metrics.TransactionProcessingSuccess.WithLabelValues(task.Type).Inc()
	} else {
		metrics.TransactionProcessingFailure.WithLabelValues(task.Type).Inc()
	}
}

// processResults processes results from the result queue
func (p *TransactionProcessorImpl) processResults() {
	for result := range p.resultQueue {
		if result == nil {
			continue
		}

		// Log result
		if result.Success {
			log.Debug().Str("task_id", result.TaskID).Msg("Task completed successfully")
		} else {
			log.Error().Str("task_id", result.TaskID).Err(result.Error).Msg("Task failed")
		}

		// Here you could add additional result processing logic
		// such as sending notifications, updating audit logs, etc.
	}
}
