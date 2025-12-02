package utils

import (
	"github.com/gocolly/colly"
)

const RequestIdHeader = "X-Request-ID"

func GetRequestIdFromResponse(r *colly.Response) string {
	return r.Request.Headers.Get(RequestIdHeader)
}
