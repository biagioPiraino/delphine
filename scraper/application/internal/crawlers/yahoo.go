package crawlers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/biagioPiraino/delphico/scraper/internal/logger"
	"github.com/biagioPiraino/delphico/scraper/internal/types"
	"github.com/biagioPiraino/delphico/scraper/internal/utils"
	"github.com/gocolly/colly"
	"github.com/google/uuid"
)

const rootBaseUrl = "https://uk.finance.yahoo.com/"
const newsBaseUrl = "https://uk.finance.yahoo.com/news"
const metadataSelector = ".mainContainer .byline .byline-attr"
const contentSelector = ".mainContainer .body-wrap .body .bodyItems-wrapper"
const readmoreSelector = ".mainContainer .body-wrap .body .read-more-wrapper"
const provider = "Yahoo Finance"

type YahooCrawler struct{}

type contextTransport struct {
	ctx   context.Context
	trans *http.Transport
}

func (t *contextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.WithContext(t.ctx)
	return t.trans.RoundTrip(req)
}

func NewYahooCrawler() *YahooCrawler {
	return &YahooCrawler{}
}

func (s *YahooCrawler) ScrapeWebsite(
	wg *sync.WaitGroup,
	ctx context.Context,
	artChan chan types.Article,
	mdChan chan types.ArticleMetadata) {
	defer wg.Done()

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(3), // leave to 0 default in production to keep scraping the site
		colly.URLFilters(regexp.MustCompile("^"+regexp.QuoteMeta(rootBaseUrl)), regexp.MustCompile("^"+regexp.QuoteMeta(newsBaseUrl))),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	c.AllowURLRevisit = false
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*uk.finance.yahoo*",
		Parallelism: 1,
	})

	c.OnRequest(func(request *colly.Request) {
		select {
		case <-ctx.Done():
			fmt.Println("ctx done, aborting request...")
			request.Abort()
		default:
			// keep scraping in default case
		}
	})

	trans := &contextTransport{
		ctx:   ctx,
		trans: &http.Transport{},
	}
	c.WithTransport(trans)

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
		mdChan <- types.ArticleMetadata{
			Url:       e.Request.URL.String(),
			Author:    author,
			Published: datePublished,
			Domain:    types.FinanceDomain,
			Provider:  provider,
		}
	})

	// extract main content
	c.OnHTML(contentSelector, func(e *colly.HTMLElement) {
		requestId := utils.GetRequestIdFromElement(e)
		logger.LogRequest(requestId, fmt.Sprintf("found content at %s", e.Request.URL))

		article := e.ChildText("p")
		article = utils.AddSpacesAfterDots(article)
		artChan <- types.Article{
			Url:     e.Request.URL.String(),
			Content: article,
			Domain:  types.FinanceDomain,
		}
	})

	// extract read more content
	c.OnHTML(readmoreSelector, func(e *colly.HTMLElement) {
		requestId := utils.GetRequestIdFromElement(e)
		logger.LogRequest(requestId, fmt.Sprintf("found read more content at %s", e.Request.URL))

		readMore := e.ChildText("p")
		readMore = utils.AddSpacesAfterDots(readMore)
		artChan <- types.Article{
			Url:     e.Request.URL.String(),
			Content: readMore,
			Domain:  types.FinanceDomain,
		}
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

	c.Visit(rootBaseUrl)
	c.Wait()
}
