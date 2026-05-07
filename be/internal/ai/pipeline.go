package ai

import (
	"context"
	"log"

	"github.com/google/uuid"
)

// Job holds the data needed for one AI processing task.
type Job struct {
	PhraseID    uuid.UUID
	TextLatin   string
	Translation string
}

// Pipeline is a channel-based worker pool for async AI processing.
type Pipeline struct {
	jobs    chan Job
	svc     *Service
	workers int
}

// NewPipeline creates a Pipeline with the given number of workers.
func NewPipeline(svc *Service, workers int) *Pipeline {
	if workers <= 0 {
		workers = 3
	}
	return &Pipeline{
		jobs:    make(chan Job, 100),
		svc:     svc,
		workers: workers,
	}
}

// Start launches the worker goroutines. It respects ctx cancellation.
func (p *Pipeline) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		go func(workerID int) {
			log.Printf("[AI pipeline] worker %d started", workerID)
			for {
				select {
				case <-ctx.Done():
					log.Printf("[AI pipeline] worker %d stopping", workerID)
					return
				case job, ok := <-p.jobs:
					if !ok {
						return
					}
					p.svc.Process(ctx, job.PhraseID, job.TextLatin, job.Translation)
				}
			}
		}(i)
	}
}

// Enqueue adds a job to the pipeline. Non-blocking: drops if queue is full.
func (p *Pipeline) Enqueue(job Job) {
	select {
	case p.jobs <- job:
	default:
		log.Printf("[AI pipeline] queue full, dropping job for phrase %s", job.PhraseID)
	}
}