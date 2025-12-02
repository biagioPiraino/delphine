package ports

import (
	"context"
	"sync"

	"github.com/biagioPiraino/delphico/scraper/internal/core/domain"
)

type Crawler interface {
	CrawlWebsite(
		wg *sync.WaitGroup,
		ctx context.Context,
		artChan chan<- domain.Article,
	)
}
