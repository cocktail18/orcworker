package orcworker

import (
	"net/http"
)

type Downloader interface {
	Fetch(seed *Seed) (*http.Response, error)
}
