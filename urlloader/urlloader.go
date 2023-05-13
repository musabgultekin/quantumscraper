package urlloader

import (
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type URLLoader struct {
	file   *os.File
	reader *csv.Reader
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

func (l *URLLoader) Close() error {
	return l.file.Close()
}
