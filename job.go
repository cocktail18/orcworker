package orcworker

import (
	"net/http"
)

type Job struct {
	fetcher   Fetcher
	seed      *Seed
	scheduler *Scheduler
	Response  *http.Response
	Result    map[string]interface{}
	done      chan error
}

func NewJob(scheduler *Scheduler, seed *Seed) *Job {
	job := &Job{}
	job.fetcher = scheduler.fetcher
	job.seed = seed
	job.scheduler = scheduler
	job.Result = make(map[string]interface{})
	job.done = make(chan error, 1)
	return job
}

func (job *Job) Run() error {
	go job.run()
	select {
	case <-job.scheduler.tomb.Dying():
		return nil
	case err := <-job.done:
		return err
	}
}

func (job *Job) run() {
	var err error
	job.Response, err = job.fetcher.Fetch(job.seed)
	if err != nil {
		goto DONE
	}
	defer job.Response.Body.Close()
	for _, processor := range job.scheduler.processors {
		var seeds []*Seed
		seeds, err = processor(job)
		if err != nil {
			goto DONE
		}
		for _, seed := range seeds {
			seed.Deep = job.seed.Deep + 1
			err = job.scheduler.AddSeed(seed)
			if err != nil {
				goto DONE
			}
		}
	}

DONE:
	job.done <- err
	return
}
