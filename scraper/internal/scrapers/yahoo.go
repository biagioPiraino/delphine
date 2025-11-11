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
const metadataSelector = ".mainContainer .byline .byline-attr"
const contentSelector = ".mainContainer .body-wrap .body .bodyItems-wrapper"
const readmoreSelector = ".mainContainer .body-wrap .body .read-more-wrapper"

type YahooScraper struct{}

func NewReutersScraper() *YahooScraper {
	return &YahooScraper{}
}

func (s *YahooScraper) ScrapeWebsite() {
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(0), // leave to 0 default in production to keep scraping the site
		colly.URLFilters(regexp.MustCompile("^"+regexp.QuoteMeta(rootWebsite))),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*uk.finance.yahoo*",
		Parallelism: 5,
		Delay:       2 * time.Second,
		RandomDelay: 4 * time.Second,
	})

	// routing and visiting callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		e.Request.Visit(link)
	})

	// extract metadata callback
	c.OnHTML(metadataSelector, func(e *colly.HTMLElement) {
		requestId := utils.GetRequestIdFromElement(e)
		logger.LogRequest(requestId, fmt.Sprintf("found metadata at %s", e.Request.URL))

		author := e.ChildText(".byline-attr-author a")
		datePublished := e.ChildText(".byline-attr-time-style .byline-attr-meta-time")
		fmt.Printf("URL: %s\n", e.Request.URL.String())
		fmt.Printf("Author: %s %s\n", author, e.Request.URL.String())
		fmt.Printf("Published: %s\n", datePublished)
		fmt.Println("---")
	})

	// extract main content
	c.OnHTML(contentSelector, func(e *colly.HTMLElement) {
		requestId := utils.GetRequestIdFromElement(e)
		logger.LogRequest(requestId, fmt.Sprintf("found content at %s", e.Request.URL))

		article := e.ChildText("p")
		article = utils.AddSpacesAfterDots(article)
		fmt.Printf("URL: %s\n", e.Request.URL.String())
		fmt.Printf("Extracted Paragraph: %s\n", article)
		fmt.Println("---")
	})

	// extract read more content
	c.OnHTML(readmoreSelector, func(e *colly.HTMLElement) {
		requestId := utils.GetRequestIdFromElement(e)
		logger.LogRequest(requestId, fmt.Sprintf("found read more content at %s", e.Request.URL))

		readMore := e.ChildText("p")
		readMore = utils.AddSpacesAfterDots(readMore)
		fmt.Printf("URL: %s\n", e.Request.URL.String())
		fmt.Printf("Extracted Read More Paragraph: %s\n", readMore)
		fmt.Println("---")
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
