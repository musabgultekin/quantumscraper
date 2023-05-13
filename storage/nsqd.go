package storage

import (
	"fmt"
	"os"
	"sync"

	"github.com/nsqio/nsq/nsqd"
)

const NsqServer = "localhost:4150"
const NsqTopic = "scraper_"
const NsqChannel = "channel"
const dataPath = "data/nsqd" // Add your data path here

type NSQDServer struct {
	nsqdInstance *nsqd.NSQD
	wg           sync.WaitGroup
	errCh        chan error
}

func NewNSQDServer() (*NSQDServer, error) {
	s := &NSQDServer{
		errCh: make(chan error, 1), // Initialize the error channel
	}

	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("mkdir all nsqd data path: %w", err)
	}

	opts := nsqd.NewOptions()
	opts.DataPath = dataPath // Set data path here
	nsqdServer, err := nsqd.New(opts)
	if err != nil {
		return nil, fmt.Errorf("new nsqd: %w", err)
	}

	s.nsqdInstance = nsqdServer // Save the instance to the struct

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := nsqdServer.Main(); err != nil {
			s.errCh <- err // Send the error to the channel
		}
	}()

	return s, nil
}

func (s *NSQDServer) Stop() {
	if s.nsqdInstance != nil {
		s.nsqdInstance.Exit()
	}
	s.wg.Wait()    // Wait for the server to fully exit
	close(s.errCh) // Close the error channel
}

func (s *NSQDServer) Error() <-chan error {
	return s.errCh
}
