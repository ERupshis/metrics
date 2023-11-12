// Package workers provides Worker pool to perform some job.
package workers

import (
	"fmt"

	"github.com/erupshis/metrics/internal/logger"
)

// Job defines type of applicable jobs.
type Job = func() error

// Pool stores communication channels for worker's manipulation and logger.
type Pool struct {
	jobs    chan Job
	results chan error

	log logger.BaseLogger
}

// CreateWorkersPool returns Pool with 'count' workers inside.
// Count param shouldn't be 0.
func CreateWorkersPool(count int64, log logger.BaseLogger) (*Pool, error) {
	if count == 0 {
		return nil, fmt.Errorf("[CreateWorkersPool] no workers")
	}
	pool := &Pool{jobs: make(chan func() error, count), results: make(chan error, count), log: log}
	pool.createWorkers(count)
	return pool, nil
}

// AddJob adds job in income channel to delegate job task to some free worker.
func (p *Pool) AddJob(job Job) {
	p.log.Info("[WorkersPool:AddJob] new job incoming.")
	p.jobs <- job
	p.log.Info("[WorkersPool:AddJob] new job added.")
}

// CloseJobsChan closes jobs channel.
// Should be called right after create function via defer.
func (p *Pool) CloseJobsChan() {
	p.log.Info("[WorkersPool:CloseJobsChan] jobs closed.")
	close(p.jobs)
}

// GetResultChan returns result channel consists of errors from workers.
func (p *Pool) GetResultChan() chan error {
	return p.results
}

// CloseResultsChan closes results channel.
// Should be called right after create function via defer.
func (p *Pool) CloseResultsChan() {
	p.log.Info("[WorkersPool:CloseJobsChan] results closed.")
	close(p.results)
}

func (p *Pool) createWorkers(count int64) {
	for i := 0; i < int(count); i++ {
		go p.worker()
	}
}

func (p *Pool) worker() {
	// worker stops when jobs channel is closed.
	for job := range p.jobs {
		p.log.Info("[WorkersPool:worker] worker starts job from queue.")
		err := job()
		p.log.Info("[WorkersPool:worker] worker is sending completed work to result queue.")
		p.results <- err
		p.log.Info("[WorkersPool:worker] worker has sent job result to result queue.")
	}
}
