package worker

import (
	"fmt"
	"github.com/musabgultekin/quantumscraper/http"
	"github.com/musabgultekin/quantumscraper/storage"
	"github.com/nsqio/go-nsq"
	"log"
)

func Worker(queue *storage.Queue) nsq.HandlerFunc {
	return func(message *nsq.Message) error {
		targetURL := string(message.Body)

		log.Println("Fetching", targetURL)

		resp, err := http.Get(targetURL)
		if err != nil {
			log.Println("http get err:", err, targetURL)
			return nil
		}

		links, err := extractLinksFromHTML(targetURL, resp)
		if err != nil {
			fmt.Println("error extract links from html", err)
			return nil
		}

		for _, link := range links {
			if err := queue.AddURL(link); err != nil {
				return fmt.Errorf("failed to add URL to visited storage: %w", err)
			}
		}

		return nil
	}
}
