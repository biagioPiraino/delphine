package consumers

import (
	"log"

	"github.com/IBM/sarama"
)

type FinanceConsumer struct {
	Topic   string
	GroupId string
}

func NewFinanceConsumer() *FinanceConsumer {
	return &FinanceConsumer{
		Topic:   "finance",
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
		// here data will be added to postgre sql
		log.Printf(
			"Received message: key=%s, value=%s, partition=%d, offset=%d\n",
			string(msg.Key),
			string(msg.Value),
			msg.Partition,
			msg.Offset)

		session.MarkMessage(msg, "")
	}
	return nil
}
