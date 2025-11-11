package scrapers

import (
	"fmt"
	"regexp"
	"time"

	"github.com/biagioPiraino/delphico/scraper/internal/logger"
	"github.com/biagioPiraino/delphico/scraper/internal/utils"
	"github.com/gocolly/colly"
	"github.com/google/uuid"
)

const rootWebsite = "https://uk.finance.yahoo.com/news"
const contentSelector = ".mainContainer .body-wrap .body .bodyItems-wrapper p"

type YahooScraper struct{}

func NewReutersScraper() *YahooScraper {
	return &YahooScraper{}
}

func (s *YahooScraper) ScrapeWebsite() {
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(10), // leave to 0 default in production to keep scraping the site
		colly.URLFilters(regexp.MustCompile("^"+regexp.QuoteMeta(rootWebsite))),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*uk.finance.yahoo*",
		Parallelism: 2,
		Delay:       2 * time.Second,
		RandomDelay: 4 * time.Second,
	})

	// routing and visiting callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})

	// extract content callback
	c.OnHTML(contentSelector, func(e *colly.HTMLElement) {
		requestId := utils.GetRequestIdFromElement(e)
		logger.LogRequest(requestId, fmt.Sprintf("found content at %s", e.Request.URL))

		//paragraphText := e.Text
		// // Clean up and print the extracted text
		// fmt.Printf("URL: %s\n", e.Request.URL.String())
		// fmt.Printf("Extracted Paragraph: %s\n", paragraphText)
		// fmt.Println("---")
	})

	c.OnRequest(func(r *colly.Request) {
		requestId := uuid.New().String()
		r.Headers.Add(utils.RequestIdHeader, requestId)
		logger.LogRequest(requestId, fmt.Sprintf("visiting %s", r.URL))
	})

	c.OnResponse(func(r *colly.Response) {
		requestId := utils.GetRequestIdFromResponse(r)
		logger.LogRequest(requestId, fmt.Sprintf("finished visiting %s - response: %d", r.Request.URL, r.StatusCode))
	})

	c.OnError(func(r *colly.Response, e error) {
		requestId := utils.GetRequestIdFromResponse(r)
		logger.LogRequest(requestId, fmt.Sprintf("error while visting %s - response: %d - details: \"%v\"", r.Request.URL, r.StatusCode, e))
	})

	c.Visit(rootWebsite)
	c.Wait()
}
