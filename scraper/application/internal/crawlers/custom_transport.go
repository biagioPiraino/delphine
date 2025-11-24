package crawlers

import (
	"context"
	"net/http"
)

type contextTransport struct {
	ctx   context.Context
	trans *http.Transport
}

func (t *contextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.WithContext(t.ctx)
	return t.trans.RoundTrip(req)
}
