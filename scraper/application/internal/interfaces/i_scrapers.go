package interfaces

import "github.com/biagioPiraino/delphico/scraper/internal/types"

type IScraper interface {
	ScrapeWebsite(artChan chan types.Article, mdChan chan types.ArticleMetadata)
}
