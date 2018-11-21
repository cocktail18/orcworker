package orcworker

import (
	"context"
	"errors"
	"github.com/google/logger"
	"gopkg.in/bsm/ratelimit.v2"
	"gopkg.in/tomb.v2"
	"io/ioutil"
	"sync"
	"sync/atomic"
	"time"
)

type Scheduler struct {
	storage       Storage
	fetcher       Fetcher
	processors    []Processor
	failJob       int64
	successJob    int64
	runningJob    int64
	lock          sync.RWMutex
	logger        *logger.Logger
	ctx           context.Context
	tomb          *tomb.Tomb
	maxDeep       int // 爬取的最大深度
	maxRunningJob int // 同时爬取数量限制
	rate          *ratelimit.RateLimiter
	runningChan   chan struct{}
}

func NewScheduler(storage Storage, fetcher Fetcher, maxDeep, maxRunningJob int, rate *ratelimit.RateLimiter) (*Scheduler, error) {
	scheduler := &Scheduler{}
	scheduler.fetcher = fetcher
	scheduler.storage = storage
	scheduler.maxDeep = maxDeep
	if maxRunningJob <= 0 {
		maxRunningJob = 1 << 30
	}
	scheduler.maxRunningJob = maxRunningJob
	scheduler.processors = make([]Processor, 0, 10)
	scheduler.logger = logger.Init("orcworker", true, false, ioutil.Discard) //@todo
	scheduler.tomb, scheduler.ctx = tomb.WithContext(context.Background())
	scheduler.rate = rate
	scheduler.runningChan = make(chan struct{}, maxRunningJob)
	return scheduler, nil
}

func (scheduler *Scheduler) Start(seeds ...*Seed) error {
	l, err := scheduler.storage.QueueCapacity()
	if err != nil {
		return err
	}
	if len(seeds) <= 0 && l <= 0 {
		return errors.New("Require seed! ")
	}
	if len(scheduler.processors) <= 0 {
		return errors.New("Require processor! ")
	}
	for _, seed := range seeds {
		err = scheduler.AddSeed(seed)
		if err != nil {
			return err
		}
	}
	scheduler.tomb.Go(func() error {
		return scheduler.run()
	})
	return scheduler.tomb.Wait()
}

func (scheduler *Scheduler) run() error {
	for {
		select {
		case <-scheduler.tomb.Dying():
			return nil
		default:
			if scheduler.isDone() {
				return nil
			}
			if scheduler.rate != nil && !scheduler.rate.Limit() {
				time.Sleep(time.Microsecond)
				continue
			}
			seed, err := scheduler.storage.DeQueue()
			if err != nil {
				if err == ERR_SEEDS_EMPTY {
					time.Sleep(time.Microsecond)
					continue
				} else {
					scheduler.logger.Warningf("get seed error: ", err.Error())
					continue
				}
			}

			select {
			case <-scheduler.tomb.Dying():
				return nil
			case scheduler.runningChan <- struct{}{}:
				// get seed
				job := NewJob(scheduler, seed)
				atomic.AddInt64(&scheduler.runningJob, 1)
				scheduler.tomb.Go(func() error {
					defer func() {
						<-scheduler.runningChan /**/
					}()
					err := job.Run()
					atomic.AddInt64(&scheduler.runningJob, -1)
					if err != nil {
						scheduler.logger.Warningf("job run error: %v", err)
						atomic.AddInt64(&scheduler.failJob, 1)
					} else {
						atomic.AddInt64(&scheduler.successJob, 1)
					}
					return nil
				})
			}
		}
	}
}

func (scheduler *Scheduler) isDone() bool {
	//scheduler.lock.RLock()
	//defer scheduler.lock.RUnlock()
	c, err := scheduler.storage.QueueCapacity()
	if err != nil {
		scheduler.logger.Warning("get seeds cap error: ", err.Error())
		return false
	}
	if c == 0 && scheduler.runningJob == 0 {
		return true
	}
	return false
}

func (scheduler *Scheduler) Stop(err error) {
	scheduler.Stat()
	scheduler.tomb.Kill(err)
}

func (scheduler *Scheduler) AddProcessor(processor ...Processor) {
	scheduler.processors = append(scheduler.processors, processor...)
}

func (scheduler *Scheduler) AddSeed(seed *Seed) error {
	if seed.Deep > scheduler.maxDeep && scheduler.maxDeep != -1 {
		return nil
	}
	b, err := scheduler.storage.IsContain(seed)
	if err != nil {
		return err
	}
	if !b {
		return scheduler.storage.EnQueue(seed)
	}
	return nil
}

func (scheduler *Scheduler) Stat() {
	scheduler.logger.Infoln("runningJob ------- ", scheduler.runningJob)
	scheduler.logger.Infoln("failJob ------- ", scheduler.failJob)
	scheduler.logger.Infoln("successJob ------- ", scheduler.successJob)
	queueCapacity, _ := scheduler.storage.QueueCapacity()
	scheduler.logger.Infoln("queue capacity ------- ", queueCapacity)
}
