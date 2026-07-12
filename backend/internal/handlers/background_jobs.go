package handlers

import (
	"runtime/debug"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type backgroundJobRunner struct {
	jobs chan func()
}

var (
	handlerBackgroundJobs = newBackgroundJobRunner(4, 128)
	backgroundJobsDropped = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "handler_background_jobs_dropped_total",
		Help: "Handler background jobs rejected because the bounded queue was full",
	})
)

func init() {
	prometheus.MustRegister(backgroundJobsDropped)
}

func newBackgroundJobRunner(workers, capacity int) *backgroundJobRunner {
	runner := &backgroundJobRunner{jobs: make(chan func(), capacity)}
	for range workers {
		go func() {
			for job := range runner.jobs {
				func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							utils.Error("Background handler job panicked", nil, map[string]interface{}{
								"panic": recovered,
								"stack": string(debug.Stack()),
							})
						}
					}()
					job()
				}()
			}
		}()
	}
	return runner
}

func (r *backgroundJobRunner) Submit(job func()) bool {
	select {
	case r.jobs <- job:
		return true
	default:
		backgroundJobsDropped.Inc()
		return false
	}
}
