package kafka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/biagioPiraino/delphico/scraper/internal/core/domain"
	"github.com/biagioPiraino/delphico/scraper/internal/core/utils"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ConfluentProducer struct {
	producer *kafka.Producer

	// used for testing only
	acksChan chan kafka.Event
}

func NewConfluentProducer(servers []string, id string) (*ConfluentProducer, error) {
	bootstraps := strings.Join(servers, ",")
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":   bootstraps,
		"delivery.timeout.ms": 3000, // report success or failure timeout
		"request.timeout.ms":  1000, // default timeout for a single request
		"retries":             1,    // harsh, only one retry to avoid sending to non-existent topics
		"client.id":           id,
		"acks":                "all",
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("[OK] producer initialised")
	return &ConfluentProducer{producer: p}, nil
}

// Run this launches goroutine for message delivery
func (p *ConfluentProducer) Run() {
	go func() {
		for e := range p.producer.Events() {
			// only used for integration testing
			if p.acksChan != nil {
				p.acksChan <- e
			}

			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("[ERR] failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					var article domain.Article
					if err := json.NewDecoder(bytes.NewReader(ev.Value)).Decode(&article); err != nil {
						fmt.Printf("error decoding article sent to kafka: %v\n", err)
					}
					fmt.Printf("[OK] produced event to topic %s: key = %-10s url = %s\n",
						*ev.TopicPartition.Topic, string(ev.Key), article.Url)
				}
			}
		}
	}()
}

func (p *ConfluentProducer) SendMessageToQueue(article domain.Article) error {
	payload, err := utils.GetArticlePayload(article)
	if err != nil {
		return err
	}
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &article.Domain, Partition: kafka.PartitionAny},
		Value:          payload,
		Key:            []byte(article.Domain),
		Timestamp:      time.Now().UTC(),
	}

	if err := p.producer.Produce(msg, nil); err != nil {
		return err
	}

	p.producer.Flush(15 * 1000)
	return nil
}

func (p *ConfluentProducer) Shutdown() {
	p.producer.Flush(15 * 1000) // flushes out all the messages
	p.producer.Close()
	// closing acks channel after testing client
	if p.acksChan != nil {
		close(p.acksChan)
	}
	fmt.Println("[OK] producer closed correctly.")
}

func (p *ConfluentProducer) SetAcksChannelForTesting(channel chan kafka.Event) {
	p.acksChan = channel
}
