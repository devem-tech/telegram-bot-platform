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

type RequestID struct{}

type Update = any

// Client defines the interface that returns incoming telegram updates.
type Client[U Update] interface {
	Updates(ctx context.Context) <-chan U
}

// Usecase represents a handler for specific types of updates.
// It determines whether it should handle a given update, and processes it.
type Usecase[U Update] interface {
	Matches(ctx context.Context, update U) (bool, error)
	Handle(ctx context.Context, update U) error
}

// Fallback defines a handler for processing errors that occur during handling.
type Fallback[U Update] interface {
	Handle(ctx context.Context, update U, err error)
}

// Job represents a cron job to be executed periodically.
type Job interface {
	Cron() string
	Handle(ctx context.Context) error
}

// JobFallback is a handler for processing errors that occur during job execution.
// It works similarly to Fallback, but for scheduled jobs.
type JobFallback interface {
	Handle(ctx context.Context, err error)
}

type HandlerFunc[U Update] func(ctx context.Context, update U) error

type Middleware[U Update] func(ctx context.Context, update U, next HandlerFunc[U]) error

// Task represents the task to process an Usecase.
type Task[U Update] struct {
	usecase  Usecase[U]
	fallback Fallback[U]
	update   U
}

// Platform encapsulates the message handling system,
// including workers, usecases, cron jobs, and a fallback strategy.
type Platform[U Update] struct {
	client      Client[U]
	usecases    []Usecase[U]
	fallback    Fallback[U]
	jobs        []Job
	jobFallback JobFallback
	middlewares []Middleware[U]
	nWorkers    int
	maxTasks    int
	location    *time.Location
}

// New constructs a new Platform instance, applying any optional configuration.
func New[U Update](options ...Option[U]) *Platform[U] {
	location, err := time.LoadLocation(defaultLocation)
	if err != nil {
		panic(err)
	}

	platform := &Platform[U]{
		client:      nil,
		usecases:    nil,
		fallback:    nil,
		jobs:        nil,
		jobFallback: nil,
		middlewares: nil,
		nWorkers:    defaultNWorkers,
		maxTasks:    defaultMaxTasks,
		location:    location,
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

		go p.worker(ctx, tasks, &wg)
	}

	// Receive and handle updates from the client.
	for update := range p.client.Updates(ctx) {
		// Dispatch the update to each usecase.
		for _, usecase := range p.usecases {
			tasks <- Task[U]{
				usecase:  usecase,
				fallback: p.fallback,
				update:   update,
			}
		}
	}

	// Close the task channel and wait for workers to finish.
	close(tasks)

	// Wait for the completion of all workers.
	wg.Wait()
}

// worker processes incoming tasks from the task channel.
func (p *Platform[U]) worker(ctx context.Context, tasks <-chan Task[U], wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		// Generate a unique request ID and store it in the context.
		ctx := context.WithValue(ctx, RequestID{}, uuid.New().String())

		core := func(ctx context.Context, _ U) error {
			// Check if the usecase matches the update.
			matches, err := task.usecase.Matches(ctx, task.update)
			if err != nil {
				if task.fallback != nil {
					task.fallback.Handle(ctx, task.update, err)
				}

				return nil
			} else if !matches {
				return nil
			}

			// Handle the update using the matched usecase.
			if err = task.usecase.Handle(ctx, task.update); err != nil {
				if task.fallback != nil {
					task.fallback.Handle(ctx, task.update, err)
				}
			}

			return nil
		}

		h := chain(p.middlewares, core)
		if err := h(ctx, task.update); err != nil {
			if task.fallback != nil {
				task.fallback.Handle(ctx, task.update, err)
			}
		}
	}
}

// cron initializes and starts all registered cron jobs.
func (p *Platform[U]) cron(ctx context.Context) *cron.Cron {
	x := cron.New(
		cron.WithLocation(p.location),
	)

	for _, job := range p.jobs {
		if _, err := x.AddFunc(job.Cron(), func() {
			if err := job.Handle(ctx); err != nil {
				if p.jobFallback != nil {
					p.jobFallback.Handle(ctx, err)
				}
			}
		}); err != nil {
			panic(err)
		}
	}

	x.Start()

	return x
}

func chain[U Update](mw []Middleware[U], endpoint HandlerFunc[U]) HandlerFunc[U] {
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
