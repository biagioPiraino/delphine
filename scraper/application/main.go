package main

import (
	"fmt"
	"sync"

	"github.com/biagioPiraino/delphico/scraper/internal/interfaces"
	"github.com/biagioPiraino/delphico/scraper/internal/logger"
	"github.com/biagioPiraino/delphico/scraper/internal/scrapers"
	"github.com/biagioPiraino/delphico/scraper/internal/types"
)

func main() {
	logger.InitLogger()

	scrapers := registerScrapers()

	articleChan := make(chan types.Article)
	metadataChan := make(chan types.ArticleMetadata)

	wg := &sync.WaitGroup{}
	wg.Add(len(scrapers))

	for i := 0; i < len(scrapers); i++ {
		scraper := scrapers[i]
		go func() {
			defer wg.Done()
			scraper.ScrapeWebsite(articleChan, metadataChan)
		}()
	}

	go func() {
		wg.Wait()
		close(articleChan)
		close(metadataChan)
	}()

	for articleChan != nil || metadataChan != nil {
		select {
		case msg1, ok := <-articleChan:
			if !ok {
				articleChan = nil
				continue
			}
			fmt.Println("Hit channel! " + msg1.Url)
		case msg2, ok := <-metadataChan:
			if !ok {
				metadataChan = nil
				continue
			}
			fmt.Println("Hit metadata channel! " + msg2.Url)
		}
	}

}

func registerScrapers() []interfaces.IScraper {
	return []interfaces.IScraper{
		scrapers.NewReutersScraper(),
	}
}
