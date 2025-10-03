package platform

import "time"

// Option defines a functional option for configuring the Platform.
type Option func(*Platform)

// WithClient sets the client used to receive updates.
func WithClient(client Client) Option {
	return func(p *Platform) {
		p.client = client
	}
}

// WithUsecases sets the usecases that will handle incoming updates.
func WithUsecases(usecases ...Usecase) Option {
	return func(p *Platform) {
		p.usecases = usecases
	}
}

// WithFallback sets the fallback handler for error processing.
func WithFallback(fallback Fallback) Option {
	return func(p *Platform) {
		p.fallback = fallback
	}
}

// WithJobs sets the cron jobs to be scheduled.
func WithJobs(jobs ...Job) Option {
	return func(p *Platform) {
		p.jobs = jobs
	}
}

// WithJobFallback sets the fallback handler for errors occurring during job execution.
func WithJobFallback(jobFallback JobFallback) Option {
	return func(p *Platform) {
		p.jobFallback = jobFallback
	}
}

func WithMiddlewares(mw ...Middleware) Option {
	return func(p *Platform) {
		p.middlewares = append(p.middlewares, mw...)
	}
}

// WithNWorkers sets the number of worker goroutines.
func WithNWorkers(nWorkers int) Option {
	return func(p *Platform) {
		p.nWorkers = nWorkers
	}
}

// WithMaxTasks sets the maximum number of queued tasks.
func WithMaxTasks(maxTasks int) Option {
	return func(p *Platform) {
		p.maxTasks = maxTasks
	}
}

// WithLocation sets the time location for scheduled jobs.
func WithLocation(location *time.Location) Option {
	return func(p *Platform) {
		p.location = location
	}
}
