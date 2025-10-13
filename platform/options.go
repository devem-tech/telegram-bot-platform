package platform

import "time"

// Option defines a functional option for configuring the Platform.
type Option[U Update] func(*Platform[U])

// WithClient sets the client used to receive incoming updates.
// The client must not be nil; otherwise, platform initialization will panic.
func WithClient[U Update](client Client[U]) Option[U] {
	return func(p *Platform[U]) {
		p.client = client
	}
}

// WithUsecases sets the list of usecases that will be evaluated against each incoming update.
// Each usecase's Matches method is called to determine if it should handle the update.
func WithUsecases[U Update](usecases ...Usecase[U]) Option[U] {
	return func(p *Platform[U]) {
		p.usecases = usecases
	}
}

// WithUsecaseFallback sets the fallback handler for errors that occur during update processing.
// This handler is invoked if an error is returned by a usecase or an update middleware.
func WithUsecaseFallback[U Update](fallback UsecaseFallback[U]) Option[U] {
	return func(p *Platform[U]) {
		p.usecaseFallback = fallback
	}
}

// WithUsecaseMiddlewares appends middleware functions to the update processing pipeline.
// These middlewares wrap the dispatch logic and are executed in the order they are added.
// Each middleware receives the update and a next handler, allowing for cross-cutting concerns
// such as logging, tracing, or validation.
func WithUsecaseMiddlewares[U Update](mw ...UsecaseMiddleware[U]) Option[U] {
	return func(p *Platform[U]) {
		p.usecaseMiddlewares = append(p.usecaseMiddlewares, mw...)
	}
}

// WithJobs sets the list of cron jobs to be scheduled and executed periodically.
// Each job must implement the Job interface, including Name(), Cron(), and Handle().
func WithJobs[U Update](jobs ...Job) Option[U] {
	return func(p *Platform[U]) {
		p.jobs = jobs
	}
}

// WithJobFallback sets the fallback handler for errors that occur during cron job execution.
// This handler is invoked if a job's Handle method returns an error.
func WithJobFallback[U Update](jobFallback JobFallback) Option[U] {
	return func(p *Platform[U]) {
		p.jobFallback = jobFallback
	}
}

// WithJobMiddlewares appends middleware functions to the cron job execution pipeline.
// These middlewares wrap each job's Handle method and are useful for tracing, logging, or metrics.
// The job name is available in the context via the JobNameKey{} key.
func WithJobMiddlewares[U Update](mw ...JobMiddleware) Option[U] {
	return func(p *Platform[U]) {
		p.jobMiddlewares = append(p.jobMiddlewares, mw...)
	}
}

// WithNWorkers sets the number of concurrent worker goroutines that process update tasks.
// Defaults to 10 if not specified via options.
func WithNWorkers[U Update](nWorkers int) Option[U] {
	return func(p *Platform[U]) {
		p.nWorkers = nWorkers
	}
}

// WithMaxTasks sets the buffer size of the task channel used to queue update processing tasks.
// If the channel is full, the platform will block on sending new tasks.
// Defaults to 100 if not specified via options.
func WithMaxTasks[U Update](maxTasks int) Option[U] {
	return func(p *Platform[U]) {
		p.maxTasks = maxTasks
	}
}

// WithLocation sets the time zone used by the cron scheduler for interpreting cron expressions.
// Defaults to "Europe/Moscow" if not specified via options.
func WithLocation[U Update](location *time.Location) Option[U] {
	return func(p *Platform[U]) {
		p.location = location
	}
}
