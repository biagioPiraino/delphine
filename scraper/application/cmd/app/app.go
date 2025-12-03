package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/biagioPiraino/delphico/scraper/internal/adapters/producers/kafka"
	"github.com/biagioPiraino/delphico/scraper/internal/core/domain"
	"github.com/biagioPiraino/delphico/scraper/internal/core/ports"
	"github.com/biagioPiraino/delphico/scraper/internal/core/services"
	"github.com/biagioPiraino/delphico/scraper/internal/logger"
)

type Config struct {
	BrokerAddress string
}

type App struct {
	config   Config
	crawlers []ports.Crawler
	producer ports.Producer
	logger   ports.Logger
}

func NewApp(config Config) *App {
	internalLogger, err := logger.NewLogger("./logs", "logs")
	if err != nil {
		log.Fatalf("[ERR] error initialising logger %v\n", err)
	}
	producer, err := kafka.NewConfluentProducer([]string{config.BrokerAddress}, "confluent_00")
	if err != nil {
		log.Fatalf("[ERR] error initialising producer %v\n", err)
	}
	engines := newCrawlers(internalLogger)

	return &App{
		logger:   internalLogger,
		crawlers: engines,
		producer: producer,
		config:   config,
	}
}

func (a *App) Run() {
	articleChan := make(chan domain.Article)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		a.producer.Shutdown()
		articleChan = nil // setting to nil to prevent panic attacks on sending article to closed channels
		a.logger.Close()
		fmt.Println("resources released, exiting...")
	}()
	a.producer.Run()

	var crawlerWg sync.WaitGroup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// launching scraper go routines
	for i := 0; i < len(a.crawlers); i++ {
		crawlerWg.Add(1)
		crawler := a.crawlers[i]
		go crawler.CrawlWebsite(&crawlerWg, ctx, articleChan)
	}

	// launch sentinel goroutine that listen for syscall term and interrupt
	go func() {
		defer cancel()
		<-sigChan
		fmt.Println("captured termination signal...exiting")
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case msg := <-articleChan:
			if err := a.producer.SendMessageToQueue(msg); err != nil {
				log.Printf("[ERR] error sending msg to queue: %v\n", err)
			}
		}
	}

	crawlerWg.Wait()
	fmt.Println("task finished, closing channels and exiting")
}

func newCrawlers(logger ports.Logger) []ports.Crawler {
	return []ports.Crawler{
		services.NewYahooCrawler(logger, services.CrawlerConfig{
			Root:         "https://uk.finance.yahoo.com/news",
			MaxDepth:     10,
			Parallelism:  2,
			AllowRevisit: false,
			DomainGlobal: "*uk.finance.yahoo*"}),

		services.NewIndependentCrawler(logger, services.CrawlerConfig{
			Root:         "https://independent.co.uk/money",
			MaxDepth:     10,
			Parallelism:  2,
			AllowRevisit: false,
			DomainGlobal: "*independent.*"}),
	}
}
