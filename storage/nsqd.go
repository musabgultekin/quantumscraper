package storage

import (
	"github.com/nsqio/nsq/nsqd"
	"log"
)

const NsqServer = "localhost:4150"
const NsqTopic = "topic"
const NsqChannel = "channel"

func StartNSQDEmbeddedServer() {
	go func() {
		nSQD, err := nsqd.New(nsqd.NewOptions())
		if err != nil {
			log.Fatal(err)
		}
		if err := nSQD.Main(); err != nil {
			log.Fatal(err)
		}
	}()
}
