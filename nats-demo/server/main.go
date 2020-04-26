package main

import (
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

func main() {
	natsOptions := nats.Options{
		Servers:        []string{"nats://0.0.0.0:4222"},
		//AllowReconnect: true,
		//MaxReconnect:   3,
		//ReconnectWait:  10,
		//// TODO: these callbacks can be set in order to allow nats to perform failure handling.
		//ClosedCB:          nil,
		//DisconnectedErrCB: nil,
	}

	nc, err := natsOptions.Connect()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error connecting to the NATS server")
		return
	}

	_, err = nc.Subscribe("consensusMessage", func(m *nats.Msg) {
		log.WithFields(log.Fields{
			"message": string(m.Data),
		}).Info("Received a message")
		nc.Publish("consensusReplyMessage", []byte("some response"))
	})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error subscribing to the subject")
	}

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			log.Info("Received an interrupt, stopping all connections...")
			//cancel()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
