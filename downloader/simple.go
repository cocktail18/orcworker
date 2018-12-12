package downloader

import (
	"github.com/cocktail18/orcworker"
	"net/http"
	"strings"
)

type SimpleDownloader struct {
	client *http.Client
}

func NewSimpleDownloader() *SimpleDownloader {
	return &SimpleDownloader{
		&http.Client{},
	}
}

func (downloader SimpleDownloader) Fetch(seed *orcworker.Seed) (*http.Response, error) {
	req, err := http.NewRequest(seed.Method, seed.URL, strings.NewReader(seed.Body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header = seed.Header
	return downloader.client.Do(req) // 这里要 close body才行
}
