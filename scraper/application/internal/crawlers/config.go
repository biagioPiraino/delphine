package crawlers

type CrawlerConfig struct {
	Root         string
	MaxDepth     int
	DomainGlobal string
	Parallelism  int
	AllowRevisit bool
}
