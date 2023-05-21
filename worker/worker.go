package worker

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/musabgultekin/quantumscraper/http"
	"github.com/musabgultekin/quantumscraper/metrics"
	"github.com/musabgultekin/quantumscraper/urlloader"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

var hostURLsQueue = make(chan []string, 1000)

type Worker struct {
	id          int
	rateLimiter *rate.Limiter
	wg          *sync.WaitGroup
	file        *os.File
}

func NewWorker(id int, wg *sync.WaitGroup) (*Worker, error) {
	filename := fmt.Sprintf("data/worker_links/%d.txt", id)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	rateLimiter := rate.NewLimiter(0.5, 1)

	return &Worker{id: id, rateLimiter: rateLimiter, wg: wg, file: file}, nil
}

func (worker *Worker) Work() error {
	defer worker.wg.Done()
	defer worker.file.Close()

	for hostUrlList := range hostURLsQueue {
		for _, targetURL := range hostUrlList {
			if err := worker.HandleUrl(targetURL); err != nil {
				// log.Println("handle url:", err, targetURL)
				continue
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

	requestStartTime := time.Now()
	metrics.RequestInFlightCount.Inc()

	resp, status, err := http.Get(targetURL)

	metrics.RequestInFlightCount.Dec()
	metrics.RequestCount.With(prometheus.Labels{"code": strconv.Itoa(status)}).Inc()
	metrics.RequestLatency.With(prometheus.Labels{"code": strconv.Itoa(status)}).Observe(time.Since(requestStartTime).Seconds())

	if err != nil {
		return fmt.Errorf("http get err: %w", err)
	}

	links, err := extractLinksFromHTML(targetURL, resp)
	if err != nil {
		return fmt.Errorf("error extract links from html: %w", err)
	}

	err = worker.SaveLinksToFile(links)
	if err != nil {
		return fmt.Errorf("error saving links to file: %w", err)
	}

	// Queue new links
	// _ = resp
	_ = links
	return nil
}

func (worker *Worker) SaveLinksToFile(links map[string]struct{}) error {
	writer := bufio.NewWriter(worker.file)
	for link := range links {
		if _, err := writer.WriteString(link + "\n"); err != nil {
			return fmt.Errorf("writing to file: %w", err)
		}
	}
	return writer.Flush()
}

func StartWorkers(urlListURL string, urlListCachePath string, parquetDir string, wg *sync.WaitGroup, concurrency int) error {
	// urlLoader, err := urlloader.New(urlListURL, urlListCachePath)
	// if err != nil {
	// 	return fmt.Errorf("url loader: %w", err)
	// }
	// defer urlLoader.Close()
	urlLoader, err := urlloader.NewParquet(parquetDir)
	if err != nil {
		return fmt.Errorf("url loader: %w", err)
	}
	defer urlLoader.Close()

	log.Println("Starting workers")
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		worker, err := NewWorker(i, wg)
		if err != nil {
			return fmt.Errorf("new worker: %w", err)
		}
		go worker.Work()
	}

	log.Println("Queuing URLs for each host")
	for {
		urlStrings, err := urlLoader.LoadNextHostURLs()
		if err != nil {
			return fmt.Errorf("url loader load next domain urls: %w", err)
		}
		if len(urlStrings) == 0 {
			break // end of file
		}
		hostURLsQueue <- urlStrings
	}
	log.Println("All URLs queued")

	// All hosts queued, we can close the queue
	close(hostURLsQueue)

	return nil
}
