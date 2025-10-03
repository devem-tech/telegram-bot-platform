package platform

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"github.com/devem-tech/telegram-bot-platform/telegram"
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

// Client defines the interface that returns incoming telegram updates.
type Client interface {
	Updates() telegram.UpdateCh
}

// Usecase represents a handler for specific types of updates.
// It determines whether it should handle a given update, and processes it.
type Usecase interface {
	Matches(ctx context.Context, update telegram.Update) (bool, error)
	Handle(ctx context.Context, update telegram.Update) error
}

// Fallback defines a handler for processing errors that occur during handling.
type Fallback interface {
	Handle(ctx context.Context, update telegram.Update, err error)
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

type HandlerFunc func(ctx context.Context, update telegram.Update) error

type Middleware func(ctx context.Context, update telegram.Update, next HandlerFunc) error

// Task represents the task to process a Usecase.
type Task struct {
	usecase  Usecase
	fallback Fallback
	update   telegram.Update
}

// Platform encapsulates the message handling system,
// including workers, usecases, cron jobs, and a fallback strategy.
type Platform struct {
	client      Client
	usecases    []Usecase
	fallback    Fallback
	jobs        []Job
	jobFallback JobFallback
	middlewares []Middleware
	nWorkers    int
	maxTasks    int
	location    *time.Location
}

// New constructs a new Platform instance, applying any optional configuration.
func New(options ...Option) *Platform {
	location, err := time.LoadLocation(defaultLocation)
	if err != nil {
		panic(err)
	}

	platform := &Platform{
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
func (p *Platform) Run(ctx context.Context) {
	// Start cron jobs.
	defer p.cron(ctx).Stop()

	// Use sync.WaitGroup to wait for all workers to complete.
	var wg sync.WaitGroup

	// Create a buffered task channel.
	tasks := make(chan Task, p.maxTasks)

	// Launch workers to process tasks concurrently.
	for range p.nWorkers {
		wg.Add(1)

		go p.worker(ctx, tasks, &wg)
	}

	// Receive and handle updates from the client.
	for update := range p.client.Updates() {
		if update.Message == nil {
			// Skip non-message updates.
			continue
		}

		// Convert the update to an internal representation.
		in := telegram.In(update)

		// Dispatch the update to each usecase.
		for _, usecase := range p.usecases {
			tasks <- Task{
				usecase:  usecase,
				fallback: p.fallback,
				update:   in,
			}
		}
	}

	// Close the task channel and wait for workers to finish.
	close(tasks)

	// Wait for the completion of all workers.
	wg.Wait()
}

// worker processes incoming tasks from the task channel.
func (p *Platform) worker(ctx context.Context, tasks <-chan Task, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		// Generate a unique request ID and store it in the context.
		ctx := context.WithValue(ctx, RequestID{}, uuid.New().String())

		core := func(ctx context.Context, _ telegram.Update) error {
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
func (p *Platform) cron(ctx context.Context) *cron.Cron {
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

func chain(mw []Middleware, endpoint HandlerFunc) HandlerFunc {
	x := endpoint
	for i := len(mw) - 1; i >= 0; i-- {
		next := x
		m := mw[i]
		x = func(ctx context.Context, update telegram.Update) error {
			return m(ctx, update, next)
		}
	}

	return x
}
