package services

import (
	"context"
	"net/http"
)

type CrawlerConfig struct {
	Root         string
	MaxDepth     int
	DomainGlobal string
	Parallelism  int
	AllowRevisit bool
}

type contextTransport struct {
	ctx   context.Context
	trans *http.Transport
}

func (t *contextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.WithContext(t.ctx)
	return t.trans.RoundTrip(req)
}
