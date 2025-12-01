package main

import (
	"linkfast/url-projector/consumer"
	"linkfast/url-projector/utils/envs"
	"log"
	"time"
)

func main() {
	log.Println("Iniciando microserviço....")

	log.Println("Waiting the anothers container startup")
	time.Sleep(15 * time.Second)

	kafkaBrokers := envs.GetEnvWithFallback("KAFKA_BROKERS", "")
	kafkaTopic := envs.GetEnvWithFallback("KAFKA_TOPIC", "")

	required := map[string]string{
		"kafkaBrokers": kafkaBrokers,
		"kafkaTopic":   kafkaTopic,
	}

	for key, value := range required {
		if value == "" {
			log.Fatalf("Environment variable %s not defined!", key)
		}
	}

	log.Println("Reading topics!")
	consumer.LinkConsumer(kafkaBrokers, kafkaTopic)
}
