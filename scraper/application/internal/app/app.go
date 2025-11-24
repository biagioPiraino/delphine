package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/biagioPiraino/delphico/scraper/internal/crawlers"
	"github.com/biagioPiraino/delphico/scraper/internal/interfaces"
	"github.com/biagioPiraino/delphico/scraper/internal/types"
)

type Config struct {
	BrokerAddress string
}

type App struct {
	config           Config
	crawlers         []interfaces.ICrawler
	ingestionChannel chan types.Article
	producer         sarama.AsyncProducer
}

func NewApp(config Config) *App {
	producer, err := newProducer(config.BrokerAddress)
	if err != nil {
		log.Println("error initialising sarama producer")
		panic(err)
	}
	engines := newCrawlers()

	return &App{
		crawlers:         engines,
		producer:         producer,
		config:           config,
		ingestionChannel: make(chan types.Article),
	}
}

func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancelling context on happy path

	wg := sync.WaitGroup{}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// launch sentinel goroutine that listen for syscall term and interrupt
	go func() {
		select {
		case <-sigChan:
			fmt.Println("captured termination signal...exiting")
			cancel() // propagate cancellation to all the children process coordinating closure of scrapers and producer
		case <-ctx.Done():
			// context cancelled by happy path, do not anything
		}
	}()

	// channel for async production of messages to kafka
	go a.setupProducerResultsChannel(ctx)

	// launching scraper go routines
	for i := 0; i < len(a.crawlers); i++ {
		scraper := a.crawlers[i]
		wg.Add(1)
		go func() {
			scraper.ScrapeWebsite(&wg, ctx, a.ingestionChannel)
		}()
	}

	// launch monitor goroutine that on cancel will wait the group to finish
	// and close channels
	go func() {
		wg.Wait()
		close(a.ingestionChannel)
	}()

	// main loop to send msg to kafka
loop:
	for {
		select {
		case article, ok := <-a.ingestionChannel:
			if !ok {
				break loop // on channel closed when scraper are done break loop
			}

			payload, err := getArticlePayload(article)
			if err != nil {
				fmt.Println("unable to create payload... continuing")
				continue
			}

			msg := &sarama.ProducerMessage{
				Topic: article.Domain,
				Key:   sarama.StringEncoder(article.Domain),
				Value: sarama.ByteEncoder(payload),
			}
			select {
			case a.producer.Input() <- msg:
			// sent
			case <-ctx.Done():
				break loop

			}
		}
	}

	// exiting gracefully
	fmt.Println("application interrupted correctly, channels and resources closed")
	os.Exit(0)
}

func getArticlePayload(article types.Article) ([]byte, error) {
	buf := types.JsonBufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	if err := json.NewEncoder(buf).Encode(article); err != nil {
		types.JsonBufferPool.Put(buf)
		return nil, err
	}

	payload := make([]byte, buf.Len())
	copy(payload, buf.Bytes()) // copied to avoid concurrency issues related to async producer
	types.JsonBufferPool.Put(buf)
	return payload, nil
}

func (a *App) setupProducerResultsChannel(ctx context.Context) {
	for {
		select {
		case <-a.producer.Successes():
			fmt.Println("message passed to kafka correctly")
		case <-a.producer.Errors():
			fmt.Println("error in sending msg to kafka")
		case <-ctx.Done():
			if err := a.producer.Close(); err != nil {
				fmt.Println("sarama producer not closed correctly")
			} else {
				fmt.Println("request have been cancelled, sarama producer closed correctly")
			}
			return
		}
	}
}

func newCrawlers() []interfaces.ICrawler {
	return []interfaces.ICrawler{
		crawlers.NewYahooCrawler(crawlers.CrawlerConfig{
			Root:         "https://uk.finance.yahoo.com/news",
			MaxDepth:     0,
			DomainGlobal: "*uk.finance.yahoo*"}),
	}
}

func newProducer(brokerAddress string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()

	// these two channels are just used in async context
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	config.Producer.Flush.Messages = 100
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	return sarama.NewAsyncProducer([]string{brokerAddress}, config)
}
