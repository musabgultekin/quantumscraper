package worker

import (
	"fmt"
	"github.com/musabgultekin/quantumscraper/storage"
	"github.com/nsqio/go-nsq"
	"log"
)

func Worker(visitedURLStorage *storage.Queue) nsq.HandlerFunc {
	return func(message *nsq.Message) error {
		log.Println(string(message.Body))

		targetURL := string(message.Body) + "&page=1"
		if err := visitedURLStorage.AddURL(targetURL); err != nil {
			return fmt.Errorf("failed to add URL to visited storage: %w", err)
		}

		return nil
	}
}
