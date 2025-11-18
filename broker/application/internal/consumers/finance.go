package consumers

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/biagioPiraino/delphico/consumer/internal/repositories"
)

type FinanceConsumer struct {
	Topics  []string
	GroupId string
}

func NewFinanceConsumer() *FinanceConsumer {
	return &FinanceConsumer{
		Topics:  []string{"finance", "finance-metadata"},
		GroupId: "finance-workers",
	}
}

func (h *FinanceConsumer) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Finance consumer setup completed. Start listening for messages...")
	return nil
}

func (h *FinanceConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Finance consumer cleanup completed.")
	return nil
}

func (h *FinanceConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		switch msg.Topic {
		case "finance":
			var article repositories.FinanceArticle
			json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&article)
			if err := repositories.AddFinanceArticle(article); err != nil {
				log.Printf("%v", err)
			}
		case "finance-metadata":
			var metadata repositories.FinanceArticleMetadata
			json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&metadata)
			if err := repositories.AddFinanceMetadata(metadata); err != nil {
				log.Printf("%v", err)
			}
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
