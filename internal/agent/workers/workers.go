package workers

import (
	"fmt"

	"github.com/erupshis/metrics/internal/logger"
)

type Job = func() error

type Pool struct {
	jobs    chan Job
	results chan error

	log logger.BaseLogger
}

func CreateWorkersPool(count int64, log logger.BaseLogger) (*Pool, error) {
	if count == 0 {
		return nil, fmt.Errorf("[CreateWorkersPool] no workers")
	}
	pool := &Pool{jobs: make(chan func() error, count), results: make(chan error, count), log: log}
	pool.createWorkers(count)
	return pool, nil
}

func (p *Pool) AddJob(job Job) {
	p.log.Info("[WorkersPool:AddJob] new job incoming.")
	p.jobs <- job
	p.log.Info("[WorkersPool:AddJob] new job added.")
}

func (p *Pool) CloseJobsChan() {
	p.log.Info("[WorkersPool:CloseJobsChan] jobs closed.")
	close(p.jobs)
}

func (p *Pool) GetResultChan() chan error {
	return p.results
}

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
	//worker stops when jobs channel is closed.
	for job := range p.jobs {
		p.log.Info("[WorkersPool:worker] worker starts job from queue.")
		err := job()
		p.log.Info("[WorkersPool:worker] worker is sending completed work to result queue.")
		p.results <- err
		p.log.Info("[WorkersPool:worker] worker has sent job result to result queue.")
	}
}
