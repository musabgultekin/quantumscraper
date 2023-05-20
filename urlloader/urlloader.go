package urlloader

import (
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/musabgultekin/quantumscraper/storage"
)

type URLLoader struct {
	file     *os.File
	reader   *csv.Reader
	lastHost string
}

func New(url string, filepath string) (*URLLoader, error) {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		log.Println("Downloading 👀", url)
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to load URL: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to load URL: status code %d", resp.StatusCode)
		}

		out, err := os.Create(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
		defer out.Close()

		if strings.HasSuffix(url, ".gz") {
			gz, err := gzip.NewReader(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to create gzip reader: %w", err)
			}
			defer gz.Close()

			_, err = io.Copy(out, gz)
			if err != nil {
				return nil, fmt.Errorf("failed to write to file: %w", err)
			}
		} else {
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to write to file: %w", err)
			}
		}
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	reader := csv.NewReader(file)

	// Read and discard the header 🗑️
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	return &URLLoader{
		file:   file,
		reader: reader,
	}, nil
}

func (l *URLLoader) Next() (string, error) {
	record, err := l.reader.Read()
	if err == io.EOF {
		return "", nil // no more lines to read
	}
	if err != nil {
		return "", fmt.Errorf("failed to read line: %w", err)
	}
	if len(record) == 0 {
		return "", errors.New("no columns in the current row")
	}
	return record[0], nil // return only the first column
}

func (l *URLLoader) GetAllURLs() (map[string][]string, error) {
	var hostURLs = make(map[string][]string)
	for {
		urlString, err := l.Next()
		if err != nil {
			return nil, fmt.Errorf("next domain: %w", err)
		}
		if urlString == "" {
			log.Println("URL Loader end of file")
			break // end of file
		}
		urlParsed, err := url.Parse(urlString)
		if err != nil {
			return nil, fmt.Errorf("parse url: %w, %s", err, urlString)
		}
		hostURLs[urlParsed.Host] = append(hostURLs[urlParsed.Host], urlString)
	}

	return hostURLs, nil
}

func (l *URLLoader) Close() error {
	return l.file.Close()
}

func StartQueueingURLs(urlListURL string, urlListPath string, queue *storage.Queue) error {
	urlLoader, err := New(urlListURL, urlListPath)
	if err != nil {
		return fmt.Errorf("url loader: %w", err)
	}
	defer urlLoader.Close()

	for {
		urlString, err := urlLoader.Next()
		if err != nil {
			return fmt.Errorf("next domain: %w", err)
		}
		if urlString == "" {
			log.Println("URL Loader end of file")
			break // end of file
		}
		if queue.IsStopped() {
			break
		}
		if err := queue.AddURL(urlString); err != nil {
			return fmt.Errorf("failed to add URL to visited storage: %w", err)
		}

	}
	return nil
}
