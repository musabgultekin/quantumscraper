package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/musabgultekin/quantumscraper/http"
	"github.com/musabgultekin/quantumscraper/storage"
	"github.com/nsqio/go-nsq"
	"golang.org/x/time/rate"
)

type QueueWorker struct {
	queue       *storage.Queue
	rateLimiter *rate.Limiter
}

func (worker *QueueWorker) HandleMessage(message *nsq.Message) error {
	if err := worker.rateLimiter.Wait(context.TODO()); err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}

	targetURL := string(message.Body)
	log.Println("Fetching", targetURL)

	resp, _, err := http.Get(targetURL)
	if err != nil {
		log.Println("http get err:", err, targetURL)
		return nil
	}

	links, err := extractLinksFromHTML(targetURL, resp)
	if err != nil {
		fmt.Println("error extract links from html", err)
		return nil
	}

	// Queue new links
	_ = links
	// for _, link := range links {
	// 	if err := worker.queue.AddURL(link); err != nil {
	// 		return fmt.Errorf("failed to add URL to visited storage: %w", err)
	// 	}
	// }

	return nil
}

func StartQueueWorkers(concurrency int, queue *storage.Queue) (consumers []*nsq.Consumer, err error) {

	// for i := 0; i < concurrency; i++ {
	// Consumer initialization
	consumerConfig := nsq.NewConfig()
	consumerConfig.MaxInFlight = 1
	// consumer, err := nsq.NewConsumer(storage.NsqTopic+strconv.Itoa(i), storage.NsqChannel, consumerConfig)
	consumer, err := nsq.NewConsumer(storage.NsqTopic, storage.NsqChannel, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("nsq new consumer: %w", err)
	}
	consumer.AddHandler(&QueueWorker{
		queue:       queue,
		rateLimiter: rate.NewLimiter(1, 1),
	})
	// Connect
	if err := consumer.ConnectToNSQD(storage.NsqServer); err != nil {
		return nil, fmt.Errorf("connect to nsqd: %w", err)
	}
	consumers = append(consumers, consumer)
	// }

	return
}
