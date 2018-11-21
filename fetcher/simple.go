package fetcher

import (
	"github.com/cocktail18/orcworker"
	"net/http"
	"strings"
)

type SimpleFetcher struct {
	client *http.Client
}

func NewSimpleFetcher() *SimpleFetcher {
	return &SimpleFetcher{
		&http.Client{},
	}
}

func (simpleFetcher SimpleFetcher) Fetch(seed *orcworker.Seed) (*http.Response, error) {
	req, err := http.NewRequest(seed.Method, seed.URL, strings.NewReader(seed.Body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header = seed.Header
	return simpleFetcher.client.Do(req) // 这里要 close body才行
}
