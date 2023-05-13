package urlloader

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
)

type URLLoader struct {
	file    *os.File
	scanner *bufio.Scanner
}

func New(url string, filepath string) (*URLLoader, error) {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
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

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to write to file: %w", err)
		}
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &URLLoader{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (l *URLLoader) Next() (string, error) {
	if l.scanner.Scan() {
		return l.scanner.Text(), nil
	}
	if err := l.scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read line: %w", err)
	}
	return "", nil // no more lines to read
}

func (l *URLLoader) Close() error {
	return l.file.Close()
}
