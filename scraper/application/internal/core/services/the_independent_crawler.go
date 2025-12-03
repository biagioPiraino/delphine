package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/biagioPiraino/delphico/scraper/internal/core/domain"
	"github.com/biagioPiraino/delphico/scraper/internal/core/utils"
	"github.com/biagioPiraino/delphico/scraper/internal/logger"
	"github.com/gocolly/colly"
	"github.com/google/uuid"
)

type IndependentCrawler struct {
	config CrawlerConfig
}

func NewIndependentCrawler(config CrawlerConfig) *IndependentCrawler {
	return &IndependentCrawler{
		config: config,
	}
}

func (cr *IndependentCrawler) CrawlWebsite(
	wg *sync.WaitGroup,
	ctx context.Context,
	artChan chan<- domain.Article) {
	defer wg.Done()

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(cr.config.MaxDepth), // leave to 0 default in production to keep scraping the site
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	c.AllowURLRevisit = cr.config.AllowRevisit

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  cr.config.DomainGlobal,
		Parallelism: cr.config.Parallelism,
	})
	if err != nil {
		fmt.Println("Unable to setup crawler limits. returning.")
		return
	}

	c.OnRequest(func(request *colly.Request) {
		select {
		case <-ctx.Done():
			fmt.Println("ctx done, aborting request from The Independent...")
			request.Abort()
		default:
			// keep scraping in default case, adding id in header to keep track of request in case of error
			requestId := uuid.New().String()
			request.Headers.Add(utils.RequestIdHeader, requestId)
			logger.LogRequest(requestId, fmt.Sprintf("visiting %s", request.URL))
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
		if e.Request.URL.Host == "www.independent.co.uk" &&
			strings.Contains(link, "/money/") ||
			strings.Contains(link, "/business/") {
			err := e.Request.Visit(link)
			if err != nil {
				return
			}
		}
	})

	// extract content from article tag
	c.OnHTML("article", func(e *colly.HTMLElement) {
		article := e.ChildText(".main-wrapper #main h2, p")
		if article == "" {
			return
		}

		title := e.ChildText("header h1")
		if title == "" {
			title = "unknown"
		}

		author := e.ChildText("header a[href*='author']")
		if author == "" {
			author = "unknown"
		}

		published := e.ChildText("#article-published-date")
		if published == "" {
			published = "unknown"
		}

		art := domain.Article{
			Url:       e.Request.URL.String(),
			Author:    author,
			Title:     title,
			Provider:  "The Independent",
			Domain:    domain.FinanceDomain.String(),
			Content:   article,
			Published: published,
		}

		if artChan != nil {
			artChan <- art
		}
	})

	c.OnError(func(r *colly.Response, e error) {
		requestId := utils.GetRequestIdFromResponse(r)
		msg := fmt.Sprintf("error while visting %s - response: %d - details: \"%v\"", r.Request.URL, r.StatusCode, e)
		logger.LogRequest(requestId, msg)
	})

	err = c.Visit(cr.config.Root)
	if err != nil {
		fmt.Printf("error visiting %s. returning...\n", cr.config.Root)
		return
	}

	<-ctx.Done()
}
