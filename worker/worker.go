package worker

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/musabgultekin/quantumscraper/http"
	"github.com/musabgultekin/quantumscraper/urlloader"
	"golang.org/x/time/rate"
)

const workerCount = 200

var hostURLsQueue = make(chan []string)

type Worker struct {
	rateLimiter *rate.Limiter
	wg          *sync.WaitGroup
}

func (worker *Worker) Work() error {
	defer worker.wg.Done()

	for hostUrlList := range hostURLsQueue {
		for _, targetURL := range hostUrlList {
			if err := worker.HandleUrl(targetURL); err != nil {
				log.Println("handle url:", err, targetURL)
			}
		}
	}
	return nil
}

func (worker *Worker) HandleUrl(targetURL string) error {
	if err := worker.rateLimiter.Wait(context.TODO()); err != nil {
		panic(err) // This should never happen, but we need to know if it happens.
	}

	// log.Println("Fetching", targetURL)

	resp, err := http.Get(targetURL)
	if err != nil {
		return fmt.Errorf("http get err: %w", err)
	}

	// links, err := extractLinksFromHTML(targetURL, resp)
	// if err != nil {
	// 	return fmt.Errorf("error extract links from html: %w", err)
	// }

	// Queue new links
	_ = resp
	// _ = links
	return nil
}

func StartWorkers(urlListURL string, urlListCachePath string, wg *sync.WaitGroup) error {
	urlLoader, err := urlloader.New(urlListURL, urlListCachePath)
	if err != nil {
		return fmt.Errorf("url loader: %w", err)
	}
	defer urlLoader.Close()

	log.Println("Loading URLs")
	allURLs, err := urlLoader.GetAllURLs()
	if err != nil {
		return fmt.Errorf("url loader get all urls: %w", err)
	}
	log.Println("URLs loaded. Host count:", len(allURLs))

	log.Println("Starting workers")
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		worker := Worker{rateLimiter: rate.NewLimiter(0.5, 1), wg: wg}
		go worker.Work()
	}

	// Send Host URLs
	for _, urlStrings := range allURLs {
		hostURLsQueue <- urlStrings
	}

	// All hosts queued, we can close the queue
	close(hostURLsQueue)

	return nil
}
