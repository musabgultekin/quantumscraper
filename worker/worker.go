package worker

import (
	"fmt"
	"log"
	"strconv"

	"github.com/musabgultekin/quantumscraper/http"
	"github.com/musabgultekin/quantumscraper/storage"
	"github.com/nsqio/go-nsq"
)

type Worker struct {
	queue *storage.Queue
}

func (worker *Worker) HandleMessage(message *nsq.Message) error {
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
		if err := worker.queue.AddURL(link); err != nil {
			return fmt.Errorf("failed to add URL to visited storage: %w", err)
		}
	}

	return nil
}

func StartWorkers(concurrency int, queue *storage.Queue) (consumers []*nsq.Consumer, err error) {
	// Consumer initialization
	for i := 0; i < concurrency; i++ {
		consumerConfig := nsq.NewConfig()
		consumerConfig.MaxInFlight = 10
		consumer, err := nsq.NewConsumer(storage.NsqTopic+strconv.Itoa(i), storage.NsqChannel, consumerConfig)
		if err != nil {
			return nil, fmt.Errorf("nsq new consumer: %w", err)
		}
		consumer.AddHandler(&Worker{queue: queue})
		if err := consumer.ConnectToNSQD(storage.NsqServer); err != nil {
			return nil, fmt.Errorf("connect to nsqd: %w", err)
		}
		consumers = append(consumers, consumer)
	}

	return
}
