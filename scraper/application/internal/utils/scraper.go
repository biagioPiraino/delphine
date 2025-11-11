package utils

import (
	"regexp"

	"github.com/gocolly/colly"
)

const RequestIdHeader = "X-Request-ID"

func GetRequestIdFromResponse(r *colly.Response) string {
	return r.Request.Headers.Get(RequestIdHeader)
}

func GetRequestIdFromElement(e *colly.HTMLElement) string {
	return e.Request.Headers.Get(RequestIdHeader)
}

func AddSpacesAfterDots(text string) string {
	trailingCharOnDot := `\.(\w+)`
	re := regexp.MustCompile(trailingCharOnDot)
	return re.ReplaceAllString(text, `. $1`)
}
