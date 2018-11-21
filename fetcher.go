package orcworker

import (
	"net/http"
)

type Fetcher interface {
	Fetch(seed *Seed) (*http.Response, error)
}
