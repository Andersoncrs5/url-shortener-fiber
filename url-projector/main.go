package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Iniciando microserviço Read API...")

	for {
		log.Println("Read API está ativo e aguardando eventos Kafka...")
		time.Sleep(10 * time.Second)
	}
}
