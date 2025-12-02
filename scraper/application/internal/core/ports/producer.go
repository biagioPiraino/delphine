package ports

import "github.com/biagioPiraino/delphico/scraper/internal/core/domain"

type Producer interface {
	Run()
	SendMessageToQueue(article domain.Article) error
	Shutdown()
}
