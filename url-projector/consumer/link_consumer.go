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

const retryDelay = 5 * time.Second

func LinkConsumer(brokers, topic string, service services.LinkService) {
	for {
		log.Printf("Tentando conectar ao Kafka em %s...", brokers)

		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  brokers,
			"group.id":           consumerGroupID,
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": true,
			"socket.timeout.ms":  3000,
		})

		if err != nil {
			log.Printf("Falha ao criar o consumer Kafka: %v. Tentando novamente em %s...", err, retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		log.Printf("Consumer Kafka criado com sucesso. Tentando se inscrever no tópico '%s'...", topic)

		if err_subscribe := consumer.Subscribe(topic, nil); err_subscribe != nil {
			consumer.Close()
			log.Printf("Falha ao se inscrever no tópico %s: %v. Tentando novamente em %s...", topic, err_subscribe, retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		log.Printf("Inscrição no tópico '%s' bem-sucedida. Começando a consumir mensagens.", topic)

		consumeLoop(consumer, service)

		log.Printf("Loop de consumo encerrado. Tentando reconectar ao Kafka em %s...", retryDelay)
		consumer.Close()
		time.Sleep(retryDelay)
	}
}

func consumeLoop(consumer *kafka.Consumer, service services.LinkService) {
	for {
		msg, err := consumer.ReadMessage(time.Second)

		if err == nil {
			var envelope cdc.Envelope

			if err := cdc.ParseToEnvelope(msg.Value, &envelope); err != nil {
				log.Printf("Erro ao parsear mensagem: %s", err.Error())
				continue
			}

			ctx := context.Background()
			service.ApplyLogic(ctx, envelope)

		} else if kafkaErr, ok := err.(kafka.Error); ok {
			if kafkaErr.Code() == kafka.ErrTimedOut {
				continue
			} else if kafkaErr.Code() == kafka.ErrAllBrokersDown || kafkaErr.IsFatal() {
				log.Printf("Erro fatal do consumer, re-tentando conexão: %v\n", err)
				return
			} else {
				log.Printf("Erro do consumer: %v\n", err)
			}
		} else {
			log.Printf("Erro inesperado durante a leitura: %v\n", err)
		}
	}
}
