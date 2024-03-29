package worker

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/musabgultekin/quantumscraper/http"
	"github.com/musabgultekin/quantumscraper/metrics"
	"github.com/musabgultekin/quantumscraper/urlloader"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

var hostURLsQueue = make(chan []string, 1000)
var foundLinks = make(map[string]struct{})
var foundLinksChan = make(chan map[string]struct{}, 5000)
var logger, _ = zap.NewDevelopment()

type Worker struct {
	id          int
	rateLimiter *rate.Limiter
	wg          *sync.WaitGroup
}

func NewWorker(id int, wg *sync.WaitGroup) (*Worker, error) {
	rateLimiter := rate.NewLimiter(0.5, 1)

	return &Worker{id: id, rateLimiter: rateLimiter, wg: wg}, nil
}

func (worker *Worker) Work() error {
	defer worker.wg.Done()

	for hostUrlList := range hostURLsQueue {
		for _, targetURL := range hostUrlList {
			if err := worker.HandleUrl(targetURL); err != nil {
				if strings.Contains(err.Error(), "no such host") {
					break // Since we dont have the host anymore, no need to continue
				}
				if strings.Contains(err.Error(), "could not connect to proxy:") &&
					strings.Contains(err.Error(), "status code: 403") {
					continue
				}
				if strings.Contains(err.Error(), "status not 200") {
					continue
				}
				if strings.Contains(err.Error(), "not HTML") {
					continue
				}
				if errors.Is(err, fasthttp.ErrConnectionClosed) {
					continue
				}
				// log.Println("handle url:", err, targetURL)
				// logger.Error("handle url", zap.Error(err))
				logger.Debug(err.Error(), zap.String("url", targetURL))
				continue
			}
		}
	}
	return nil
}

func (worker *Worker) HandleUrl(targetURL string) error {
	// if err := worker.rateLimiter.Wait(context.TODO()); err != nil {
	// 	panic(err) // This should never happen, but we need to know if it happens.
	// }

	// log.Println("Fetching", targetURL)
	// logger.Debug("Fetching", zap.String("url", targetURL))

	requestStartTime := time.Now()
	metrics.RequestInFlightCount.Inc()

	resp, status, err := http.GetFast(targetURL)

	metrics.RequestInFlightCount.Dec()
	metrics.RequestCount.With(prometheus.Labels{"code": strconv.Itoa(status)}).Inc()
	metrics.RequestLatency.With(prometheus.Labels{"code": strconv.Itoa(status)}).Observe(time.Since(requestStartTime).Seconds())

	// if status == fasthttp.StatusTooManyRequests {
	// 	time.Sleep(time.Second * 5)
	// }

	if err != nil {
		return fmt.Errorf("http get err: %w", err)
	}

	links, err := extractLinksFromHTML(targetURL, resp)
	if err != nil {
		return fmt.Errorf("error extract links from html: %w", err)
	}

	foundLinksChan <- links

	// if err := worker.SaveLinks(links); err != nil {
	// 	return fmt.Errorf("save links: %w", err)
	// }

	// Queue new links
	// _ = resp
	_ = links
	return nil
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

	// Save loaded URLs
	go func() {
		for linksBatch := range foundLinksChan {
			for link := range linksBatch {
				foundLinks[link] = struct{}{}
			}
			length := len(foundLinks)
			metrics.FoundURLsCount.Set(float64(length))
			if length >= 100_000_000 {
				// 	if err := storage.WriteLinksToFileRandomFilename(foundLinks, "data/"); err != nil {
				// 		panic(err)
				// 	}
				foundLinks = make(map[string]struct{})
			}
		}
	}()

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
