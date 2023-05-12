package storage

import (
	"fmt"
	"github.com/nsqio/nsq/nsqd"
	"log"
	"os"
)

const NsqServer = "localhost:4150"
const NsqTopic = "topic"
const NsqChannel = "channel"
const dataPath = "data/nsqd" // Add your data path here

func StartNSQDEmbeddedServer() error {
	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir all nsqd data path: %w", err)
	}
	go func() {
		opts := nsqd.NewOptions()
		opts.DataPath = dataPath // Set data path here
		nsqdServer, err := nsqd.New(opts)
		if err != nil {
			log.Fatal(err)
		}
		if err := nsqdServer.Main(); err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}
