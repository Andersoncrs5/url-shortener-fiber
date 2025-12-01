package consumer

import (
	"context"
	"linkfast/url-projector/cdc"
	"linkfast/url-projector/services"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var consumerGroupID = "link_fast_group"

func LinkConsumer(brokers, topic string, service services.LinkService) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           consumerGroupID,
		"auto.offset.reset":  "latest",
		"enable.auto.commit": true,
	})

	if err != nil {
		log.Fatalf("Falha ao criar o consumer Kafka: %v", err)
	}

	defer consumer.Close()

	if err_subscribe := consumer.Subscribe(topic, nil); err_subscribe != nil {
		log.Fatalf("Falha ao se inscrever no tópico %s: %v", topic, err_subscribe)
	}

	for {
		msg, err := consumer.ReadMessage(time.Second)
		if err == nil {
			var envelope cdc.Envelope

			if err := cdc.ParseToEnvelope(msg.Value, &envelope); err != nil {
				log.Printf("%s", err.Error())
				return
			}

			ctx := context.Background()

			service.ApplyLogic(ctx, envelope)

		} else {
			if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() != kafka.ErrTimedOut {
				log.Printf("Erro do consumer: %v\n", err)
			}
		}
	}
}
