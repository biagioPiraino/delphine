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

func worker(ctx context.Context, jobs <-chan types.Article, results chan<- *sarama.ProducerMessage) {
	for j := range jobs {
		select {
		case <-ctx.Done(): // context is cancelled from parent, closing worker
			return
		default:
			time.Sleep(1 * time.Second)
			payload, err := getArticlePayload(j)
			if err != nil {
				fmt.Println("unable to create payload... continuing")
				continue
			}

			msg := &sarama.ProducerMessage{
				Topic: j.Domain,
				Key:   sarama.StringEncoder(j.Domain),
				Value: sarama.ByteEncoder(payload),
			}
			results <- msg
		}
	}
}

func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(1 * time.Second) // async results channel requires some buffer to be closed correclty
		fmt.Println("resources released, exiting...")
	}()

	resultsChannel := make(chan *sarama.ProducerMessage)

	for w := 0; w < 10; w++ {
		go worker(ctx, a.ingestionChannel, resultsChannel)
	}

	var crawlerWg sync.WaitGroup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// channel for async production of messages to kafka
	go a.setupProducerResultsChannel(ctx)

	// launching scraper go routines
	for i := 0; i < len(a.crawlers); i++ {
		crawlerWg.Add(1)
		crawler := a.crawlers[i]
		go crawler.ScrapeWebsite(&crawlerWg, ctx, a.ingestionChannel)
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
		case msg := <-resultsChannel:
			a.producer.Input() <- msg // sending results from channel to kafka
		}
	}

	crawlerWg.Wait()
	fmt.Println("task finished, closing channels and exiting")
	close(a.ingestionChannel)
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
		case <-ctx.Done():
			fmt.Println("closing sarama producer")
			if err := a.producer.Close(); err != nil {
				fmt.Println("sarama producer not closed correctly")
			} else {
				fmt.Println("sarama producer closed correctly")
			}
			return
		case <-a.producer.Successes():
			fmt.Println("[OK] message correctly sent to kafka")
		case err := <-a.producer.Errors():
			fmt.Printf("[ERR] error sending msg to kafka %s\n", err.Error())
		}
	}
}

func newCrawlers() []interfaces.ICrawler {
	return []interfaces.ICrawler{
		crawlers.NewYahooCrawler(crawlers.CrawlerConfig{
			Root:         "https://uk.finance.yahoo.com/news",
			MaxDepth:     10,
			Parallelism:  2,
			AllowRevisit: false,
			DomainGlobal: "*uk.finance.yahoo*"}),

		crawlers.NewIndependentCrawler(crawlers.CrawlerConfig{
			Root:         "https://independent.co.uk/money",
			MaxDepth:     10,
			Parallelism:  2,
			AllowRevisit: false,
			DomainGlobal: "*independent*"}),
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
