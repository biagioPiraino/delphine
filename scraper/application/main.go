package main

import (
	"github.com/biagioPiraino/delphico/scraper/internal/interfaces"
	"github.com/biagioPiraino/delphico/scraper/internal/logger"
	"github.com/biagioPiraino/delphico/scraper/internal/scrapers"
)

func main() {
	logger.InitLogger()

	scrapers := registerScrapers()
	for _, scraper := range scrapers {
		scraper.ScrapeWebsite()
	}
}

func registerScrapers() []interfaces.IScraper {
	return []interfaces.IScraper{
		scrapers.NewReutersScraper(),
	}
}
