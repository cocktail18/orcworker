package orcworker

import (
	"net/url"
	"net/http"
	"errors"
	"crypto/sha256"
	"bytes"
	"strings"
)

var ERR_SEEDS_EMPTY = errors.New("Empty Seed ")

type Seed struct {
	URL    string
	Method string
	Body   url.Values
	Header http.Header
	Deep   int // 深度
}

func NewSeed(rawURL string, method string, body url.Values, header http.Header) (*Seed, error) {
	seed := &Seed{}
	seed.URL = rawURL
	seed.Method = strings.ToUpper(method)
	seed.Body = body
	seed.Header = header
	seed.Deep = 0
	return seed, nil
}

func (seed Seed) Sha256() (string, error) {
	var h = sha256.New()
	buf := bytes.NewBuffer(nil)
	buf.WriteString(seed.URL)
	buf.WriteString(seed.Method)
	buf.WriteString(seed.Body.Encode())
	err := seed.Header.Write(buf)
	if err != nil {
		return "", err
	}
	h.Write(buf.Bytes())
	return string(h.Sum(nil)), nil
}

type Storage interface {
	EnQueue(seed *Seed) error
	DeQueue() (*Seed, error)
	IsContain(seed *Seed) (bool, error)
	QueueCapacity() (int, error)
}
