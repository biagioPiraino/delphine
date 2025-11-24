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
	crawlers []interfaces.ICrawler

	articleChannel  chan types.Article
	metadataChannel chan types.ArticleMetadata

	producer sarama.AsyncProducer

	config Config
}

func NewApp(config Config) *App {
	producer, err := newProducer(config.BrokerAddress)
	if err != nil {
		log.Println("error initialising sarama producer")
		panic(err)
	}
	engines := newCrawlers()

	return &App{
		crawlers:        engines,
		producer:        producer,
		config:          config,
		articleChannel:  make(chan types.Article),
		metadataChannel: make(chan types.ArticleMetadata),
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
		<-sigChan
		fmt.Println("captured termination signal...exiting")
		cancel() // propagate cancellation to all the children process coordinating closure of scrapers and producer
	}()

	// launch monitor goroutine that on cancel will wait the group to finish
	// and close channels
	go func() {
		wg.Wait()
		close(a.articleChannel)
		close(a.metadataChannel)
	}()

	// channel for async production of messages to kafka
	wg.Add(1)
	go a.setupProducerResultsChannel(&wg, ctx)

	// launching scraper go routines
	for i := 0; i < len(a.crawlers); i++ {
		scraper := a.crawlers[i]
		wg.Add(1)
		go func() {
			scraper.ScrapeWebsite(&wg, ctx, a.articleChannel, a.metadataChannel)
		}()
	}

	// setup article and metadata channels receivers
	for a.articleChannel != nil || a.metadataChannel != nil {
		select {
		case article, ok := <-a.articleChannel:
			if !ok {
				a.articleChannel = nil
				continue
			}

			payload, err := getArticlePayload(article)
			if err != nil {
				fmt.Println("unable to create payload... continuing")
				continue
			}

			msg := &sarama.ProducerMessage{
				Topic: article.Domain.ToString(),
				Key:   sarama.StringEncoder(article.Domain.ToString()),
				Value: sarama.ByteEncoder(payload),
			}
			a.producer.Input() <- msg
		case metadata, ok := <-a.metadataChannel:
			if !ok {
				a.metadataChannel = nil
				continue
			}

			payload, err := getArticleMetadataPayload(metadata)
			if err != nil {
				fmt.Println("unable to create payload... continuing")
				continue
			}

			msg := &sarama.ProducerMessage{
				Topic: metadata.Domain.ToString() + "-metadata",
				Key:   sarama.StringEncoder(metadata.Domain.ToString()),
				Value: sarama.ByteEncoder(payload),
			}
			a.producer.Input() <- msg
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

func getArticleMetadataPayload(metadata types.ArticleMetadata) ([]byte, error) {
	buf := types.JsonBufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	if err := json.NewEncoder(buf).Encode(metadata); err != nil {
		types.JsonBufferPool.Put(buf)
		return nil, err
	}

	payload := make([]byte, buf.Len())
	copy(payload, buf.Bytes()) // copied to avoid concurrency issues related to async producer
	types.JsonBufferPool.Put(buf)
	return payload, nil
}

func (a *App) setupProducerResultsChannel(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
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
		crawlers.NewYahooCrawler(),
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
