package platform

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

const (
	// defaultNWorkers defines the number of workers to process tasks.
	defaultNWorkers = 10

	// defaultMaxTasks defines the channel capacity for processing tasks.
	defaultMaxTasks = 100

	// defaultLocation is the default timezone used by the cron scheduler.
	defaultLocation = "Europe/Moscow"
)

type (
	// RequestID is the context key for unique request identifiers.
	// Value type: string.
	RequestID struct{}
)

// Update is a type alias for any incoming event (e.g., Telegram update).
// It allows the platform to be generic over different update types.
type Update = any

// Client defines the interface that returns a channel of incoming updates.
// The channel must be closed when the client is shutting down or the context is cancelled.
// Implementations should respect the provided context for cancellation.
type Client[U Update] interface {
	Updates(ctx context.Context) <-chan U
}

// Usecase represents a handler for specific types of updates.
// It first checks if it should handle a given update via Matches,
// and if so, processes it via Handle.
type Usecase[U Update] interface {
	// Matches determines whether this usecase should handle the given update.
	// It may return an error if the matching logic fails (e.g., due to I/O).
	// Returning (false, nil) means the update is ignored by this usecase.
	Matches(ctx context.Context, update U) (bool, error)

	// Handle processes the update.
	// It is only called if Matches returned (true, nil).
	// Errors returned here will be passed to the UsecaseFallback, if configured.
	Handle(ctx context.Context, update U) error
}

// UsecaseFallback defines a handler for errors that occur during update processing.
// It is invoked when an Usecase's Matches or Handle method returns an error,
// or when a UsecaseMiddleware returns an error.
type UsecaseFallback[U Update] interface {
	// Handle receives the original update and the error that occurred.
	// The context may contain request-scoped values such as RequestID.
	Handle(ctx context.Context, update U, err error)
}

// HandlerFunc is a function type that processes an update.
// It is used as the terminal handler in the middleware chain for updates.
type HandlerFunc[U Update] func(ctx context.Context, update U) error

// UsecaseMiddleware is a middleware function that wraps update handling logic.
// It receives the current context, the update, and the next handler in the chain.
// Middleware can perform cross-cutting concerns such as logging, tracing, metrics, or validation.
type UsecaseMiddleware[U Update] func(ctx context.Context, update U, next HandlerFunc[U]) error

// Job represents a cron job to be executed periodically.
// Each job must provide a human-readable name, a cron schedule, and an execution handler.
type Job interface {
	// Cron returns the cron expression defining the job's schedule (e.g., "@daily", "0 2 * * *").
	// The expression is interpreted in the timezone configured via WithLocation.
	Cron() string

	// Handle contains the job's business logic.
	// Errors returned here will be passed to the JobFallback, if configured.
	Handle(ctx context.Context) error
}

// JobFallback is a handler for errors that occur during job execution.
// It works similarly to UsecaseFallback, but for scheduled cron jobs.
type JobFallback interface {
	// Handle receives the error that occurred during job execution.
	// The context may contain request-scoped values such as RequestID and JobNameKey.
	Handle(ctx context.Context, err error)
}

// JobHandlerFunc is a function type that executes a cron job's logic.
// It is used as the terminal handler in the job middleware chain.
type JobHandlerFunc func(ctx context.Context) error

// JobMiddleware is a middleware function that wraps cron job execution.
// It receives the current context and the next handler in the chain.
// The job name is available in the context via JobNameKey{}.
// Useful for tracing, logging, panic recovery, or metrics.
type JobMiddleware func(ctx context.Context, next JobHandlerFunc) error

// Task represents a unit of work to be processed by a worker goroutine.
// It bundles an usecase, its fallback, the update, and the associated context.
// The context is used to propagate request-scoped values (e.g., RequestID) and cancellation signals.
type Task[U Update] struct {
	ctx      context.Context //nolint:containedctx // Propagates request-scoped values and cancellation.
	usecase  Usecase[U]
	fallback UsecaseFallback[U]
	update   U
}

// Platform encapsulates the message handling system,
// including workers, usecases, cron jobs, and fallback strategies.
// It is configured via functional options and started with the Run method.
type Platform[U Update] struct {
	client             Client[U]
	usecases           []Usecase[U]
	usecaseFallback    UsecaseFallback[U]
	usecaseMiddlewares []UsecaseMiddleware[U]
	jobs               []Job
	jobFallback        JobFallback
	jobMiddlewares     []JobMiddleware
	nWorkers           int
	maxTasks           int
	location           *time.Location
}

