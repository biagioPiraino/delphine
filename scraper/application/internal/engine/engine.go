package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"github.com/biagioPiraino/delphico/scraper/internal/interfaces"
	"github.com/biagioPiraino/delphico/scraper/internal/logger"
	"github.com/biagioPiraino/delphico/scraper/internal/scrapers"
	"github.com/biagioPiraino/delphico/scraper/internal/types"
)

type ScrapingEngine struct {
	scrapers        []interfaces.IScraper
	scrapersCount   int
	articleChannel  chan types.Article
	metadataChannel chan types.ArticleMetadata
	producer        sarama.AsyncProducer
}

// run from plaintext port on kafka broker
// var brokerAddress = "host.docker.internal:" + os.Getenv("KAFKA_PORT")
const brokerAddress = "localhost:9094"

func InitialiseEngine() *ScrapingEngine {
	scrapers := initialiseScrapers()
	producer, err := initialiseProducer()
	if err != nil {
		logger.LogRequest(strconv.Itoa(os.Getpid()), fmt.Sprintf("error initialising sarama producer %v, exiting...", err))
		os.Exit(1)
	}
	return &ScrapingEngine{
		scrapers:        scrapers,
		producer:        producer,
		scrapersCount:   len(scrapers),
		articleChannel:  make(chan types.Article),
		metadataChannel: make(chan types.ArticleMetadata),
	}
}

func initialiseScrapers() []interfaces.IScraper {
	return []interfaces.IScraper{
		scrapers.NewReutersScraper(),
	}
}

func initialiseProducer() (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	// enables success channel
	config.Producer.Return.Successes = true
	// enables errors channel
	config.Producer.Return.Errors = true
	config.Producer.Flush.Messages = 100
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	return sarama.NewAsyncProducer([]string{brokerAddress}, config)
}

func (e ScrapingEngine) Run() {
	defer e.producer.Close()

	wg := &sync.WaitGroup{}
	wg.Add(e.scrapersCount)

	for i := 0; i < e.scrapersCount; i++ {
		scraper := e.scrapers[i]
		go func() {
			defer wg.Done()
			scraper.ScrapeWebsite(e.articleChannel, e.metadataChannel)
		}()
	}

	go func() {
		wg.Wait()
		close(e.articleChannel)
		close(e.metadataChannel)
	}()

	// setup channel for kafka consumer
	go func() {
		for {
			select {
			case success := <-e.producer.Successes():
				logger.LogRequest(strconv.Itoa(os.Getpid()), fmt.Sprintf("message delivered correctly to topic %s", success.Topic))
				continue
			case err := <-e.producer.Errors():
				logger.LogRequest(strconv.Itoa(os.Getpid()), fmt.Sprintf("error sending the message to kafka queue %v", err))
			}
		}
	}()

	// setup article and metadata channels
	for e.articleChannel != nil || e.metadataChannel != nil {
		select {
		case article, ok := <-e.articleChannel:
			if !ok {
				e.articleChannel = nil
				continue
			}
			json_article, err := json.Marshal(article)
			if err != nil {
				logger.LogRequest(strconv.Itoa(os.Getpid()), "error while marshalling the article before sending to the queue")
				continue
			}
			msg := &sarama.ProducerMessage{
				Topic: article.Domain.ToString(),
				Key:   sarama.StringEncoder(article.Domain.ToString()),
				Value: sarama.ByteEncoder(json_article),
			}
			e.producer.Input() <- msg
		case metadata, ok := <-e.metadataChannel:
			if !ok {
				e.metadataChannel = nil
				continue
			}
			json_metadata, err := json.Marshal(metadata)
			if err != nil {
				logger.LogRequest(strconv.Itoa(os.Getpid()), "error while marshalling the metadata before sending to the queue")
				continue
			}
			msg := &sarama.ProducerMessage{
				Topic: metadata.Domain.ToString() + "-metadata",
				Key:   sarama.StringEncoder(metadata.Domain.ToString()),
				Value: sarama.ByteEncoder(json_metadata),
			}
			e.producer.Input() <- msg
		}
	}
}
