package interfaces

import (
	"context"

	"github.com/biagioPiraino/delphico/scraper/internal/types"
)

type IScraper interface {
	ScrapeWebsite(ctx context.Context, artChan chan types.Article, mdChan chan types.ArticleMetadata)
}