// New constructs a new Platform instance, applying any optional configuration.
// At minimum, a Client must be provided via WithClient; otherwise, New panics.
func New[U Update](options ...Option[U]) *Platform[U] {
	location, err := time.LoadLocation(defaultLocation)
	if err != nil {
		panic(err)
	}

	platform := &Platform[U]{
		client:             nil,
		usecases:           nil,
		usecaseFallback:    nil,
		usecaseMiddlewares: nil,
		jobs:               nil,
		jobFallback:        nil,
		jobMiddlewares:     nil,
		nWorkers:           defaultNWorkers,
		maxTasks:           defaultMaxTasks,
		location:           location,
	}

	for _, opt := range options {
		opt(platform)
	}

	if platform.client == nil {
		panic("platform: client is nil")
	}

	return platform
}

// Run starts processing incoming updates and scheduled jobs.
// It blocks until the context is cancelled or the client's update channel is closed.
// On exit, it ensures all workers complete and the cron scheduler is stopped.
func (p *Platform[U]) Run(ctx context.Context) {
	// Start cron jobs.
	defer p.cron(ctx).Stop()

	// Use sync.WaitGroup to wait for all workers to complete.
	var wg sync.WaitGroup

	// Create a buffered task channel.
	tasks := make(chan Task[U], p.maxTasks)

	// Launch workers to process tasks concurrently.
	for range p.nWorkers {
		wg.Add(1)

		go p.worker(tasks, &wg)
	}

	// Dispatch the update to each usecase.
	dispatch := func(ctx context.Context, update U) error {
		for _, usecase := range p.usecases {
			tasks <- Task[U]{
				ctx:      ctx,
				usecase:  usecase,
				fallback: p.usecaseFallback,
				update:   update,
			}
		}

		return nil
	}

	x := chainUsecases(p.usecaseMiddlewares, dispatch)

	// Receive and handle updates from the client.
	for update := range p.client.Updates(ctx) {
		// Generate a unique request ID and store it in the context.
		ctx := context.WithValue(ctx, RequestID{}, uuid.New().String())

		if err := x(ctx, update); err != nil && p.usecaseFallback != nil {
			p.usecaseFallback.Handle(ctx, update, err)
		}
	}

	// Close the task channel and wait for workers to finish.
	close(tasks)

	// Wait for the completion of all workers.
	wg.Wait()
}

// worker processes incoming tasks from the task channel.
// It evaluates each usecase's Matches method and invokes Handle if matched.
// Errors from Matches or Handle are forwarded to the task's fallback handler.
func (p *Platform[U]) worker(tasks <-chan Task[U], wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		ctx := task.ctx

		matches, err := task.usecase.Matches(ctx, task.update)
		if err != nil {
			if task.fallback != nil {
				task.fallback.Handle(ctx, task.update, err)
			}

			continue
		} else if !matches {
			continue
		}

		if err = task.usecase.Handle(ctx, task.update); err != nil {
			if task.fallback != nil {
				task.fallback.Handle(ctx, task.update, err)
			}
		}
	}
}

// cron initializes and starts all registered cron jobs.
// Each job is wrapped with job middlewares and executed in its own goroutine by the cron scheduler.
// The job's name and a unique request ID are injected into the context before execution.
func (p *Platform[U]) cron(ctx context.Context) *cron.Cron {
	x := cron.New(
		cron.WithLocation(p.location),
	)

	for _, job := range p.jobs {
		ch := chainJobs(p.jobMiddlewares, func(ctx context.Context) error {
			if err := job.Handle(ctx); err != nil && p.jobFallback != nil {
				p.jobFallback.Handle(ctx, err)
			}

			return nil
		})

		if _, err := x.AddFunc(job.Cron(), func() {
			ctx := context.WithValue(ctx, RequestID{}, uuid.New().String())
			_ = ch(ctx)
		}); err != nil {
			panic(err)
		}
	}

	x.Start()

	return x
}

// chainUsecases builds a middleware chain that wraps the endpoint.
// Middlewares are applied in reverse order (last added executes first).
func chainUsecases[U Update](mw []UsecaseMiddleware[U], endpoint HandlerFunc[U]) HandlerFunc[U] {
	x := endpoint
	for i := len(mw) - 1; i >= 0; i-- {
		next := x
		m := mw[i]
		x = func(ctx context.Context, update U) error {
			return m(ctx, update, next)
		}
	}

	return x
}

// chainJobs builds a middleware chain for cron job execution.
// Middlewares are applied in reverse order (last added executes first).
func chainJobs(mw []JobMiddleware, endpoint JobHandlerFunc) JobHandlerFunc {
	x := endpoint
	for i := len(mw) - 1; i >= 0; i-- {
		next := x
		m := mw[i]
		x = func(ctx context.Context) error {
			return m(ctx, next)
		}
	}

	return x
}
