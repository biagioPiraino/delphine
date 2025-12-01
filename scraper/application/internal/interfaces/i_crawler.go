package interfaces

import (
	"context"
	"sync"

	"github.com/biagioPiraino/delphico/scraper/internal/types"
)

type ICrawler interface {
	ScrapeWebsite(
		wg *sync.WaitGroup,
		ctx context.Context,
		artChan chan<- types.Article,
	)
}
