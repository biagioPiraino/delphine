package consumers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/biagioPiraino/delphico/consumer/internal/databases"
	"github.com/biagioPiraino/delphico/consumer/internal/repositories"
)

type FinanceConsumer struct {
	Topics     []string
	GroupId    string
	Repository *repositories.FinanceRepository
}

func NewFinanceConsumer(db *databases.DelphineDb) *FinanceConsumer {
	return &FinanceConsumer{
		Topics:     []string{"finance", "finance-metadata"},
		GroupId:    "finance-workers",
		Repository: repositories.NewFinanceRepository(db),
	}
}

func (h FinanceConsumer) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Finance consumer setup completed. Start listening for messages...")
	return nil
}

func (h FinanceConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Finance consumer cleanup completed.")
	return nil
}

func (h FinanceConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		switch msg.Topic {
		case "finance":
			var article repositories.Article
			if err := json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&article); err != nil {
				fmt.Printf("error decoding finance article, continuing...\n%v", err)
				continue
			}
			if err := h.Repository.AddFinanceArticle(article); err != nil {
				fmt.Printf("%v\n", err)
			}
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
