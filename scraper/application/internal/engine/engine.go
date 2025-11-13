package engine

import (
	"fmt"
	"sync"

	"github.com/biagioPiraino/delphico/scraper/internal/interfaces"
	"github.com/biagioPiraino/delphico/scraper/internal/scrapers"
	"github.com/biagioPiraino/delphico/scraper/internal/types"
)

type ScrapingEngine struct {
	scrapers        []interfaces.IScraper
	scrapersCount   int
	articleChannel  chan types.Article
	metadataChannel chan types.ArticleMetadata
}

func InitialiseEngine() *ScrapingEngine {
	scrapers := initialiseScrapers()
	return &ScrapingEngine{
		scrapers:        scrapers,
		scrapersCount:   len(scrapers),
		articleChannel:  make(chan types.Article),
		metadataChannel: make(chan types.ArticleMetadata),
	}
}

func initialiseScrapers() []interfaces.IScraper {
	return []interfaces.IScraper{
		scrapers.NewReutersScraper(),
	}
}

func (e *ScrapingEngine) Run() {
	wg := &sync.WaitGroup{}
	wg.Add(e.scrapersCount)

	for i := 0; i < e.scrapersCount; i++ {
		scraper := e.scrapers[i]
		go func() {
			defer wg.Done()
			scraper.ScrapeWebsite(e.articleChannel, e.metadataChannel)
		}()
	}

	go func() {
		wg.Wait()
		close(e.articleChannel)
		close(e.metadataChannel)
	}()

	for e.articleChannel != nil || e.metadataChannel != nil {
		select {
		case msg1, ok := <-e.articleChannel:
			if !ok {
				e.articleChannel = nil
				continue
			}
			fmt.Println("Hit channel! " + msg1.Url)
		case msg2, ok := <-e.metadataChannel:
			if !ok {
				e.metadataChannel = nil
				continue
			}
			fmt.Println("Hit metadata channel! " + msg2.Url)
		}
	}
}
