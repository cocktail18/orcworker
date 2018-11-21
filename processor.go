package orcworker

type Processor func(job *Job) (seeds []*Seed, err error)
