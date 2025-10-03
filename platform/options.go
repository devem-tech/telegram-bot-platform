package platform

import "time"

// Option defines a functional option for configuring the Platform.
type Option[U Update] func(*Platform[U])

// WithClient sets the client used to receive updates.
func WithClient[U Update](client Client[U]) Option[U] {
	return func(p *Platform[U]) {
		p.client = client
	}
}

// WithUsecases sets the usecases that will handle incoming updates.
func WithUsecases[U Update](usecases ...Usecase[U]) Option[U] {
	return func(p *Platform[U]) {
		p.usecases = usecases
	}
}

// WithFallback sets the fallback handler for error processing.
func WithFallback[U Update](fallback Fallback[U]) Option[U] {
	return func(p *Platform[U]) {
		p.fallback = fallback
	}
}

// WithJobs sets the cron jobs to be scheduled.
func WithJobs[U Update](jobs ...Job) Option[U] {
	return func(p *Platform[U]) {
		p.jobs = jobs
	}
}

// WithJobFallback sets the fallback handler for errors occurring during job execution.
func WithJobFallback[U Update](jobFallback JobFallback) Option[U] {
	return func(p *Platform[U]) {
		p.jobFallback = jobFallback
	}
}

func WithMiddlewares[U Update](mw ...Middleware[U]) Option[U] {
	return func(p *Platform[U]) {
		p.middlewares = append(p.middlewares, mw...)
	}
}

// WithNWorkers sets the number of worker goroutines.
func WithNWorkers[U Update](nWorkers int) Option[U] {
	return func(p *Platform[U]) {
		p.nWorkers = nWorkers
	}
}

// WithMaxTasks sets the maximum number of queued tasks.
func WithMaxTasks[U Update](maxTasks int) Option[U] {
	return func(p *Platform[U]) {
		p.maxTasks = maxTasks
	}
}

// WithLocation sets the time location for scheduled jobs.
func WithLocation[U Update](location *time.Location) Option[U] {
	return func(p *Platform[U]) {
		p.location = location
	}
}
