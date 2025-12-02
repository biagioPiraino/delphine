package utils

import (
	"bytes"
	"encoding/json"

	"github.com/biagioPiraino/delphico/scraper/internal/core/domain"
	"github.com/gocolly/colly"
)

const RequestIdHeader = "X-Request-ID"

func GetRequestIdFromResponse(r *colly.Response) string {
	return r.Request.Headers.Get(RequestIdHeader)
}

func GetArticlePayload(article domain.Article) ([]byte, error) {
	buf := domain.JsonBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer domain.JsonBufferPool.Put(buf)

	if err := json.NewEncoder(buf).Encode(article); err != nil {
		return nil, err
	}

	payload := make([]byte, buf.Len())
	copy(payload, buf.Bytes()) // copied to avoid concurrency issues related to async producer
	return payload, nil
}
