package worker

import (
	"fmt"
	"github.com/musabgultekin/quantumscraper/storage"
	"github.com/nsqio/go-nsq"
	"log"
)

func Worker(visitedURLStorage *storage.VisitedURLStorage, producer *nsq.Producer) nsq.HandlerFunc {
	return func(message *nsq.Message) error {
		log.Println(string(message.Body))

		targetURL := string(message.Body) + "&page=1"
		added, err := visitedURLStorage.AddURL(targetURL)
		if err != nil {
			return fmt.Errorf("failed to add URL to visited storage: %w", err)
		}
		if !added {
			log.Println("URL already exists:", targetURL)
			return nil
		}

		if err := producer.Publish(storage.NsqTopic, []byte(targetURL)); err != nil {
			return fmt.Errorf("producer publish: %w", err)
		}

		return nil
	}
}
