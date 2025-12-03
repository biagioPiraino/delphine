package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/biagioPiraino/delphico/scraper/internal/core/domain"
	"github.com/biagioPiraino/delphico/scraper/internal/core/ports"
	"github.com/biagioPiraino/delphico/scraper/internal/core/utils"
	"github.com/gocolly/colly"
	"github.com/google/uuid"
)

type YahooCrawler struct {
	logger ports.Logger
	config CrawlerConfig
}

func NewYahooCrawler(logger ports.Logger, config CrawlerConfig) *YahooCrawler {
	return &YahooCrawler{
		logger: logger,
		config: config,
	}
}

func (yc *YahooCrawler) CrawlWebsite(
	wg *sync.WaitGroup,
	ctx context.Context,
	artChan chan<- domain.Article) {
	defer wg.Done()

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(yc.config.MaxDepth), // leave to 0 default in production to keep scraping the site
		colly.URLFilters(regexp.MustCompile("^"+regexp.QuoteMeta(yc.config.Root))),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)
	c.AllowURLRevisit = yc.config.AllowRevisit

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  yc.config.DomainGlobal,
		Parallelism: yc.config.Parallelism,
	})
	if err != nil {
		fmt.Println("Unable to setup crawler limits. returning.")
		return
	}

	c.OnRequest(func(request *colly.Request) {
		select {
		case <-ctx.Done():
			fmt.Println("ctx done, aborting request from Yahoo Finance...")
			request.Abort()
		default:
			// keep scraping in default case, adding id in header to keep track of request in case of error
			requestId := uuid.New().String()
			request.Headers.Add(utils.RequestIdHeader, requestId)
			yc.logger.LogDebug(requestId, fmt.Sprintf("visiting %s", request.URL))
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
		err := e.Request.Visit(link)
		if err != nil {
			return
		}
	})

	// extract content from article tag
	c.OnHTML("article", func(e *colly.HTMLElement) {
		article := e.ChildText(".article-wrap p")
		if article == "" {
			return
		}

		author := strings.Trim(strings.Split(e.ChildText(".byline-attr-author"), "·")[0], " ")
		if author == "" {
			author = "unknown"
		}

		title := e.ChildText(".cover-title")
		if title == "" {
			title = "unknown"
		}

		published := e.ChildText(".byline-attr-meta-time")
		if published == "" {
			published = "unknown"
		}

		art := domain.Article{
			Url:       e.Request.URL.String(),
			Author:    author,
			Title:     title,
			Provider:  "Yahoo Finance",
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
		errMsg := fmt.Sprintf("error while visting %s - response: %d - details: \"%v\"", r.Request.URL, r.StatusCode, e)
		err := errors.New(errMsg)
		yc.logger.LogError(requestId, err)
	})

	err = c.Visit(yc.config.Root)
	if err != nil {
		fmt.Printf("error visiting %s. returning...\n", yc.config.Root)
		return
	}

	<-ctx.Done()
}
